package mahjong

import (
	"msg"
)

func init() {
	msg.MsgRegister(&C2S_StartGDMatching{})
	msg.MsgRegister(&S2C_DecideGDJoker{})
	msg.MsgRegister(&S2C_GDRoundResult{})
	msg.MsgRegister(&S2C_GDTotalResult{})
	msg.MsgRegister(&S2C_UpdateGDTotalScore{})
	msg.MsgRegister(&C2S_GetGDIOSProductList{})
	msg.MsgRegister(&S2C_SitDown{})
	msg.MsgRegister(&C2S_GetGDAndroidProductList{})
	msg.MsgRegister(&S2C_GDIOSProductList{})
	msg.MsgRegister(&S2C_GDBuyHorse{})
	msg.MsgRegister(&S2C_GDDisCardTing{})
	msg.MsgRegister(&C2S_CreateGDRoom{})
	msg.MsgRegister(&C2S_GetAllPlayers{})
	msg.MsgRegister(&S2C_CreateRoom{})
	msg.MsgRegister(&C2S_EnterRoom{})
	msg.MsgRegister(&S2C_EnterRoom{})
	msg.MsgRegister(&S2C_StandUp{})
	msg.MsgRegister(&C2S_ExitOrDisbandRoom{})
	msg.MsgRegister(&S2C_ExitRoom{})
	msg.MsgRegister(&S2C_DisbandRoom{})
	msg.MsgRegister(&S2C_ActionDisbandRoom{})
	msg.MsgRegister(&C2S_AgreeDisbandRoom{})
	msg.MsgRegister(&S2C_AgreeDisbandRoom{})
	msg.MsgRegister(&C2S_RefuseDisbandRoom{})

}

type C2S_StartGDMatching struct {
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
	RoundResults []GDPlayerRoundResult
	ContinueGame bool // 是否继续游戏
}

// 总成绩
type S2C_GDTotalResult struct {
	TotalResults []GDPlayerTotalResult
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

type ProductInfo struct {
	ID    string
	Desc  string
	Price int
}

type C2S_CreateGDRoom struct {
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

const (
	S2C_CreateRoom_OK              = 0
	S2C_CreateRoom_InnerError      = 1 // 创建房间出错，请稍后重试
	S2C_CreateRoom_CreateRepeated  = 2 // "房间: " + S2C_CreateRoom.RoomNumber + " 已存在"
	S2C_CreateRoom_InOtherRoom     = 3 // 正在其他房间对局，是否回去？
	S2C_CreateRoom_LackOfRoomCards = 4 // 房卡不足，需要 + S2C_S2C_CreateRoom.RoomCards + 张房卡才能游戏
	S2C_CreateRoom_RuleError       = 5 // 规则错误，请稍后重试
	S2C_CreateRoom_LocationError   = 6 // 定位参数错误，请检查GPS
)

type S2C_CreateRoom struct {
	Error     int
	RoomCards int // 需要的房卡数
}

type C2S_GetAllPlayers struct{}

type C2S_EnterRoom struct {
	RoomNumber string
	GPS        bool // 是否开启GPS
	Location   []float64
}

const (
	S2C_EnterRoom_OK              = 0
	S2C_EnterRoom_NotCreated      = 1 // "房间: " + S2C_EnterRoom.RoomNumber + " 未创建"
	S2C_EnterRoom_Full            = 2 // "房间: " + S2C_EnterRoom.RoomNumber + " 玩家人数已满"
	S2C_EnterRoom_Unknown         = 3 // 进入房间出错，请稍后重试
	S2C_EnterRoom_LackOfRoomCards = 4 // 房卡不足，需要 + S2C_EnterRoom.RoomCards + 张房卡才能进入
	S2C_EnterRoom_IPConflict      = 5 // IP重复，无法进入
	S2C_EnterRoom_GPSNotOpen      = 6 // 定位失败，请检查GPS是否开启
	S2C_EnterRoom_LocationError   = 7 // 定位参数错误，请检查GPS
	S2C_EnterRoom_NotRightNow     = 8 // 比赛暂未开始，请到时再来

)

type S2C_EnterRoom struct {
	Error         int
	RoomType      int // 房间类型， 0 练习，1 房卡匹配，2 私人
	RedPacketType int // 红包种类(元): 1、10、100、999
	RoomNumber    string
	Position      int
	RoomDesc      string
	MaxPlayers    int  // 最大玩家数
	MaxRounds     int  // 总局数
	NeedJoker     bool // 癞子
	BuyHorse      int  // 买马 1匹马、2匹马
	RoomCards     int  // 进入房间需要的房卡数
	GamePlaying   bool // 游戏是否进行中
}

type S2C_SitDown struct {
	Position   int
	AccountID  int
	LoginIP    string
	Nickname   string
	Headimgurl string
	Sex        int
	Owner      bool
	Ready      bool
	Location   []float64
}

type S2C_StandUp struct {
	Position int
}

type C2S_ExitOrDisbandRoom struct{}

const (
	S2C_ExitRoom_OK          = 0
	S2C_ExitRoom_GamePlaying = 1 // 游戏进行中，不能退出房间
)

type S2C_ExitRoom struct {
	Error    int
	Position int
}

const (
	S2C_DisbandRoom_OK           = 0
	S2C_DisbandRoom_PlayerRefuse = 1 // 玩家拒绝
)

type S2C_DisbandRoom struct {
	Error            int
	RoomNumber       string
	OwnerNickName    string // 房主
	RejecterNickName string // 拒绝者
}

type S2C_ActionDisbandRoom struct {
	ApplicantNickname  string                // 申请者
	PlayerDisbandInfos []GDPlayerDisbandInfo // 解散信息
	Enable             bool                  // 前端是否显示同意、拒绝
	WaitingTime        int                   // 等待时间 300 秒
}

type C2S_AgreeDisbandRoom struct{}

type S2C_AgreeDisbandRoom struct {
	Position int
	Nickname string
}

type C2S_RefuseDisbandRoom struct{}
