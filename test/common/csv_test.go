package common

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aoyo/qp/internal/common"
	"github.com/stretchr/testify/assert"
)

func TestCSVReader(t *testing.T) {
	// 创建测试CSV文件
	testFile := createTestCSVFile(t)
	defer os.Remove(testFile)

	// 创建CSV读取器
	reader := common.NewCSVReader(testFile)

	// 测试ReadCSV方法
	testReadCSV(t, reader)

	// 测试ReadCSVAsMap方法
	testReadCSVAsMap(t, reader)

	// 测试ReadCSVAsMapWithHeader方法
	testReadCSVAsMapWithHeader(t, reader)
}

func createTestCSVFile(t *testing.T) string {
	// 创建临时文件
	file, err := os.CreateTemp("", "test_*.csv")
	assert.NoError(t, err)

	// 写入测试数据
	data := `id,name,age
1,Alice,25
2,Bob,30
3,Charlie,35`
	_, err = file.WriteString(data)
	assert.NoError(t, err)

	// 关闭文件
	err = file.Close()
	assert.NoError(t, err)

	return file.Name()
}

func testReadCSV(t *testing.T, reader *common.CSVReader) {
	// 读取CSV文件
	rows, err := reader.ReadCSV()
	assert.NoError(t, err)

	// 验证行数
	assert.Len(t, rows, 4)

	// 验证表头
	assert.Equal(t, []string{"id", "name", "age"}, rows[0])

	// 验证数据行
	assert.Equal(t, []string{"1", "Alice", "25"}, rows[1])
	assert.Equal(t, []string{"2", "Bob", "30"}, rows[2])
	assert.Equal(t, []string{"3", "Charlie", "35"}, rows[3])
}

func testReadCSVAsMap(t *testing.T, reader *common.CSVReader) {
	// 读取CSV文件并构建map
	csvMap, err := reader.ReadCSVAsMap()
	assert.NoError(t, err)

	// 验证map大小
	assert.Len(t, csvMap, 4)

	// 验证map内容
	assert.Equal(t, []string{"id", "name", "age"}, csvMap["id"])
	assert.Equal(t, []string{"1", "Alice", "25"}, csvMap["1"])
	assert.Equal(t, []string{"2", "Bob", "30"}, csvMap["2"])
	assert.Equal(t, []string{"3", "Charlie", "35"}, csvMap["3"])
}

func testReadCSVAsMapWithHeader(t *testing.T, reader *common.CSVReader) {
	// 读取CSV文件并构建map
	csvMap, err := reader.ReadCSVAsMapWithHeader()
	assert.NoError(t, err)

	// 验证map大小
	assert.Len(t, csvMap, 3)

	// 验证map内容
	assert.Equal(t, map[string]string{"id": "1", "name": "Alice", "age": "25"}, csvMap["1"])
	assert.Equal(t, map[string]string{"id": "2", "name": "Bob", "age": "30"}, csvMap["2"])
	assert.Equal(t, map[string]string{"id": "3", "name": "Charlie", "age": "35"}, csvMap["3"])
}

func TestReadCSVFromDir(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "test_csv_dir")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 创建测试CSV文件
	createTestCSVFileInDir(t, tempDir, "test1.csv")
	createTestCSVFileInDir(t, tempDir, "test2.csv")

	// 读取目录中的CSV文件
	result, err := common.ReadCSVFromDir(tempDir)
	assert.NoError(t, err)

	// 验证结果
	assert.Len(t, result, 2)
	assert.Contains(t, result, "test1")
	assert.Contains(t, result, "test2")
}

func createTestCSVFileInDir(t *testing.T, dirPath, fileName string) {
	// 创建文件路径
	filePath := filepath.Join(dirPath, fileName)

	// 写入测试数据
	data := `id,name
1,Alice
2,Bob`
	err := os.WriteFile(filePath, []byte(data), 0644)
	assert.NoError(t, err)
}
