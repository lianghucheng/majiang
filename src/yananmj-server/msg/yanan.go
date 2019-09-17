package data_struct

import (
	"yananmj-server/game/mahjong"
)

type S2C_DecideYananJoker struct {
	WildCard int   // 混儿
	Jokers   []int // 宝
}

//创建陕西麻将房间
type C2S_CreateYananRoom struct {
	mahjong.YananRule
}

//下炮子
type S2C_ActionSetGun struct { ////通知前端下炮子
	Countdown int
}

type C2S_SetGun struct {
	Gun int //下了多少炮
}

type S2C_SetGun struct { //下炮子结果
	Position int
	Gun      int
}

//单局成绩
type S2C_YananRoundResult struct {
	Result       int
	RoomDesc     string
	Jokers       []int
	RoundResults []mahjong.YananPlayerRoundResult
	ContinueGame bool // 是否继续游戏
}

//总成绩
type S2C_YananTotalResult struct {
	TotalResults []mahjong.YananPlayerTotalResult
}

//更新总分
type S2C_UpdateYananToTalScore struct {
	Position   int
	TotalScore int //总分
}

//开始练习或房卡匹配房
type C2S_StartYananMatching struct {
	RoomType      int // 0 练习、1 房卡匹配场、2 私人房、 3 红包匹配、 4 红包私人
	RoomCards     int
	RedPacketType int // 红包种类(元): 1、10
}

//获取ios、andriod商品列表
type C2S_GetYananIOSProductList struct{}

type S2C_YananIOSProductList struct {
	Infos []ProductInfo
}

type C2S_GetYananAndriodProductList struct{}

type S2C_YananAndriodProductList struct {
	Infos []ProductInfo
}
