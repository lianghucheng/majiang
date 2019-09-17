package room

// 玩家状态
const (
	_               = iota
	GdReady         // 1 准备
	GdWaiting       // 2 等待
	GdActionDiscard // 3 前端显示出牌动作
	GdActionClaim   // 4 前端显示胡、杠、碰、吃动作
	GdWin           // 5 胡
	GdKong          // 6 杠
	GdPong          // 7 碰
	GdChow          // 8 吃
)

// 倒计时
const (
	Cd_gdDiscard = 20
	Cd_gdClaim   = 20
)

// 房间状态
const (
	RoomIdle    = iota // 0 空闲
	RoomGame           // 1 游戏中
	RoomGameEnd        // 2 游戏结束
)

// 房间类型
const (
	RoomPractice          = iota // 0 练习
	RoomRoomCardMatch            // 1 房卡匹配
	RoomPrivate                  // 2 私人
	RoomRedPacketMatching        // 3 红包匹配
	RoomRedPacketPrivate         // 4 红包私人
)

// 玩家解散房间动作码
const (
	_                    = iota
	ActionWaitingDisband // 1 等待解散
	ActionAgreeDisband   // 2 同意解散
)

var (
	roomNumbers = []int{}
	roomCounter = 0
)
