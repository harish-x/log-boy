package metric_service

import (
	"gRPC-gateway/config"
	metricProtogen "gRPC-gateway/internal/services/genproto/metrics"
	"io"
	"log"

	"github.com/IBM/sarama"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MetricsServiceServer struct {
	metricProtogen.UnimplementedMetricsServiceServer
	producer        sarama.SyncProducer
	protoSerializer *config.ProtobufSerializer
}

func NewMetricsServiceServer(kafka sarama.SyncProducer, protoSerializer *config.ProtobufSerializer) *MetricsServiceServer {
	return &MetricsServiceServer{
		metricProtogen.UnimplementedMetricsServiceServer{},
		kafka,
		protoSerializer,
	}
}

func (s *MetricsServiceServer) ReceiveMetrics(stream grpc.ClientStreamingServer[metricProtogen.Metrics, metricProtogen.Res]) error {
	log.Println("New client stream connected")

	for {
		metricsMessage, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				log.Println("Client stream finished")
				return stream.SendAndClose(&metricProtogen.Res{
					Ack: true,
				})
			}
			log.Printf("Failed to receive metrics: %v", err)
			return status.Errorf(codes.Unknown, "failed to receive metrics: %v", err)
		}

		log.Printf("Received metrics: Service=%s, CPU usage=%v, memory usage=%v\n", metricsMessage.GetServiceName(), metricsMessage.GetCpuUsage(), metricsMessage.GetMemoryUsage())
		topic := metricsMessage.GetServiceName()
		if topic == "" {
			log.Println("Received log with empty serviceName, skipping")
			continue
		}
		topic = "metrics-" + topic
		kafkaValue, err := s.protoSerializer.Serialize("Metrics-value", metricsMessage)
		if err != nil {
			log.Printf("Failed to serialize protobuf message for topic %s: %v", topic, err)
			continue
		}
		// Create a Sarama producer message
		msg := &sarama.ProducerMessage{
			Topic: topic,
			Value: sarama.ByteEncoder(kafkaValue),
		}
		// Send a message synchronously
		partition, offset, err := s.producer.SendMessage(msg)
		if err != nil {
			log.Printf("Failed to produce message to Kafka for topic %s: %v", topic, err)
			continue
		}
		log.Printf("Delivered message to topic %s [%d] at offset %v\n",
			topic, partition, offset)

	}

}
