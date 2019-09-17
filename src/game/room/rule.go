package room

import (
	"msg/room/mahjong"
	"time"
)

type GDRule struct {
	RoomType      int       // 0 练习、1 房卡匹配场、2 私人房、 3 红包匹配、 4 红包私人
	MaxRounds     int       // 局数 4、8、16
	MaxPlayers    int       // 人数 2、3、4
	MustSelfDraw  bool      // true 只能自摸，false 可以点炮，默认false
	BaseScore     int       // 底分，1
	BuyHorse      int       // 买马 1匹马、2匹马
	WithHonors    bool      // 是否带风牌
	NeedJoker     bool      // 癞子
	RoomCards     int       // 需要的房卡数量
	IPAntiCheat   bool      // IP 防作弊
	GPSAntiCheat  bool      // GPS 防作弊
	RedPacketType int       // 红包种类(元): 1、5、10、50、100、200
	Location      []float64 // 房主的经纬度
}

func NewRule(info *mahjong.C2S_CreateGDRoom, neencard int) *GDRule {
	return &GDRule{
		RoomType:     RoomPrivate,
		MaxRounds:    info.MaxRounds,
		MaxPlayers:   info.MaxPlayers,
		MustSelfDraw: info.MustSelfDraw,
		BaseScore:    1,
		BuyHorse:     info.BuyHorse,
		WithHonors:   info.WithHonors,
		NeedJoker:    info.NeedJoker,
		RoomCards:    neencard,
		IPAntiCheat:  info.IPAntiCheat,
	}
}

func Start() bool {
	nowTime := time.Now()
	noon := time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day(), 12, 0, 0, 0, time.Local)
	_13oClock := time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day(), 13, 0, 0, 0, time.Local)
	_20oClock := time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day(), 20, 0, 0, 0, time.Local)
	_22oClock := time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day(), 22, 0, 0, 0, time.Local)
	// 小于12点、大于13点且小于20点、大于22点
	if nowTime.Unix() < noon.Unix() || nowTime.Unix() > _13oClock.Unix() && nowTime.Unix() < _20oClock.Unix() || nowTime.Unix() > _22oClock.Unix() {
		return false
	}
	return true
}
