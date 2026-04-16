package grpc

// 消息号定义
const (
	// 基础消息
	MessageIDLoginRequest          = 1001
	MessageIDLoginResponse         = 1002
	MessageIDGetRoleInfoRequest    = 1003
	MessageIDGetRoleInfoResponse   = 1004

	// 角色消息
	MessageIDActorCreateRequest    = 2001
	MessageIDActorUseRequest       = 2002
	MessageIDActorUseWithNameRequest = 2003
	MessageIDActorUseResponse      = 2004

	// 背包消息
	MessageIDGetBagRequest         = 3001
	MessageIDGetBagResponse        = 3002
	MessageIDBagItemUseRequest     = 3003
	MessageIDBagItemUseResponse    = 3004
	MessageIDSyncBagItemChange     = 3005

	// 装备消息
	MessageIDGetEquipRequest       = 4001
	MessageIDGetEquipResponse      = 4002
	MessageIDUpgradeEquipRequest   = 4003
	MessageIDUpgradeEquipResponse  = 4004

	// 英雄消息
	MessageIDGetHeroesRequest      = 5001
	MessageIDGetHeroesResponse     = 5002
	MessageIDRecruitHeroRequest    = 5003
	MessageIDRecruiHeroesResponse  = 5004
	MessageIDUpStarHeroRequest     = 5005
	MessageIDUpStarHeroesResponse  = 5006
	MessageIDOpenSkillHeroRequest  = 5007
	MessageIDUpSkillHeroRequest    = 5008
	MessageIDOpenSkillHeroesResponse = 5009

	// 货币消息
	MessageIDGetGameMoneyRequest   = 6001
	MessageIDGetGameMoneyResponse  = 6002
	MessageIDSyncGameMoneyChange   = 6003
)

// GetMessageIDRange 获取消息ID范围
func GetMessageIDRange() (int32, int32) {
	return MessageIDLoginRequest, MessageIDSyncGameMoneyChange
}
