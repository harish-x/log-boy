package log_service

import (
	metricProtogen "gRPC-gateway/internal/services/genproto/metrics"
)

type MetricsServcieServer struct {
	metricProtogen.UnimplementedMetricsServiceServer
}
