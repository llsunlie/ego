.PHONY: proto

.PHONY: proto proto-go proto-dart

proto: proto-go proto-dart

proto-go:
	@PATH="$$PATH:$$(go env GOPATH)/bin" protoc \
		--go_out=server/proto \
		--go_opt=paths=source_relative \
		--go-grpc_out=server/proto \
		--go-grpc_opt=paths=source_relative \
		proto/ego/api.proto
	@echo "proto (Go) generated"

proto-dart:
	@PATH="$$PATH:$$(go env GOPATH)/bin:$$HOME/.pub-cache/bin" protoc \
		--dart_out=client/lib/data/generated \
		proto/ego/api.proto
	@echo "proto (Dart) generated"
