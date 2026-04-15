package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 配置结构
type Config struct {
	Sandbox bool `yaml:"sandbox"`
	Server  struct {
		Logger struct {
			UdpPort int `yaml:"udp_port"`
		} `yaml:"logger"`
	} `yaml:"server"`
}

// LogMessage 日志消息结构
type LogMessage struct {
	ServerURI string    `json:"server_uri"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

func main() {
	// 加载配置
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 打印欢迎日志
	printWelcomeLog("Logger", 0, config.Server.Logger.UdpPort, "", 0, "")

	// 初始化日志文件
	logFile, err := initLogFile()
	if err != nil {
		log.Fatalf("Failed to initialize log file: %v", err)
	}
	defer logFile.Close()

	// 创建日志通道
	logChan := make(chan string, 1000)

	// 启动日志写入线程
	go func() {
		defer recoverPanic("Logger Write Thread")
		writeLogsToFile(logChan, logFile)
	}()

	// 启动UDP服务器
	defer recoverPanic("Logger UDP Server")
	startUDPServer(config.Server.Logger.UdpPort, logChan)
}

// loadConfig 加载配置文件
func loadConfig() (*Config, error) {
	file, err := os.Open("configs/config.yaml")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// initLogFile 初始化日志文件
func initLogFile() (*os.File, error) {
	// 创建logs目录
	logsDir := "logs"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return nil, err
	}

	// 创建日志文件
	logFileName := filepath.Join(logsDir, fmt.Sprintf("server_logs_%s.log", time.Now().Format("2006-01-02")))
	logFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return logFile, nil
}

// writeLogsToFile 写入日志到文件
func writeLogsToFile(logChan chan string, logFile *os.File) {
	writer := bufio.NewWriter(logFile)
	defer writer.Flush()

	for logMsg := range logChan {
		_, err := writer.WriteString(logMsg + "\n")
		if err != nil {
			log.Printf("Error writing log to file: %v", err)
		}
		// 每100条日志刷新一次
		writer.Flush()
	}
}

// startUDPServer 启动UDP服务器
func startUDPServer(port int, logChan chan string) {
	addr := fmt.Sprintf(":%d", port)
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Fatalf("Failed to resolve UDP address: %v", err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatalf("Failed to start UDP server: %v", err)
	}
	defer conn.Close()

	log.Printf("UDP log server started on port %d", port)

	// 缓冲区大小
	buffer := make([]byte, 4096)

	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("Error reading from UDP: %v", err)
			continue
		}

		// 处理收到的日志消息
		logMsg := string(buffer[:n])
		logChan <- logMsg

		// 打印到控制台
		log.Printf("Received log: %s", logMsg)
	}
}

// recoverPanic 恢复panic并打印错误信息和调用堆栈
func recoverPanic(serverName string) {
	if r := recover(); r != nil {
		// 捕获panic信息
		panicMsg := fmt.Sprintf("Panic recovered in %s server: %v\n%s", serverName, r, string(debug.Stack()))

		// 打印到控制台
		log.Printf("ERROR: %s", panicMsg)
	}
}

// printWelcomeLog 打印欢迎日志
func printWelcomeLog(serverType string, httpPort, udpPort int, dbHost string, dbPort int, dbName string) {
	// 打印欢迎日志
	log.Println("")
	log.Println("===============================================================")
	log.Printf("🎉 %s Server Welcome! 🎉", serverType)
	log.Println("===============================================================")
	log.Printf("🌐 Server Type: %s", serverType)
	if httpPort > 0 {
		log.Printf("🚪 HTTP Port: %d", httpPort)
	}
	if udpPort > 0 {
		log.Printf("🔗 UDP Port: %d", udpPort)
	}
	if dbHost != "" {
		log.Printf("🗄️  Database: %s:%d/%s", dbHost, dbPort, dbName)
	}
	log.Println("===============================================================")
	log.Println("")
}
