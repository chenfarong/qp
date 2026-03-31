package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	// 获取当前工作目录
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current working directory: %v", err)
	}

	// 定义路径（基于当前工作目录）
	protoDir := filepath.Join(cwd, "proto")
	outputDir := filepath.Join(cwd, "kits", "proto")

	// 检查proto目录是否存在
	if _, err := os.Stat(protoDir); os.IsNotExist(err) {
		log.Fatalf("Proto directory does not exist: %s", protoDir)
	}

	// 检查output目录是否存在
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		log.Fatalf("Output directory does not exist: %s", outputDir)
	}

	// 读取proto目录下的所有.proto文件
	protoFiles, err := filepath.Glob(filepath.Join(protoDir, "*.proto"))
	if err != nil {
		log.Fatalf("Error finding proto files: %v", err)
	}

	if len(protoFiles) == 0 {
		log.Fatalf("No proto files found in %s", protoDir)
	}

	// 生成Go代码
	log.Println("Generating Go code...")
	for _, protoFile := range protoFiles {
		if err := generateGoCode(protoFile, protoDir); err != nil {
			log.Printf("Error generating Go code for %s: %v", protoFile, err)
		}
	}

	// 生成CocosJS代码
	log.Println("Generating CocosJS code...")
	for _, protoFile := range protoFiles {
		if err := generateJSCode(protoFile, protoDir, outputDir); err != nil {
			log.Printf("Error generating CocosJS code for %s: %v", protoFile, err)
		}
	}

	log.Println("Generation completed successfully!")
}

// generateGoCode 生成Go代码
func generateGoCode(protoFile, protoDir string) error {
	// 使用protoc命令生成Go代码
	cmd := exec.Command(
		"protoc",
		"--go_out=.",
		"--go_opt=paths=source_relative",
		filepath.Base(protoFile),
	)

	// 设置工作目录为protoDir
	cmd.Dir = protoDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// 消息类型结构
type MessageType struct {
	Name  string
	Value int
}

// 消息结构
type Message struct {
	Name   string
	Fields []Field
}

// 字段结构
type Field struct {
	Name string
	Type string
}

// parseProtoFile 解析proto文件
func parseProtoFile(protoFile string) ([]MessageType, []Message, error) {
	// 读取proto文件内容
	content, err := os.ReadFile(protoFile)
	if err != nil {
		return nil, nil, err
	}

	// 解析消息类型
	messageTypes := parseMessageTypes(string(content))
	// 解析消息
	messages := parseMessages(string(content))

	return messageTypes, messages, nil
}

// parseMessageTypes 解析消息类型枚举
func parseMessageTypes(content string) []MessageType {
	var messageTypes []MessageType

	// 查找MessageType枚举
	lines := strings.Split(content, "\n")
	inEnum := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "enum MessageType") {
			inEnum = true
			continue
		}

		if inEnum {
			if line == "}" {
				break
			}

			// 提取枚举值
			parts := strings.Split(line, "=")
			if len(parts) >= 2 {
				name := strings.TrimSpace(parts[0])
				if name != "" {
					// 提取值
					valueStr := strings.TrimSpace(parts[1])
					valueStr = strings.Split(valueStr, ";")[0] // 移除分号
					value := 0
					fmt.Sscanf(valueStr, "%d", &value)
					messageTypes = append(messageTypes, MessageType{
						Name:  name,
						Value: value,
					})
				}
			}
		}
	}

	return messageTypes
}

// parseMessages 解析消息结构
func parseMessages(content string) []Message {
	var messages []Message

	// 查找message定义
	lines := strings.Split(content, "\n")
	var currentMessage *Message
	inOneOf := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 开始新消息
		if strings.HasPrefix(line, "message ") {
			parts := strings.Split(line, " ")
			if len(parts) >= 2 {
				messageName := strings.TrimSuffix(parts[1], " {")
				currentMessage = &Message{
					Name:   messageName,
					Fields: []Field{},
				}
			}
			inOneOf = false
			continue
		}

		// 结束消息
		if line == "}" && currentMessage != nil {
			messages = append(messages, *currentMessage)
			currentMessage = nil
			inOneOf = false
			continue
		}

		// 检查是否进入oneof
		if strings.HasPrefix(line, "oneof ") {
			inOneOf = true
			continue
		}

		// 检查是否退出oneof
		if inOneOf && line == "}" {
			inOneOf = false
			continue
		}

		// 解析字段
		if currentMessage != nil && !strings.HasPrefix(line, "//") && line != "" {
			// 跳过oneof字段（在CocosJS中我们使用简单的JSON结构）
			if inOneOf {
				continue
			}

			// 简单解析字段
			parts := strings.Split(line, " ")
			if len(parts) >= 3 {
				// 处理repeated字段
				fieldType := parts[0]
				fieldName := parts[1]
				if fieldType == "repeated" {
					if len(parts) >= 4 {
						fieldType = parts[1]
						fieldName = parts[2]
					}
				}
				// 移除分号
				fieldName = strings.TrimSuffix(fieldName, ";")
				currentMessage.Fields = append(currentMessage.Fields, Field{
					Name: fieldName,
					Type: fieldType,
				})
			}
		}
	}

	return messages
}

// generateJSCode 生成CocosJS代码（基于JSON）
func generateJSCode(protoFile, protoDir, outputDir string) error {
	// 获取文件名（不含扩展名）
	fileName := filepath.Base(protoFile)
	baseName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	jsOutputFile := filepath.Join(outputDir, baseName+".js")

	// 解析proto文件
	messageTypes, messages, err := parseProtoFile(protoFile)
	if err != nil {
		return err
	}

	// 生成基于JSON的CocosJS代码
	jsContent := generateJSONBasedJSCode(baseName, messageTypes, messages)

	// 写入JS文件
	return os.WriteFile(jsOutputFile, []byte(jsContent), 0644)
}

// generateJSONBasedJSCode 生成基于JSON的CocosJS代码
func generateJSONBasedJSCode(baseName string, messageTypes []MessageType, messages []Message) string {
	// 生成消息类型代码
	messageTypeCode := ""
	for i, mt := range messageTypes {
		messageTypeCode += fmt.Sprintf("%s: %d", mt.Name, mt.Value)
		if i < len(messageTypes)-1 {
			messageTypeCode += ",\n\t\t"
		}
	}

	// 生成消息构造函数代码
	messageConstructorsCode := ""
	for _, msg := range messages {
		// 生成参数列表
		params := ""
		assignments := ""
		for i, field := range msg.Fields {
			params += field.Name
			assignments += fmt.Sprintf("\t\tthis.%s = %s;", field.Name, field.Name)
			if i < len(msg.Fields)-1 {
				params += ", "
				assignments += "\n"
			}
		}

		// 生成构造函数
		messageConstructorsCode += fmt.Sprintf(`
	// %s message
	%s: function(%s) {
%s
	},`, msg.Name, msg.Name, params, assignments)
	}

	return fmt.Sprintf(`// CocosJS code for %s
// Generated by proto generator
// Based on JSON serialization/deserialization

if (typeof window === 'undefined') {
	window = {};
}

if (!window.proto) {
	window.proto = {};
}

// %s namespace
window.proto.%s = {
	// Serialize message to JSON
	serialize: function(message) {
		return JSON.stringify(message);
	},
	
	// Deserialize JSON to message
	deserialize: function(json) {
		return JSON.parse(json);
	},
	
	// Message types
	MessageType: {
		%s
	},
	
	// Message wrapper
	Message: function(type, data) {
		this.type = type;
		this.data = data;
	},
	%s
};

// Expose to CocosJS
try {
	if (cc && cc.Class) {
		cc.proto = window.proto;
	}
} catch (e) {
	// CocosJS not found, continue with window.proto
}
`, baseName, baseName, baseName, messageTypeCode, messageConstructorsCode)
}
