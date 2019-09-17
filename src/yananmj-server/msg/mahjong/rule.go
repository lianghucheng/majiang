package mahjong

type Rule struct {
	RoomType          int       // 房间类型 0 练习、1 房卡匹配、2 私人
	MaxRounds         int       // 局数 8、16
	MaxPlayers        int       // 人数 4
	BaseScore         int       // 底分 1
	RoomCards         int       // 需要房卡数
	RedDragonJoker    bool      // 红中癞子
	MustSelfDraw      bool      // true 只能自摸，false 可以点炮，默认false
	WithHonors        bool      // 是否带风牌
	Gun               bool      // 是否下炮子
	IPAntiCheat       bool      // IP 防作弊
	GPSAntiCheat      bool      // GPS 防作弊
	RedPacketType     int       // 红包种类(元): 1、5、10、50、100、200
	Location          []float64 // 房主的经纬度
	BuyHorse          int       //广东麻将   买马 1匹马、2匹马
	NeedJoker         bool      //广东麻将   癞子
	DistinguishDealer bool      //湖南转转麻将   true 分庄闲(庄家翻倍)，false 通庄，默认false
	Birds             int       //湖南转转麻将   抓鸟数，2、4、6
}
