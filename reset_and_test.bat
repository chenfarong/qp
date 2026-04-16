@echo off
echo === 清除测试缓存并重新执行用户注册、角色创建 ===



echo.
echo 步骤2: 清理Go缓存...
go clean -modcache
go clean -cache
go clean -testcache
echo Go缓存已清理



echo.
echo 步骤5: 执行注册和角色创建测试...
go run -mod=mod ./test/registration_and_role.go -username=resetuser -password=resetpass -actor=resethero

echo.
echo === 测试完成 ===
pause