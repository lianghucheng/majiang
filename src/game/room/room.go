package room

import (
	"algorithm"
	. "db"
	"fmt"
	. "game"

	. "game/player"
	. "msg/room/mahjong"
	"util"

	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/timer"
)

type Room struct {
	State                   int
	LoginIPs                map[string]bool
	PositionUserIDs         map[int]int // key: 座位号, value: userID
	CreatorUserID           int         // 创建者 userID
	OwnerUserID             int         // 房主 userID
	Number                  string
	Desc                    string
	StartTimestamp          int64 // 开始时间
	EachRoundStartTimestamp int64 // 每一局开始时间
	EndTimestamp            int64 // 结束时间
}

func init() {
	for i := 0; i < 1000000; i++ {
		roomNumbers = append(roomNumbers, i)
	}
	roomNumbers = util.Shuffle(roomNumbers)
}

func getRoomNumber() string {
	log.Debug("房间计数器: %v", roomCounter)
	roomNumber := fmt.Sprintf("%06d", roomNumbers[roomCounter])
	roomCounter = (roomCounter + 1) % 1000000
	return roomNumber
}

func toRelativePosition(pos int, zeroPos int, maxPlayers int) int {
	return (maxPlayers - zeroPos + pos) % maxPlayers
}

func upsertRobotData(id string, update interface{}) {
	Skeleton.Go(func() {
		db := MongoDB.Ref()
		defer MongoDB.UnRef(db)
		_, err := db.DB(DB).C("robot").UpsertId(id, update)
		if err != nil {
			log.Error("upsert %v error: %v", id, err)
		}
	}, nil)
}

type GDRoom struct {
	Room
	GameType          int
	Rule              *GDRule
	Useridplayerdatas map[int]*GDPlayerData // key: userid
	Tiles             []int                 // 洗好的牌
	Currentround      int                   // 第几局
	Dealeruserid      int                   // 庄家 userid
	Wildcard          int                   // 混儿，只有一张
	Jokers            []int                 // 宝牌
	Discards          []int                 // 玩家打的牌
	Rests             []int                 // 剩余的牌
	Draweruserid      int                   // 最近一次摸牌的人 userid
	Discarderuserid   int                   // 最近一次出牌的人 userid

	Actionwinusers  map[int]int // 可以胡的玩家 key: userid, value: 1
	Actionkongusers map[int]int // 可以杠的玩家 key: userid, value: 1
	Actionpongusers map[int]int // 可以碰的玩家 key: userid, value: 1
	Actionchowusers map[int]int // 可以吃的玩家 key: userid, value: 1

	Discardtimer *timer.Timer // 出牌定时器
	Claimtimer   *timer.Timer // 打牌后可以吃,碰,杠,胡得玩家操作
	Disbandtimer *timer.Timer // 房间解散定时器
	Claimuserid  int          // 当前谁可以操作

	Countwinstreak         int // 连庄次数
	Disbandapplicantuserid int // 申请解散房间者 userid
	Winneruserids          []int
}

// 玩家数据
type GDPlayerData struct {
	User            *User
	TotalResultData *TotalResultData
	RoundResultData *RoundResultData
	State           int
	Position        int // 用户在桌子上的位置，从 0 开始

	Owner           bool  // 房主
	Dealer          bool  // 庄家
	Draw            int   // 摸的一张牌
	Hands           []int //手牌
	HorseTile       []int // 马牌
	Discards        []int // 打出的牌
	Analyzer        *algorithm.GDAnalyzer
	ClaimActionCode int     //吃,碰,杠得状态码
	Claims          [][]int // 吃、碰、杠到的牌
	ActionTimestamp int64   // 记录操作时间戳
	Managed         bool    //是否托管
	DiscardsCount   int     // 记录自动出牌的次数

	Claim       int // 待吃、碰、杠的牌
	Quadruplet  []int
	Quadruplets [][]int //可以杠的牌型 一维表示长度 二维表示杠的牌
	Triplet     []int
	Sequence    []int
	Sequences   [][]int

	KongType    int
	WinType     int
	RoundResult *GDPlayerRoundResult
	TotalResult *GDPlayerTotalResult

	DisbandActionCode int
	WinTiles          []int
}

func initPlayer(playerData *GDPlayerData) {
	playerData.Draw = -1
	playerData.Hands = []int{}
	playerData.HorseTile = []int{}
	playerData.Discards = []int{}
	playerData.Claims = [][]int{}
	playerData.ActionTimestamp = 0
	playerData.DiscardsCount = 0
	playerData.Managed = false

	roundResult := playerData.RoundResult
	roundResult.WinType = 0
	roundResult.WinScore = 0
	roundResult.CatchHorseScore = 0
	roundResult.ExposedKongScore = 0
	roundResult.PongKongScore = 0
	roundResult.HiddenKongScore = 0
	roundResult.TotalScore = 0
}
func NewGDRoom(rule *GDRule) *GDRoom {
	gdRoom := new(GDRoom)
	gdRoom.State = RoomIdle
	gdRoom.LoginIPs = make(map[string]bool)
	gdRoom.PositionUserIDs = make(map[int]int)
	gdRoom.Useridplayerdatas = make(map[int]*GDPlayerData)
	gdRoom.Actionwinusers = make(map[int]int)
	gdRoom.Actionkongusers = make(map[int]int)
	gdRoom.Actionpongusers = make(map[int]int)
	gdRoom.Actionchowusers = make(map[int]int)

	gdRoom.Currentround = 1
	gdRoom.Rule = rule

	win := "点炮胡"
	if rule.MustSelfDraw {
		win = "自摸胡"
	}
	redDragonJoker := ""
	if rule.NeedJoker {
		redDragonJoker = "癞子"
	}
	switch gdRoom.Rule.RoomType {
	case RoomPrivate:
		gdRoom.Desc = fmt.Sprintf("%v %v %v局 %v人 底分%v分 %v匹马", win, redDragonJoker, rule.MaxRounds, rule.MaxPlayers, rule.BaseScore, rule.BuyHorse)
	case RoomRoomCardMatch:
		gdRoom.Desc = fmt.Sprintf("%v人 底注%v房卡 %v %v", rule.MaxPlayers, rule.RoomCards, redDragonJoker, win)
	case RoomRedPacketMatching, RoomRedPacketPrivate:
		gdRoom.Desc = fmt.Sprintf("%v人 %v元红包 %v %v", rule.MaxPlayers, rule.RedPacketType, redDragonJoker, win)
	}
	if gdRoom.Rule.IPAntiCheat {
		gdRoom.Desc += " IP防作弊"
	}
	if gdRoom.Rule.GPSAntiCheat {
		gdRoom.Desc += " GPS防作弊"
	}
	return gdRoom
}

func (r *GDRoom) Full() bool {
	return len(r.PositionUserIDs) == r.Rule.MaxPlayers
}
func (r *GDRoom) AllReady() bool {
	count := 0
	if r.Full() {
		for _, userID := range r.PositionUserIDs {
			playerData := r.Useridplayerdatas[userID]
			if playerData.State == GdReady {
				count++
			}
		}
		if count == r.Rule.MaxPlayers {
			return true
		}
		return false
	}
	return false
}

func (r *GDRoom) JoinNumber() int {
	return len(r.PositionUserIDs)
}
func (r *GDRoom) Empty() bool {
	return len(r.PositionUserIDs) == 0
}
func (r *GDRoom) Clean() {
	for _, uid := range r.PositionUserIDs {
		GetRoomMgr().DelPerson(uid)
	}
	r.PositionUserIDs = make(map[int]int)
	for _, playerData := range r.Useridplayerdatas {
		playerData.User.Location = []float64{}
	}
	r.Useridplayerdatas = make(map[int]*GDPlayerData)
	if r.Claimtimer != nil {
		r.Claimtimer.Stop()
		r.Claimtimer = nil
	}
	if r.Discardtimer != nil {
		r.Discardtimer.Stop()
		r.Discardtimer = nil
	}
}
func (r *GDRoom) AllAgree() bool {
	count := 0
	for _, userID := range r.PositionUserIDs {
		playerData := r.Useridplayerdatas[userID]
		if playerData.DisbandActionCode == ActionAgreeDisband {
			count++
		}
	}
	if count == r.Rule.MaxPlayers {
		return true
	}
	return false
}

func (r *GDRoom) Broadcast(msg interface{}, pos int) {

	for key := range r.Useridplayerdatas {
		if r.Useridplayerdatas[key].Position != pos && r.Useridplayerdatas[key].User.State == UserLogout {
			r.Useridplayerdatas[key].User.WriteMsg(msg)
		}
	}
}

func (r *GDRoom) BroadcastAll(msg interface{}) {
	for key := range r.Useridplayerdatas {
		if r.Useridplayerdatas[key].User.State != UserLogout {
			r.Useridplayerdatas[key].User.WriteMsg(msg)
		}
	}
}

func (r *GDRoom) ResetActionClaimUsers() {
	if r.Claimtimer != nil {
		r.Claimtimer.Stop()
		r.Claimtimer = nil
	}
	for _, userID := range r.PositionUserIDs {
		r.DeleteActionClaimUsers(userID)
	}
}

func (r *GDRoom) DeleteActionClaimUsers(userID int) {
	playerData := r.Useridplayerdatas[userID]
	if playerData.State == GdActionClaim {
		playerData.State = GdWaiting
		playerData.ClaimActionCode = 0
	}
	delete(r.Actionwinusers, userID)
	delete(r.Actionkongusers, userID)
	delete(r.Actionpongusers, userID)
	delete(r.Actionchowusers, userID)
}

func (r *GDRoom) ActionClaimUsersEmpty() bool {
	if len(r.Actionwinusers) == 0 && len(r.Actionkongusers) == 0 &&
		len(r.Actionpongusers) == 0 && len(r.Actionchowusers) == 0 {
		return true
	}
	return false
}
