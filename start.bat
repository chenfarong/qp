@echo off

echo Starting SSO Auth Service...
start "SSO Auth Service" bin\ssoauth.exe

echo Starting Gateway Service...
start "Gateway Service" bin\gateway.exe

echo Starting Game Logic Service...
start "Game Logic Service" bin\gamelogic.exe

echo All services started successfully!
