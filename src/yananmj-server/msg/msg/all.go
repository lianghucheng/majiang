package msg

import "yananmj-server/msg/mahjong"

type S2C_DecideYananJoker struct {
	WildCard int   // 混儿
	Jokers   []int // 宝
}

//创建陕西麻将房间
type C2S_CreateRoom struct {
	mahjong.Rule
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
type S2C_RoundResult struct {
	Result       int
	RoomDesc     string
	Jokers       []int
	RoundResults []mahjong.YananPlayerRoundResult
	ContinueGame bool // 是否继续游戏
}

//总成绩
type S2C_TotalResult struct {
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

type S2C_IOSProductList struct {
	Infos []ProductInfo
}

type C2S_GetYananAndriodProductList struct{}

type S2C_AndriodProductList struct {
	Infos []ProductInfo
}

//广东#####################################
type C2S_StartMatching struct {
	RoomType      int // 0 练习、1 房卡匹配场、2 私人房、 3 红包匹配、 4 红包私人
	RoomCards     int
	RedPacketType int // 红包种类(元): 1、10
}

type S2C_DecideJoker struct {
	WildCard int   // 混儿
	Jokers   []int // 宝
}

type S2C_UpdateTotalScore struct {
	Position   int
	TotalScore int // 总分
}

type C2S_GetIOSProductList struct {
}

type C2S_GetAndroidProductList struct {
}

type S2C_AndroidProductList struct {
	Infos []ProductInfo
}

type S2C_BuyHorse struct {
	Position int
	Tiles    []int
}

//湖南##################################
type S2C_CatchBird struct {
	Position int
	Tiles    []int
	Birds    []int
}
