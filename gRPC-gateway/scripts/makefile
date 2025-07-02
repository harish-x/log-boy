# windows powershell or cmd
genLogsProto:
	@echo "Generating logs.pb.go and logs_grpc.pb.go"
	@echo "----------------------------------------"
	@if not exist ..\internal\services\genproto\logs mkdir ..\internal\services\genproto\logs
	protoc \
		--proto_path=../internal/protobuf ../internal/protobuf/log.proto \
		--go_out=../internal/services/genproto/logs --go_opt=paths=source_relative \
		--go-grpc_out=../internal/services/genproto/logs --go-grpc_opt=paths=source_relative

