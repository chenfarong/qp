#!/bin/bash

echo "Building SSO Auth Service..."
go build -o bin/ssoauth.exe ./internet/ssoauth/cmd

echo "Building Gateway Service..."
go build -o bin/gateway.exe ./internet/gateway/cmd

echo "Building Game Logic Service..."
go build -o bin/gamelogic.exe ./inside/gamelogic

echo "Build completed successfully!"
