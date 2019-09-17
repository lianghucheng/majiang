package agent

import (
	"msg"
)

func init() {
	msg.MsgRegister(&C2S_GetAllAgentInfo{})
	msg.MsgRegister(&S2C_AllAgentInfo{})
	msg.MsgRegister(&C2S_GetAllUserInfo{})
	msg.MsgRegister(&S2C_AllUserInfo{})
	msg.MsgRegister(&C2S_GetBlackList{})
	msg.MsgRegister(&S2C_BlackList{})
}

type C2S_GetAllAgentInfo struct {
	PageNumber int // 页码数
	PageSize   int // 条数
	StartTime  int64
	EndTime    int64
}

type AgentInfo struct {
	JoinAgencyTime string // 加入代理时间
	Role           int
	AccountID      int    // 玩家ID
	Nickname       string // 玩家昵称
	RoomCards      int    // 持卡数量
	Total          int
	PageNumber     int // 页码
}

type S2C_AllAgentInfo struct {
	Infos []AgentInfo
}

type C2S_GetAllUserInfo struct {
	Nickname   string // 玩家昵称
	PageNumber int    // 页码数
	PageSize   int    // 条数
}

type S2C_AllUserInfo struct {
	Infos []UserInfo
}

type UserInfo struct {
	AccountID          int // 玩家ID
	Headimgurl         string
	Nickname           string
	Sex                int
	RoomCards          int
	GameScore          int    // 游戏积分
	ConsumedRoomCards  int    // 消耗的房卡
	PurchasedRoomCards int    // 一共购买的房卡
	NewUserYesterday   int    // 昨日新增人数
	OnlineUser         int    // 在线人数
	Total              int    // 用户总数
	Role               int    // 角色
	LastLogin          string // 上一次登录
	PageNumber         int    // 页码
}

type C2S_GetBlackList struct {
	PageNumber int // 页码数
	PageSize   int // 条数
}

type S2C_BlackList struct {
	Infos []UserInfo
}
