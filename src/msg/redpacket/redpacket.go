package redpacket

import (
	"msg"

	"gopkg.in/mgo.v2/bson"
)

func init() {
	msg.MsgRegister(&C2S_GetRedPacketMatchRecord{})
	msg.MsgRegister(&S2C_RedPacketMatchRecord{})
	msg.MsgRegister(&C2S_TakeRedPacketMatchPrize{})
	msg.MsgRegister(&S2C_TakeRedPacketMatchPrize{})
	msg.MsgRegister(&S2C_UpdateUntakenRedPacketMatchPrizeNumber{})
	msg.MsgRegister(&S2C_UpdateRedPacketMatchOnlineNumber{})
}

// 获取红包比赛记录
type C2S_GetRedPacketMatchRecord struct {
	PageNumber int // 页码数
	PageSize   int // 一页显示的条数
}

type RedPacketMatchRecordItem struct {
	ID            bson.ObjectId
	RedPacketType int
	RedPacket     float64
	Taken         bool
	Date          string
}

type S2C_RedPacketMatchRecord struct {
	Items      []RedPacketMatchRecordItem
	Total      int // 总数
	PageNumber int // 页码数
	PageSize   int // 一页显示的条数
}

// 领取红包比赛奖励
type C2S_TakeRedPacketMatchPrize struct {
	ID bson.ObjectId
}

const (
	S2C_TakeRedPacketMatchPrize_OK              = 0 // 恭喜领取 S2C_TakeRedPacketMatchPrize.RedPacket元红包奖励，请至“圈圈”查看
	S2C_TakeRedPacketMatchPrize_IDInvalid       = 1 // 比赛记录ID无效
	S2C_TakeRedPacketMatchPrize_NotYetWon       = 2 // 离获奖还差一点点，请继续努力吧
	S2C_TakeRedPacketMatchPrize_TakeRepeated    = 3 // S2C_TakeRedPacketMatchPrize.RedPacket元红包奖励已被领取，请勿重复操作
	S2C_TakeRedPacketMatchPrize_CircleIDInvalid = 4 // 圈圈ID无效
	S2C_TakeRedPacketMatchPrize_Error           = 5 // 领取出错，请稍后重试

)

// 领取红包比赛奖励
type S2C_TakeRedPacketMatchPrize struct {
	Error     int
	ID        bson.ObjectId
	RedPacket float64
}

// 更新未领取的红包比赛奖励数量
type S2C_UpdateUntakenRedPacketMatchPrizeNumber struct {
	Number int
}

// 更新红包比赛在线人数
type S2C_UpdateRedPacketMatchOnlineNumber struct {
	Numbers []int
}
