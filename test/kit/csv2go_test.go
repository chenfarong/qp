package kit

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// runCSV2Go 运行csv2go工具
func runCSV2Go() error {
	cmd := exec.Command("go", "run", "res/kit/csv2go.go")
	return cmd.Run()
}

func TestCSV2Go(t *testing.T) {
	// 创建测试CSV文件
	testDir := "res/kit/csv"
	testFile := filepath.Join(testDir, "test.csv")

	// 写入测试数据
	data := `id,name,age,is_active,score
1,test,25,true,95.5`
	err := os.WriteFile(testFile, []byte(data), 0644)
	assert.NoError(t, err)
	defer os.Remove(testFile)

	// 运行csv2go工具
	err = runCSV2Go()
	assert.NoError(t, err)

	// 检查生成的文件
	generatedFile := "res/kit/generated/test_gen.go"
	assert.FileExists(t, generatedFile)
	defer os.Remove(generatedFile)

	// 读取生成的文件内容
	content, err := os.ReadFile(generatedFile)
	assert.NoError(t, err)

	// 验证生成的内容
	expectedContent := `package generated

// Test 自动生成的结构体定义
type Test struct {
	Id int ` + "`json:\"id\"`" + `
	Name string ` + "`json:\"name\"`" + `
	Age int ` + "`json:\"age\"`" + `
	IsActive bool ` + "`json:\"is_active\"`" + `
	Score float64 ` + "`json:\"score\"`" + `
}`
	assert.Contains(t, string(content), expectedContent)
}
