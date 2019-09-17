package mahjong

import (
	"msg"
)

func init() {
	msg.MsgRegister(&C2S_GetTotalResults{})
	msg.MsgRegister(&S2C_TotalResults{})
	msg.MsgRegister(&C2S_GetRoundResults{})
	msg.MsgRegister(&S2C_RoundResults{})
}

// 玩家单局成绩
type GDPlayerRoundResult struct {
	Nickname         string // 昵称
	Headimgurl       string // 头像
	Dealer           bool   // 庄家
	Hands            []int
	Claims           [][]int
	LastTile         int
	WinType          int     // 胡牌类型
	WinScore         int     // 胡牌得分
	CatchHorseScore  int     // 抓马得分
	ExposedKongScore int     // 明杠得分
	PongKongScore    int     // 碰杠得分
	HiddenKongScore  int     // 暗杠得分
	TotalScore       int     // 总分
	RoomCards        int     // (房卡匹配场有效)
	RedPacket        float64 // 红包种类(元): 1、5、10、50、100、200 (红包场有效)
}

// 玩家总成绩
type GDPlayerTotalResult struct {
	Nickname   string // 昵称
	Headimgurl string // 头像
	Owner      bool   // 房主
	AccountID  int    // 账户ID
	Scores     []int  // 每一轮得分
	TotalScore int    // 每一局得分总和
}

// 玩家解散信息
type GDPlayerDisbandInfo struct {
	Nickname   string // 昵称
	ActionCode int    // 0 等待 1 同意
}

type C2S_GetTotalResults struct{}

// 玩家成绩
type PlayerResult struct {
	Nickname  string // 昵称
	Score     int    // 分数
	RoomCards int    // (房卡匹配场有效)
}

type TotalResult struct {
	TotalResultID int
	RoomType      int    // 房间类型
	RoomNumber    string // 房号
	RoomDesc      string // 房间描述
	Result        int    // 0 输 1 赢 2 平
	Duration      string // 时长
	Position      int
	PlayerResults []PlayerResult
}

type S2C_TotalResults struct {
	Results []TotalResult
}

type C2S_GetRoundResults struct {
	TotalResultID int
}

type RoundResult struct {
	Round         int
	Duration      string // 时长
	Position      int
	PlayerResults []PlayerResult
}

type S2C_RoundResults struct {
	Results []RoundResult
}
