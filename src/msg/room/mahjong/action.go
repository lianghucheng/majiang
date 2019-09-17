package mahjong

import (
	"msg"
)

func init() {
	msg.MsgRegister(&C2S_Prepare{})
	msg.MsgRegister(&S2C_Prepare{})
	msg.MsgRegister(&S2C_GameStart{})
	msg.MsgRegister(&S2C_UpdateMahjongHands{})
	msg.MsgRegister(&S2C_UpdateMahjongDiscads{})
	msg.MsgRegister(&S2C_UpdateMahjongClaims{})
	msg.MsgRegister(&S2C_UpdateMahjongRestsNumber{})
	msg.MsgRegister(&S2C_UpdateMahjongCurrentRound{})
	msg.MsgRegister(&S2C_MahjongDraw{})
	msg.MsgRegister(&S2C_ActionMahjongDiscard{})
	msg.MsgRegister(&C2S_MahjongDiscard{})
	msg.MsgRegister(&S2C_ActionMahjongClaim{})
	msg.MsgRegister(&C2S_MahjongWin{})
	msg.MsgRegister(&C2S_MahjongKong{})
	msg.MsgRegister(&C2S_MahjongPong{})
	msg.MsgRegister(&C2S_MahjongChow{})
	msg.MsgRegister(&C2S_MahjongPass{})
	msg.MsgRegister(&S2C_MahjongWin{})
	msg.MsgRegister(&S2C_MahjongKong{})
	msg.MsgRegister(&S2C_MahjongPong{})
	msg.MsgRegister(&S2C_MahjongChow{})
	msg.MsgRegister(&S2C_UpdateMahjongDiscardCusor{})
	msg.MsgRegister(&S2C_DecideDealer{})
	msg.MsgRegister(&C2S_MahjongManaged{})
	msg.MsgRegister(&S2C_ManagedMahjongPass{})
	msg.MsgRegister(&S2C_MahjongDiscard{})
	msg.MsgRegister(&S2C_MahjongManaged{})
	msg.MsgRegister(&S2C_UpdateWinTiles{})
}

type S2C_UpdateWinTiles struct {
	Tiles []int
}
type C2S_Prepare struct{}

type S2C_Prepare struct {
	Position int
	Ready    bool
}

type S2C_GameStart struct{}

//更新玩家的手牌消息
type S2C_UpdateMahjongHands struct {
	Position      int
	Hands         []int // 手牌
	NumberOfHands int   // 手牌数量
}

//更新玩家打出的牌墙
type S2C_UpdateMahjongDiscads struct {
	Position int
	Discards []int // 打出的牌
}

//更新玩家吃碰杠的牌墙
type S2C_UpdateMahjongClaims struct {
	Position int
	Claims   [][]int // 吃、碰、杠到的牌
}

//更新牌局剩下的牌数
type S2C_UpdateMahjongRestsNumber struct {
	NumberOfRests int // 剩余牌数
}

//更新当前房间的局数
type S2C_UpdateMahjongCurrentRound struct {
	CurrentRound int // 当前局数
}

//更新玩家摸牌操作
type S2C_MahjongDraw struct {
	Position      int
	Tile          int // 摸起来的一张牌
	NumberOfHands int // 手牌数量
}

//更新玩家倒计时出牌操作
type S2C_ActionMahjongDiscard struct {
	Position  int
	Countdown int // 倒计时
}

type C2S_MahjongDiscard struct {
	Tile int
}

//更新玩家出的牌
type S2C_MahjongDiscard struct {
	Position int
	Tile     int
}

// 要牌动作（有杠的情况下就有碰，故3、10、11的情况不存在）
//更新玩家的要牌操作
type S2C_ActionMahjongClaim struct {
	Position    int
	ActionCode  int     // 1胡 2暗杠、碰杠 3胡杠 4碰 5胡碰 6杠碰 7胡杠碰 8吃 9胡吃 10杠吃 11胡杠吃 12碰吃 13胡碰吃 14杠碰吃 15胡杠碰吃
	Countdown   int     // 倒计时
	Sequences   [][]int // 所有可以吃的牌
	Quadruplets [][]int // 所有可以暗杠的牌
}

type C2S_MahjongWin struct{}

type C2S_MahjongKong struct {
	Meld []int
}

type C2S_MahjongPong struct{}

type C2S_MahjongChow struct {
	Meld []int
}

type C2S_MahjongPass struct{}

//更新玩家胡的消息
type S2C_MahjongWin struct {
	Position int
	WinType  int
}

type S2C_MahjongKong struct {
	Position int
}

type S2C_MahjongPong struct {
	Position int
}

type S2C_MahjongChow struct {
	Position int
}

//更新玩家出牌牌墙的牌
type S2C_UpdateMahjongDiscardCusor struct {
	Position int
	Index    int
}

//更新庄家的位置
type S2C_DecideDealer struct {
	Position int
}

//取消托管
type C2S_MahjongManaged struct {
	Managed bool
}

// 托管
type S2C_MahjongManaged struct {
	Managed bool
}

type S2C_ManagedMahjongPass struct{}
