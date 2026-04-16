package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// 全局变量存储命令行参数
var (
	authServerURL = flag.String("auth", "http://localhost:8080", "验证服务器URL")
	gameServerURL = flag.String("game", "http://localhost:8081", "游戏服务器URL")
	username      = flag.String("username", "za_admin", "用户名")
	password      = flag.String("password", "za_admin", "密码")
	actorName     = flag.String("actor", "", "角色名称（可选）")
)

// LoginResponse 登录响应结构
type LoginResponse struct {
	Success bool   `json:"success"`
	Session string `json:"session"`
	Token   string `json:"token"`
	Message string `json:"message"`
}

// ActorInfo 角色信息结构
type ActorInfo struct {
	ActorId   string `json:"actor_id"`
	Name      string `json:"name"`
	Level     int32  `json:"level"`
	Realm     string `json:"realm"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
	OnlineAt  int64  `json:"online_at"`
	OfflineAt int64  `json:"offline_at"`
}

// ActorListResponse 角色列表响应结构
type ActorListResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    []ActorInfo `json:"data"`
}

// ActorCreateRequest 创建角色请求结构
type ActorCreateRequest struct {
	Name string `json:"name"`
}

// ActorCreateResponse 创建角色响应结构
type ActorCreateResponse struct {
	Success bool      `json:"success"`
	Message string    `json:"message"`
	Data    ActorInfo `json:"data"`
}

// GatewayInfo 网关信息结构
type GatewayInfo struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// GatewayResponse 网关响应结构
type GatewayResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    GatewayInfo `json:"data"`
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

	fmt.Println("=== 游戏逻辑测试客户端 ===")
	fmt.Printf("验证服务器: %s\n", *authServerURL)
	fmt.Printf("游戏服务器: %s\n", *gameServerURL)
	fmt.Printf("用户名: %s\n", *username)
	fmt.Printf("密码: %s\n", *password)
	fmt.Printf("角色名称: %s\n", *actorName)

	// 1. 通过HTTP登录验证获得session
	fmt.Println("\n=== 步骤1: 登录验证 ===")
	session, token, err := login()
	if err != nil {
		fmt.Printf("登录失败: %v\n", err)
		return
	}
	fmt.Printf("登录成功! Session: %s\n", session)
	fmt.Printf("Token: %s\n", token)

	// 2. 获取已经创建的角色列表
	fmt.Println("\n=== 步骤2: 获取角色列表 ===")
	actors, err := getActorList(token)
	if err != nil {
		fmt.Printf("获取角色列表失败: %v\n", err)
		return
	}

	// 3. 处理角色选择或创建
	fmt.Println("\n=== 步骤3: 角色选择或创建 ===")
	var selectedActor *ActorInfo
	if len(actors) == 0 {
		// 没有角色，需要创建
		fmt.Println("没有找到角色，需要创建新角色")
		selectedActor, err = createNewActor(token)
		if err != nil {
			fmt.Printf("创建角色失败: %v\n", err)
			return
		}
	} else if len(actors) == 1 {
		// 只有一个角色，直接选择
		selectedActor = &actors[0]
		fmt.Printf("自动选择角色: %s (ID: %s)\n", selectedActor.Name, selectedActor.ActorId)
	} else {
		// 多个角色，需要选择
		if *actorName != "" {
			// 参数提供了角色名称，尝试查找
			for i := range actors {
				if actors[i].Name == *actorName {
					selectedActor = &actors[i]
					break
				}
			}
			if selectedActor == nil {
				fmt.Printf("未找到名称为 %s 的角色\n", *actorName)
				selectedActor = selectActor(actors)
			}
		} else {
			// 需要用户选择
			selectedActor = selectActor(actors)
		}
	}

	fmt.Printf("\n已选择角色: %s (ID: %s, 等级: %d)\n",
		selectedActor.Name, selectedActor.ActorId, selectedActor.Level)

	// 4. 等待更多的游戏角色功能测试
	fmt.Println("\n=== 步骤4: 等待更多游戏功能测试 ===")
	fmt.Println("测试客户端已准备就绪，等待更多游戏功能测试...")
	fmt.Println("当前角色信息:")
	fmt.Printf("  角色ID: %s\n", selectedActor.ActorId)
	fmt.Printf("  角色名称: %s\n", selectedActor.Name)
	fmt.Printf("  角色等级: %d\n", selectedActor.Level)
	fmt.Printf("  所在服区: %s\n", selectedActor.Realm)
}

// login 登录验证
func login() (string, string, error) {
	// 构造登录请求
	loginReq := map[string]string{
		"username": *username,
		"password": *password,
	}

	jsonData, err := json.Marshal(loginReq)
	if err != nil {
		return "", "", fmt.Errorf("构造登录请求失败: %v", err)
	}

	// 发送登录请求
	resp, err := http.Post(*authServerURL+"/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", "", fmt.Errorf("发送登录请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 解析响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("读取登录响应失败: %v", err)
	}

	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return "", "", fmt.Errorf("解析登录响应失败: %v", err)
	}

	if !loginResp.Success {
		return "", "", fmt.Errorf("登录失败: %s", loginResp.Message)
	}

	return loginResp.Session, loginResp.Token, nil
}

// getActorList 获取角色列表
func getActorList(token string) ([]ActorInfo, error) {
	// 构造请求
	req, err := http.NewRequest("GET", *gameServerURL+"/actor/list", nil)
	if err != nil {
		return nil, fmt.Errorf("构造请求失败: %v", err)
	}

	// 添加Authorization header
	req.Header.Set("Authorization", "Bearer "+token)

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 解析响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var actorListResp ActorListResponse
	if err := json.Unmarshal(body, &actorListResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if !actorListResp.Success {
		return nil, fmt.Errorf("获取角色列表失败: %s", actorListResp.Message)
	}

	fmt.Printf("找到 %d 个角色\n", len(actorListResp.Data))
	for i, actor := range actorListResp.Data {
		fmt.Printf("  %d. %s (ID: %s, 等级: %d)\n", i+1, actor.Name, actor.ActorId, actor.Level)
	}

	return actorListResp.Data, nil
}

// selectActor 选择角色
func selectActor(actors []ActorInfo) *ActorInfo {
	fmt.Println("\n请选择一个角色:")
	for i, actor := range actors {
		fmt.Printf("  %d. %s (ID: %s, 等级: %d)\n", i+1, actor.Name, actor.ActorId, actor.Level)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("请输入角色编号: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		var choice int
		if _, err := fmt.Sscanf(input, "%d", &choice); err != nil {
			fmt.Println("输入无效，请输入数字")
			continue
		}

		if choice < 1 || choice > len(actors) {
			fmt.Printf("输入无效，请输入 1-%d 之间的数字\n", len(actors))
			continue
		}

		return &actors[choice-1]
	}
}

// createNewActor 创建新角色
func createNewActor(token string) (*ActorInfo, error) {
	// 如果没有提供角色名称，使用默认名称
	actorNameInput := *actorName
	if actorNameInput == "" {
		// 使用默认角色名称
		actorNameInput = "test_actor"
		fmt.Printf("未提供角色名称，使用默认名称: %s\n", actorNameInput)
	}

	// 构造创建角色请求
	createReq := ActorCreateRequest{
		Name: actorNameInput,
	}

	jsonData, err := json.Marshal(createReq)
	if err != nil {
		return nil, fmt.Errorf("构造创建角色请求失败: %v", err)
	}

	// 构造请求
	req, err := http.NewRequest("POST", *gameServerURL+"/actor/create", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("构造请求失败: %v", err)
	}

	// 添加Authorization header
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 解析响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var createResp ActorCreateResponse
	if err := json.Unmarshal(body, &createResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if !createResp.Success {
		return nil, fmt.Errorf("创建角色失败: %s", createResp.Message)
	}

	fmt.Printf("创建角色成功: %s (ID: %s)\n", createResp.Data.Name, createResp.Data.ActorId)
	return &createResp.Data, nil
}

// printHelp 打印帮助信息
func printHelp() {
	fmt.Println("游戏逻辑测试客户端")
	fmt.Println("Usage:")
	fmt.Println("  gamelogic_test.exe [help] [-auth=auth_url] [-game=game_url] [-username=username] [-password=password] [-actor=actor_name]")
	fmt.Println("\nOptions:")
	fmt.Println("  help                Show this help message")
	fmt.Println("  -auth=auth_url      Auth server URL (default: http://localhost:8080)")
	fmt.Println("  -game=game_url      Game server URL (default: http://localhost:8081)")
	fmt.Println("  -username=username  Username (default: za_admin)")
	fmt.Println("  -password=password  Password (default: za_admin)")
	fmt.Println("  -actor=actor_name   Actor name (optional)")
	fmt.Println("\nExamples:")
	fmt.Println("  gamelogic_test.exe help")
	fmt.Println("  gamelogic_test.exe")
	fmt.Println("  gamelogic_test.exe -username=testuser -password=testpassword")
	fmt.Println("  gamelogic_test.exe -actor=myactor")
}
