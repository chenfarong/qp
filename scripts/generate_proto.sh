#!/bin/bash

# 设置输出目录
SERVERS_OUT=pkg/proto/servers

# 创建输出目录
mkdir -p $SERVERS_OUT

# 生成servers目录下的gRPC代码
protoc --go_out=. --go-grpc_out=. proto/servers/gateway.proto
protoc --go_out=. --go-grpc_out=. proto/servers/auth.proto
protoc --go_out=. --go-grpc_out=. proto/servers/gamelogic.proto
protoc --go_out=. --go-grpc_out=. proto/servers/bill.proto
protoc --go_out=. --go-grpc_out=. proto/servers/chat.proto

# 移动生成的Go文件到正确的目录
mv proto/servers/*.pb.go $SERVERS_OUT/ 2>/dev/null || true
mv proto/servers/*.grpc.pb.go $SERVERS_OUT/ 2>/dev/null || true

# 生成原始目录下的gRPC代码
protoc --go_out=. --go-grpc_out=. proto/auth.proto
protoc --go_out=. --go-grpc_out=. proto/game.proto
protoc --go_out=. --go-grpc_out=. proto/bill.proto
protoc --go_out=. --go-grpc_out=. proto/chat.proto
protoc --go_out=. --go-grpc_out=. proto/gateway.proto

echo "Proto files generated successfully!"