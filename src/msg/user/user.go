package user

import (
	"msg"
)

func init() {
	msg.MsgRegister(&C2S_GetUserInfo{})
	msg.MsgRegister(&S2C_UserInfo{})
	msg.MsgRegister(&C2S_SetUsernamePassword{})
}

type C2S_GetUserInfo struct {
	AccountID int
}

const (
	S2C_UserInfo_OK               = 0
	S2C_UserInfo_AccountIDInvalid = 1 // 账户ID无效
)

type S2C_UserInfo struct {
	Error              int
	AccountID          int
	Nickname           string
	Headimgurl         string
	Sex                int
	RoomCards          int    // 持卡数量
	JoinAgencyTime     string // 加入代理时间
	Role               int    // 角色 1 玩家、2 代理、3 管理员、4 超管
	GameScore          int    // 游戏积分
	ConsumedRoomCards  int    // 消耗的房卡
	PurchasedRoomCards int    // 一共购买的房卡
	LastLogin          string // 上一次登录
}

type C2S_SetUsernamePassword struct {
	Username string
	Password string
}
