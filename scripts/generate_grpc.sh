#!/bin/sh
protoc \
  --go_out=internal/transport/grpc/pb --go_opt=paths=source_relative \
  --go-grpc_out=internal/transport/grpc/pb --go-grpc_opt=paths=source_relative \
  api/grpc/pvz.proto
