# QP Game Server

QP Game Server是一个基于Go语言开发的游戏服务器，包含Gateway、GameLogic和SsoAuth三个核心模块，使用PostgreSQL作为数据库。

## 项目架构

### 模块说明

1. **Gateway** - 网关服务，负责处理客户端连接请求，将请求转发到相应的服务
2. **GameLogic** - 游戏逻辑服务，处理游戏核心逻辑，如角色创建、升级、战斗等
3. **SsoAuth** - 单点登录服务，负责用户账号创建、登录和认证

### 技术栈

- **语言**: Go 1.20
- **Web框架**: Gin
- **数据库**: PostgreSQL
- **认证**: JWT
- **通信**: HTTP/gRPC

## 目录结构

```
qp/
├── cmd/              # 命令行入口
│   ├── gateway/      # 网关服务
│   ├── gamelogic/    # 游戏逻辑服务
│   └── ssoauth/      # 单点登录服务
├── configs/          # 配置文件
├── internal/         # 内部包
│   ├── gateway/      # 网关相关代码
│   ├── gamelogic/    # 游戏逻辑相关代码
│   ├── ssoauth/      # 单点登录相关代码
│   └── common/       # 公共代码
├── pkg/              # 公共包
│   ├── db/           # 数据库相关代码
│   ├── proto/        # gRPC proto文件
│   └── utils/        # 工具函数
├── migrations/       # 数据库迁移脚本
├── scripts/          # 脚本文件
└── README.md         # 项目说明
```

## 安装与运行

### 前置条件

- Go 1.20+
- PostgreSQL 13+

### 安装步骤

1. 克隆项目

```bash
git clone https://github.com/aoyo/qp.git
cd qp
```

2. 安装依赖

```bash
go mod tidy
```

3. 初始化数据库

```bash
# 执行数据库初始化脚本
psql -U postgres -f migrations/init.sql
```

4. 配置文件

编辑 `configs/config.yaml` 文件，根据实际情况修改配置。

### 运行服务

1. 启动SsoAuth服务

```bash
go run cmd/ssoauth/main.go
```

2. 启动GameLogic服务

```bash
go run cmd/gamelogic/main.go
```

3. 启动Gateway服务

```bash
go run cmd/gateway/main.go
```

## API文档

### 认证相关API

- `POST /api/auth/register` - 用户注册
- `POST /api/auth/login` - 用户登录
- `GET /api/auth/validate` - 验证令牌

### 游戏相关API

- `POST /api/game/characters` - 创建角色
- `GET /api/game/characters` - 获取用户角色列表
- `GET /api/game/characters/:id` - 获取角色详情
- `PUT /api/game/characters/:id/status` - 更新角色状态
- `POST /api/game/battle` - 战斗

## 测试

运行测试用例：

```bash
go test ./...
```

## 部署

### 构建

```bash
go build -o bin/gateway cmd/gateway/main.go
go build -o bin/gamelogic cmd/gamelogic/main.go
go build -o bin/ssoauth cmd/ssoauth/main.go
```

### 运行

```bash
./bin/ssoauth
./bin/gamelogic
./bin/gateway
```

## 许可证

MIT
