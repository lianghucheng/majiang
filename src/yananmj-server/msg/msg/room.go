package msg

//创建房间返回常量
const (
	S2C_CreateRoom_OK              = 0
	S2C_CreateRoom_InnerError      = 1 // 创建房间出错，请稍后重试
	S2C_CreateRoom_CreateRepeated  = 2 // "房间: " + S2C_CreateRoom.RoomNumber + " 已存在"
	S2C_CreateRoom_InOtherRoom     = 3 // 正在其他房间对局，是否回去？
	S2C_CreateRoom_LackOfRoomCards = 4 // 需要 + S2C_S2C_CreateRoom.RoomCards + 张房卡才能创建
	S2C_CreateRoom_RuleError       = 5 // 规则错误，请稍后重试
	S2C_CreateRoom_LocationError   = 6 // 定位参数错误，请检查GPS
)

type S2C_CreateRoom struct {
	Error      int
	RoomNumber string //房号
	RoomCards  int
}

//进入房间
type C2S_EnterRoom struct {
	RoomNumber string
	GPS        bool // 是否开启GPS
	Location   []float64
}

//进入房间返回常量
const (
	S2C_EnterRoom_Ok                = 0 // 进入房间成功
	S2C_EnterRoom_NotCreated        = 1 // 房间+"S2C_EnterRoom.RoomNumber"+未创建
	S2C_EnterRoom_NotAllowBystander = 2 // 房间已满
	S2C_EnterRoom_Unknow            = 3 // 进入房间出错 请稍后重试
	S2C_EnterRoom_LackOfRoomCards   = 4 // 需要n张房卡才能进入
	S2C_EnterRoom_IPConflict        = 5 // IP重复，无法进入
	S2C_EnterRoom_GPSNotOpen        = 6 // 定位失败，请检查GPS是否开启
	S2C_EnterRoom_LocationError     = 7 // 定位参数错误，请检查GPS
	S2C_EnterRoom_NotRightNow       = 8 // 比赛暂未开始，请到时再来
)

type S2C_EnterRoom struct {
	Error          int
	RoomType       int
	RoomNumber     string
	RedPacketType  int // 红包种类(元): 1、10、100、999
	Position       int
	RoomDesc       string
	MaxPlayers     int  // 最大玩家数
	MaxRounds      int  // 总局数
	RoomCards      int  // 需要的房卡数量
	Gun            bool // 是否下炮子
	RedDragonJoker bool // 红中癞子
	MustSelfDraw   bool // true 只能自摸，false 可以点炮，默认false
	GamePlaying    bool
}

const (
	S2C_ExitRoom_OK          = 0
	S2C_ExitRoom_GamePlaying = 1 // 游戏进行中，不能退出房间
)

//退出房间
type C2S_ExitOrDisbandRoom struct{}

const (
	S2C_DisbandRoom_OK           = 0
	S2C_DisbandRoom_PlayerRefuse = 1 // 玩家拒绝
)

type S2C_ExitRoom struct {
	Error    int
	Position int
}

type S2C_DisbandRoom struct {
	Error            int
	RoomNumber       string
	OwnerNickName    string // 房主
	RejecterNickName string // 拒绝者
}

//获取所有玩家
type C2S_GetAllPlayers struct{}

type S2C_StandUp struct {
	Position int
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
