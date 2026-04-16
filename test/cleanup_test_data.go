package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	// 数据库连接参数
	connStr := "host=localhost port=5432 user=postgres password=postgres dbname=zgame sslmode=disable"

	// 连接数据库
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("无法连接到数据库: %v", err)
	}
	defer db.Close()

	// 测试连接
	err = db.Ping()
	if err != nil {
		log.Fatalf("数据库连接测试失败: %v", err)
	}

	fmt.Println("连接到数据库成功")

	// 删除测试角色
	result1, err := db.Exec("DELETE FROM actors WHERE name LIKE 'test%' OR name LIKE 'new%'")
	if err != nil {
		log.Fatalf("删除测试角色失败: %v", err)
	}
	rowsAffected1, _ := result1.RowsAffected()
	fmt.Printf("删除了 %d 个测试角色\n", rowsAffected1)

	// 删除测试用户
	result2, err := db.Exec("DELETE FROM users WHERE username LIKE 'test%' OR username LIKE 'new%'")
	if err != nil {
		log.Fatalf("删除测试用户失败: %v", err)
	}
	rowsAffected2, _ := result2.RowsAffected()
	fmt.Printf("删除了 %d 个测试用户\n", rowsAffected2)

	fmt.Println("测试数据清理完成")
}
