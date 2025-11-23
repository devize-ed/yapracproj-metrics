package server

import (
	"context"
	"net"

	models "github.com/devize-ed/yapracproj-metrics.git/internal/model"
	"github.com/devize-ed/yapracproj-metrics.git/internal/repository"
	pb "github.com/devize-ed/yapracproj-metrics.git/pkg/api"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Server is the gRPC server for the metrics service.
type Server struct {
	pb.UnimplementedMetricsServer
	storage repository.Repository
	logger  *zap.SugaredLogger
}

// NewServer creates a new gRPC server.
func NewServer(storage repository.Repository, logger *zap.SugaredLogger) *Server {
	return &Server{
		storage: storage,
		logger:  logger,
	}
}

// Run starts the gRPC server.
func (s *Server) Serve(ctx context.Context, host string, subnet string) {
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(ipFilterInterceptor(subnet)),
	)
	// Register the service.
	grpcServer.RegisterService(&pb.Metrics_ServiceDesc, s)
	lis, err := net.Listen("tcp", host)
	if err != nil {
		s.logger.Errorf("failed to listen: %w", err)
		return
	}
	s.logger.Infof("gRPC server listening on %s", host)
	// Serve the gRPC server.
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			s.logger.Errorf("failed to serve: %w", err)
			return
		}
	}()
	// Wait for shutdown signal.
	<-ctx.Done()
	s.logger.Info("Stop signal received, shutting down the server...")
	grpcServer.GracefulStop()
	s.logger.Debug("gRPC server stopped")
}

// UpdateMetrics updates the metrics in the database.
func (s *Server) UpdateMetrics(ctx context.Context, req *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	s.logger.Debug("Updating metrics from gRPC request")
	var metrics []models.Metrics
	// Convert the metrics to the model struct
	for _, metric := range req.GetMetrics() {
		s.logger.Debugf("Updating metric: ID = %s, Type = %s, Value = %v, Delta = %v", metric.GetId(), metric.GetType(), metric.GetValue(), metric.GetDelta())
		
		// Convert enum to string
		var mType string
		switch metric.GetType() {
		case pb.Metric_GAUGE:
			mType = models.Gauge
		case pb.Metric_COUNTER:
			mType = models.Counter
		default:
			return nil, status.Errorf(codes.InvalidArgument, "unknown metric type: %v", metric.GetType())
		}

		// Create metric with proper nil handling
		m := models.Metrics{
			ID:    metric.GetId(),
			MType: mType,
		}

		// Set value for gauge metrics
		if metric.GetType() == pb.Metric_GAUGE {
			value := metric.GetValue()
			m.Value = &value
		}

		// Set delta for counter metrics
		if metric.GetType() == pb.Metric_COUNTER {
			delta := metric.GetDelta()
			m.Delta = &delta
		}

		metrics = append(metrics, m)
	}
	// Save the metrics to the database
	if err := s.storage.SaveBatch(ctx, metrics); err != nil {
		s.logger.Error("failed to save batch", zap.Error(err))
		return nil, err
	}
	s.logger.Debug("Metrics updated successfully")
	return &pb.UpdateMetricsResponse{}, nil
}

// ipFilterInterceptor is a gRPC interceptor that filters requests by IP address.
func ipFilterInterceptor(subnet string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// If the trusted subnet is not configured, skip the check.
		if subnet == "" {
			return handler(ctx, req)
		}

		var ip string
		// Get the metadata from the context.
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			// Get the X-Real-IP header.
			values := md.Get("x-real-ip")
			if len(values) > 0 {
				// The X-Real-IP header contains a slice of strings, get the first string.
				ip = values[0]
			}
		}
		// If the IP address is empty, return an error.
		if ip == "" {
			return nil, status.Errorf(codes.Unauthenticated, "missing x-real-ip header")
		}
		// Parse the subnet.
		_, subnetNet, err := net.ParseCIDR(subnet)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid subnet: %v", err)
		}
		// Parse the IP address.
		ipAddr := net.ParseIP(ip)
		if ipAddr == nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid IP address: %s", ip)
		}
		// Check if the IP address is in the subnet.
		if !subnetNet.Contains(ipAddr) {
			return nil, status.Errorf(codes.PermissionDenied, "IP address %s is not in trusted subnet", ip)
		}
		// Call the next handler.
		return handler(ctx, req)
	}
}
