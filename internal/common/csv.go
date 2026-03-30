package common

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
)

// CSVReader CSV读取器
type CSVReader struct {
	filePath string
}

// NewCSVReader 创建CSV读取器实例
func NewCSVReader(filePath string) *CSVReader {
	return &CSVReader{
		filePath: filePath,
	}
}

// ReadCSV 读取CSV文件并返回所有行
func (r *CSVReader) ReadCSV() ([][]string, error) {
	// 检查文件是否存在
	if _, err := os.Stat(r.filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", r.filePath)
	}

	// 打开文件
	file, err := os.Open(r.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// 创建CSV读取器
	reader := csv.NewReader(file)

	// 读取所有行
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read csv: %v", err)
	}

	return rows, nil
}

// ReadCSVAsMap 读取CSV文件并将第一列作为key，整行数据作为value构建map
func (r *CSVReader) ReadCSVAsMap() (map[string][]string, error) {
	rows, err := r.ReadCSV()
	if err != nil {
		return nil, err
	}

	// 构建map
	result := make(map[string][]string)
	for i, row := range rows {
		if len(row) == 0 {
			continue
		}

		// 第一列作为key
		key := row[0]
		result[key] = row
	}

	return result, nil
}

// ReadCSVAsMapWithHeader 读取CSV文件，将第一行作为表头，后续行以第一列作为key构建map
func (r *CSVReader) ReadCSVAsMapWithHeader() (map[string]map[string]string, error) {
	rows, err := r.ReadCSV()
	if err != nil {
		return nil, err
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("csv file must have at least header and one data row")
	}

	// 第一行作为表头
	header := rows[0]

	// 构建map
	result := make(map[string]map[string]string)
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) == 0 {
			continue
		}

		// 第一列作为key
		key := row[0]

		// 构建行数据map
		rowMap := make(map[string]string)
		for j, value := range row {
			if j < len(header) {
				rowMap[header[j]] = value
			}
		}

		result[key] = rowMap
	}

	return result, nil
}

// ReadCSVFromDir 从目录中读取所有CSV文件并构建map
func ReadCSVFromDir(dirPath string) (map[string]map[string][]string, error) {
	// 检查目录是否存在
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("directory not found: %s", dirPath)
	}

	// 读取目录中的所有文件
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %v", err)
	}

	// 构建结果map
	result := make(map[string]map[string][]string)

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
		filePath := filepath.Join(dirPath, file.Name())

		// 读取CSV文件
		reader := NewCSVReader(filePath)
		csvMap, err := reader.ReadCSVAsMap()
		if err != nil {
			return nil, fmt.Errorf("failed to read csv file %s: %v", file.Name(), err)
		}

		// 将文件名（不含扩展名）作为key
		fileName := filepath.Base(file.Name())
		fileNameWithoutExt := fileName[:len(fileName)-len(filepath.Ext(fileName))]
		result[fileNameWithoutExt] = csvMap
	}

	return result, nil
}
