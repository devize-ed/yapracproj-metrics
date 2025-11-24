package server

import (
	"context"
	"net"
	"testing"

	"github.com/devize-ed/yapracproj-metrics.git/internal/repository"
	mstorage "github.com/devize-ed/yapracproj-metrics.git/internal/repository/mstorage"
	pb "github.com/devize-ed/yapracproj-metrics.git/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func setupTestServer(t *testing.T, storage repository.Repository, subnet string) (pb.MetricsClient, func()) {
	bufSize := 1024 * 1024
	lis := bufconn.Listen(bufSize)

	logger := zap.NewNop().Sugar()
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(ipFilterInterceptor(subnet)),
	)

	srv := NewServer(storage, logger)
	pb.RegisterMetricsServer(grpcServer, srv)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	client := pb.NewMetricsClient(conn)

	cleanup := func() {
		grpcServer.Stop()
		lis.Close()
		conn.Close()
	}

	return client, cleanup
}

func TestUpdateMetrics(t *testing.T) {
	ms := mstorage.NewMemStorage()
	client, cleanup := setupTestServer(t, ms, "")
	defer cleanup()

	tests := []struct {
		name         string
		request      *pb.UpdateMetricsRequest
		expectedErr  bool
		expectedCode codes.Code
	}{
		{
			name: "update_counter",
			request: func() *pb.UpdateMetricsRequest {
				metric := pb.Metric_builder{
					Id:    "testCounter",
					Type:  pb.Metric_COUNTER,
					Delta: 5,
				}.Build()
				return pb.UpdateMetricsRequest_builder{
					Metrics: []*pb.Metric{metric},
				}.Build()
			}(),
			expectedErr:  false,
			expectedCode: codes.OK,
		},
		{
			name: "update_gauge",
			request: func() *pb.UpdateMetricsRequest {
				metric := pb.Metric_builder{
					Id:    "testGauge",
					Type:  pb.Metric_GAUGE,
					Value: 123.45,
				}.Build()
				return pb.UpdateMetricsRequest_builder{
					Metrics: []*pb.Metric{metric},
				}.Build()
			}(),
			expectedErr:  false,
			expectedCode: codes.OK,
		},
		{
			name: "update_multiple_metrics",
			request: func() *pb.UpdateMetricsRequest {
				gauge := pb.Metric_builder{
					Id:    "testGauge",
					Type:  pb.Metric_GAUGE,
					Value: 123.45,
				}.Build()
				counter := pb.Metric_builder{
					Id:    "testCounter",
					Type:  pb.Metric_COUNTER,
					Delta: 5,
				}.Build()
				return pb.UpdateMetricsRequest_builder{
					Metrics: []*pb.Metric{gauge, counter},
				}.Build()
			}(),
			expectedErr:  false,
			expectedCode: codes.OK,
		},
		{
			name: "unknown_metric_type",
			request: func() *pb.UpdateMetricsRequest {
				metric := pb.Metric_builder{
					Id:   "testMetric",
					Type: pb.Metric_MType(999),
				}.Build()
				return pb.UpdateMetricsRequest_builder{
					Metrics: []*pb.Metric{metric},
				}.Build()
			}(),
			expectedErr:  true,
			expectedCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			md := metadata.New(map[string]string{"x-real-ip": "192.168.1.1"})
			ctx = metadata.NewOutgoingContext(ctx, md)

			resp, err := client.UpdateMetrics(ctx, tt.request)

			if tt.expectedErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.IsType(t, &pb.UpdateMetricsResponse{}, resp)
			}
		})
	}
}

func TestUpdateMetrics_Conversion(t *testing.T) {
	ms := mstorage.NewMemStorage()
	client, cleanup := setupTestServer(t, ms, "")
	defer cleanup()

	request := func() *pb.UpdateMetricsRequest {
		gauge := pb.Metric_builder{
			Id:    "gaugeMetric",
			Type:  pb.Metric_GAUGE,
			Value: 99.99,
		}.Build()
		counter := pb.Metric_builder{
			Id:    "counterMetric",
			Type:  pb.Metric_COUNTER,
			Delta: 42,
		}.Build()
		return pb.UpdateMetricsRequest_builder{
			Metrics: []*pb.Metric{gauge, counter},
		}.Build()
	}()

	ctx := context.Background()
	md := metadata.New(map[string]string{"x-real-ip": "192.168.1.1"})
	ctx = metadata.NewOutgoingContext(ctx, md)

	resp, err := client.UpdateMetrics(ctx, request)
	require.NoError(t, err)
	require.NotNil(t, resp)

	gauge, err := ms.GetGauge(ctx, "gaugeMetric")
	require.NoError(t, err)
	require.NotNil(t, gauge)
	assert.Equal(t, 99.99, *gauge)

	counter, err := ms.GetCounter(ctx, "counterMetric")
	require.NoError(t, err)
	require.NotNil(t, counter)
	assert.Equal(t, int64(42), *counter)
}

func TestIPFilterInterceptor(t *testing.T) {
	ms := mstorage.NewMemStorage()

	tests := []struct {
		name         string
		subnet       string
		ip           string
		expectedErr  bool
		expectedCode codes.Code
	}{
		{
			name:         "empty_subnet_allows_all",
			subnet:       "",
			ip:           "192.168.1.1",
			expectedErr:  false,
			expectedCode: codes.OK,
		},
		{
			name:         "valid_ip_in_subnet",
			subnet:       "192.168.1.0/24",
			ip:           "192.168.1.10",
			expectedErr:  false,
			expectedCode: codes.OK,
		},
		{
			name:         "missing_ip_header",
			subnet:       "192.168.1.0/24",
			ip:           "",
			expectedErr:  true,
			expectedCode: codes.Unauthenticated,
		},
		{
			name:         "invalid_ip_address",
			subnet:       "192.168.1.0/24",
			ip:           "invalid-ip",
			expectedErr:  true,
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "ip_not_in_subnet",
			subnet:       "192.168.1.0/24",
			ip:           "10.0.0.1",
			expectedErr:  true,
			expectedCode: codes.PermissionDenied,
		},
		{
			name:         "invalid_subnet",
			subnet:       "invalid-subnet",
			ip:           "192.168.1.10",
			expectedErr:  true,
			expectedCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, cleanup := setupTestServer(t, ms, tt.subnet)
			defer cleanup()

			request := func() *pb.UpdateMetricsRequest {
				metric := pb.Metric_builder{
					Id:    "testGauge",
					Type:  pb.Metric_GAUGE,
					Value: 123.45,
				}.Build()
				return pb.UpdateMetricsRequest_builder{
					Metrics: []*pb.Metric{metric},
				}.Build()
			}()

			ctx := context.Background()
			if tt.ip != "" {
				md := metadata.New(map[string]string{"x-real-ip": tt.ip})
				ctx = metadata.NewOutgoingContext(ctx, md)
			}

			resp, err := client.UpdateMetrics(ctx, request)

			if tt.expectedErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
			}
		})
	}
}
