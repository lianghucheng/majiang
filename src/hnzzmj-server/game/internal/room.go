package internal

import (
	"fmt"
	"github.com/name5566/leaf/log"
	"hnzzmj-server/common"
)

// 房间状态
const (
	roomIdle    = iota // 0 空闲
	roomGame           // 1 游戏中
	roomGameEnd        // 2 游戏结束
)

// 房间类型
const (
	roomPractice          = iota // 0 练习
	roomRoomCardMatch            // 1 房卡匹配
	roomPrivate                  // 2 私人
	roomRedPacketMatching        // 3 红包匹配
	roomRedPacketPrivate         // 4 红包私人
)

// 玩家解散房间动作码
const (
	_                    = iota
	actionWaitingDisband // 1 等待解散
	actionAgreeDisband   // 2 同意解散
)

var (
	roomNumbers = []int{}
	roomCounter = 0
)

type Room interface {
	Enter(user *User) bool
	Exit(user *User)
	SitDown(user *User, pos int)
	StandUp(user *User, pos int)
	GetAllPlayers(user *User)
	Disband(user *User)
	StartGame()
	EndGame()
}

type room struct {
	state                   int
	loginIPs                map[string]bool
	positionUserIDs         map[int]int // key: 座位号, value: userID
	creatorUserID           int         // 创建者 userID
	ownerUserID             int         // 房主 userID
	number                  string
	desc                    string
	startTimestamp          int64 // 开始时间
	eachRoundStartTimestamp int64 // 每一局开始时间
	endTimestamp            int64 // 结束时间
}

func init() {
	for i := 0; i < 1000000; i++ {
		roomNumbers = append(roomNumbers, i)
	}
	roomNumbers = common.Shuffle(roomNumbers)
}

func getRoomNumber() string {
	log.Debug("房间计数器: %v", roomCounter)
	roomNumber := fmt.Sprintf("%06d", roomNumbers[roomCounter])
	roomCounter = (roomCounter + 1) % 1000000
	return roomNumber
}

func broadcast(msg interface{}, positionUserIDs map[int]int, pos int) {
	for position, userID := range positionUserIDs {
		if position == pos {
			continue
		}
		if user, ok := userIDUsers[userID]; ok {
			if user.state != userLogout {
				user.WriteMsg(msg)
			}
		}
	}
}

func broadcastAll(msg interface{}) {
	for _, user := range userIDUsers {
		if user.state != userLogout {
			user.WriteMsg(msg)
		}
	}
}

func toRelativePosition(pos int, zeroPos int, maxPlayers int) int {
	return (maxPlayers - zeroPos + pos) % maxPlayers
}

func toRealPosition(relativePos int, zeroPos int, maxPlayers int) int {
	return (maxPlayers + zeroPos + relativePos) % maxPlayers
}

func upsertRobotData(id string, update interface{}) {
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		_, err := db.DB(DB).C("robot").UpsertId(id, update)
		if err != nil {
			log.Error("upsert %v error: %v", id, err)
		}
	}, nil)
}
