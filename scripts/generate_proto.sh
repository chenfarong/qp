#!/bin/bash

# 生成gRPC代码
protoc --go_out=. --go-grpc_out=. proto/auth.proto
protoc --go_out=. --go-grpc_out=. proto/game.proto
protoc --go_out=. --go-grpc_out=. proto/bill.proto
protoc --go_out=. --go-grpc_out=. proto/chat.proto
protoc --go_out=. --go-grpc_out=. proto/gateway.proto

echo "Proto files generated successfully!"
