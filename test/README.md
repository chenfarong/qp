# 测试执行程序使用说明

## 执行程序位置
测试执行程序位于 `test/bin` 目录下：
- Windows: `test/bin/ssoauth_test.exe`

## 运行方式

### 基本运行
直接运行执行程序，使用默认参数：

```bash
test/bin/ssoauth_test.exe
```

### 使用命令行参数
可以通过命令行参数指定服务器URL、用户名和密码：

```bash
test/bin/ssoauth_test.exe -url=http://localhost:8080 -username=testuser -password=testpassword
```

## 可用的命令行参数

| 参数 | 描述 | 默认值 |
|------|------|--------|
| `-url` | 服务器URL | `http://localhost:8080` |
| `-username` | 用户名 | `testuser` |
| `-password` | 密码 | `testpassword` |
| `-method` | 测试方法名 | `all` |

### 测试方法说明

| 方法名 | 描述 |
|--------|------|
| `all` | 运行所有测试（注册、登录、个人信息） |
| `register` | 只运行注册测试 |
| `login` | 只运行登录测试 |
| `profile` | 只运行个人信息测试（会先登录获取token） |

### 示例用法

运行所有测试：
```bash
test/bin/ssoauth_test.exe
```

只运行登录测试：
```bash
test/bin/ssoauth_test.exe -method=login
```

指定服务器URL和只运行注册测试：
```bash
test/bin/ssoauth_test.exe -url=http://localhost:8080 -method=register
```

## 测试流程

执行程序会按照以下流程进行测试：

1. **注册测试**：向服务器发送注册请求，测试注册功能
2. **登录测试**：向服务器发送登录请求，测试登录功能并获取session和token
3. **个人信息测试**：使用获取到的token访问个人信息接口，测试JWT验证功能

## 测试输出

执行程序会在控制台打印详细的测试信息，包括：

- 命令行参数值
- 每个测试步骤的请求内容
- 每个测试步骤的响应内容
- 测试结果

## 示例输出

```
Testing SSO Auth Server
Server URL: http://localhost:8080
Username: testuser
Password: testpassword
=== Testing Register ===
Register request: {"password":"testpassword","username":"testuser"}
Register response: {
  "message": "Username already exists",
  "success": false
}
Register failed: Username already exists

=== Testing Login ===
Login request: {"password":"testpassword","username":"testuser"}
Login response: {
  "message": "Login successful",
  "session": "q0cx6bk8nhydj2v3hn9p6aqyc8eg9w1f",
  "success": true,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6InRlc3R1c2VyIiwiZXhwIjoxNzc2MzM0ODMwLCJuYmYiOjE3NzYyNDg0MzAsImlhdCI6MTc3NjI0ODQzMH0.I67_acgFcc_LiS2PZDq-mfj25JV9ufEc8DMcQyzPQi8"
}
Login successful
Session: q0cx6bk8nhydj2v3hn9p6aqyc8eg9w1f
Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6InRlc3R1c2VyIiwiZXhwIjoxNzc2MzM0ODMwLCJuYmYiOjE3NzYyNDg0MzAsImlhdCI6MTc3NjI0ODQzMH0.I67_acgFcc_LiS2PZDq-mfj25JV9ufEc8DMcQyzPQi8

=== Testing Profile ===
Profile request URL: http://localhost:8080/auth/profile
Profile request Authorization header: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6InRlc3R1c2VyIiwiZXhwIjoxNzc2MzM0ODMwLCJuYmYiOjE3NzYyNDg0MzAsImlhdCI6MTc3NjI0ODQzMH0.I67_acgFcc_LiS2PZDq-mfj25JV9ufEc8DMcQyzPQi8
Profile response: {
  "message": "Profile retrieved successfully",
  "success": true,
  "username": "testuser"
}
Profile request successful
Username: testuser

Testing completed
```

## 注意事项

- 确保ssoauth服务器正在运行，并且监听在指定的URL上
- 注册测试可能会失败，因为用户可能已经存在，这是正常现象
- 登录测试需要使用正确的用户名和密码
- 个人信息测试需要服务器支持JWT验证功能