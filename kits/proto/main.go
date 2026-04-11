package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// 支持的输出语言类型
const (
	LangGo    = "go"
	LangCpp   = "cpp"
	LangJS    = "js"
	LangCocos = "cocos"
	LangTS    = "ts"
)

func main() {
	// 自定义帮助信息
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s --proto-dir=proto --output-dir=bin --lang=go\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --help\n", os.Args[0])
	}

	// 定义命令行参数
	var (
		protoDir  = flag.String("proto-dir", "", "Protocol buffer definition directory (default: proto)")
		outputDir = flag.String("output-dir", "", "Output directory (default: bin)")
		lang      = flag.String("lang", "go", "Output language type: go, cpp, js, cocos, ts (default: go)")
	)
	flag.Parse()

	// 设置默认 proto 目录
	if *protoDir == "" {
		*protoDir = "proto"
		log.Printf("Using default proto directory: %s", *protoDir)
	}

	// 设置默认输出目录
	if *outputDir == "" {
		*outputDir = "bin"
		log.Printf("Using default output directory: %s", *outputDir)
	}

	// 将相对路径转换为绝对路径
	var err error
	*protoDir, err = filepath.Abs(*protoDir)
	if err != nil {
		log.Fatalf("Failed to get absolute path for proto directory: %v", err)
	}

	*outputDir, err = filepath.Abs(*outputDir)
	if err != nil {
		log.Fatalf("Failed to get absolute path for output directory: %v", err)
	}

	log.Printf("Absolute proto directory: %s", *protoDir)
	log.Printf("Absolute output directory: %s", *outputDir)

	// 验证语言类型
	validLangs := map[string]bool{
		LangGo:    true,
		LangCpp:   true,
		LangJS:    true,
		LangCocos: true,
		LangTS:    true,
	}
	if !validLangs[*lang] {
		log.Fatalf("Error: unsupported language type '%s'. Supported types: go, cpp, js, cocos, ts", *lang)
	}

	// 检查proto目录是否存在，不存在则创建
	if _, err := os.Stat(*protoDir); os.IsNotExist(err) {
		if err := os.MkdirAll(*protoDir, 0755); err != nil {
			log.Fatalf("Failed to create proto directory: %s", *protoDir)
		}
		log.Printf("Created proto directory: %s", *protoDir)
	}

	// 检查output目录是否存在，不存在则创建
	if _, err := os.Stat(*outputDir); os.IsNotExist(err) {
		if err := os.MkdirAll(*outputDir, 0755); err != nil {
			log.Fatalf("Failed to create output directory: %s", *outputDir)
		}
		log.Printf("Created output directory: %s", *outputDir)
	}

	// 读取proto目录下的所有.proto文件
	protoFiles, err := filepath.Glob(filepath.Join(*protoDir, "*.proto"))
	if err != nil {
		log.Fatalf("Error finding proto files: %v", err)
	}

	if len(protoFiles) == 0 {
		log.Fatalf("No proto files found in %s", *protoDir)
	}

	log.Printf("Found %d proto files in %s", len(protoFiles), *protoDir)
	log.Printf("Output language: %s", *lang)
	log.Printf("Output directory: %s", *outputDir)

	// 根据语言类型生成代码
	switch *lang {
	case LangGo:
		log.Println("Generating Go code...")
		for _, protoFile := range protoFiles {
			if err := generateGoCode(protoFile, *protoDir, *outputDir); err != nil {
				log.Printf("Error generating Go code for %s: %v", protoFile, err)
			}
		}
	case LangCpp:
		log.Println("Generating C++ header files...")
		for _, protoFile := range protoFiles {
			if err := generateCppCode(protoFile, *protoDir, *outputDir); err != nil {
				log.Printf("Error generating C++ code for %s: %v", protoFile, err)
			}
		}
	case LangJS, LangCocos:
		log.Println("Generating CocosJS code...")
		for _, protoFile := range protoFiles {
			if err := generateJSCode(protoFile, *protoDir, *outputDir); err != nil {
				log.Printf("Error generating CocosJS code for %s: %v", protoFile, err)
			}
		}
	case LangTS:
		log.Println("Generating TypeScript code...")
		for _, protoFile := range protoFiles {
			if err := generateTSCode(protoFile, *protoDir, *outputDir); err != nil {
				log.Printf("Error generating TypeScript code for %s: %v", protoFile, err)
			}
		}
	}

	log.Println("Generation completed successfully!")
}

// generateGoCode 生成Go代码
func generateGoCode(protoFile, protoDir, outputDir string) error {
	// 使用protoc命令生成Go代码和gRPC代码
	cmd := exec.Command(
		"protoc",
		"--go_out="+outputDir,
		"--go_opt=paths=source_relative",
		"--go-grpc_out="+outputDir,
		"--go-grpc_opt=paths=source_relative",
		filepath.Base(protoFile),
	)

	// 设置工作目录为protoDir
	cmd.Dir = protoDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// generateCppCode 生成C++头文件
func generateCppCode(protoFile, protoDir, outputDir string) error {
	// 获取文件名（不含扩展名）
	fileName := filepath.Base(protoFile)
	baseName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	hOutputFile := filepath.Join(outputDir, baseName+".h")

	// 解析proto文件
	messageTypes, messages, enums, imports, packageName, err := parseProtoFileDetailed(protoFile)
	if err != nil {
		return err
	}

	// 生成C++头文件内容
	hContent := generateCppHeader(baseName, messageTypes, messages, enums, imports, packageName)

	// 写入.h文件
	return os.WriteFile(hOutputFile, []byte(hContent), 0644)
}

// 消息类型结构
type MessageType struct {
	Name  string
	Value int
}

// 枚举结构
type Enum struct {
	Name   string
	Values []EnumValue
}

// 枚举值结构
type EnumValue struct {
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
	Name     string
	Type     string
	Number   int
	Repeated bool
	Optional bool
}

// parseProtoFileDetailed 详细解析proto文件
func parseProtoFileDetailed(protoFile string) ([]MessageType, []Message, []Enum, []string, string, error) {
	// 读取proto文件内容
	content, err := os.ReadFile(protoFile)
	if err != nil {
		return nil, nil, nil, nil, "", err
	}

	contentStr := string(content)

	// 解析包名
	packageName := parsePackageName(contentStr)

	// 解析导入语句
	imports := parseImports(contentStr)

	// 解析消息类型
	messageTypes := parseMessageTypes(contentStr)

	// 解析枚举
	enums := parseEnums(contentStr)

	// 解析消息
	messages := parseMessagesDetailed(contentStr)

	return messageTypes, messages, enums, imports, packageName, nil
}

// parsePackageName 解析包名
func parsePackageName(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "package ") {
			parts := strings.Split(line, " ")
			if len(parts) >= 2 {
				return strings.TrimSuffix(parts[1], ";")
			}
		}
	}
	return ""
}

// parseImports 解析导入语句
func parseImports(content string) []string {
	var imports []string
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "import ") {
			// 提取导入路径
			start := strings.Index(line, "\"")
			end := strings.LastIndex(line, "\"")
			if start != -1 && end != -1 && start < end {
				imports = append(imports, line[start+1:end])
			}
		}
	}
	return imports
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

// parseEnums 解析枚举定义
func parseEnums(content string) []Enum {
	var enums []Enum

	lines := strings.Split(content, "\n")
	var currentEnum *Enum
	inEnum := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 开始新枚举
		if strings.HasPrefix(line, "enum ") {
			parts := strings.Split(line, " ")
			if len(parts) >= 2 {
				enumName := strings.TrimSuffix(parts[1], " {")
				currentEnum = &Enum{
					Name:   enumName,
					Values: []EnumValue{},
				}
			}
			inEnum = true
			continue
		}

		// 结束枚举
		if line == "}" && currentEnum != nil {
			enums = append(enums, *currentEnum)
			currentEnum = nil
			inEnum = false
			continue
		}

		// 解析枚举值
		if inEnum && currentEnum != nil && !strings.HasPrefix(line, "//") && line != "" {
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
					currentEnum.Values = append(currentEnum.Values, EnumValue{
						Name:  name,
						Value: value,
					})
				}
			}
		}
	}

	return enums
}

// parseMessagesDetailed 详细解析消息结构
func parseMessagesDetailed(content string) []Message {
	var messages []Message

	lines := strings.Split(content, "\n")
	var currentMessage *Message
	inMessage := false
	inOneOf := false
	braceCount := 0
	oneOfBraceCount := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 跳过枚举定义
		if strings.HasPrefix(line, "enum ") {
			continue
		}

		// 开始新消息
		if strings.HasPrefix(line, "message ") {
			parts := strings.Split(line, " ")
			if len(parts) >= 2 {
				messageName := strings.TrimSuffix(parts[1], " {")
				if !strings.HasSuffix(line, "{") {
					// 需要继续读取
					continue
				}
				currentMessage = &Message{
					Name:   messageName,
					Fields: []Field{},
				}
			}
			inMessage = true
			braceCount = 1
			continue
		}

		// 处理消息内容
		if inMessage && currentMessage != nil {
			// 检查是否进入 oneof
			if strings.HasPrefix(line, "oneof ") {
				inOneOf = true
				oneOfBraceCount = 1
				continue
			}

			// 如果在 oneof 内部，跟踪括号
			if inOneOf {
				oneOfBraceCount += strings.Count(line, "{")
				oneOfBraceCount -= strings.Count(line, "}")
				if oneOfBraceCount == 0 {
					inOneOf = false
				}
				continue // 跳过 oneof 内部的字段
			}

			braceCount += strings.Count(line, "{")
			braceCount -= strings.Count(line, "}")

			// 解析字段（跳过 oneof、注释、空行、嵌套消息和枚举）
			if !strings.HasPrefix(line, "//") && line != "" &&
				!strings.HasPrefix(line, "message ") && !strings.HasPrefix(line, "enum ") &&
				!strings.HasPrefix(line, "oneof ") {
				field := parseField(line)
				if field.Name != "" && field.Type != "oneof" {
					currentMessage.Fields = append(currentMessage.Fields, field)
				}
			}

			// 结束消息
			if braceCount == 0 {
				messages = append(messages, *currentMessage)
				currentMessage = nil
				inMessage = false
			}
		}
	}

	return messages
}

// parseField 解析字段
func parseField(line string) Field {
	field := Field{}

	// 移除分号
	line = strings.TrimSuffix(line, ";")

	// 处理 repeated 和 optional
	parts := strings.Fields(line)
	isRepeated := false
	isOptional := false

	for i, part := range parts {
		if part == "repeated" {
			isRepeated = true
			continue
		}
		if part == "optional" {
			isOptional = true
			continue
		}

		// 解析字段 - 格式: [repeated|optional] 类型 字段名 = 字段号
		// i 指向类型，i+1 指向字段名
		if i+1 >= len(parts) {
			continue
		}

		field.Type = part
		field.Name = parts[i+1]
		field.Repeated = isRepeated
		field.Optional = isOptional

		// 查找字段号 (格式: "= 1" 或 "=1")
		for j := i + 2; j < len(parts)-1; j++ {
			if parts[j] == "=" {
				field.Number, _ = strconv.Atoi(strings.TrimSpace(parts[j+1]))
				break
			}
		}
		// 如果上面没找到，尝试查找包含 "=" 的部分
		if field.Number == 0 {
			for j := i + 2; j < len(parts); j++ {
				if strings.Contains(parts[j], "=") {
					fieldParts := strings.Split(parts[j], "=")
					if len(fieldParts) >= 2 {
						field.Number, _ = strconv.Atoi(strings.TrimSpace(fieldParts[1]))
					}
					break
				}
			}
		}

		break
	}

	return field
}

// generateCppHeader 生成C++头文件
func generateCppHeader(baseName string, messageTypes []MessageType, messages []Message, enums []Enum, imports []string, packageName string) string {
	var buffer bytes.Buffer

	// 生成头文件保护宏
	guardMacro := strings.ToUpper(baseName) + "_H"
	buffer.WriteString(fmt.Sprintf("#ifndef %s\n", guardMacro))
	buffer.WriteString(fmt.Sprintf("#define %s\n\n", guardMacro))

	// 包含标准头文件
	buffer.WriteString("#include <cstdint>\n")
	buffer.WriteString("#include <string>\n")
	buffer.WriteString("#include <vector>\n")
	buffer.WriteString("#include <memory>\n")
	buffer.WriteString("#include <stdexcept>\n\n")

	// 包含导入的头文件
	for _, imp := range imports {
		// 提取文件名
		impBase := filepath.Base(imp)
		impBase = strings.TrimSuffix(impBase, ".proto")
		buffer.WriteString(fmt.Sprintf("#include \"%s.h\"\n", impBase))
	}
	if len(imports) > 0 {
		buffer.WriteString("\n")
	}

	// 命名空间
	if packageName != "" {
		buffer.WriteString(fmt.Sprintf("namespace %s {\n\n", packageName))
	}

	// 生成枚举定义（排除 MessageType，因为它会单独生成）
	for _, enum := range enums {
		if enum.Name != "MessageType" {
			buffer.WriteString(generateCppEnum(enum))
			buffer.WriteString("\n")
		}
	}

	// 生成消息类型枚举（只生成一次）
	if len(messageTypes) > 0 {
		buffer.WriteString("// Message types\n")
		buffer.WriteString("enum class MessageType {\n")
		for i, mt := range messageTypes {
			buffer.WriteString(fmt.Sprintf("    %s = %d", mt.Name, mt.Value))
			if i < len(messageTypes)-1 {
				buffer.WriteString(",")
			}
			buffer.WriteString("\n")
		}
		buffer.WriteString("};\n\n")
	}

	// 生成消息结构体
	for _, msg := range messages {
		buffer.WriteString(generateCppMessage(msg, enums))
		buffer.WriteString("\n")
	}

	// 关闭命名空间
	if packageName != "" {
		buffer.WriteString(fmt.Sprintf("} // namespace %s\n\n", packageName))
	}

	// 结束头文件保护
	buffer.WriteString(fmt.Sprintf("#endif // %s\n", guardMacro))

	return buffer.String()
}

// generateCppEnum 生成C++枚举定义
func generateCppEnum(enum Enum) string {
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("// %s enum\n", enum.Name))
	buffer.WriteString(fmt.Sprintf("enum class %s {\n", enum.Name))

	for i, ev := range enum.Values {
		buffer.WriteString(fmt.Sprintf("    %s = %d", ev.Name, ev.Value))
		if i < len(enum.Values)-1 {
			buffer.WriteString(",")
		}
		buffer.WriteString("\n")
	}

	buffer.WriteString("};\n")

	return buffer.String()
}

// generateCppMessage 生成C++消息结构体
func generateCppMessage(msg Message, enums []Enum) string {
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("// %s message\n", msg.Name))
	buffer.WriteString(fmt.Sprintf("struct %s {\n", msg.Name))

	// 生成字段定义
	for _, field := range msg.Fields {
		isEnum := isEnumType(field.Type, enums)
		cppType := protoTypeToCppType(field.Type, isEnum)
		if field.Repeated {
			buffer.WriteString(fmt.Sprintf("    std::vector<%s> %s;\n", cppType, field.Name))
		} else {
			buffer.WriteString(fmt.Sprintf("    %s %s;\n", cppType, field.Name))
		}
	}

	buffer.WriteString("\n")

	// 生成默认构造函数
	buffer.WriteString(fmt.Sprintf("    // Default constructor\n"))
	buffer.WriteString(fmt.Sprintf("    %s() = default;\n\n", msg.Name))

	// 生成序列化方法
	buffer.WriteString(generateCppSerialize(msg, enums))
	buffer.WriteString("\n")

	// 生成反序列化方法
	buffer.WriteString(generateCppDeserialize(msg, enums))

	buffer.WriteString("};\n")

	return buffer.String()
}

// protoTypeToCppType 将proto类型转换为C++类型
func protoTypeToCppType(protoType string, isEnum bool) string {
	// 如果是枚举类型，直接返回枚举名
	if isEnum {
		return protoType
	}

	switch protoType {
	case "double":
		return "double"
	case "float":
		return "float"
	case "int32", "sint32", "sfixed32":
		return "int32_t"
	case "int64", "sint64", "sfixed64":
		return "int64_t"
	case "uint32", "fixed32":
		return "uint32_t"
	case "uint64", "fixed64":
		return "uint64_t"
	case "bool":
		return "bool"
	case "string":
		return "std::string"
	case "bytes":
		return "std::vector<uint8_t>"
	default:
		// 自定义类型（可能是消息或枚举）
		return protoType
	}
}

// isEnumType 检查类型是否是枚举
func isEnumType(protoType string, enums []Enum) bool {
	for _, enum := range enums {
		if enum.Name == protoType {
			return true
		}
	}
	return false
}

// generateCppSerialize 生成C++序列化方法
func generateCppSerialize(msg Message, enums []Enum) string {
	var buffer bytes.Buffer

	buffer.WriteString("    // Serialize to byte array\n")
	buffer.WriteString("    std::vector<uint8_t> Serialize() const {\n")
	buffer.WriteString("        std::vector<uint8_t> buffer;\n")
	buffer.WriteString("        size_t offset = 0;\n\n")

	for _, field := range msg.Fields {
		fieldNum := field.Number
		isEnum := isEnumType(field.Type, enums)
		wireType := getWireType(field.Type, isEnum)
		fieldKey := (fieldNum << 3) | wireType

		if field.Repeated {
			// 处理 repeated 字段
			buffer.WriteString(fmt.Sprintf("        // Serialize %s (repeated)\n", field.Name))
			buffer.WriteString(fmt.Sprintf("        for (const auto& item : %s) {\n", field.Name))
			buffer.WriteString(fmt.Sprintf("            // Field key: %d (field number %d, wire type %d)\n", fieldKey, fieldNum, wireType))
			buffer.WriteString(fmt.Sprintf("            buffer.push_back(%d);\n", fieldKey))
			buffer.WriteString(serializeCppField("item", field.Type, true, isEnum))
			buffer.WriteString("        }\n")
		} else {
			// 处理普通字段
			buffer.WriteString(fmt.Sprintf("        // Serialize %s\n", field.Name))
			buffer.WriteString(fmt.Sprintf("        buffer.push_back(%d); // Field key: %d (field number %d, wire type %d)\n", fieldKey, fieldKey, fieldNum, wireType))
			buffer.WriteString(serializeCppField(field.Name, field.Type, false, isEnum))
		}
		buffer.WriteString("\n")
	}

	buffer.WriteString("        return buffer;\n")
	buffer.WriteString("    }\n")

	return buffer.String()
}

// getWireType 获取字段的wire type
func getWireType(protoType string, isEnum bool) int {
	if isEnum {
		return 0 // 枚举使用 Varint
	}

	switch protoType {
	case "int32", "int64", "uint32", "uint64", "sint32", "sint64", "bool":
		return 0 // Varint
	case "fixed64", "sfixed64", "double":
		return 1 // 64-bit
	case "string", "bytes", "message":
		return 2 // Length-delimited
	case "fixed32", "sfixed32", "float":
		return 5 // 32-bit
	default:
		return 2 // 默认为Length-delimited（嵌套消息）
	}
}

// serializeCppField 生成C++字段序列化代码
func serializeCppField(fieldName, protoType string, isRepeated bool, isEnum bool) string {
	var buffer bytes.Buffer

	// 如果是枚举类型，使用 Varint 编码
	if isEnum {
		buffer.WriteString(fmt.Sprintf("            // Enum encoding for %s\n", fieldName))
		buffer.WriteString(fmt.Sprintf("            uint64_t value = static_cast<uint64_t>(%s);\n", fieldName))
		buffer.WriteString("            while (value > 127) {\n")
		buffer.WriteString("                buffer.push_back(static_cast<uint8_t>((value & 0x7F) | 0x80));\n")
		buffer.WriteString("                value >>= 7;\n")
		buffer.WriteString("            }\n")
		buffer.WriteString("            buffer.push_back(static_cast<uint8_t>(value));\n")
		return buffer.String()
	}

	switch protoType {
	case "int32", "int64", "uint32", "uint64", "sint32", "sint64", "bool":
		// Varint编码
		buffer.WriteString(fmt.Sprintf("            // Varint encoding for %s\n", protoType))
		buffer.WriteString(fmt.Sprintf("            uint64_t value = static_cast<uint64_t>(%s);\n", fieldName))
		buffer.WriteString("            while (value > 127) {\n")
		buffer.WriteString("                buffer.push_back(static_cast<uint8_t>((value & 0x7F) | 0x80));\n")
		buffer.WriteString("                value >>= 7;\n")
		buffer.WriteString("            }\n")
		buffer.WriteString("            buffer.push_back(static_cast<uint8_t>(value));\n")

	case "fixed32", "sfixed32", "float":
		// 32-bit little-endian
		buffer.WriteString(fmt.Sprintf("            // 32-bit little-endian encoding for %s\n", protoType))
		buffer.WriteString(fmt.Sprintf("            union { %s value; uint32_t bits; } converter;\n", protoType))
		buffer.WriteString(fmt.Sprintf("            converter.value = %s;\n", fieldName))
		buffer.WriteString("            buffer.push_back(static_cast<uint8_t>(converter.bits & 0xFF));\n")
		buffer.WriteString("            buffer.push_back(static_cast<uint8_t>((converter.bits >> 8) & 0xFF));\n")
		buffer.WriteString("            buffer.push_back(static_cast<uint8_t>((converter.bits >> 16) & 0xFF));\n")
		buffer.WriteString("            buffer.push_back(static_cast<uint8_t>((converter.bits >> 24) & 0xFF));\n")

	case "fixed64", "sfixed64", "double":
		// 64-bit little-endian
		buffer.WriteString(fmt.Sprintf("            // 64-bit little-endian encoding for %s\n", protoType))
		buffer.WriteString(fmt.Sprintf("            union { %s value; uint64_t bits; } converter;\n", protoType))
		buffer.WriteString(fmt.Sprintf("            converter.value = %s;\n", fieldName))
		buffer.WriteString("            for (int i = 0; i < 8; ++i) {\n")
		buffer.WriteString("                buffer.push_back(static_cast<uint8_t>((converter.bits >> (i * 8)) & 0xFF));\n")
		buffer.WriteString("            }\n")

	case "string":
		// 字符串编码
		buffer.WriteString("            // String encoding\n")
		buffer.WriteString(fmt.Sprintf("            uint64_t length = %s.size();\n", fieldName))
		buffer.WriteString("            while (length > 127) {\n")
		buffer.WriteString("                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));\n")
		buffer.WriteString("                length >>= 7;\n")
		buffer.WriteString("            }\n")
		buffer.WriteString("            buffer.push_back(static_cast<uint8_t>(length));\n")
		buffer.WriteString(fmt.Sprintf("            buffer.insert(buffer.end(), %s.begin(), %s.end());\n", fieldName, fieldName))

	case "bytes":
		// 字节数组编码
		buffer.WriteString("            // Bytes encoding\n")
		buffer.WriteString(fmt.Sprintf("            uint64_t length = %s.size();\n", fieldName))
		buffer.WriteString("            while (length > 127) {\n")
		buffer.WriteString("                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));\n")
		buffer.WriteString("                length >>= 7;\n")
		buffer.WriteString("            }\n")
		buffer.WriteString("            buffer.push_back(static_cast<uint8_t>(length));\n")
		buffer.WriteString(fmt.Sprintf("            buffer.insert(buffer.end(), %s.begin(), %s.end());\n", fieldName, fieldName))

	default:
		// 自定义消息类型
		buffer.WriteString(fmt.Sprintf("            // Nested message encoding for %s\n", protoType))
		buffer.WriteString(fmt.Sprintf("            std::vector<uint8_t> nestedData = %s.Serialize();\n", fieldName))
		buffer.WriteString("            uint64_t length = nestedData.size();\n")
		buffer.WriteString("            while (length > 127) {\n")
		buffer.WriteString("                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));\n")
		buffer.WriteString("                length >>= 7;\n")
		buffer.WriteString("            }\n")
		buffer.WriteString("            buffer.push_back(static_cast<uint8_t>(length));\n")
		buffer.WriteString("            buffer.insert(buffer.end(), nestedData.begin(), nestedData.end());\n")
	}

	return buffer.String()
}

// generateCppDeserialize 生成C++反序列化方法
func generateCppDeserialize(msg Message, enums []Enum) string {
	var buffer bytes.Buffer

	buffer.WriteString("    // Deserialize from byte array\n")
	buffer.WriteString("    static std::unique_ptr<" + msg.Name + "> Deserialize(const std::vector<uint8_t>& buffer) {\n")
	buffer.WriteString("        auto msg = std::make_unique<" + msg.Name + ">();\n")
	buffer.WriteString("        size_t offset = 0;\n\n")
	buffer.WriteString("        while (offset < buffer.size()) {\n")
	buffer.WriteString("            uint8_t key = buffer[offset++];\n")
	buffer.WriteString("            int fieldNumber = key >> 3;\n")
	buffer.WriteString("            int wireType = key & 0x07;\n\n")
	buffer.WriteString("            switch (fieldNumber) {\n")

	for _, field := range msg.Fields {
		isEnum := isEnumType(field.Type, enums)
		buffer.WriteString(fmt.Sprintf("                case %d: { // %s\n", field.Number, field.Name))
		buffer.WriteString(deserializeCppField(field.Name, field.Type, field.Repeated, isEnum, enums))
		buffer.WriteString("                    break;\n")
		buffer.WriteString("                }\n")
	}

	buffer.WriteString("                default:\n")
	buffer.WriteString("                    // Skip unknown fields\n")
	buffer.WriteString("                    if (wireType == 0) { // Varint\n")
	buffer.WriteString("                        while (offset < buffer.size() && (buffer[offset] & 0x80)) ++offset;\n")
	buffer.WriteString("                        if (offset < buffer.size()) ++offset;\n")
	buffer.WriteString("                    } else if (wireType == 2) { // Length-delimited\n")
	buffer.WriteString("                        uint64_t length = 0;\n")
	buffer.WriteString("                        int shift = 0;\n")
	buffer.WriteString("                        while (offset < buffer.size() && (buffer[offset] & 0x80)) {\n")
	buffer.WriteString("                            length |= (buffer[offset++] & 0x7F) << shift;\n")
	buffer.WriteString("                            shift += 7;\n")
	buffer.WriteString("                        }\n")
	buffer.WriteString("                        if (offset < buffer.size()) {\n")
	buffer.WriteString("                            length |= buffer[offset++] << shift;\n")
	buffer.WriteString("                        }\n")
	buffer.WriteString("                        offset += length;\n")
	buffer.WriteString("                    } else if (wireType == 5) { // 32-bit\n")
	buffer.WriteString("                        offset += 4;\n")
	buffer.WriteString("                    } else if (wireType == 1) { // 64-bit\n")
	buffer.WriteString("                        offset += 8;\n")
	buffer.WriteString("                    }\n")
	buffer.WriteString("                    break;\n")
	buffer.WriteString("            }\n")
	buffer.WriteString("        }\n\n")
	buffer.WriteString("        return msg;\n")
	buffer.WriteString("    }\n")

	return buffer.String()
}

// deserializeCppField 生成C++字段反序列化代码
func deserializeCppField(fieldName, protoType string, isRepeated bool, isEnum bool, enums []Enum) string {
	var buffer bytes.Buffer

	// 如果是枚举类型，使用 Varint 解码
	if isEnum {
		buffer.WriteString("                    // Enum decoding\n")
		buffer.WriteString("                    uint64_t value = 0;\n")
		buffer.WriteString("                    int shift = 0;\n")
		buffer.WriteString("                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {\n")
		buffer.WriteString("                        value |= (buffer[offset++] & 0x7F) << shift;\n")
		buffer.WriteString("                        shift += 7;\n")
		buffer.WriteString("                    }\n")
		buffer.WriteString("                    if (offset < buffer.size()) {\n")
		buffer.WriteString("                        value |= buffer[offset++] << shift;\n")
		buffer.WriteString("                    }\n")
		if isRepeated {
			buffer.WriteString(fmt.Sprintf("                    msg->%s.push_back(static_cast<%s>(value));\n", fieldName, protoTypeToCppType(protoType, true)))
		} else {
			buffer.WriteString(fmt.Sprintf("                    msg->%s = static_cast<%s>(value);\n", fieldName, protoTypeToCppType(protoType, true)))
		}
		return buffer.String()
	}

	switch protoType {
	case "int32", "int64", "uint32", "uint64", "sint32", "sint64", "bool":
		// Varint解码
		buffer.WriteString("                    // Varint decoding\n")
		buffer.WriteString("                    uint64_t value = 0;\n")
		buffer.WriteString("                    int shift = 0;\n")
		buffer.WriteString("                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {\n")
		buffer.WriteString("                        value |= (buffer[offset++] & 0x7F) << shift;\n")
		buffer.WriteString("                        shift += 7;\n")
		buffer.WriteString("                    }\n")
		buffer.WriteString("                    if (offset < buffer.size()) {\n")
		buffer.WriteString("                        value |= buffer[offset++] << shift;\n")
		buffer.WriteString("                    }\n")
		if isRepeated {
			buffer.WriteString(fmt.Sprintf("                    msg->%s.push_back(static_cast<%s>(value));\n", fieldName, protoTypeToCppType(protoType, false)))
		} else {
			buffer.WriteString(fmt.Sprintf("                    msg->%s = static_cast<%s>(value);\n", fieldName, protoTypeToCppType(protoType, false)))
		}

	case "fixed32", "sfixed32", "float":
		// 32-bit little-endian解码
		buffer.WriteString("                    // 32-bit little-endian decoding\n")
		buffer.WriteString("                    if (offset + 4 <= buffer.size()) {\n")
		buffer.WriteString("                        uint32_t bits = buffer[offset] | (buffer[offset+1] << 8) | (buffer[offset+2] << 16) | (buffer[offset+3] << 24);\n")
		buffer.WriteString("                        offset += 4;\n")
		buffer.WriteString(fmt.Sprintf("                        union { %s value; uint32_t bits; } converter;\n", protoType))
		buffer.WriteString("                        converter.bits = bits;\n")
		if isRepeated {
			buffer.WriteString(fmt.Sprintf("                        msg->%s.push_back(converter.value);\n", fieldName))
		} else {
			buffer.WriteString(fmt.Sprintf("                        msg->%s = converter.value;\n", fieldName))
		}
		buffer.WriteString("                    }\n")

	case "fixed64", "sfixed64", "double":
		// 64-bit little-endian解码
		buffer.WriteString("                    // 64-bit little-endian decoding\n")
		buffer.WriteString("                    if (offset + 8 <= buffer.size()) {\n")
		buffer.WriteString("                        uint64_t bits = 0;\n")
		buffer.WriteString("                        for (int i = 0; i < 8; ++i) {\n")
		buffer.WriteString("                            bits |= static_cast<uint64_t>(buffer[offset+i]) << (i * 8);\n")
		buffer.WriteString("                        }\n")
		buffer.WriteString("                        offset += 8;\n")
		buffer.WriteString(fmt.Sprintf("                        union { %s value; uint64_t bits; } converter;\n", protoType))
		buffer.WriteString("                        converter.bits = bits;\n")
		if isRepeated {
			buffer.WriteString(fmt.Sprintf("                        msg->%s.push_back(converter.value);\n", fieldName))
		} else {
			buffer.WriteString(fmt.Sprintf("                        msg->%s = converter.value;\n", fieldName))
		}
		buffer.WriteString("                    }\n")

	case "string":
		// 字符串解码
		buffer.WriteString("                    // String decoding\n")
		buffer.WriteString("                    uint64_t length = 0;\n")
		buffer.WriteString("                    int shift = 0;\n")
		buffer.WriteString("                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {\n")
		buffer.WriteString("                        length |= (buffer[offset++] & 0x7F) << shift;\n")
		buffer.WriteString("                        shift += 7;\n")
		buffer.WriteString("                    }\n")
		buffer.WriteString("                    if (offset < buffer.size()) {\n")
		buffer.WriteString("                        length |= buffer[offset++] << shift;\n")
		buffer.WriteString("                    }\n")
		buffer.WriteString("                    if (offset + length <= buffer.size()) {\n")
		if isRepeated {
			buffer.WriteString(fmt.Sprintf("                        msg->%s.emplace_back(buffer.begin() + offset, buffer.begin() + offset + length);\n", fieldName))
		} else {
			buffer.WriteString(fmt.Sprintf("                        msg->%s = std::string(buffer.begin() + offset, buffer.begin() + offset + length);\n", fieldName))
		}
		buffer.WriteString("                        offset += length;\n")
		buffer.WriteString("                    }\n")

	case "bytes":
		// 字节数组解码
		buffer.WriteString("                    // Bytes decoding\n")
		buffer.WriteString("                    uint64_t length = 0;\n")
		buffer.WriteString("                    int shift = 0;\n")
		buffer.WriteString("                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {\n")
		buffer.WriteString("                        length |= (buffer[offset++] & 0x7F) << shift;\n")
		buffer.WriteString("                        shift += 7;\n")
		buffer.WriteString("                    }\n")
		buffer.WriteString("                    if (offset < buffer.size()) {\n")
		buffer.WriteString("                        length |= buffer[offset++] << shift;\n")
		buffer.WriteString("                    }\n")
		buffer.WriteString("                    if (offset + length <= buffer.size()) {\n")
		if isRepeated {
			buffer.WriteString(fmt.Sprintf("                        msg->%s.emplace_back(buffer.begin() + offset, buffer.begin() + offset + length);\n", fieldName))
		} else {
			buffer.WriteString(fmt.Sprintf("                        msg->%s = std::vector<uint8_t>(buffer.begin() + offset, buffer.begin() + offset + length);\n", fieldName))
		}
		buffer.WriteString("                        offset += length;\n")
		buffer.WriteString("                    }\n")

	default:
		// 自定义消息类型
		buffer.WriteString(fmt.Sprintf("                    // Nested message decoding for %s\n", protoType))
		buffer.WriteString("                    uint64_t length = 0;\n")
		buffer.WriteString("                    int shift = 0;\n")
		buffer.WriteString("                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {\n")
		buffer.WriteString("                        length |= (buffer[offset++] & 0x7F) << shift;\n")
		buffer.WriteString("                        shift += 7;\n")
		buffer.WriteString("                    }\n")
		buffer.WriteString("                    if (offset < buffer.size()) {\n")
		buffer.WriteString("                        length |= buffer[offset++] << shift;\n")
		buffer.WriteString("                    }\n")
		buffer.WriteString("                    if (offset + length <= buffer.size()) {\n")
		buffer.WriteString("                        std::vector<uint8_t> nestedData(buffer.begin() + offset, buffer.begin() + offset + length);\n")
		if isRepeated {
			buffer.WriteString(fmt.Sprintf("                        msg->%s.push_back(*%s::Deserialize(nestedData));\n", fieldName, protoType))
		} else {
			buffer.WriteString(fmt.Sprintf("                        msg->%s = *%s::Deserialize(nestedData);\n", fieldName, protoType))
		}
		buffer.WriteString("                        offset += length;\n")
		buffer.WriteString("                    }\n")
	}

	return buffer.String()
}

// generateJSCode 生成CocosJS代码
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

// generateTSCode 生成TypeScript代码
func generateTSCode(protoFile, protoDir, outputDir string) error {
	// 获取文件名（不含扩展名）
	fileName := filepath.Base(protoFile)
	baseName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	tsOutputFile := filepath.Join(outputDir, baseName+".ts")

	// 解析proto文件
	messageTypes, messages, err := parseProtoFile(protoFile)
	if err != nil {
		return err
	}

	// 生成TypeScript代码
	tsContent := generateTypeScriptCode(baseName, messageTypes, messages)

	// 写入TS文件
	return os.WriteFile(tsOutputFile, []byte(tsContent), 0644)
}

// generateTypeScriptCode 生成TypeScript代码
func generateTypeScriptCode(baseName string, messageTypes []MessageType, messages []Message) string {
	// 生成消息类型代码
	messageTypeCode := ""
	for i, mt := range messageTypes {
		messageTypeCode += fmt.Sprintf("%s = %d", mt.Name, mt.Value)
		if i < len(messageTypes)-1 {
			messageTypeCode += ",\n\t	"
		}
	}

	// 生成消息类型定义代码
	messageTypeDefinitions := ""
	for _, msg := range messages {
		// 生成字段类型定义
		fields := ""
		for i, field := range msg.Fields {
			tsType := protoTypeToTsType(field.Type)
			fields += fmt.Sprintf("\t\t%s: %s", field.Name, tsType)
			if i < len(msg.Fields)-1 {
				fields += ",\n"
			}
		}

		// 生成接口定义
		messageTypeDefinitions += fmt.Sprintf(`
	// %s message
	export interface %s {
%s
	}`, msg.Name, msg.Name, fields)
	}

	return fmt.Sprintf(`// TypeScript code for %s
// Generated by proto generator
// Based on binary serialization/deserialization

// %s namespace
export namespace %s {
	// Message types
	export enum MessageType {
		%s
	}

	// Message wrapper
	export interface Message {
		type: MessageType;
		data: any;
	}

	// BinaryReader 用于从二进制流中读取数据
	export class BinaryReader {
		private buffer: Uint8Array;
		private offset: number;

		constructor(buffer: Uint8Array) {
			this.buffer = buffer;
			this.offset = 0;
		}

		// 读取 varint
		public readVarint(): number {
			let value = 0;
			let shift = 0;
			let byte;
			
			do {
				byte = this.buffer[this.offset++];
				value |= (byte & 0x7F) << shift;
				shift += 7;
			} while (byte & 0x80);
			
			return value;
		}

		// 读取 32-bit little-endian
		public readUInt32(): number {
			const value = (
				this.buffer[this.offset] |
				(this.buffer[this.offset + 1] << 8) |
				(this.buffer[this.offset + 2] << 16) |
				(this.buffer[this.offset + 3] << 24)
			);
			this.offset += 4;
			return value;
		}

		// 读取 64-bit little-endian
		public readUInt64(): bigint {
			let value = 0n;
			for (let i = 0; i < 8; i++) {
				value |= BigInt(this.buffer[this.offset + i]) << BigInt(i * 8);
			}
			this.offset += 8;
			return value;
		}

		// 读取 float
		public readFloat(): number {
			const bytes = new Uint8Array(4);
			for (let i = 0; i < 4; i++) {
				bytes[i] = this.buffer[this.offset + i];
			}
			this.offset += 4;
			return new Float32Array(bytes.buffer)[0];
		}

		// 读取 double
		public readDouble(): number {
			const bytes = new Uint8Array(8);
			for (let i = 0; i < 8; i++) {
				bytes[i] = this.buffer[this.offset + i];
			}
			this.offset += 8;
			return new Float64Array(bytes.buffer)[0];
		}

		// 读取 boolean
		public readBool(): boolean {
			return this.readVarint() !== 0;
		}

		// 读取 string
		public readString(): string {
			const length = this.readVarint();
			const bytes = this.buffer.subarray(this.offset, this.offset + length);
			this.offset += length;
			return new TextDecoder().decode(bytes);
		}

		// 读取 bytes
		public readBytes(): Uint8Array {
			const length = this.readVarint();
			const bytes = this.buffer.subarray(this.offset, this.offset + length);
			this.offset += length;
			return bytes;
		}

		// 检查是否已读取完毕
		public isEOF(): boolean {
			return this.offset >= this.buffer.length;
		}
	}

	// BinaryWriter 用于将数据写入二进制流
	export class BinaryWriter {
		private buffer: Uint8Array;
		private offset: number;

		constructor(capacity: number = 1024) {
			this.buffer = new Uint8Array(capacity);
			this.offset = 0;
		}

		// 确保缓冲区有足够空间
		private ensureCapacity(additional: number): void {
			if (this.offset + additional > this.buffer.length) {
				const newCapacity = Math.max(this.buffer.length * 2, this.offset + additional);
				const newBuffer = new Uint8Array(newCapacity);
				newBuffer.set(this.buffer);
				this.buffer = newBuffer;
			}
		}

		// 写入 varint
		public writeVarint(value: number): void {
			this.ensureCapacity(10); // 最大 10 字节
			
			do {
				let byte = value & 0x7F;
				value >>= 7;
				if (value > 0) {
					byte |= 0x80;
				}
				this.buffer[this.offset++] = byte;
			} while (value > 0);
		}

		// 写入 32-bit little-endian
		public writeUInt32(value: number): void {
			this.ensureCapacity(4);
			this.buffer[this.offset++] = value & 0xFF;
			this.buffer[this.offset++] = (value >> 8) & 0xFF;
			this.buffer[this.offset++] = (value >> 16) & 0xFF;
			this.buffer[this.offset++] = (value >> 24) & 0xFF;
		}

		// 写入 64-bit little-endian
		public writeUInt64(value: bigint): void {
			this.ensureCapacity(8);
			for (let i = 0; i < 8; i++) {
				this.buffer[this.offset++] = Number((value >> BigInt(i * 8)) & 0xFFn);
			}
		}

		// 写入 float
		public writeFloat(value: number): void {
			this.ensureCapacity(4);
			const bytes = new Uint8Array(new Float32Array([value]).buffer);
			for (let i = 0; i < 4; i++) {
				this.buffer[this.offset++] = bytes[i];
			}
		}

		// 写入 double
		public writeDouble(value: number): void {
			this.ensureCapacity(8);
			const bytes = new Uint8Array(new Float64Array([value]).buffer);
			for (let i = 0; i < 8; i++) {
				this.buffer[this.offset++] = bytes[i];
			}
		}

		// 写入 boolean
		public writeBool(value: boolean): void {
			this.writeVarint(value ? 1 : 0);
		}

		// 写入 string
		public writeString(value: string): void {
			const bytes = new TextEncoder().encode(value);
			this.writeVarint(bytes.length);
			this.ensureCapacity(bytes.length);
			this.buffer.set(bytes, this.offset);
			this.offset += bytes.length;
		}

		// 写入 bytes
		public writeBytes(value: Uint8Array): void {
			this.writeVarint(value.length);
			this.ensureCapacity(value.length);
			this.buffer.set(value, this.offset);
			this.offset += value.length;
		}

		// 获取最终的二进制数据
		public toArray(): Uint8Array {
			return this.buffer.subarray(0, this.offset);
		}
	}

	// Serialize message to binary
	export function serialize(message: any): Uint8Array {
		const writer = new BinaryWriter();
		// 这里需要根据具体消息类型实现序列化逻辑
		// 简化实现：将消息转换为 JSON 后再转换为字节
		const jsonStr = JSON.stringify(message);
		writer.writeString(jsonStr);
		return writer.toArray();
	}
	
	// Deserialize from binary
	export function deserialize(buffer: Uint8Array): any {
		const reader = new BinaryReader(buffer);
		// 这里需要根据具体消息类型实现反序列化逻辑
		// 简化实现：从字节中读取 JSON 字符串后解析
		const jsonStr = reader.readString();
		return JSON.parse(jsonStr);
	}

%s
}
`, baseName, baseName, baseName, messageTypeCode, messageTypeDefinitions)
}

// protoTypeToTsType 将proto类型转换为TypeScript类型
func protoTypeToTsType(protoType string) string {
	switch protoType {
	case "double", "float":
		return "number"
	case "int32", "int64", "uint32", "uint64", "sint32", "sint64", "fixed32", "fixed64", "sfixed32", "sfixed64":
		return "number"
	case "bool":
		return "boolean"
	case "string":
		return "string"
	case "bytes":
		return "Uint8Array"
	default:
		// 自定义类型
		return protoType
	}
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
