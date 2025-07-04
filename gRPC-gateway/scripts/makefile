# windows powershell or cmd
genLogsProto:
	@echo "Generating logs.pb.go and logs_grpc.pb.go"
	@echo "----------------------------------------"
	@if not exist ..\internal\services\genproto\logs mkdir ..\internal\services\genproto\logs
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
		-d "{\"schemaType\":\"PROTOBUF\",\"schema\":\"syntax = \\\"proto3\\\";\\n\\nimport \\\"google/protobuf/timestamp.proto\\\";\\n\\nmessage BuildDetails {\\n  string nodeVersion = 1;\\n  string appVersion = 2;\\n}\\n\\nmessage Log {\\n  string serviceName = 1;\\n  BuildDetails buildDetails = 2;\\n  string level = 3;\\n  string message = 4;\\n  string stack = 5;\\n  string requestId = 6;\\n  string requestUrl = 7;\\n  string requestMethod = 8;\\n  string userAgent = 9;\\n  string remoteIp = 10;\\n  google.protobuf.Timestamp timestamp = 11;\\n}\"}"
