package msg

import (
	"gdmj-server/game/mahjong"
)

type C2S_StartGDMatching struct {
	GameType      int
	RoomType      int // 0 练习、1 房卡匹配场、2 私人房、 3 红包匹配、 4 红包私人
	RoomCards     int
	RedPacketType int // 红包种类(元): 1、10
}

type S2C_DecideGDJoker struct {
	WildCard int   // 混儿
	Jokers   []int // 宝
}

// 单局成绩
type S2C_GDRoundResult struct {
	Result       int // 失败、胜利、流局
	RoomDesc     string
	Jokers       []int
	RoundResults []mahjong.GDPlayerRoundResult
	ContinueGame bool // 是否继续游戏
}

// 总成绩
type S2C_GDTotalResult struct {
	TotalResults []mahjong.GDPlayerTotalResult
}

type S2C_UpdateGDTotalScore struct {
	Position   int
	TotalScore int // 总分
}

type C2S_GetGDIOSProductList struct {
}

type C2S_GetGDAndroidProductList struct {
}

type S2C_GDIOSProductList struct {
	Infos []ProductInfo
}

type S2C_GDAndroidProductList struct {
	Infos []ProductInfo
}

type S2C_GDBuyHorse struct {
	Position int
	Tiles    []int
}

type S2C_GDDisCardTing struct {
	Ting  []TingCard
	Count int
}
type TingCard struct {
	Card   int
	HuCard []int
}

//湖南转转麻将
type S2C_HNZZCatchBird struct {
	Position int
	Tiles    []int
	Birds    []int
}