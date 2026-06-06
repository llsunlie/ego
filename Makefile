.PHONY: proto proto-go proto-dart sqlc

proto: proto-go proto-dart

proto-go:
	@PATH="$$PATH:$$(go env GOPATH)/bin" protoc \
		--proto_path=proto \
		--go_out=server/proto \
		--go_opt=paths=source_relative \
		--go-grpc_out=server/proto \
		--go-grpc_opt=paths=source_relative \
		ego/api.proto
	@echo "proto (Go) generated"

proto-dart:
	@PATH="$$PATH:$$HOME/.pub-cache/bin" protoc \
		--proto_path=proto \
		--dart_out=grpc:client/lib/data/generated \
		proto/ego/api.proto
	@mv client/lib/data/generated/ego/*.dart client/lib/data/generated/
	@rm -rf client/lib/data/generated/ego/
	@echo "proto (Dart) generated"

sqlc:
	@PATH="$$PATH:$$(go env GOPATH)/bin" sqlc generate
	@echo "sqlc generated"
