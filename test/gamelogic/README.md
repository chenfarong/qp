# 游戏逻辑测试客户端

## 功能说明

该测试客户端用于测试游戏逻辑服务的功能，包括：
- 账号注册和登录
- 角色创建
- 进入游戏
- 获取背包信息
- 获取装备信息
- 获取英雄信息
- 获取游戏金币

## 编译方法

在 `test/gamelogic` 目录下执行以下命令：

```bash
go build -mod=mod -o gamelogic_test.exe main.go
```

## 配置文件

测试客户端会按照以下顺序加载配置文件：
1. 当前目录下的 `config.yml`
2. 当前目录下的 `test_config.yml`
3. 上级目录下的 `config.yml`
4. 上上级目录下的 `config.yml`
5. 上级目录下的 `test_config.yml`（相对于 `test/gamelogic` 目录）

配置文件示例 (`test_config.yml`)：

```yaml
# 测试配置文件
auth:
  host: "localhost"
  port: 8080

gateway:
  host: "localhost"
  ws_port: 8061
  grpc_port: 8082
```

## 命令行参数

测试客户端支持以下命令行参数：

- `-auth`：验证服务器 URL，默认值为 `http://localhost:8080`
- `-gateway`：网关服务器 WebSocket URL，默认值为 `ws://localhost:8081`
- `-username`：用户名，默认值为空（随机生成）
- `-password`：密码，默认值为空（随机生成）
- `-actor`：角色名称，默认值为空（随机生成）
- `-interval`：定时发送 GetGameMoneyRequest 请求的间隔（秒），默认值为 0（不开启）

## 使用样例

### 1. 使用默认配置和随机生成的账号密码

```bash
./gamelogic_test.exe
```

### 2. 指定服务器地址

```bash
./gamelogic_test.exe -auth http://localhost:8080 -gateway ws://localhost:8061
```

### 3. 指定账号密码和角色名称

```bash
./gamelogic_test.exe -username test_user -password test_pass -actor test_actor
```

### 4. 开启定时发送金币请求

```bash
./gamelogic_test.exe -interval 5
```

### 5. 完整参数示例

```bash
./gamelogic_test.exe -auth http://localhost:8080 -gateway ws://localhost:8061 -username test_user -password test_pass -actor test_actor -interval 5
```

## 测试流程

1. **登录验证**：通过 HTTP POST 向 ssoauth 提交账号和密码，获得 session
2. **网关验证**：通过 WebSocket 连接 gateway 服务器，连接成功后提交 session
3. **创建角色**：向 gateway 发送创建角色请求
4. **进入游戏**：向 gateway 发送进入游戏请求
5. **测试游戏功能**：获取背包、装备、英雄和游戏金币信息

## 注意事项

- 运行测试前，确保 ssoauth、gateway 和 gamelogic 服务已经启动
- 如果没有提供用户名和密码，测试客户端会自动随机生成
- 如果登录失败，测试客户端会尝试注册账号，然后再次登录
