package log_service

import (
	protogen "gRPC-gateway/internal/services/genproto/logs"
	"io"
	"log"

	"google.golang.org/grpc"
)

type LogServiceServer struct {
	protogen.UnimplementedLogServiceServer
}

func NewLogServiceServer() *LogServiceServer {
	return &LogServiceServer{
		protogen.UnimplementedLogServiceServer{},
	}
}

func (s *LogServiceServer) ReceiveLogsStream(stream grpc.ClientStreamingServer[protogen.Log, protogen.Response]) error {
	log.Println("New client stream connected")

	var logCount int32 = 0

	// Continuously read from the stream
	for {
		logMsg, err := stream.Recv()
		if err == io.EOF {
			// Client has finished sending, send response and close
			log.Printf("Stream ended. Received %d logs total", logCount)
			return stream.SendAndClose(&protogen.Response{
				Ack: true,
			})
		}
		if err != nil {
			log.Printf("Error receiving from stream: %v", err)
			return err
		}

		logCount++

		// Process the log message
		log.Printf("Received log #%d from service: %s, level: %s, message: %s",
			logCount, logMsg.ServiceName, logMsg.Level, logMsg.Message)

		// Here you can add your log processing logic:
		// - Save to database
		// - Forward to other services
		// - Apply filters/transformations
		// - etc.

		// Optional: You can add some processing logic here
		if err := s.processLog(logMsg); err != nil {
			log.Printf("Error processing log: %v", err)
			// Decide whether to continue or return error
		}
	}
}

func (s *LogServiceServer) processLog(logMsg *protogen.Log) error {
	// Add your log processing logic here
	// For example:
	// - Validate log data
	// - Save to database
	// - Send to monitoring systems
	// - Apply business logic

	log.Print("Processing log:",
		logMsg.ServiceName, logMsg.Level, logMsg.Message, logMsg.RequestId, logMsg.BuildDetails, logMsg.RemoteIp, logMsg.Stack, logMsg.UserAgent, logMsg.RequestMethod, logMsg.RequestUrl, logMsg.Timestamp)

	return nil
}
