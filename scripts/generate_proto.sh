#!/bin/bash

# 生成gRPC代码
protoc --go_out=. --go-grpc_out=. pkg/proto/auth.proto
protoc --go_out=. --go-grpc_out=. pkg/proto/game.proto
protoc --go_out=. --go-grpc_out=. pkg/proto/bill.proto
protoc --go_out=. --go-grpc_out=. pkg/proto/chat.proto
protoc --go_out=. --go-grpc_out=. pkg/proto/gateway.proto

echo "Proto files generated successfully!"
