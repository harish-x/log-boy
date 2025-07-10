genLogsProtoPS:
	@echo "Generating logs.pb.go and logs_grpc.pb.go"
	@echo "----------------------------------------"
	@mkdir -p ../internal/services/genproto/metrics 2>/dev/null || mkdir "../internal/services/genproto/metrics" 2>nul || true
	protoc \
		--proto_path=../internal/protobuf ../internal/protobuf/log.proto \
		--go_out=../internal/services/genproto/logs --go_opt=paths=source_relative \
		--go-grpc_out=../internal/services/genproto/logs --go-grpc_opt=paths=source_relative

genAVROschemaLog:
	@echo "Generating avro from .avsc files"
	gogen-avro -package avro_log -short-unions ../internal/services/avro_log log.avsc


CreateKafkaLogSchema:
	@echo "Creating schema on kafka schema registry"
	curl -X POST http://localhost:8081/subjects/Logs-value/versions \
		-H "Content-Type: application/vnd.schemaregistry.v1+json" \
		-d @logs.proto.schema.json

CreateKafkaMetricsSchema:
	@echo "Creating metrics schema on kafka schema registry"
	curl -X POST http://localhost:8081/subjects/Metrics-value/versions \
		-H "Content-Type: application/vnd.schemaregistry.v1+json" \
		-d @metrics.proto.schema.json

genMetricsProto:
	@echo "generating metrics.pb.go and metrics_grpc.pb.go"
	@echo "----------------------------------------------"
	@mkdir -p ../internal/services/genproto/metrics 2>/dev/null || mkdir "../internal/services/genproto/metrics" 2>nul || true
	protoc \
		--proto_path=../internal/protobuf ../internal/protobuf/metrics.proto \
		--go_out=../internal/services/genproto/metrics --go_opt=paths=source_relative \
		--go-grpc_out=../internal/services/genproto/metrics --go-grpc_opt=paths=source_relative