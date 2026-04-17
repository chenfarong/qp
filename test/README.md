# 测试脚本使用说明

## 概述

本目录包含游戏服务器的测试脚本，用于测试注册账号、创建角色、获取背包数据等功能。

## 测试脚本

### 1. 游戏逻辑测试客户端

**文件路径**: `test/gamelogic/main.go`

**功能**: 测试游戏逻辑服务的各种功能，包括登录验证、创建角色、获取背包数据等。

**使用方法**:

1. **编译脚本**:
   ```bash
   cd test/gamelogic
   go build -mod=mod -o gamelogic_test.exe main.go
   ```

2. **运行脚本**:
   ```bash
   # 使用配置文件中的服务器地址
   ./gamelogic_test.exe
   
   # 使用指定的服务器地址
   ./gamelogic_test.exe -auth=http://localhost:7080 -gateway=ws://localhost:7061
   
   # 使用指定的用户名和密码
   ./gamelogic_test.exe -username=testuser -password=testpassword
   
   # 使用指定的角色名称
   ./gamelogic_test.exe -actor=testactor
   
   # 组合使用多个参数
   ./gamelogic_test.exe -username=testuser -password=testpassword -actor=testactor
   ```

**参数说明**:

- `help`: 显示帮助信息
- `-auth=auth_url`: 验证服务器URL (默认: 从配置文件读取)
- `-gateway=gateway_url`: 网关服务器WebSocket URL (默认: 从配置文件读取)
- `-username=username`: 用户名 (默认: za_admin)
- `-password=password`: 密码 (默认: za_admin)
- `-actor=actor_name`: 角色名称 (可选)
- `-interval=seconds`: 定时发送GetGameMoneyRequest请求的间隔（秒，0表示不开启）

**配置文件**:

脚本会自动读取 `../../config.yml` 文件中的服务器配置:

- `auth.host` 和 `auth.port`: 验证服务器地址和端口
- `gateway.host` 和 `gateway.ws_port`: 网关服务器地址和WebSocket端口

如果没有提供命令行参数，脚本会使用配置文件中的服务器地址。

### 2. 注册和角色测试脚本

**文件路径**: `test/registration_and_role.go`

**功能**: 测试注册账号和创建角色的流程。

**使用方法**:

1. **运行脚本**:
   ```bash
   go run registration_and_role.go
   ```

## 测试流程

1. **启动服务**:
   - 启动 ssoauth 服务（验证服务器）
   - 启动 gateway 服务（网关服务器）
   - 启动 gamelogic 服务（游戏逻辑服务器）

2. **运行测试脚本**:
   - 执行注册账号操作
   - 执行创建角色操作
   - 执行获取背包数据操作

3. **查看测试结果**:
   - 脚本会输出测试过程中的各种信息
   - 检查是否成功完成所有测试步骤

## 注意事项

- 确保所有服务都已启动并正常运行
- 确保配置文件中的服务器地址和端口正确
- 确保网络连接正常，没有防火墙阻止连接
- 测试完成后，可以按 Ctrl+C 停止测试脚本

## 故障排除

- **连接失败**: 检查服务器是否启动，端口是否正确，网络是否正常
- **登录失败**: 检查用户名和密码是否正确，验证服务器是否正常运行
- **创建角色失败**: 检查游戏逻辑服务器是否正常运行，角色名称是否符合要求
- **获取背包数据失败**: 检查游戏逻辑服务器是否正常运行，角色是否已成功创建