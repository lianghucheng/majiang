package msg

import "hnzzmj-server/game/mahjong"

type S2C_DecideHNZZJoker struct {
	WildCard int   // 混儿
	Jokers   []int // 宝
}

// 单局成绩
type S2C_HNZZRoundResult struct {
	Result       int // 失败、胜利、流局
	RoomDesc     string
	Jokers       []int
	RoundResults []mahjong.HNZZPlayerRoundResult
	ContinueGame bool // 是否继续游戏
}

// 总成绩
type S2C_HNZZTotalResult struct {
	TotalResults []mahjong.HNZZPlayerTotalResult
}

type S2C_UpdateHNZZTotalScore struct {
	Position   int
	TotalScore int // 总分
}

type C2S_StartHNZZMatching struct {
	RoomType      int // 0 练习、1 房卡匹配场、2 私人房、 3 红包匹配、 4 红包私人
	RoomCards     int
	RedPacketType int // 红包种类(元): 1、10
}

type C2S_GetHNZZIOSProductList struct {
}

type C2S_GetHNZZAndroidProductList struct {
}

type S2C_HNZZIOSProductList struct {
	Infos []ProductInfo
}

type S2C_HNZZAndroidProductList struct {
	Infos []ProductInfo
}

type S2C_HNZZCatchBird struct {
	Position int
	Tiles    []int
	Birds    []int
}
