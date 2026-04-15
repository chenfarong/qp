package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
)

// 全局变量存储命令行参数
var (
	serverURL  = flag.String("url", "http://localhost:8080", "服务器URL")
	username   = flag.String("username", "testuser", "用户名")
	password   = flag.String("password", "testpassword", "密码")
	testMethod = flag.String("method", "all", "测试方法名 (all/register/login/profile)")
)

// 测试注册功能
func testRegister() {
	fmt.Println("=== Testing Register ===")

	// 注册请求
	registerReq := map[string]string{
		"username": *username,
		"password": *password,
	}

	// 将请求转换为JSON
	jsonData, err := json.Marshal(registerReq)
	if err != nil {
		fmt.Printf("Failed to marshal register request: %v\n", err)
		return
	}

	// 打印请求内容
	fmt.Printf("Register request: %s\n", string(jsonData))

	// 发送POST请求到注册接口
	resp, err := http.Post(*serverURL+"/register", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Failed to send register request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// 解析响应
	var registerResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&registerResp); err != nil {
		fmt.Printf("Failed to decode register response: %v\n", err)
		return
	}

	// 打印完整响应内容
	respJson, _ := json.MarshalIndent(registerResp, "", "  ")
	fmt.Printf("Register response: %s\n", string(respJson))

	// 检查响应
	if success, ok := registerResp["success"].(bool); ok && success {
		fmt.Println("Register successful")
	} else {
		fmt.Printf("Register failed: %v\n", registerResp["message"])
	}
}

// 测试登录功能
func testLogin() (string, string) {
	fmt.Println("\n=== Testing Login ===")

	// 登录请求
	loginReq := map[string]string{
		"username": *username,
		"password": *password,
	}

	// 将请求转换为JSON
	jsonData, err := json.Marshal(loginReq)
	if err != nil {
		fmt.Printf("Failed to marshal login request: %v\n", err)
		return "", ""
	}

	// 打印请求内容
	fmt.Printf("Login request: %s\n", string(jsonData))

	// 发送POST请求到登录接口
	resp, err := http.Post(*serverURL+"/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Failed to send login request: %v\n", err)
		return "", ""
	}
	defer resp.Body.Close()

	// 解析响应
	var loginResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		fmt.Printf("Failed to decode login response: %v\n", err)
		return "", ""
	}

	// 打印完整响应内容
	respJson, _ := json.MarshalIndent(loginResp, "", "  ")
	fmt.Printf("Login response: %s\n", string(respJson))

	// 检查响应
	if success, ok := loginResp["success"].(bool); ok && success {
		fmt.Println("Login successful")
		session, _ := loginResp["session"].(string)
		token, _ := loginResp["token"].(string)
		fmt.Printf("Session: %s\n", session)
		fmt.Printf("Token: %s\n", token)
		return session, token
	} else {
		fmt.Printf("Login failed: %v\n", loginResp["message"])
		return "", ""
	}
}

// 测试使用JWT token访问受保护的接口
func testProfile(token string) {
	fmt.Println("\n=== Testing Profile ===")

	// 使用token访问profile接口
	req, err := http.NewRequest("GET", *serverURL+"/auth/profile", nil)
	if err != nil {
		fmt.Printf("Failed to create profile request: %v\n", err)
		return
	}

	// 添加Authorization header
	req.Header.Set("Authorization", "Bearer "+token)

	// 打印请求信息
	fmt.Printf("Profile request URL: %s\n", req.URL.String())
	fmt.Printf("Profile request Authorization header: %s\n", req.Header.Get("Authorization"))

	// 发送请求
	client := &http.Client{}
	profileResp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Failed to send profile request: %v\n", err)
		return
	}
	defer profileResp.Body.Close()

	// 解析响应
	var profileData map[string]interface{}
	if err := json.NewDecoder(profileResp.Body).Decode(&profileData); err != nil {
		fmt.Printf("Failed to decode profile response: %v\n", err)
		return
	}

	// 打印完整响应内容
	respJson, _ := json.MarshalIndent(profileData, "", "  ")
	fmt.Printf("Profile response: %s\n", string(respJson))

	// 检查响应
	if success, ok := profileData["success"].(bool); ok && success {
		fmt.Println("Profile request successful")
		username, _ := profileData["username"].(string)
		fmt.Printf("Username: %s\n", username)
	} else {
		fmt.Printf("Profile request failed: %v\n", profileData["message"])
	}
}

func main() {
	// 检查是否有help参数
	for _, arg := range os.Args[1:] {
		if arg == "help" {
			printHelp()
			return
		}
	}

	// 解析命令行参数
	flag.Parse()

	fmt.Println("Testing SSO Auth Server")
	fmt.Printf("Server URL: %s\n", *serverURL)
	fmt.Printf("Username: %s\n", *username)
	fmt.Printf("Password: %s\n", *password)
	fmt.Printf("Test Method: %s\n", *testMethod)

	switch *testMethod {
	case "register":
		// 只测试注册
		testRegister()
		break
	case "login":
		// 只测试登录
		testLogin()
		break
	case "profile":
		// 只测试个人信息
		// 先登录获取token
		_, token := testLogin()
		if token != "" {
			testProfile(token)
		}
		break
	default:
		// 测试注册
		testRegister()

		// 测试登录
		_, token := testLogin()

		// 测试访问受保护的接口
		if token != "" {
			testProfile(token)
		}
		break
	}

	fmt.Println("\nTesting completed")
}

// printHelp 打印帮助信息
func printHelp() {
	fmt.Println("SSO Auth Server Test Tool")
	fmt.Println("Usage:")
	fmt.Println("  ssoauth_test.exe [help] [-url=server_url] [-username=username] [-password=password] [-method=test_method]")
	fmt.Println("\nOptions:")
	fmt.Println("  help                Show this help message")
	fmt.Println("  -url=server_url     Server URL (default: http://localhost:8080)")
	fmt.Println("  -username=username  Username (default: testuser)")
	fmt.Println("  -password=password  Password (default: testpassword)")
	fmt.Println("  -method=test_method Test method (default: all)")
	fmt.Println("\nTest Methods:")
	fmt.Println("  all         Run all tests (register, login, profile)")
	fmt.Println("  register    Run only register test")
	fmt.Println("  login       Run only login test")
	fmt.Println("  profile     Run only profile test (will login first)")
	fmt.Println("\nExamples:")
	fmt.Println("  ssoauth_test.exe help")
	fmt.Println("  ssoauth_test.exe")
	fmt.Println("  ssoauth_test.exe -method=login")
	fmt.Println("  ssoauth_test.exe -url=http://localhost:8080 -username=testuser -password=testpassword")
}
