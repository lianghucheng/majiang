package card

import (
	"msg"
)

func init() {
	msg.MsgRegister(&C2S_GetRoomCards{})
	msg.MsgRegister(&S2C_UpdateRoomCards{})
	msg.MsgRegister(&C2S_TransferRoomCard{})
	msg.MsgRegister(&S2C_TransferRoomCard{})
	msg.MsgRegister(&C2S_IAPReceiptData{})
	msg.MsgRegister(&C2S_GetTransferRoomCardRecord{})
	msg.MsgRegister(&S2C_TransferRoomCardRecord{})
	msg.MsgRegister(&S2C_AllTransferRoomCardRecord{})
	msg.MsgRegister(&S2C_UpdateRoomCardsMatchOnlineNumber{})
	msg.MsgRegister(&C2S_GetAllTransferRoomCardRecord{})
}

//获取房卡数量
type C2S_GetRoomCards struct {
}

type S2C_UpdateRoomCards struct {
	RoomCards int //房卡数量
}

type C2S_TransferRoomCard struct {
	AccountID int
	RoomCards int
}

const (
	S2C_TransferRoomCard_OK               = 0
	S2C_TransferRoomCard_AccountIDInvalid = 1 // 账户ID无效
	S2C_TransferRoomCard_NotYourself      = 2 // 不能转给自己
	S2C_TransferRoomCard_RoomCardsInvalid = 3 // 房卡数量 + S2C_TransferRoomCard.RoomCards + 无效
	S2C_TransferRoomCard_PermissionDenied = 4 // 没有权限
)

type S2C_TransferRoomCard struct {
	Error     int
	RoomCards int
}

type TransferRoomCardUserInfo struct {
	FromAccountID  int
	FromNickName   string
	FromHeadimgurl string
	FromRole       int
	ToAccountID    int
	ToNickName     string
	ToHeadimgurl   string
	ToRole         int
	RoomCards      int
	Date           string
	Total          int // 一共多少条记录
	PageNumber     int // 页码
}

type C2S_IAPReceiptData struct {
	ReceiptData string
}

type C2S_GetTransferRoomCardRecord struct {
	AccountID  int
	PageNumber int
	PageSize   int
}

const (
	S2C_TransferRoomCardRecord_OK               = 0
	S2C_TransferRoomCardRecord_AccountIDInvalid = 1 // 账户ID无效
	S2C_TransferRoomCardRecord_PermissionDenied = 2 // 没有权限
)

type S2C_TransferRoomCardRecord struct {
	Error int
	Infos []TransferRoomCardUserInfo
}

type C2S_GetAllTransferRoomCardRecord struct {
	PageNumber int //页码数
	PageSize   int //条数
	StartTime  int64
	EndTime    int64
}

type S2C_AllTransferRoomCardRecord struct {
	Infos []TransferRoomCardUserInfo
}

// 更新房卡比赛在线人数
type S2C_UpdateRoomCardsMatchOnlineNumber struct {
	Numbers []int
}
