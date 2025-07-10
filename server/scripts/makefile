genLogsProto:
	@echo "Generating logs.pb.go"
	@echo "----------------------------------------"
	@if not exist ..\internal\services\proto\logs mkdir ..\internal\services\proto\logs
	protoc \
		--proto_path=./protobuf ./protobuf/logs.proto \
		--go_out=../internal/services/proto/logs --go_opt=paths=source_relative \
		--go-grpc_out=../internal/services/proto/logs --go-grpc_opt=paths=source_relative

genMetricsProto:
	@echo "Generating metrics.pb.go"
	@if not exist ..\internal\services\proto\metrics mkdir ..\internal\services\proto\metrics
	protoc \
		--proto_path=./protobuf ./protobuf/metrics.proto \
		--go_out=../internal/services/proto/metrics --go_opt=paths=source_relative \
		--go-grpc_out=../internal/services/proto/metrics --go-grpc_opt=paths=source_relative