.PHONY: proto-gen sqlc-gen migrate run

proto-gen:
	protoc --go_out=proto --go_opt=paths=source_relative \
	       --go-grpc_out=proto --go-grpc_opt=paths=source_relative \
	       --proto_path=../proto ../proto/ego/api.proto

sqlc-gen:
	sqlc generate

migrate:
	go run cmd/migrate/main.go

run:
	go run cmd/ego/main.go
