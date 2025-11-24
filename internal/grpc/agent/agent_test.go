package agent

import (
	"context"
	"net"
	"testing"

	models "github.com/devize-ed/yapracproj-metrics.git/internal/model"
	pb "github.com/devize-ed/yapracproj-metrics.git/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
)

type mockMetricsServer struct {
	pb.UnimplementedMetricsServer
	receivedRequests []*pb.UpdateMetricsRequest
	receivedMetadata []metadata.MD
	errorToReturn    error
}

func (m *mockMetricsServer) UpdateMetrics(ctx context.Context, req *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	if m.errorToReturn != nil {
		return nil, m.errorToReturn
	}
	m.receivedRequests = append(m.receivedRequests, req)
	md, _ := metadata.FromIncomingContext(ctx)
	m.receivedMetadata = append(m.receivedMetadata, md)
	return &pb.UpdateMetricsResponse{}, nil
}

func setupTestClient(t *testing.T, mockServer *mockMetricsServer) (*Client, func()) {
	bufSize := 1024 * 1024
	lis := bufconn.Listen(bufSize)

	grpcServer := grpc.NewServer()
	pb.RegisterMetricsServer(grpcServer, mockServer)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	logger := zap.NewNop().Sugar()
	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	client := &Client{
		MetricsClient: pb.NewMetricsClient(conn),
		logger:        logger,
	}

	cleanup := func() {
		grpcServer.Stop()
		lis.Close()
		conn.Close()
	}

	return client, cleanup
}

func TestUpdateMetrics(t *testing.T) {
	tests := []struct {
		name            string
		metrics         []models.Metrics
		batchSize       int
		expectedBatches int
		expectedErr     bool
		errorContains   string
	}{
		{
			name: "update_single_gauge",
			metrics: []models.Metrics{
				{
					ID:    "testGauge",
					MType: models.Gauge,
					Value: func() *float64 { v := 123.45; return &v }(),
				},
			},
			batchSize:       10,
			expectedBatches: 1,
			expectedErr:     false,
		},
		{
			name: "update_single_counter",
			metrics: []models.Metrics{
				{
					ID:    "testCounter",
					MType: models.Counter,
					Delta: func() *int64 { d := int64(10); return &d }(),
				},
			},
			batchSize:       10,
			expectedBatches: 1,
			expectedErr:     false,
		},
		{
			name: "update_multiple_metrics",
			metrics: []models.Metrics{
				{
					ID:    "testGauge",
					MType: models.Gauge,
					Value: func() *float64 { v := 123.45; return &v }(),
				},
				{
					ID:    "testCounter",
					MType: models.Counter,
					Delta: func() *int64 { d := int64(10); return &d }(),
				},
			},
			batchSize:       10,
			expectedBatches: 1,
			expectedErr:     false,
		},
		{
			name: "batching_25_metrics",
			metrics: func() []models.Metrics {
				metrics := make([]models.Metrics, 25)
				for i := 0; i < 25; i++ {
					value := float64(i)
					metrics[i] = models.Metrics{
						ID:    "testGauge",
						MType: models.Gauge,
						Value: &value,
					}
				}
				return metrics
			}(),
			batchSize:       10,
			expectedBatches: 3,
			expectedErr:     false,
		},
		{
			name: "unknown_metric_type",
			metrics: []models.Metrics{
				{
					ID:    "testMetric",
					MType: "unknown",
				},
			},
			batchSize:       10,
			expectedBatches: 0,
			expectedErr:     true,
			errorContains:   "unknown metric type",
		},
		{
			name: "server_error",
			metrics: []models.Metrics{
				{
					ID:    "testGauge",
					MType: models.Gauge,
					Value: func() *float64 { v := 123.45; return &v }(),
				},
			},
			batchSize:       10,
			expectedBatches: 0,
			expectedErr:     true,
			errorContains:   "failed to update metrics batch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServer := &mockMetricsServer{}
			if tt.name == "server_error" {
				mockServer.errorToReturn = assert.AnError
			}
			client, cleanup := setupTestClient(t, mockServer)
			defer cleanup()

			ctx := context.Background()
			err := client.UpdateMetrics(ctx, tt.metrics, tt.batchSize)

			if tt.expectedErr {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Len(t, mockServer.receivedRequests, tt.expectedBatches)
			} else {
				require.NoError(t, err)
				assert.Len(t, mockServer.receivedRequests, tt.expectedBatches)

				if len(mockServer.receivedMetadata) > 0 {
					ipValues := mockServer.receivedMetadata[0].Get("x-real-ip")
					require.NotEmpty(t, ipValues, "x-real-ip should be present in metadata")
					assert.NotEmpty(t, ipValues[0], "IP address should not be empty")
				}
			}
		})
	}
}

func TestUpdateMetrics_Conversion(t *testing.T) {
	tests := []struct {
		name          string
		metric        models.Metrics
		expectedID    string
		expectedType  pb.Metric_MType
		expectedValue float64
		expectedDelta int64
	}{
		{
			name: "gauge_conversion",
			metric: models.Metrics{
				ID:    "testGauge",
				MType: models.Gauge,
				Value: func() *float64 { v := 99.99; return &v }(),
			},
			expectedID:    "testGauge",
			expectedType:  pb.Metric_GAUGE,
			expectedValue: 99.99,
		},
		{
			name: "counter_conversion",
			metric: models.Metrics{
				ID:    "testCounter",
				MType: models.Counter,
				Delta: func() *int64 { d := int64(42); return &d }(),
			},
			expectedID:    "testCounter",
			expectedType:  pb.Metric_COUNTER,
			expectedDelta: 42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServer := &mockMetricsServer{}
			client, cleanup := setupTestClient(t, mockServer)
			defer cleanup()

			ctx := context.Background()
			err := client.UpdateMetrics(ctx, []models.Metrics{tt.metric}, 10)

			require.NoError(t, err)
			require.Len(t, mockServer.receivedRequests, 1)
			require.Len(t, mockServer.receivedRequests[0].GetMetrics(), 1)

			metric := mockServer.receivedRequests[0].GetMetrics()[0]
			assert.Equal(t, tt.expectedID, metric.GetId())
			assert.Equal(t, tt.expectedType, metric.GetType())

			if tt.expectedValue != 0 {
				assert.Equal(t, tt.expectedValue, metric.GetValue())
			}
			if tt.expectedDelta != 0 {
				assert.Equal(t, tt.expectedDelta, metric.GetDelta())
			}
		})
	}
}

func TestUpdateMetrics_NilValues(t *testing.T) {
	mockServer := &mockMetricsServer{}
	client, cleanup := setupTestClient(t, mockServer)
	defer cleanup()

	metrics := []models.Metrics{
		{
			ID:    "testGauge",
			MType: models.Gauge,
			Value: nil,
		},
		{
			ID:    "testCounter",
			MType: models.Counter,
			Delta: nil,
		},
	}

	ctx := context.Background()
	err := client.UpdateMetrics(ctx, metrics, 10)

	require.NoError(t, err)
	require.Len(t, mockServer.receivedRequests, 1)
	assert.Len(t, mockServer.receivedRequests[0].GetMetrics(), 2)
}

func TestUpdateMetrics_IPAddressInMetadata(t *testing.T) {
	mockServer := &mockMetricsServer{}
	client, cleanup := setupTestClient(t, mockServer)
	defer cleanup()

	metrics := []models.Metrics{
		{
			ID:    "testGauge",
			MType: models.Gauge,
			Value: func() *float64 { v := 123.45; return &v }(),
		},
	}

	ctx := context.Background()
	err := client.UpdateMetrics(ctx, metrics, 10)

	require.NoError(t, err)
	require.Len(t, mockServer.receivedMetadata, 1)

	ipValues := mockServer.receivedMetadata[0].Get("x-real-ip")
	require.NotEmpty(t, ipValues, "x-real-ip should be present in metadata")
	assert.NotEmpty(t, ipValues[0], "IP address should not be empty")
}
