package bag

import "github.com/aoyo/qp/internal/gamelogic/actor"

// Item 背包物品，与 actor.Character.Items 同构
type Item = actor.InventoryItem

// InventoryPayload API 中的背包片段（来自角色文档的 items 字段）
type InventoryPayload struct {
	CharacterID string                `json:"character_id"`
	Items       []actor.InventoryItem `json:"items"`
}
