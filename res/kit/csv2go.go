package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// 主函数
func main() {
	// 定义目录路径
	csvDir := "res/csv"
	outputDir := "res/go"

	// 创建输出目录
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Failed to create output directory: %v\n", err)
		return
	}

	// 读取CSV目录下的所有文件
	files, err := os.ReadDir(csvDir)
	if err != nil {
		fmt.Printf("Failed to read CSV directory: %v\n", err)
		return
	}

	// 处理每个CSV文件
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// 检查文件扩展名
		if filepath.Ext(file.Name()) != ".csv" {
			continue
		}

		// 构建文件路径
		filePath := filepath.Join(csvDir, file.Name())

		// 解析CSV文件并生成Go结构定义
		if err := generateGoStruct(filePath, outputDir); err != nil {
			fmt.Printf("Failed to generate Go struct for %s: %v\n", file.Name(), err)
			continue
		}

		fmt.Printf("Generated Go struct for %s\n", file.Name())
	}

	fmt.Println("All CSV files processed successfully!")
}

// generateGoStruct 根据CSV文件生成Go结构定义
func generateGoStruct(csvPath, outputDir string) error {
	// 打开CSV文件
	file, err := os.Open(csvPath)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %v", err)
	}
	defer file.Close()

	// 创建CSV读取器
	reader := csv.NewReader(file)

	// 读取表头
	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read CSV header: %v", err)
	}

	// 读取第一行数据，用于推断类型
	dataRow, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read CSV data row: %v", err)
	}

	// 生成结构体名称
	fileName := filepath.Base(csvPath)
	structName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	structName = toPascalCase(structName)

	// 生成Go结构定义
	structDef := generateStructDefinition(structName, header, dataRow)

	// 生成文件名
	outputFile := filepath.Join(outputDir, strings.ToLower(structName)+"_gen.go")

	// 写入文件
	if err := os.WriteFile(outputFile, []byte(structDef), 0644); err != nil {
		return fmt.Errorf("failed to write output file: %v", err)
	}

	return nil
}

// generateStructDefinition 生成结构体定义
func generateStructDefinition(structName string, header []string, dataRow []string) string {
	var sb strings.Builder

	// 写入包声明
	sb.WriteString("package gogores\n\n")

	// 写入结构体定义
	sb.WriteString(fmt.Sprintf("// %s 自动生成的结构体定义\n", structName))
	sb.WriteString(fmt.Sprintf("type %s struct {\n", structName))

	// 写入字段定义
	for i, fieldName := range header {
		// 转换字段名为PascalCase
		pascalFieldName := toPascalCase(fieldName)

		// 推断字段类型
		fieldType := "string"
		if i < len(dataRow) {
			fieldType = inferType(dataRow[i])
		}

		// 写入字段定义
		sb.WriteString(fmt.Sprintf("\t%s %s `json:\"%s\"`\n", pascalFieldName, fieldType, fieldName))
	}

	sb.WriteString("}\n")

	return sb.String()
}

// toPascalCase 将字符串转换为PascalCase
func toPascalCase(s string) string {
	words := strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	var sb strings.Builder
	for _, word := range words {
		if len(word) > 0 {
			sb.WriteString(strings.ToUpper(word[:1]))
			sb.WriteString(strings.ToLower(word[1:]))
		}
	}

	result := sb.String()
	if result == "" {
		result = "Unknown"
	}

	return result
}

// inferType 根据值推断Go类型
func inferType(value string) string {
	// 尝试解析为整数
	if isInteger(value) {
		return "int"
	}

	// 尝试解析为浮点数
	if isFloat(value) {
		return "float64"
	}

	// 尝试解析为布尔值
	if isBoolean(value) {
		return "bool"
	}

	// 默认返回字符串类型
	return "string"
}

// isInteger 判断字符串是否为整数
func isInteger(s string) bool {
	for i, r := range s {
		if i == 0 && r == '-' {
			continue
		}
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return len(s) > 0
}

// isFloat 判断字符串是否为浮点数
func isFloat(s string) bool {
	hasDot := false
	for i, r := range s {
		if i == 0 && r == '-' {
			continue
		}
		if r == '.' {
			if hasDot {
				return false
			}
			hasDot = true
			continue
		}
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return len(s) > 0 && (hasDot || isInteger(s))
}

// isBoolean 判断字符串是否为布尔值
func isBoolean(s string) bool {
	lower := strings.ToLower(s)
	return lower == "true" || lower == "false"
}
