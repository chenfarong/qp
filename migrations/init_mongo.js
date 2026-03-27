// MongoDB初始化脚本

// 连接到admin数据库
use admin;

// 创建管理员用户
db.createUser({
  user: "admin",
  pwd: "password",
  roles: ["userAdminAnyDatabase", "dbAdminAnyDatabase", "readWriteAnyDatabase"]
});

// 连接到qp_game数据库
use qp_game;

// 创建用户集合
db.createCollection("users");
// 创建角色集合
db.createCollection("characters");

// 创建索引
db.users.createIndex({ "username": 1 }, { unique: true });
db.users.createIndex({ "email": 1 }, { unique: true });
db.characters.createIndex({ "user_id": 1 });
db.characters.createIndex({ "name": 1 }, { unique: true });

print("MongoDB initialization completed successfully!");
