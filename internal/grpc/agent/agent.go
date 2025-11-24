package agent

import (
	"context"
	"fmt"
	"net"

	models "github.com/devize-ed/yapracproj-metrics.git/internal/model"
	pb "github.com/devize-ed/yapracproj-metrics.git/pkg/api"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// Agent is the gRPC agent for the metrics service.
type Client struct {
	pb.MetricsClient
	logger *zap.SugaredLogger
}

func NewClient(host string, logger *zap.SugaredLogger) *Client {
	conn, err := grpc.NewClient(host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Errorf("failed to create gRPC client: %w", err)
		return nil
	}
	return &Client{pb.NewMetricsClient(conn), logger}
}

// UpdateMetrics sends a batch of metrics to the server.
func (c *Client) UpdateMetrics(ctx context.Context, metrics []models.Metrics, batchSize int) error {
	c.logger.Debug("Updating metrics from gRPC request")

	// Divide metrics into batches of N metrics (const) and send to client
	for i := 0; i < len(metrics); i += batchSize {
		end := i + batchSize
		if end > len(metrics) {
			end = len(metrics)
		}
		batch := metrics[i:end]

		// Convert models.Metrics to pb.Metric
		pbMetrics := make([]*pb.Metric, 0, len(batch))
		for _, metric := range batch {
			// Convert MType string to enum
			var mType pb.Metric_MType
			switch metric.MType {
			case models.Gauge:
				mType = pb.Metric_GAUGE
			case models.Counter:
				mType = pb.Metric_COUNTER
			default:
				return fmt.Errorf("unknown metric type: %s", metric.MType)
			}

			// Build the pb.Metric using builder
			pbMetricBuilder := pb.Metric_builder{
				Id:   metric.ID,
				Type: mType,
			}
			if metric.Delta != nil {
				pbMetricBuilder.Delta = *metric.Delta
			}
			if metric.Value != nil {
				pbMetricBuilder.Value = *metric.Value
			}

			pbMetrics = append(pbMetrics, pbMetricBuilder.Build())
		}

		// Create UpdateMetricsRequest using builder
		reqBuilder := pb.UpdateMetricsRequest_builder{
			Metrics: pbMetrics,
		}
		req := reqBuilder.Build()

		// Get IP address.
		ip, err := getIPAddress()
		if err != nil {
			return fmt.Errorf("failed to get IP address: %w", err)
		}
		// Set the X-Real-IP to the metadata.
		md := metadata.New(map[string]string{"x-real-ip": ip})
		ctx = metadata.NewOutgoingContext(ctx, md)

		// Call the gRPC client method (embedded via pb.MetricsClient)
		resp, err := c.MetricsClient.UpdateMetrics(ctx, req)
		if err != nil {
			return fmt.Errorf("failed to update metrics batch: %w", err)
		}
		c.logger.Debugf("Updated metrics batch: %v", resp)
	}
	return nil
}

// getIPAddress gets the IP address from the system.
func getIPAddress() (string, error) {
	// Get all network interfaces.
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("failed to get network interfaces: %w", err)
	}

	// Find the first non-loopback interface with an IPv4 address.
	for _, iface := range interfaces {
		// Skip loopback interfaces.
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		// Skip interfaces that are down.
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// Skip if not IPv4 or is loopback.
			if ip == nil || ip.IsLoopback() {
				continue
			}
			// Prefer IPv4 addresses.
			if ip.To4() != nil {
				return ip.String(), nil
			}
		}
	}

	// Fallback: if no non-loopback interface found, return localhost.
	return "127.0.0.1", nil
}
