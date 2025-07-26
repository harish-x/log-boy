package server

import (
	"context"
	"gRPC-gateway/config"

	"gRPC-gateway/internal/services"
	protogen "gRPC-gateway/internal/services/genproto/logs"
	metricProtogen "gRPC-gateway/internal/services/genproto/metrics"
	"gRPC-gateway/internal/services/log_service"
	"gRPC-gateway/internal/services/metric_service"
	"log"
	"net"

	"github.com/IBM/sarama"
	"google.golang.org/grpc"
)

type Kfk struct {
	Producer        sarama.SyncProducer
	ProtoSerializer *config.ProtobufSerializer
}

func StartNewgRPCServer(ctx context.Context, cfg *config.AppConfig, kfk *Kfk) error {
	lis, err := net.Listen("tcp", cfg.ServerPort)

	if err != nil {
		return err
	}
	pg, err := config.NewPostgres(cfg.PostgresDb, 10, 10, "5m")

	if err != nil {
		return err
	}
	s := grpc.NewServer(
		grpc.StreamInterceptor(services.NewAuthStreamInterceptor(pg, cfg.GRPCSecret)),
	)

	log.Println("Server started on port", cfg.ServerPort)

	logService := log_service.NewLogServiceServer(kfk.Producer, kfk.ProtoSerializer)
	protogen.RegisterLogServiceServer(s, logService)
	metricService := metric_service.NewMetricsServiceServer(kfk.Producer, kfk.ProtoSerializer)
	metricProtogen.RegisterMetricsServiceServer(s, metricService)
	go func() {
		<-ctx.Done()
		log.Println("Shutting down gRPC server...")
		s.GracefulStop()
	}()
	return s.Serve(lis)
}
