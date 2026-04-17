package proto

// 消息号定义
const (
	// 基础消息
	MSG_LoginRequest         = 1001
	MSG_LoginResponse        = 1002
	MSG_GetRoleInfoRequest   = 1003
	MSG_GetRoleInfoResponse  = 1004
	MSG_GetActorListRequest  = 1005
	MSG_GetActorListResponse = 1006

	// 角色消息
	MSG_ActorCreateRequest      = 2001
	MSG_ActorUseRequest         = 2002
	MSG_ActorUseWithNameRequest = 2003
	MSG_ActorUseResponse        = 2004

	// 背包消息
	MSG_GetBagRequest      = 3001
	MSG_GetBagResponse     = 3002
	MSG_BagItemUseRequest  = 3003
	MSG_BagItemUseResponse = 3004
	MSG_SyncBagItemChange  = 3005

	// 装备消息
	MSG_GetEquipRequest      = 4001
	MSG_GetEquipResponse     = 4002
	MSG_UpgradeEquipRequest  = 4003
	MSG_UpgradeEquipResponse = 4004

	// 英雄消息
	MSG_GetHeroesRequest        = 5001
	MSG_GetHeroesResponse       = 5002
	MSG_RecruitHeroRequest      = 5003
	MSG_RecruiHeroesResponse    = 5004
	MSG_UpStarHeroRequest       = 5005
	MSG_UpStarHeroesResponse    = 5006
	MSG_OpenSkillHeroRequest    = 5007
	MSG_UpSkillHeroRequest      = 5008
	MSG_OpenSkillHeroesResponse = 5009

	// 货币消息
	MSG_GetGameMoneyRequest  = 6001
	MSG_GetGameMoneyResponse = 6002
	MSG_SyncGameMoneyChange  = 6003
)

// MessageIDToName 消息ID到消息名的映射
var MessageIDToName = map[int32]string{
	1001: "MSG_LoginRequest",
	1002: "MSG_LoginResponse",
	1003: "MSG_GetRoleInfoRequest",
	1004: "MSG_GetRoleInfoResponse",
	1005: "MSG_GetActorListRequest",
	1006: "MSG_GetActorListResponse",
	2001: "MSG_ActorCreateRequest",
	2002: "MSG_ActorUseRequest",
	2003: "MSG_ActorUseWithNameRequest",
	2004: "MSG_ActorUseResponse",
	3001: "MSG_GetBagRequest",
	3002: "MSG_GetBagResponse",
	3003: "MSG_BagItemUseRequest",
	3004: "MSG_BagItemUseResponse",
	3005: "MSG_SyncBagItemChange",
	4001: "MSG_GetEquipRequest",
	4002: "MSG_GetEquipResponse",
	4003: "MSG_UpgradeEquipRequest",
	4004: "MSG_UpgradeEquipResponse",
	5001: "MSG_GetHeroesRequest",
	5002: "MSG_GetHeroesResponse",
	5003: "MSG_RecruitHeroRequest",
	5004: "MSG_RecruiHeroesResponse",
	5005: "MSG_UpStarHeroRequest",
	5006: "MSG_UpStarHeroesResponse",
	5007: "MSG_OpenSkillHeroRequest",
	5008: "MSG_UpSkillHeroRequest",
	5009: "MSG_OpenSkillHeroesResponse",
	6001: "MSG_GetGameMoneyRequest",
	6002: "MSG_GetGameMoneyResponse",
	6003: "MSG_SyncGameMoneyChange",
}

// GetMessageName 根据消息ID获取消息名
func GetMessageName(msgID int32) string {
	if name, ok := MessageIDToName[msgID]; ok {
		return name
	}
	return "UnknownMessage"
}

// GetMessageIDRange 获取消息ID范围
func GetMessageIDRange() (int32, int32) {
	return MSG_LoginRequest, MSG_SyncGameMoneyChange
}
