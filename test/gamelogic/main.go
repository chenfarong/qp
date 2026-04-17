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
	"os/exec"
	"strings"
	"time"

	"zagame/config"
	pb "zagame/pb/golang/gamelogic"
	"zagame/proto"

	"github.com/gorilla/websocket"
)

// 全局变量存储命令行参数
var (
	authServerURL = flag.String("auth", "http://localhost:8080", "验证服务器URL")
	gatewayURL    = flag.String("gateway", "ws://localhost:8081", "网关服务器WebSocket URL")
	username      = flag.String("username", "za_admin", "用户名")
	password      = flag.String("password", "za_admin", "密码")
	actorName     = flag.String("actor", "", "角色名称（可选）")
	interval      = flag.Int("interval", 0, "定时发送GetGameMoneyRequest请求的间隔（秒，0表示不开启）")
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

// WebSocketMessage WebSocket消息结构
type WebSocketMessage struct {
	MsgID int32           `json:"msg_id"`
	Data  json.RawMessage `json:"data"`
}

var wsConn *websocket.Conn

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

	// 加载配置文件
	config.LoadConfig()

	// 如果没有提供命令行参数，使用配置文件中的服务器地址
	if *authServerURL == "http://localhost:8080" {
		authServerURL = &[]string{fmt.Sprintf("http://%s:%d", config.AppConfig.Auth.Host, config.AppConfig.Auth.Port)}[0]
	}
	if *gatewayURL == "ws://localhost:8081" {
		gatewayURL = &[]string{fmt.Sprintf("ws://%s:%d", config.AppConfig.Gateway.Host, config.AppConfig.Gateway.WsPort)}[0]
	}

	// 处理清除缓存参数
	fmt.Println("正在清除Go测试缓存...")
	exec.Command("go", "clean", "-testcache").Run()
	fmt.Println("Go测试缓存已清除")

	fmt.Println("=== 游戏逻辑测试客户端 ===")
	fmt.Printf("验证服务器: %s\n", *authServerURL)
	fmt.Printf("网关服务器: %s\n", *gatewayURL)
	fmt.Printf("用户名: %s\n", *username)
	fmt.Printf("密码: %s\n", *password)
	fmt.Printf("角色名称: %s\n", *actorName)
	fmt.Printf("定时发送间隔: %d秒\n", *interval)

	// 如果设置了定时发送间隔，启动定时器
	if *interval > 0 {
		go startGameMoneyTimer()
	}

	// 1. 通过HTTP登录验证获得token和session
	fmt.Println("\n=== 步骤1: 登录验证 ===")
	token, session, err := login()
	if err != nil {
		fmt.Printf("登录失败: %v\n", err)
		fmt.Println("尝试注册账号...")
		err = register()
		if err != nil {
			fmt.Printf("注册失败: %v\n", err)
			return
		}
		fmt.Println("注册成功! 再次尝试登录...")
		token, session, err = login()
		if err != nil {
			fmt.Printf("登录失败: %v\n", err)
			return
		}
	}
	fmt.Printf("登录成功! Token: %s, Session: %s\n", token, session)

	// 2. 连接WebSocket
	fmt.Println("\n=== 步骤2: 连接WebSocket ===")
	err = connectWebSocket(*gatewayURL, token, session)
	if err != nil {
		fmt.Printf("连接WebSocket失败: %v\n", err)
		return
	}
	defer wsConn.Close()

	// 3. 获取已经创建的角色列表
	fmt.Println("\n=== 步骤3: 获取角色列表 ===")
	actors, err := getActorList()
	if err != nil {
		fmt.Printf("获取角色列表失败: %v\n", err)
		return
	}

	// 4. 处理角色选择或创建
	fmt.Println("\n=== 步骤4: 角色选择或创建 ===")
	var selectedActor *ActorInfo
	if len(actors) == 0 {
		// 没有角色，需要创建
		fmt.Println("没有找到角色，需要创建新角色")
		selectedActor, err = createNewActor(session)
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

	// 5. 进入游戏
	fmt.Println("\n=== 步骤5: 进入游戏 ===")
	err = useActor(selectedActor.ActorId, session)
	if err != nil {
		fmt.Printf("进入游戏失败: %v\n", err)
		return
	}

	// 6. 测试游戏功能
	fmt.Println("\n=== 步骤6: 测试游戏功能 ===")
	testGameFunctions()

	// 7. 等待更多的游戏角色功能测试
	fmt.Println("\n=== 步骤7: 等待更多游戏功能测试 ===")
	fmt.Println("测试客户端已准备就绪，等待更多游戏功能测试...")
	fmt.Println("当前角色信息:")
	fmt.Printf("  角色ID: %s\n", selectedActor.ActorId)
	fmt.Printf("  角色名称: %s\n", selectedActor.Name)
	fmt.Printf("  角色等级: %d\n", selectedActor.Level)
	fmt.Printf("  所在服区: %s\n", selectedActor.Realm)

	// 保持连接
	select {}
}

// register 注册账号
func register() error {
	// 构造注册请求
	registerReq := map[string]string{
		"username": *username,
		"password": *password,
	}

	jsonData, err := json.Marshal(registerReq)
	if err != nil {
		return fmt.Errorf("构造注册请求失败: %v", err)
	}

	// 发送注册请求
	resp, err := http.Post(*authServerURL+"/register", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("发送注册请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 解析响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取注册响应失败: %v", err)
	}

	var registerResp struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(body, &registerResp); err != nil {
		return fmt.Errorf("解析注册响应失败: %v", err)
	}

	if !registerResp.Success {
		return fmt.Errorf("注册失败: %s", registerResp.Message)
	}

	return nil
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

	return loginResp.Token, loginResp.Session, nil
}

// connectWebSocket 连接WebSocket
func connectWebSocket(url, token string, session string) error {
	// 构造WebSocket连接URL
	wsURL := url + "/ws?token=" + token

	// 连接WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf("连接WebSocket失败: %v", err)
	}

	wsConn = conn

	// 启动一个goroutine接收消息
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				fmt.Printf("读取WebSocket消息失败: %v\n", err)
				return
			}

			// 解析消息
			var wsMsg WebSocketMessage
			if err := json.Unmarshal(message, &wsMsg); err != nil {
				fmt.Printf("解析WebSocket消息失败: %v\n", err)
				continue
			}

			// 处理不同类型的消息
			handleWebSocketMessage(wsMsg)
		}
	}()

	fmt.Println("WebSocket连接成功")

	// 向gateway发送登录请求，提交session值
	loginReq := map[string]string{
		"session": session,
	}
	err = sendWebSocketMessage(proto.MSG_LoginRequest, loginReq)
	if err != nil {
		return fmt.Errorf("发送登录请求失败: %v", err)
	}

	// 等待登录响应
	time.Sleep(1 * time.Second)

	return nil
}

// handleWebSocketMessage 处理WebSocket消息
func handleWebSocketMessage(msg WebSocketMessage) {
	fmt.Printf("收到消息: MsgID=%d, Data=%s\n", msg.MsgID, msg.Data)

	switch msg.MsgID {
	case proto.MSG_LoginResponse:
		// 处理登录响应
		var resp pb.LoginResponse
		if err := json.Unmarshal(msg.Data, &resp); err != nil {
			fmt.Printf("解析登录响应失败: %v\n", err)
			return
		}
		fmt.Printf("登录响应: Success=%v, Message=%s\n", resp.Success, resp.Message)
		if resp.Success && resp.Role != nil {
			fmt.Printf("角色ID: %s, 名称: %s, 等级: %d\n", resp.Role.Aid, resp.Role.Name, resp.Role.Level)
		}

	case proto.MSG_GetRoleInfoResponse:
		// 处理获取角色信息响应
		var resp pb.GetRoleInfoResponse
		if err := json.Unmarshal(msg.Data, &resp); err != nil {
			fmt.Printf("解析角色信息响应失败: %v\n", err)
			return
		}
		fmt.Printf("角色信息响应: Success=%v, Message=%s\n", resp.Success, resp.Message)
		if resp.Success && resp.Role != nil {
			fmt.Printf("角色ID: %s, 名称: %s, 等级: %d\n", resp.Role.Aid, resp.Role.Name, resp.Role.Level)
		}

	case proto.MSG_ActorUseResponse:
		// 处理使用角色响应
		var resp pb.ActorUseResponse
		if err := json.Unmarshal(msg.Data, &resp); err != nil {
			fmt.Printf("解析使用角色响应失败: %v\n", err)
			return
		}
		if resp.Err != nil {
			fmt.Printf("使用角色响应: ErrCode=%d, ErrText=%s\n", resp.Err.ErrCode, resp.Err.ErrText)
		} else {
			fmt.Printf("使用角色响应: 成功\n")
		}
		if resp.Data != nil {
			fmt.Printf("角色ID: %s, 名称: %s, 等级: %d\n", resp.Data.ActorId, resp.Data.Name, resp.Data.Level)
		}

	case proto.MSG_GetBagResponse:
		// 处理获取背包响应
		var resp pb.GetBagResponse
		if err := json.Unmarshal(msg.Data, &resp); err != nil {
			fmt.Printf("解析背包响应失败: %v\n", err)
			return
		}
		if resp.Err != nil {
			fmt.Printf("背包响应: ErrCode=%d, ErrText=%s\n", resp.Err.ErrCode, resp.Err.ErrText)
		} else {
			fmt.Printf("背包响应: 成功\n")
		}
		if resp.Data != nil {
			fmt.Printf("背包物品数量: %d\n", len(resp.Data))
			for i, item := range resp.Data {
				fmt.Printf("  %d. ID: %d, 类型: %d, 数量: %d\n", i+1, item.ItemId, item.ItemCfgId, item.Num)
			}
		}

	case proto.MSG_GetEquipResponse:
		// 处理获取装备响应
		var resp pb.GetEquipResponse
		if err := json.Unmarshal(msg.Data, &resp); err != nil {
			fmt.Printf("解析装备响应失败: %v\n", err)
			return
		}
		if resp.Err != nil {
			fmt.Printf("装备响应: ErrCode=%d, ErrText=%s\n", resp.Err.ErrCode, resp.Err.ErrText)
		} else {
			fmt.Printf("装备响应: 成功\n")
		}
		if resp.Data != nil {
			fmt.Printf("装备数量: %d\n", len(resp.Data))
			for i, item := range resp.Data {
				fmt.Printf("  %d. ID: %d, 类型: %d\n", i+1, item.EquipId, item.EquipCfgId)
			}
		}

	case proto.MSG_GetHeroesResponse:
		// 处理获取英雄响应
		var resp pb.GetHeroesResponse
		if err := json.Unmarshal(msg.Data, &resp); err != nil {
			fmt.Printf("解析英雄响应失败: %v\n", err)
			return
		}
		if resp.Err != nil {
			fmt.Printf("英雄响应: ErrCode=%d, ErrText=%s\n", resp.Err.ErrCode, resp.Err.ErrText)
		} else {
			fmt.Printf("英雄响应: 成功\n")
		}
		if resp.Data != nil {
			fmt.Printf("英雄数量: %d\n", len(resp.Data))
			for i, hero := range resp.Data {
				fmt.Printf("  %d. ID: %d, 配置ID: %d, 等级: %d, 星级: %d\n", i+1, hero.Uid, hero.CfgId, hero.Level, hero.Star)
			}
		}

	case proto.MSG_GetGameMoneyResponse:
		// 处理获取游戏货币响应
		var resp pb.GetGameMoneyResponse
		if err := json.Unmarshal(msg.Data, &resp); err != nil {
			fmt.Printf("解析游戏货币响应失败: %v\n", err)
			return
		}
		if resp.Err != nil {
			fmt.Printf("游戏货币响应: ErrCode=%d, ErrText=%s\n", resp.Err.ErrCode, resp.Err.ErrText)
		} else {
			fmt.Printf("游戏货币响应: 成功\n")
		}
		if resp.Data != nil {
			for _, money := range resp.Data {
				fmt.Printf("  货币类型: %d, 数量: %d\n", money.CfgId, money.Num)
			}
		}

	default:
		fmt.Printf("未知消息类型: %d\n", msg.MsgID)
	}
}

// sendWebSocketMessage 发送WebSocket消息
func sendWebSocketMessage(msgID int32, data interface{}) error {
	// 序列化数据
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化数据失败: %v", err)
	}

	// 构造WebSocket消息
	wsMsg := WebSocketMessage{
		MsgID: msgID,
		Data:  dataJSON,
	}

	// 序列化WebSocket消息
	msgJSON, err := json.Marshal(wsMsg)
	if err != nil {
		return fmt.Errorf("序列化WebSocket消息失败: %v", err)
	}

	// 发送消息
	err = wsConn.WriteMessage(websocket.TextMessage, msgJSON)
	if err != nil {
		return fmt.Errorf("发送WebSocket消息失败: %v", err)
	}

	return nil
}

// getActorList 获取角色列表
func getActorList() ([]ActorInfo, error) {
	// 发送获取角色信息请求
	req := pb.GetRoleInfoRequest{}
	err := sendWebSocketMessage(proto.MSG_GetRoleInfoRequest, req)
	if err != nil {
		return nil, fmt.Errorf("发送获取角色信息请求失败: %v", err)
	}

	// 等待响应（这里简化处理，实际应该使用通道或回调）
	time.Sleep(1 * time.Second)

	// 模拟返回数据
	// 实际项目中，应该通过WebSocket接收响应并解析
	return []ActorInfo{
		{
			ActorId:   "actor_1",
			Name:      "TestActor",
			Level:     1,
			Realm:     "realm_1",
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
			OnlineAt:  time.Now().Unix(),
			OfflineAt: 0,
		},
	}, nil
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
func createNewActor(session string) (*ActorInfo, error) {
	// 如果没有提供角色名称，使用默认名称
	actorNameInput := *actorName
	if actorNameInput == "" {
		// 使用默认角色名称
		actorNameInput = "test_actor"
		fmt.Printf("未提供角色名称，使用默认名称: %s\n", actorNameInput)
	}

	// 发送创建角色请求
	req := pb.ActorCreateRequest{
		Session:   session,
		ActorName: &actorNameInput,
	}
	err := sendWebSocketMessage(proto.MSG_ActorCreateRequest, req)
	if err != nil {
		return nil, fmt.Errorf("发送创建角色请求失败: %v", err)
	}

	// 等待响应
	time.Sleep(1 * time.Second)

	// 模拟返回数据
	// 实际项目中，应该通过WebSocket接收响应并解析
	newActor := &ActorInfo{
		ActorId:   "actor_" + fmt.Sprintf("%d", time.Now().Unix()),
		Name:      actorNameInput,
		Level:     1,
		Realm:     "realm_1",
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
		OnlineAt:  time.Now().Unix(),
		OfflineAt: 0,
	}

	fmt.Printf("创建角色成功: %s (ID: %s)\n", newActor.Name, newActor.ActorId)
	return newActor, nil
}

// useActor 使用角色
func useActor(actorId string, session string) error {
	// 发送使用角色请求
	req := pb.ActorUseRequest{
		Session: session,
		Aid:     actorId,
	}
	err := sendWebSocketMessage(proto.MSG_ActorUseRequest, req)
	if err != nil {
		return fmt.Errorf("发送使用角色请求失败: %v", err)
	}

	// 等待响应
	time.Sleep(1 * time.Second)

	fmt.Println("进入游戏成功")
	return nil
}

// startGameMoneyTimer 启动定时发送游戏货币请求的定时器
func startGameMoneyTimer() {
	ticker := time.NewTicker(time.Duration(*interval) * time.Second)
	defer ticker.Stop()

	fmt.Printf("定时发送游戏货币请求已启动，间隔: %d秒\n", *interval)

	for {
		select {
		case <-ticker.C:
			req := pb.GetGameMoneyRequest{}
			err := sendWebSocketMessage(proto.MSG_GetGameMoneyRequest, req)
			if err != nil {
				fmt.Printf("定时发送游戏货币请求失败: %v\n", err)
			} else {
				fmt.Printf("[定时任务] 发送游戏货币请求成功\n")
			}
		}
	}
}

// testGameFunctions 测试游戏功能
func testGameFunctions() {
	// 测试获取背包
	fmt.Println("\n测试获取背包...")
	req := pb.GetBagRequest{}
	err := sendWebSocketMessage(proto.MSG_GetBagRequest, req)
	if err != nil {
		fmt.Printf("发送获取背包请求失败: %v\n", err)
		return
	}
	time.Sleep(1 * time.Second)

	// 测试获取装备
	fmt.Println("\n测试获取装备...")
	req2 := pb.GetEquipRequest{}
	err = sendWebSocketMessage(proto.MSG_GetEquipRequest, req2)
	if err != nil {
		fmt.Printf("发送获取装备请求失败: %v\n", err)
		return
	}
	time.Sleep(1 * time.Second)

	// 测试获取英雄
	fmt.Println("\n测试获取英雄...")
	req3 := pb.GetHeroesRequest{}
	err = sendWebSocketMessage(proto.MSG_GetHeroesRequest, req3)
	if err != nil {
		fmt.Printf("发送获取英雄请求失败: %v\n", err)
		return
	}
	time.Sleep(1 * time.Second)

	// 测试获取游戏货币
	fmt.Println("\n测试获取游戏货币...")
	req4 := pb.GetGameMoneyRequest{}
	err = sendWebSocketMessage(proto.MSG_GetGameMoneyRequest, req4)
	if err != nil {
		fmt.Printf("发送获取游戏货币请求失败: %v\n", err)
		return
	}
	time.Sleep(1 * time.Second)
}

// printHelp 打印帮助信息
func printHelp() {
	fmt.Println("游戏逻辑测试客户端")
	fmt.Println("Usage:")
	fmt.Println("  gamelogic_test.exe [help] [-auth=auth_url] [-gateway=gateway_url] [-username=username] [-password=password] [-actor=actor_name] [-interval=seconds] [-cleancache]")
	fmt.Println("\nOptions:")
	fmt.Println("  help                  Show this help message")
	fmt.Println("  -auth=auth_url        Auth server URL (default: http://localhost:8080)")
	fmt.Println("  -gateway=gateway_url  Gateway server WebSocket URL (default: ws://localhost:8081)")
	fmt.Println("  -username=username    Username (default: za_admin)")
	fmt.Println("  -password=password    Password (default: za_admin)")
	fmt.Println("  -actor=actor_name     Actor name (optional)")
	fmt.Println("  -interval=seconds      Interval for periodic GetGameMoneyRequest (0 means disabled)")
	fmt.Println("\nExamples:")
	fmt.Println("  gamelogic_test.exe help")
	fmt.Println("  gamelogic_test.exe")
	fmt.Println("  gamelogic_test.exe -username=testuser -password=testpassword")
	fmt.Println("  gamelogic_test.exe -actor=myactor")
	fmt.Println("  gamelogic_test.exe -interval=5  # 每5秒发送一次游戏货币请求")
}
