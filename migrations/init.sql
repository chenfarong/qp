-- 创建数据库
CREATE DATABASE qp_game;

-- 连接到数据库
\c qp_game;

-- 创建用户表
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    username VARCHAR(50) NOT NULL UNIQUE,
    password VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    nickname VARCHAR(50),
    avatar VARCHAR(255),
    status INTEGER DEFAULT 1
);

-- 创建角色表
CREATE TABLE IF NOT EXISTS characters (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    user_id INTEGER NOT NULL REFERENCES users(id),
    name VARCHAR(50) NOT NULL,
    level INTEGER DEFAULT 1,
    exp INTEGER DEFAULT 0,
    hp INTEGER DEFAULT 100,
    mp INTEGER DEFAULT 50,
    strength INTEGER DEFAULT 10,
    agility INTEGER DEFAULT 10,
    intelligence INTEGER DEFAULT 10,
    gold INTEGER DEFAULT 0,
    status INTEGER DEFAULT 1
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_characters_user_id ON characters(user_id);
