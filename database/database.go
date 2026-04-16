package database

import (
	"database/sql"
	"fmt"
	"log"

	"zgame/config"

	_ "github.com/lib/pq"
)

// DB 全局数据库连接
var DB *sql.DB

// InitDatabase 初始化数据库连接
func InitDatabase() error {
	// 从配置中获取数据库信息
	dbConfig := config.AppConfig.Database

	// 构建连接字符串
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.Dbname, dbConfig.Sslmode,
	)

	// 连接数据库
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("无法连接到数据库: %v", err)
	}

	// 测试连接
	err = db.Ping()
	if err != nil {
		return fmt.Errorf("数据库连接测试失败: %v", err)
	}

	// 存储全局连接
	DB = db

	log.Println("数据库连接成功")

	// 创建必要的表
	err = createTables()
	if err != nil {
		return fmt.Errorf("创建表失败: %v", err)
	}

	return nil
}

// createTables 创建必要的表结构
func createTables() error {
	// 创建用户表
	userTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(50) UNIQUE NOT NULL,
		password VARCHAR(100) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err := DB.Exec(userTableSQL)
	if err != nil {
		return fmt.Errorf("创建用户表失败: %v", err)
	}

	// 创建角色表
	actorTableSQL := `
	CREATE TABLE IF NOT EXISTS actors (
		id SERIAL PRIMARY KEY,
		actor_id VARCHAR(50) UNIQUE NOT NULL,
		user_id INTEGER NOT NULL REFERENCES users(id),
		name VARCHAR(50) NOT NULL,
		level INTEGER DEFAULT 1,
		realm VARCHAR(50) DEFAULT 'realm_1',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		online_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		offline_at TIMESTAMP
	);
	`

	_, err = DB.Exec(actorTableSQL)
	if err != nil {
		return fmt.Errorf("创建角色表失败: %v", err)
	}

	log.Println("表结构创建成功")
	return nil
}

// CloseDatabase 关闭数据库连接
func CloseDatabase() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
