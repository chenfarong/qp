@echo off
echo === 清除测试缓存并重新执行用户注册、角色创建 ===

echo.
echo 步骤1: 停止所有服务...
taskkill /f /im ssoauth.exe >nul 2>&1
taskkill /f /im gateway.exe >nul 2>&1
taskkill /f /im gamelogic.exe >nul 2>&1
echo 服务已停止

echo.
echo 步骤2: 清理Go缓存...
go clean -modcache
go clean -cache
go clean -testcache
echo Go缓存已清理

echo.
echo 步骤3: 重新构建服务...
go build -mod=mod -o bin/ssoauth.exe ./outside/ssoauth/cmd
go build -mod=mod -o bin/gateway.exe ./outside/gateway/cmd
go build -mod=mod -o bin/gamelogic.exe ./inside/gamelogic/cmd
echo 服务重新构建完成

echo.
echo 步骤4: 启动服务...
start "SSO Auth Service" bin\ssoauth.exe
start "Gateway Service" bin\gateway.exe
start "Game Logic Service" bin\gamelogic.exe
timeout /t 3 /nobreak >nul
echo 服务已启动

echo.
echo 步骤5: 执行注册和角色创建测试...
go run -mod=mod ./test/registration_and_role.go -username=resetuser -password=resetpass -actor=resethero

echo.
echo === 测试完成 ===
pause