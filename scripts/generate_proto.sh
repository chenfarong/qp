#!/bin/bash

# 生成gRPC代码
protoc --go_out=. --go-grpc_out=. pkg/proto/auth.proto
protoc --go_out=. --go-grpc_out=. pkg/proto/game.proto

echo "Proto files generated successfully!"
