package battle

// BattleRequest 战斗请求
type BattleRequest struct {
	CharacterID string `json:"character_id" binding:"required"`
	EnemyLevel  int    `json:"enemy_level" binding:"required,min=1"`
}

// BattleResponse 战斗响应
type BattleResponse struct {
	Victory    bool   `json:"victory"`
	ExpGained  int    `json:"exp_gained"`
	GoldGained int    `json:"gold_gained"`
	Message    string `json:"message"`
}
