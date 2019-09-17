package msg

import (
	"github.com/name5566/leaf/network/json"
	"gopkg.in/mgo.v2/bson"
)

var Processor = json.NewProcessor()

func init() {
	Processor.Register(&C2S_Heartbeat{})
	Processor.Register(&C2S_WeChatLogin{})
	Processor.Register(&C2S_TokenLogin{})
	Processor.Register(&C2S_UsernamePasswordLogin{})
	Processor.Register(&C2S_SetUsernamePassword{})
	Processor.Register(&C2S_SetUserRole{})
	Processor.Register(&C2S_SetGDConfig{})
	Processor.Register(&C2S_SetSystemOn{})
	Processor.Register(&C2S_CreateGDRoom{})
	Processor.Register(&C2S_StartGDMatching{})
	Processor.Register(&C2S_EnterRoom{})
	Processor.Register(&C2S_GetAllPlayers{})
	Processor.Register(&C2S_ExitOrDisbandRoom{})
	Processor.Register(&C2S_Prepare{})
	Processor.Register(&C2S_MahjongDiscard{})
	Processor.Register(&C2S_MahjongWin{})
	Processor.Register(&C2S_MahjongKong{})
	Processor.Register(&C2S_MahjongPong{})
	Processor.Register(&C2S_MahjongChow{})
	Processor.Register(&C2S_MahjongPass{})
	Processor.Register(&C2S_AgreeDisbandRoom{})
	Processor.Register(&C2S_RefuseDisbandRoom{})
	Processor.Register(&C2S_GetRoomCards{})
	Processor.Register(&C2S_TransferRoomCard{})
	Processor.Register(&C2S_GetTotalResults{})
	Processor.Register(&C2S_GetRoundResults{})
	Processor.Register(&C2S_CompleteDailyShare{})
	Processor.Register(&C2S_TextMessage{})
	Processor.Register(&C2S_ExpressionMessage{})
	Processor.Register(&C2S_GCloudVoiceMessage{})
	Processor.Register(&C2S_GetGDAndroidProductList{})
	Processor.Register(&C2S_GetGDIOSProductList{})
	Processor.Register(&C2S_IAPReceiptData{})
	Processor.Register(&C2S_GetUserInfo{})
	Processor.Register(&C2S_GetTransferRoomCardRecord{})
	Processor.Register(&C2S_GetAllTransferRoomCardRecord{})
	Processor.Register(&C2S_GetAllAgentInfo{})
	Processor.Register(&C2S_GetAllUserInfo{})
	Processor.Register(&C2S_GetBlackList{})
	Processor.Register(&C2S_MahjongManaged{})
	Processor.Register(&C2S_GetRedPacketMatchRecord{})
	Processor.Register(&C2S_TakeRedPacketMatchPrize{})
	Processor.Register(&C2S_FakeWXPay{})
	Processor.Register(&C2S_GetCircleLoginCode{})

	Processor.Register(&S2C_Heartbeat{})
	Processor.Register(&S2C_Login{})
	Processor.Register(&S2C_Close{})
	Processor.Register(&S2C_SetGDConfig{})
	Processor.Register(&S2C_SetUserRole{})
	Processor.Register(&S2C_UpdateWinTiles{})
	Processor.Register(&S2C_UpdateNotice{})
	Processor.Register(&S2C_UpdateRadio{})
	Processor.Register(&S2C_UpdateRoomCards{})
	Processor.Register(&S2C_SetSystemOn{})
	Processor.Register(&S2C_CreateRoom{})
	Processor.Register(&S2C_EnterRoom{})
	Processor.Register(&S2C_SitDown{})
	Processor.Register(&S2C_StandUp{})
	Processor.Register(&S2C_ExitRoom{})
	Processor.Register(&S2C_DisbandRoom{})
	Processor.Register(&S2C_ActionDisbandRoom{})
	Processor.Register(&S2C_AgreeDisbandRoom{})
	Processor.Register(&S2C_Prepare{})
	Processor.Register(&S2C_GameStart{})
	Processor.Register(&S2C_UpdateMahjongHands{})
	Processor.Register(&S2C_GDBuyHorse{})
	Processor.Register(&S2C_UpdateMahjongDiscads{})
	Processor.Register(&S2C_UpdateMahjongDiscardCusor{})
	Processor.Register(&S2C_UpdateMahjongClaims{})
	Processor.Register(&S2C_UpdateMahjongRestsNumber{})
	Processor.Register(&S2C_UpdateMahjongCurrentRound{})
	Processor.Register(&S2C_MahjongDraw{})
	Processor.Register(&S2C_ActionMahjongDiscard{})
	Processor.Register(&S2C_MahjongDiscard{})
	Processor.Register(&S2C_MahjongWin{})
	Processor.Register(&S2C_MahjongKong{})
	Processor.Register(&S2C_MahjongPong{})
	Processor.Register(&S2C_MahjongChow{})
	Processor.Register(&S2C_ActionMahjongClaim{})
	Processor.Register(&S2C_DecideDealer{})
	Processor.Register(&S2C_DecideGDJoker{})
	Processor.Register(&S2C_GDRoundResult{})
	Processor.Register(&S2C_GDTotalResult{})
	Processor.Register(&S2C_UpdateGDTotalScore{})
	Processor.Register(&S2C_TransferRoomCard{})
	Processor.Register(&S2C_TotalResults{})
	Processor.Register(&S2C_RoundResults{})
	Processor.Register(&S2C_CompleteDailyShare{})
	Processor.Register(&S2C_TextMessage{})
	Processor.Register(&S2C_ExpressionMessage{})
	Processor.Register(&S2C_GCloudVoiceMessage{})
	Processor.Register(&S2C_GDAndroidProductList{})
	Processor.Register(&S2C_GDIOSProductList{})
	Processor.Register(&S2C_UserInfo{})
	Processor.Register(&S2C_TransferRoomCardRecord{})
	Processor.Register(&S2C_AllTransferRoomCardRecord{})
	Processor.Register(&S2C_AllAgentInfo{})
	Processor.Register(&S2C_AllUserInfo{})
	Processor.Register(&S2C_BlackList{})
	Processor.Register(&S2C_RedPacketMatchRecord{})
	Processor.Register(&S2C_TakeRedPacketMatchPrize{})
	Processor.Register(&S2C_UpdateUntakenRedPacketMatchPrizeNumber{})
	Processor.Register(&S2C_UpdateRedPacketMatchOnlineNumber{})
	Processor.Register(&S2C_UpdateRoomCardsMatchOnlineNumber{})
	Processor.Register(&S2C_MahjongManaged{})
	Processor.Register(&S2C_ManagedMahjongPass{})
	Processor.Register(&S2C_PayOK{})
	Processor.Register(&S2C_UpdateCircleLoginCode{})
	Processor.Register(&S2C_GDDisCardTing{})

	// manage robot
	Processor.Register(&C2S_SetRobotData{})
	Processor.Register(&C2S_SetGun{})
	Processor.Register(&S2C_ActionSetGun{})
	Processor.Register(&S2C_SetGun{})
}

type C2S_Heartbeat struct{}

type S2C_Heartbeat struct{}

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

// 系统开关
type C2S_SetSystemOn struct {
	On bool
}

const (
	S2C_SetSystemOn_OK               = 0
	S2C_SetSystemOn_PermissionDenied = 1 // 没有权限
)

type S2C_SetSystemOn struct {
	Error int
	On    bool
}

type C2S_SetUsernamePassword struct {
	Username string
	Password string
}

type C2S_SetUserRole struct {
	AccountID int
	Role      int //-1 黑名单 0 机器人 1 玩家 2 代理 3 管理员 4 超管
}

const (
	S2C_SetUserRole_OK               = 0
	S2C_SetUserRole_AccountIDInvalid = 1 // 账户ID无效
	S2C_SetUserRole_NotYourself      = 2 // 不能设置自己
	S2C_SetUserRole_RoleInvalid      = 3 // 角色 + S2C_SetUserRole.Role + 无效
	S2C_SetUserRole_PermissionDenied = 4 // 没有权限
	S2C_SetUserRole_SetRepeated      = 5 // 用户已经是 S2C_SetUserRole.Role(1 玩家 2 二级代理 3 一级代理)
)

type S2C_SetUserRole struct {
	Error int
	Role  int
}

// 更新胡牌提示
type S2C_UpdateWinTiles struct {
	Tiles []int
}

// 更新公告
type S2C_UpdateNotice struct {
	Notice string
}

// 更新广播
type S2C_UpdateRadio struct {
	Radio string
}

//获取房卡数量
type C2S_GetRoomCards struct {
}

type S2C_UpdateRoomCards struct {
	RoomCards int //房卡数量
}

type C2S_TransferRoomCard struct {
	AccountID int
	RoomCards int
}

const (
	S2C_TransferRoomCard_OK               = 0
	S2C_TransferRoomCard_AccountIDInvalid = 1 // 账户ID无效
	S2C_TransferRoomCard_NotYourself      = 2 // 不能转给自己
	S2C_TransferRoomCard_RoomCardsInvalid = 3 // 房卡数量 + S2C_TransferRoomCard.RoomCards + 无效
	S2C_TransferRoomCard_PermissionDenied = 4 // 没有权限
)

type S2C_TransferRoomCard struct {
	Error     int
	RoomCards int
}

type TransferRoomCardUserInfo struct {
	FromAccountID  int
	FromNickName   string
	FromHeadimgurl string
	FromRole       int
	ToAccountID    int
	ToNickName     string
	ToHeadimgurl   string
	ToRole         int
	RoomCards      int
	Date           string
	Total          int // 一共多少条记录
	PageNumber     int // 页码
}

type C2S_GetTotalResults struct{}

// 玩家成绩
type PlayerResult struct {
	Nickname  string // 昵称
	Score     int    // 分数
	RoomCards int    // (房卡匹配场有效)
}

type TotalResult struct {
	TotalResultID int
	RoomType      int    // 房间类型
	RoomNumber    string // 房号
	RoomDesc      string // 房间描述
	Result        int    // 0 输 1 赢 2 平
	Duration      string // 时长
	Position      int
	PlayerResults []PlayerResult
}

type S2C_TotalResults struct {
	Results []TotalResult
}

type C2S_GetRoundResults struct {
	TotalResultID int
}

type RoundResult struct {
	Round         int
	Duration      string // 时长
	Position      int
	PlayerResults []PlayerResult
}

type S2C_RoundResults struct {
	Results []RoundResult
}

type C2S_CompleteDailyShare struct{}

type S2C_CompleteDailyShare struct {
	RoomCards int
}

type C2S_TextMessage struct {
	Message string
}

type S2C_TextMessage struct {
	Position int
	Message  string
}

type C2S_ExpressionMessage struct {
	Expression int
}

type S2C_ExpressionMessage struct {
	Position   int
	Expression int
}

type C2S_GCloudVoiceMessage struct {
	FileID string
}

type S2C_GCloudVoiceMessage struct {
	Position int
	FileID   string
}

type ProductInfo struct {
	ID    string
	Desc  string
	Price int
}

type C2S_IAPReceiptData struct {
	ReceiptData string
}

type C2S_GetUserInfo struct {
	AccountID int
}

const (
	S2C_UserInfo_OK               = 0
	S2C_UserInfo_AccountIDInvalid = 1 // 账户ID无效
)

type S2C_UserInfo struct {
	Error              int
	AccountID          int
	Nickname           string
	Headimgurl         string
	Sex                int
	RoomCards          int    // 持卡数量
	JoinAgencyTime     string // 加入代理时间
	Role               int    // 角色 1 玩家、2 代理、3 管理员、4 超管
	GameScore          int    // 游戏积分
	ConsumedRoomCards  int    // 消耗的房卡
	PurchasedRoomCards int    // 一共购买的房卡
	LastLogin          string // 上一次登录
}

type C2S_GetTransferRoomCardRecord struct {
	AccountID  int
	PageNumber int
	PageSize   int
}

const (
	S2C_TransferRoomCardRecord_OK               = 0
	S2C_TransferRoomCardRecord_AccountIDInvalid = 1 // 账户ID无效
	S2C_TransferRoomCardRecord_PermissionDenied = 2 // 没有权限
)

type S2C_TransferRoomCardRecord struct {
	Error int
	Infos []TransferRoomCardUserInfo
}

type C2S_GetAllTransferRoomCardRecord struct {
	PageNumber int //页码数
	PageSize   int //条数
	StartTime  int64
	EndTime    int64
}

type S2C_AllTransferRoomCardRecord struct {
	Infos []TransferRoomCardUserInfo
}

type C2S_GetAllAgentInfo struct {
	PageNumber int // 页码数
	PageSize   int // 条数
	StartTime  int64
	EndTime    int64
}

type AgentInfo struct {
	JoinAgencyTime string // 加入代理时间
	Role           int
	AccountID      int    // 玩家ID
	Nickname       string // 玩家昵称
	RoomCards      int    // 持卡数量
	Total          int
	PageNumber     int // 页码
}

type S2C_AllAgentInfo struct {
	Infos []AgentInfo
}

type C2S_GetAllUserInfo struct {
	Nickname   string // 玩家昵称
	PageNumber int    // 页码数
	PageSize   int    // 条数
}

type S2C_AllUserInfo struct {
	Infos []UserInfo
}

type UserInfo struct {
	AccountID          int // 玩家ID
	Headimgurl         string
	Nickname           string
	Sex                int
	RoomCards          int
	GameScore          int    // 游戏积分
	ConsumedRoomCards  int    // 消耗的房卡
	PurchasedRoomCards int    // 一共购买的房卡
	NewUserYesterday   int    // 昨日新增人数
	OnlineUser         int    // 在线人数
	Total              int    // 用户总数
	Role               int    // 角色
	LastLogin          string // 上一次登录
	PageNumber         int    // 页码
}

type C2S_GetBlackList struct {
	PageNumber int // 页码数
	PageSize   int // 条数
}

type S2C_BlackList struct {
	Infos []UserInfo
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

// 获取红包比赛记录
type C2S_GetRedPacketMatchRecord struct {
	PageNumber int // 页码数
	PageSize   int // 一页显示的条数
}

type RedPacketMatchRecordItem struct {
	ID            bson.ObjectId
	RedPacketType int
	RedPacket     float64
	Taken         bool
	Date          string
}

type S2C_RedPacketMatchRecord struct {
	Items      []RedPacketMatchRecordItem
	Total      int // 总数
	PageNumber int // 页码数
	PageSize   int // 一页显示的条数
}

// 领取红包比赛奖励
type C2S_TakeRedPacketMatchPrize struct {
	ID bson.ObjectId
}

const (
	S2C_TakeRedPacketMatchPrize_OK              = 0 // 恭喜领取 S2C_TakeRedPacketMatchPrize.RedPacket元红包奖励，请至“圈圈”查看
	S2C_TakeRedPacketMatchPrize_IDInvalid       = 1 // 比赛记录ID无效
	S2C_TakeRedPacketMatchPrize_NotYetWon       = 2 // 离获奖还差一点点，请继续努力吧
	S2C_TakeRedPacketMatchPrize_TakeRepeated    = 3 // S2C_TakeRedPacketMatchPrize.RedPacket元红包奖励已被领取，请勿重复操作
	S2C_TakeRedPacketMatchPrize_CircleIDInvalid = 4 // 圈圈ID无效
	S2C_TakeRedPacketMatchPrize_Error           = 5 // 领取出错，请稍后重试

)

// 领取红包比赛奖励
type S2C_TakeRedPacketMatchPrize struct {
	Error     int
	ID        bson.ObjectId
	RedPacket float64
}

// 更新未领取的红包比赛奖励数量
type S2C_UpdateUntakenRedPacketMatchPrizeNumber struct {
	Number int
}

// 更新红包比赛在线人数
type S2C_UpdateRedPacketMatchOnlineNumber struct {
	Numbers []int
}

// 更新房卡比赛在线人数
type S2C_UpdateRoomCardsMatchOnlineNumber struct {
	Numbers []int
}

type C2S_FakeWXPay struct {
	TotalFee int
}

// 购买S2C_PayOK.RoomCards 房卡成功
type S2C_PayOK struct {
	RoomCards int
}

type C2S_GetCircleLoginCode struct{}

const (
	S2C_UpdateCircleLoginCode_OK    = 0
	S2C_UpdateCircleLoginCode_Error = 1 // 圈圈授权出错，请稍后重试
)

type S2C_UpdateCircleLoginCode struct {
	Error     int
	LoginCode string
}

// robot
type C2S_SetRobotData struct {
	LoginIP   string
	RoomCards int
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
