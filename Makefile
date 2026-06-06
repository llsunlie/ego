.PHONY: proto proto-go proto-dart

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
	@PATH="$$PATH:$$(go env GOPATH)/bin:$$HOME/.pub-cache/bin" protoc \
		--proto_path=proto \
		--dart_out=client/lib/data/generated \
		ego/api.proto
	@echo "proto (Dart) generated"

.PHONY: sqlc
sqlc:
	@PATH="$$PATH:$$(go env GOPATH)/bin" sqlc generate
	@echo "sqlc generated"
