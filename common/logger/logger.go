package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	PANIC
	FATAL
)

var levelNames = map[Level]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
	PANIC: "PANIC",
	FATAL: "FATAL",
}

func (l Level) String() string {
	if name, ok := levelNames[l]; ok {
		return name
	}
	return "UNKNOWN"
}

type OutputType int

const (
	Console OutputType = iota
	UDP
	File
)

var outputTypeNames = map[OutputType]string{
	Console: "Console",
	UDP:     "UDP",
	File:    "File",
}

func (o OutputType) String() string {
	if name, ok := outputTypeNames[o]; ok {
		return name
	}
	return "UNKNOWN"
}

type Config struct {
	ServerName string
	Level      Level
	Outputs    []OutputConfig
	UDPServer  string
	UDPPort    int
}

type OutputConfig struct {
	Type OutputType
	Addr string
}

var (
	defaultLogger *Logger
	once          sync.Once
)

type Logger struct {
	config     Config
	consoleLog *log.Logger
	fileLog    *log.Logger
	udpConn    *net.UDPConn
	mu         sync.Mutex
}

type LogMessage struct {
	Time       string                 `json:"time"`
	Level      string                 `json:"level"`
	ServerName string                 `json:"server_name"`
	Message    string                 `json:"message"`
	Source     string                 `json:"source"`
	Line       int                    `json:"line"`
	Fields     map[string]interface{} `json:"fields,omitempty"`
}

func New(config Config) *Logger {
	l := &Logger{
		config: config,
	}

	l.consoleLog = log.New(os.Stdout, "", 0)

	// 初始化文件日志
	for _, output := range config.Outputs {
		if output.Type == File {
			// 确保 logs 目录存在
			logDir := "logs"
			if _, err := os.Stat(logDir); os.IsNotExist(err) {
				err = os.MkdirAll(logDir, 0755)
				if err != nil {
					l.consoleLog.Printf("Failed to create log directory: %v\n", err)
					continue
				}
			}

			// 创建日志文件
			logFile := filepath.Join(logDir, fmt.Sprintf("%s_%s.log", config.ServerName, time.Now().Format("2006-01-02")))
			file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				l.consoleLog.Printf("Failed to open log file: %v\n", err)
				continue
			}

			l.fileLog = log.New(file, "", 0)
			break
		}
	}

	if config.UDPServer != "" && config.UDPPort > 0 {
		addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", config.UDPServer, config.UDPPort))
		if err == nil {
			conn, err := net.DialUDP("udp", nil, addr)
			if err == nil {
				l.udpConn = conn
			}
		}
	}

	return l
}

func Init(config Config) {
	once.Do(func() {
		defaultLogger = New(config)
	})
}

func Default() *Logger {
	if defaultLogger == nil {
		defaultLogger = New(Config{
			ServerName: "Unknown",
			Level:      INFO,
			Outputs:    []OutputConfig{{Type: Console}},
		})
	}
	return defaultLogger
}

func (l *Logger) shouldLog(level Level) bool {
	return level >= l.config.Level
}

func (l *Logger) formatMessage(level Level, v ...interface{}) (string, string, int) {
	msg := fmt.Sprint(v...)
	return l.formatString(level, msg)
}

func (l *Logger) formatMessagef(level Level, format string, v ...interface{}) (string, string, int) {
	msg := fmt.Sprintf(format, v...)
	return l.formatString(level, msg)
}

func (l *Logger) formatString(level Level, msg string) (string, string, int) {
	pc, file, line, ok := runtime.Caller(5)
	if !ok {
		return l.formatLogMsg(level, msg, "unknown", 0, nil), "", 0
	}

	filename := filepath.Base(file)

	funcName := "unknown"
	if fn := runtime.FuncForPC(pc); fn != nil {
		funcName = filepath.Base(fn.Name())
	}

	source := fmt.Sprintf("%s:%s", filename, funcName)

	return l.formatLogMsg(level, msg, source, line, nil), source, line
}

func (l *Logger) formatLogMsg(level Level, msg string, source string, line int, fields map[string]interface{}) string {
	logMsg := LogMessage{
		Time:       time.Now().Format("2006-01-02 15:04:05.000"),
		Level:      level.String(),
		ServerName: l.config.ServerName,
		Message:    msg,
		Source:     source,
		Line:       line,
		Fields:     fields,
	}

	jsonBytes, err := json.Marshal(logMsg)
	if err != nil {
		return fmt.Sprintf("[%s] [%s] [%s] [%s:%d] %s",
			logMsg.Time, logMsg.Level, logMsg.ServerName, source, line, logMsg.Message)
	}
	return string(jsonBytes)
}

func (l *Logger) log(level Level, v ...interface{}) {
	if !l.shouldLog(level) {
		return
	}

	msg, source, line := l.formatMessage(level, v...)

	l.mu.Lock()
	defer l.mu.Unlock()

	for _, output := range l.config.Outputs {
		switch output.Type {
		case Console:
			// 控制台使用非JSON格式
			consoleMsg := l.formatConsoleLog(level, fmt.Sprint(v...), source, line)
			l.consoleLog.Println(consoleMsg)
		case UDP:
			// UDP使用JSON格式
			l.sendUDP(msg, source, line)
		case File:
			// 文件使用JSON格式
			if l.fileLog != nil {
				l.fileLog.Println(msg)
			}
		}
	}
}

func (l *Logger) logf(level Level, format string, v ...interface{}) {
	if !l.shouldLog(level) {
		return
	}

	msg, source, line := l.formatMessagef(level, format, v...)

	l.mu.Lock()
	defer l.mu.Unlock()

	for _, output := range l.config.Outputs {
		switch output.Type {
		case Console:
			// 控制台使用非JSON格式
			consoleMsg := l.formatConsoleLog(level, fmt.Sprintf(format, v...), source, line)
			l.consoleLog.Println(consoleMsg)
		case UDP:
			// UDP使用JSON格式
			l.sendUDP(msg, source, line)
		case File:
			// 文件使用JSON格式
			if l.fileLog != nil {
				l.fileLog.Println(msg)
			}
		}
	}
}

// formatConsoleLog 格式化控制台日志
func (l *Logger) formatConsoleLog(level Level, msg string, source string, line int) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	return fmt.Sprintf("%s %s %s %s %s:%d",
		timestamp, l.config.ServerName, level.String(), msg, source, line)
}

// formatConsoleLogKV 格式化控制台日志（支持 key-value）
func (l *Logger) formatConsoleLogKV(level Level, msg string, source string, line int, fields map[string]interface{}) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	baseMsg := fmt.Sprintf("%s %s %s %s %s:%d",
		timestamp, l.config.ServerName, level.String(), msg, source, line)

	// 添加 key-value 对
	if len(fields) > 0 {
		var kvPairs []string
		for k, v := range fields {
			kvPairs = append(kvPairs, fmt.Sprintf("%s=%v", k, v))
		}
		baseMsg += " " + strings.Join(kvPairs, " ")
	}

	return baseMsg
}

// logkv 处理 key-value 格式的日志
func (l *Logger) logkv(level Level, msg string, keysAndValues ...interface{}) {
	if !l.shouldLog(level) {
		return
	}

	// 解析 key-value 对
	fields := make(map[string]interface{})
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			if key, ok := keysAndValues[i].(string); ok {
				fields[key] = keysAndValues[i+1]
			}
		}
	}

	// 获取源代码位置
	pc, file, line, ok := runtime.Caller(3)
	if !ok {
		file = "unknown"
		line = 0
	}

	filename := filepath.Base(file)

	funcName := "unknown"
	if fn := runtime.FuncForPC(pc); fn != nil {
		funcName = filepath.Base(fn.Name())
	}

	source := fmt.Sprintf("%s:%s", filename, funcName)

	l.mu.Lock()
	defer l.mu.Unlock()

	for _, output := range l.config.Outputs {
		switch output.Type {
		case Console:
			// 控制台使用非JSON格式，显示 key-value
			consoleMsg := l.formatConsoleLogKV(level, msg, source, line, fields)
			l.consoleLog.Println(consoleMsg)
		case UDP:
			// UDP使用JSON格式
			msgJSON := l.formatLogMsg(level, msg, source, line, fields)
			l.sendUDP(msgJSON, source, line)
		case File:
			// 文件使用JSON格式
			if l.fileLog != nil {
				msgJSON := l.formatLogMsg(level, msg, source, line, fields)
				l.fileLog.Println(msgJSON)
			}
		}
	}
}

func (l *Logger) sendUDP(msg string, source string, line int) {
	if l.udpConn == nil {
		return
	}

	_, err := l.udpConn.Write([]byte(msg))
	if err != nil {
		l.consoleLog.Printf("Failed to send UDP log: %v\n", err)
	}
}

func (l *Logger) Debug(v ...interface{}) {
	l.log(DEBUG, v...)
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	l.logf(DEBUG, format, v...)
}

func (l *Logger) Info(v ...interface{}) {
	l.log(INFO, v...)
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.logf(INFO, format, v...)
}

func (l *Logger) Warn(v ...interface{}) {
	l.log(WARN, v...)
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	l.logf(WARN, format, v...)
}

func (l *Logger) Error(v ...interface{}) {
	l.log(ERROR, v...)
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.logf(ERROR, format, v...)
}

func (l *Logger) Panic(v ...interface{}) {
	l.log(DEBUG, v...)
}

func (l *Logger) Panicf(format string, v ...interface{}) {
	l.logf(DEBUG, format, v...)
}

// 支持 key-value 格式的日志方法
func (l *Logger) DebugKV(msg string, keysAndValues ...interface{}) {
	l.logkv(DEBUG, msg, keysAndValues...)
}

func (l *Logger) InfoKV(msg string, keysAndValues ...interface{}) {
	l.logkv(INFO, msg, keysAndValues...)
}

func (l *Logger) WarnKV(msg string, keysAndValues ...interface{}) {
	l.logkv(WARN, msg, keysAndValues...)
}

func (l *Logger) ErrorKV(msg string, keysAndValues ...interface{}) {
	l.logkv(ERROR, msg, keysAndValues...)
}

func (l *Logger) PanicKV(msg string, keysAndValues ...interface{}) {
	l.logkv(PANIC, msg, keysAndValues...)
}

func (l *Logger) FatalKV(msg string, keysAndValues ...interface{}) {
	l.logkv(FATAL, msg, keysAndValues...)
	os.Exit(1)
}

func (l *Logger) Fatal(v ...interface{}) {
	l.log(FATAL, v...)
	os.Exit(1)
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.logf(FATAL, format, v...)
	os.Exit(1)
}

func (l *Logger) Close() {
	if l.udpConn != nil {
		l.udpConn.Close()
	}
	// 关闭文件日志
	if l.fileLog != nil {
		// 尝试关闭文件
		if file, ok := l.fileLog.Writer().(*os.File); ok {
			file.Close()
		}
	}
}

func SetServerName(name string) {
	if defaultLogger != nil {
		defaultLogger.config.ServerName = name
	}
}

func SetLevel(level Level) {
	if defaultLogger != nil {
		defaultLogger.config.Level = level
	}
}

func SetOutputs(outputs []OutputConfig) {
	if defaultLogger != nil {
		defaultLogger.config.Outputs = outputs
	}
}

func Debug(v ...interface{}) {
	Default().Debug(v...)
}

func Debugf(format string, v ...interface{}) {
	Default().Debugf(format, v...)
}

func Info(v ...interface{}) {
	Default().Info(v...)
}

func Infof(format string, v ...interface{}) {
	Default().Infof(format, v...)
}

func Warn(v ...interface{}) {
	Default().Warn(v...)
}

func Warnf(format string, v ...interface{}) {
	Default().Warnf(format, v...)
}

func Error(v ...interface{}) {
	Default().Error(v...)
}

func Errorf(format string, v ...interface{}) {
	Default().Errorf(format, v...)
}

func Panic(v ...interface{}) {
	Default().Panic(v...)
}

func Panicf(format string, v ...interface{}) {
	Default().Panicf(format, v...)
}

func Fatal(v ...interface{}) {
	Default().Fatal(v...)
}

func Fatalf(format string, v ...interface{}) {
	Default().Fatalf(format, v...)
}

// 支持 key-value 格式的全局日志方法
func DebugKV(msg string, keysAndValues ...interface{}) {
	Default().DebugKV(msg, keysAndValues...)
}

func InfoKV(msg string, keysAndValues ...interface{}) {
	Default().InfoKV(msg, keysAndValues...)
}

func WarnKV(msg string, keysAndValues ...interface{}) {
	Default().WarnKV(msg, keysAndValues...)
}

func ErrorKV(msg string, keysAndValues ...interface{}) {
	Default().ErrorKV(msg, keysAndValues...)
}

func PanicKV(msg string, keysAndValues ...interface{}) {
	Default().PanicKV(msg, keysAndValues...)
}

func FatalKV(msg string, keysAndValues ...interface{}) {
	Default().FatalKV(msg, keysAndValues...)
}

func Close() {
	Default().Close()
}

func ParseLevel(s string) Level {
	s = strings.ToUpper(s)
	switch s {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN", "WARNING":
		return WARN
	case "ERROR":
		return ERROR
	case "PANIC":
		return PANIC
	case "FATAL":
		return FATAL
	default:
		return INFO
	}
}

func ParseOutputType(s string) OutputType {
	s = strings.ToUpper(s)
	switch s {
	case "CONSOLE":
		return Console
	case "UDP":
		return UDP
	case "FILE":
		return File
	default:
		return Console
	}
}
