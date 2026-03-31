package common

// 游戏状态常量
const (
	// 角色状态
	CharacterStatusActive   = 1 // 活跃
	CharacterStatusInactive = 0 // 非活跃

	// 战斗结果
	BattleResultVictory = "victory" // 胜利
	BattleResultDefeat  = "defeat"  // 失败

	// 物品类型
	ItemTypeWeapon     = "weapon"     // 武器
	ItemTypeArmor      = "armor"      // 防具
	ItemTypePotion     = "potion"     // 药水
	ItemTypeAccessory  = "accessory"  // 饰品
	ItemTypeConsumable = "consumable" // 消耗品
	ItemTypeMaterial   = "material"   // 材料

	// 会话状态
	SessionStatusActive   = 1 // 活跃
	SessionStatusInactive = 0 // 非活跃
)

// 错误码常量
const (
	// 通用错误
	ErrCodeSuccess        = 0    // 成功
	ErrCodeInternalError  = 5000 // 内部错误
	ErrCodeInvalidRequest = 4000 // 请求无效
	ErrCodeUnauthorized   = 4010 // 未授权
	ErrCodeForbidden      = 4030 // 禁止访问
	ErrCodeNotFound       = 4040 // 资源不存在

	// 用户相关错误
	ErrCodeUsernameExists     = 1001 // 用户名已存在
	ErrCodeEmailExists        = 1002 // 邮箱已存在
	ErrCodeInvalidCredentials = 1003 // 无效的凭证
	ErrCodeUserDisabled       = 1004 // 用户已禁用

	// 角色相关错误
	ErrCodeCharacterLimit      = 2001 // 角色数量达到上限
	ErrCodeCharacterNotFound   = 2002 // 角色不存在
	ErrCodeCharacterNameExists = 2003 // 角色名已存在

	// 背包相关错误
	ErrCodeInventoryFull    = 3001 // 背包已满
	ErrCodeItemNotFound     = 3002 // 物品不存在
	ErrCodeInsufficientItem = 3003 // 物品数量不足

	// 战斗相关错误
	ErrCodeBattleFailed  = 4001 // 战斗失败
	ErrCodeCharacterBusy = 4002 // 角色繁忙
)

// 配置相关常量
const (
	// 服务器配置
	DefaultGatewayPort   = 8080 // 默认网关端口
	DefaultSsoAuthPort   = 9090 // 默认SSO认证端口
	DefaultGameLogicPort = 9000 // 默认游戏逻辑端口
	DefaultBillPort      = 9100 // 默认账单服务端口

	// 数据库配置
	DefaultMongoDBHost     = "localhost" // 默认MongoDB主机
	DefaultMongoDBPort     = 27017       // 默认MongoDB端口
	DefaultMongoDBUser     = "admin"     // 默认MongoDB用户名
	DefaultMongoDBPassword = "password"  // 默认MongoDB密码
	DefaultMongoDBDBName   = "qp_game"   // 默认MongoDB数据库名

	// JWT配置
	DefaultJWTExpireHours          = 24  // 默认JWT过期时间（小时）
	DefaultRefreshTokenExpireHours = 168 // 默认刷新令牌过期时间（小时）

	// 会话配置
	DefaultSessionExpireHours = 72 // 默认会话过期时间（小时）

	// WebSocket配置
	DefaultWebSocketReadBufferSize  = 1024 // 默认WebSocket读取缓冲区大小
	DefaultWebSocketWriteBufferSize = 1024 // 默认WebSocket写入缓冲区大小

	// 游戏配置
	MaxCharactersPerUser         = 3   // 每个用户最大角色数量
	DefaultCharacterLevel        = 1   // 默认角色等级
	DefaultCharacterHP           = 100 // 默认角色生命值
	DefaultCharacterMP           = 50  // 默认角色魔法值
	DefaultCharacterStrength     = 10  // 默认角色力量
	DefaultCharacterAgility      = 10  // 默认角色敏捷
	DefaultCharacterIntelligence = 10  // 默认角色智力
	DefaultCharacterGold         = 0   // 默认角色金币

	// 战斗配置
	BaseEnemyPower        = 50 // 基础敌人力量
	VictoryExpMultiplier  = 20 // 胜利经验倍数
	DefeatExpMultiplier   = 5  // 失败经验倍数
	VictoryGoldMultiplier = 5  // 胜利金币倍数

	// 升级配置
	BaseExpToLevelUp            = 100 // 基础升级经验
	LevelUpHPIncrease           = 10  // 升级生命值增加
	LevelUpMPIncrease           = 5   // 升级魔法值增加
	LevelUpStrengthIncrease     = 2   // 升级力量增加
	LevelUpAgilityIncrease      = 1   // 升级敏捷增加
	LevelUpIntelligenceIncrease = 1   // 升级智力增加

	// 背包配置
	DefaultInventorySize = 50 // 默认背包大小
)

// 消息类型常量
const (
	// 认证消息类型
	MsgTypeAuthRegister = "auth_register" // 注册
	MsgTypeAuthLogin    = "auth_login"    // 登录
	MsgTypeAuthValidate = "auth_validate" // 验证
	MsgTypeAuthRefresh  = "auth_refresh"  // 刷新令牌
	MsgTypeAuthLogout   = "auth_logout"   // 注销

	// 角色消息类型
	MsgTypeCharacterCreate = "character_create" // 创建角色
	MsgTypeCharacterList   = "character_list"   // 获取角色列表
	MsgTypeCharacterDetail = "character_detail" // 获取角色详情
	MsgTypeCharacterUpdate = "character_update" // 更新角色
	MsgTypeCharacterDelete = "character_delete" // 删除角色

	// 背包消息类型
	MsgTypeInventoryGet    = "inventory_get"    // 获取背包
	MsgTypeInventoryAdd    = "inventory_add"    // 添加物品
	MsgTypeInventoryUse    = "inventory_use"    // 使用物品
	MsgTypeInventoryRemove = "inventory_remove" // 删除物品
	MsgTypeInventoryEquip  = "inventory_equip"  // 装备物品

	// 战斗消息类型
	MsgTypeBattleStart  = "battle_start"  // 开始战斗
	MsgTypeBattleResult = "battle_result" // 战斗结果

	// 代币消息类型
	MsgTypeTokenAdd      = "token_add"      // 添加代币
	MsgTypeTokenRemove   = "token_remove"   // 减少代币
	MsgTypeTokenTransfer = "token_transfer" // 转移代币

	// 支付消息类型
	MsgTypePaymentCreate   = "payment_create"   // 创建支付
	MsgTypePaymentCallback = "payment_callback" // 支付回调
	MsgTypePaymentQuery    = "payment_query"    // 查询支付
)

// 响应状态常量
const (
	StatusSuccess = "success" // 成功
	StatusError   = "error"   // 错误
)

// 代币类型常量
const (
	// 主要代币
	TokenTypeGold       = "gold"       // 金币
	TokenTypeDiamond    = "diamond"    // 钻石
	TokenTypeExp        = "exp"        // 经验值
	TokenTypeHonor      = "honor"      // 荣誉值
	TokenTypeReputation = "reputation" // 声望值

	// 特殊代币
	TokenTypeEnergy     = "energy"      // 体力
	TokenTypeStamina    = "stamina"     // 耐力
	TokenTypeSkillPoint = "skill_point" // 技能点
	TokenTypeLuck       = "luck"        // 幸运值
)

// 代币操作类型常量
const (
	TokenOperationAdd      = "add"      // 添加
	TokenOperationRemove   = "remove"   // 减少
	TokenOperationLock     = "lock"     // 锁定
	TokenOperationUnlock   = "unlock"   // 解锁
	TokenOperationTransfer = "transfer" // 转移
)

// 道具品质常量
const (
	ItemQualityCommon    = "common"    // 普通
	ItemQualityUncommon  = "uncommon"  // 优秀
	ItemQualityRare      = "rare"      // 稀有
	ItemQualityEpic      = "epic"      // 史诗
	ItemQualityLegendary = "legendary" // 传说
)

// 道具等级常量
const (
	ItemLevel1  = 1  // 1级
	ItemLevel2  = 2  // 2级
	ItemLevel3  = 3  // 3级
	ItemLevel4  = 4  // 4级
	ItemLevel5  = 5  // 5级
	ItemLevel6  = 6  // 6级
	ItemLevel7  = 7  // 7级
	ItemLevel8  = 8  // 8级
	ItemLevel9  = 9  // 9级
	ItemLevel10 = 10 // 10级
)

// 装备部位常量
const (
	EquipSlotHead     = "head"     // 头部
	EquipSlotShoulder = "shoulder" // 肩部
	EquipSlotChest    = "chest"    // 胸部
	EquipSlotWaist    = "waist"    // 腰部
	EquipSlotLegs     = "legs"     // 腿部
	EquipSlotFeet     = "feet"     // 脚部
	EquipSlotHands    = "hands"    // 手部
	EquipSlotWeapon   = "weapon"   // 武器
	EquipSlotShield   = "shield"   // 盾牌
	EquipSlotNeck     = "neck"     // 项链
	EquipSlotRing     = "ring"     // 戒指
	EquipSlotTrinket  = "trinket"  // 饰品
)

// 药水效果常量
const (
	PotionEffectHP       = "hp"       // 生命值
	PotionEffectMP       = "mp"       // 魔法值
	PotionEffectStamina  = "stamina"  // 耐力
	PotionEffectSpeed    = "speed"    // 速度
	PotionEffectStrength = "strength" // 力量
	PotionEffectDefense  = "defense"  // 防御
)

// 材料类型常量
const (
	MaterialTypeOre     = "ore"     // 矿石
	MaterialTypeHerb    = "herb"    // 草药
	MaterialTypeLeather = "leather" // 皮革
	MaterialTypeCloth   = "cloth"   // 布料
	MaterialTypeMetal   = "metal"   // 金属
	MaterialTypeWood    = "wood"    // 木材
	MaterialTypeCrystal = "crystal" // 水晶
	MaterialTypeGem     = "gem"     // 宝石
)

// 消耗品类型常量
const (
	ConsumableTypePotion  = "potion"   // 药水
	ConsumableTypeFood    = "food"     // 食物
	ConsumableTypeScroll  = "scroll"   // 卷轴
	ConsumableTypeTrap    = "trap"     // 陷阱
	ConsumableTypeKey     = "key"      // 钥匙
	ConsumableTypeTicket  = "ticket"   // 票券
	ConsumableTypeGiftBox = "gift_box" // 礼盒
)
