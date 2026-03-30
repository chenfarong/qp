package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/xuri/excelize/v2"
)

func main() {
	// 定义目录路径
	xlsxDir := "../../xlsx"
	csvDir := "../../res/csv"

	// 确保csv目录存在
	if err := os.MkdirAll(csvDir, 0755); err != nil {
		log.Fatalf("Failed to create csv directory: %v", err)
	}

	// 遍历xlsx目录下的所有xlsx文件
	err := filepath.Walk(xlsxDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 只处理xlsx文件
		if !info.IsDir() && filepath.Ext(path) == ".xlsx" {
			// 提取文件名（不含扩展名）
			fileName := filepath.Base(path)
			nameWithoutExt := fileName[:len(fileName)-len(filepath.Ext(fileName))]

			// 处理xlsx文件
			if err := convertXLSXToCSV(path, csvDir, nameWithoutExt); err != nil {
				log.Printf("Failed to convert %s: %v", path, err)
			}
		}

		return nil
	})

	if err != nil {
		log.Fatalf("Error walking xlsx directory: %v", err)
	}

	fmt.Println("Conversion completed!")
}

// convertXLSXToCSV 将xlsx文件转换为csv文件
func convertXLSXToCSV(xlsxPath, csvDir, baseName string) error {
	// 打开xlsx文件
	f, err := excelize.OpenFile(xlsxPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// 获取所有工作表
	sheetNames := f.GetSheetList()

	// 为每个工作表生成一个csv文件
	for _, sheetName := range sheetNames {
		// 生成csv文件名：xlsx文件名_sheet名字.csv
		csvPath := filepath.Join(csvDir, baseName+"_"+sheetName+".csv")

		// 获取工作表的所有行
		rows, err := f.GetRows(sheetName)
		if err != nil {
			log.Printf("Failed to get rows for sheet %s: %v", sheetName, err)
			continue
		}

		// 创建csv文件
		csvFile, err := os.Create(csvPath)
		if err != nil {
			log.Printf("Failed to create csv file %s: %v", csvPath, err)
			continue
		}

		// 写入csv文件
		for _, row := range rows {
			for i, cell := range row {
				if i > 0 {
					if _, err := csvFile.WriteString(","); err != nil {
						csvFile.Close()
						log.Printf("Failed to write to csv file %s: %v", csvPath, err)
						continue
					}
				}
				if _, err := csvFile.WriteString(cell); err != nil {
					csvFile.Close()
					log.Printf("Failed to write to csv file %s: %v", csvPath, err)
					continue
				}
			}
			if _, err := csvFile.WriteString("\n"); err != nil {
				csvFile.Close()
				log.Printf("Failed to write to csv file %s: %v", csvPath, err)
				continue
			}
		}

		csvFile.Close()
		fmt.Printf("Converted %s (sheet: %s) to %s\n", xlsxPath, sheetName, csvPath)

		// 生成对应的Go数据结构
		if err := generateGoStruct(csvPath, baseName, sheetName); err != nil {
			log.Printf("Failed to generate Go struct for %s: %v", csvPath, err)
		}
	}

	return nil
}

// generateGoStruct 根据csv文件生成Go数据结构
func generateGoStruct(csvPath, baseName, sheetName string) error {
	// 打开csv文件
	csvFile, err := os.Open(csvPath)
	if err != nil {
		return err
	}
	defer csvFile.Close()

	// 读取第一行作为字段名
	var headers []string
	buffer := make([]byte, 1024)
	n, err := csvFile.Read(buffer)
	if err != nil {
		return err
	}

	// 解析第一行，提取字段名
	content := string(buffer[:n])
	lines := strings.Split(content, "\n")
	if len(lines) > 0 {
		headers = strings.Split(lines[0], ",")
	}

	// 如果没有字段名，跳过
	if len(headers) == 0 {
		return nil
	}

	// 生成Go结构体名称
	structName := toPascalCase(baseName) + toPascalCase(sheetName)

	// 生成Go结构体代码
	goCode := generateStructCode(structName, headers)

	// 生成输出文件路径
	goOutputDir := "../../res/go"
	os.MkdirAll(goOutputDir, 0755)
	goOutputPath := filepath.Join(goOutputDir, baseName+"_"+sheetName+"_gen.go")

	// 写入Go文件
	if err := os.WriteFile(goOutputPath, []byte(goCode), 0644); err != nil {
		return err
	}

	fmt.Printf("Generated Go struct for %s to %s\n", csvPath, goOutputPath)
	return nil
}

// generateStructCode 生成Go结构体代码
func generateStructCode(structName string, fields []string) string {
	var sb strings.Builder

	// 写入包声明
	sb.WriteString("package gogores\n\n")

	// 写入结构体定义
	sb.WriteString(fmt.Sprintf("// %s 数据结构\n", structName))
	sb.WriteString(fmt.Sprintf("type %s struct {\n", structName))

	// 写入字段
	for _, field := range fields {
		if field == "" {
			continue
		}
		// 转换字段名为Go风格
		fieldName := toPascalCase(field)
		// 默认为string类型
		sb.WriteString(fmt.Sprintf("\t%s string `json:\"%s\"`\n", fieldName, field))
	}

	sb.WriteString("}\n")
	return sb.String()
}

// toPascalCase 将字符串转换为PascalCase
func toPascalCase(s string) string {
	// 移除特殊字符，转换为PascalCase
	var sb strings.Builder
	words := strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})

	for _, word := range words {
		if len(word) > 0 {
			// 首字母大写，其余小写
			sb.WriteRune(unicode.ToUpper(rune(word[0])))
			if len(word) > 1 {
				sb.WriteString(strings.ToLower(word[1:]))
			}
		}
	}

	// 如果结果为空，返回默认名称
	if sb.Len() == 0 {
		return "Default"
	}

	return sb.String()
}
