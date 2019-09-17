package internal

import (
	"encoding/json"
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/timer"
	"gopkg.in/mgo.v2/bson"
	"hnzzmj-server/common"
	"hnzzmj-server/msg"
	"math/rand"
	"reflect"
	"strconv"
	"time"
)

// 用户状态
const (
	userLogin  = iota
	userLogout // 1
)

const (
	roleRobot  = -2
	roleBlack  = -1
	rolePlayer = 1
	roleAgent  = 2
	roleAdmin  = 3
	roleRoot   = 4
)

var (
	userIDUsers = make(map[int]*User)

	userIDRooms     = make(map[int]interface{})
	roomNumberRooms = make(map[string]interface{})

	hnzzPracticeRooms      = make(map[int]interface{}) // key: userID
	hnzzRoomCardMatchRooms = make(map[int]interface{}) // 房卡匹配场（前端叫房卡比赛场）

	systemOn = true // 系统开关

	accountIDs         = []int{}
	accountIDCounter   = 0
	reservedAccountIDs = []int{6666666, 8888888, 9999999}

	roomCardMatchOnlineNumber  = []int{0, 0, 0, 0} // 房卡比赛场在线人数
	redPacketMatchOnlineNumber = []int{0, 0, 0, 0} // 红包比赛在线人数
)

type User struct {
	gate.Agent
	state          int
	data           *BaseData
	heartbeatTimer *timer.Timer
	heartbeatStop  bool
	location       []float64 // 定位
	robot          bool
}

type BaseData struct {
	userData    *UserData
	ownerUserID int // 所在房间的房主
}

func init() {
	rand.Seed(time.Now().UnixNano())

	result := new(UserData)

	m := make(map[int]bool)
	for _, v := range reservedAccountIDs {
		m[v] = true
	}

	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)

		iter := db.DB(DB).C("users").Find(nil).Iter()
		for iter.Next(&result) {
			m[result.AccountID] = true
		}
		if err := iter.Close(); err != nil {
			log.Error("iter close error: %v", err)
		}
	}, func() {
		for i := 1000000; i < 10000000; i++ {
			if !m[i] {
				accountIDs = append(accountIDs, i)
			}
		}
		accountIDs = common.Shuffle(accountIDs)
		// log.Debug("%v %v", len(m), len(accountIDs))
	})

	cronExpr, _ := timer.NewCronExpr("10 0 0 * * *")
	skeleton.CronFunc(cronExpr, func() {
		roomCardMatchOnlineNumber[0] = 0
		roomCardMatchOnlineNumber[1] = 0
		roomCardMatchOnlineNumber[2] = 0
		roomCardMatchOnlineNumber[3] = 0
	})
}

// 生成7位数的账号ID
func getAccountID() int {
	log.Debug("账号ID计数器: %v", accountIDCounter)
	accountID := accountIDs[accountIDCounter]
	accountIDCounter++
	return accountID
}

func newUser(a gate.Agent) *User {
	user := new(User)
	user.Agent = a
	user.state = userLogin
	user.data = new(BaseData)
	user.data.userData = new(UserData)
	user.location = []float64{}
	return user
}

// 检测红包比赛是否到开赛时间
func checkRedPacketMatchingTime() bool {
	nowTime := time.Now()
	noon := time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day(), 12, 0, 0, 0, time.Local)
	_13oClock := time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day(), 13, 0, 0, 0, time.Local)
	_20oClock := time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day(), 20, 0, 0, 0, time.Local)
	_22oClock := time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day(), 22, 0, 0, 0, time.Local)
	// 小于12点、大于13点且小于20点、大于22点
	if nowTime.Unix() < noon.Unix() || nowTime.Unix() > _13oClock.Unix() && nowTime.Unix() < _20oClock.Unix() || nowTime.Unix() > _22oClock.Unix() {
		log.Debug("未到开赛时间, 当前时间: %v", nowTime.Format("2006/01/02 15:04:05"))
		return false
	}
	return true
}

// 计算红包比赛在线人数
func calculateRedPacketMatchOnlineNumber(redPacketType int) {
	switch redPacketType {
	case 1: // 红包匹配场
		redPacketMatchOnlineNumber[0]++
	case 10: // 红包匹配场
		redPacketMatchOnlineNumber[1]++
	case 100: // 红包私人房
		redPacketMatchOnlineNumber[2]++
	case 999: // 红包私人房
		redPacketMatchOnlineNumber[3]++
	}
}

// 计算房卡比赛在线人数
func calculateRoomCardMatchOnlineNumber(roomCards int, exit bool) {
	switch roomCards {
	case 1:
		if exit {
			roomCardMatchOnlineNumber[0]--
		} else {
			roomCardMatchOnlineNumber[0]++
		}
	case 10:
		if exit {
			roomCardMatchOnlineNumber[1]--
		} else {
			roomCardMatchOnlineNumber[1]++
		}
	case 50:
		if exit {
			roomCardMatchOnlineNumber[2]--
		} else {
			roomCardMatchOnlineNumber[2]++
		}
	case 100:
		if exit {
			roomCardMatchOnlineNumber[3]--
		} else {
			roomCardMatchOnlineNumber[3]++
		}
	}
}

func toRoleString(role int) string {
	switch role {
	case roleRoot:
		return "超管"
	case roleAdmin:
		return "管理员"
	case roleAgent:
		return "代理"
	case rolePlayer:
		return "玩家"
	//case roleRobot:
	//	return "机器人"
	case roleBlack:
		return "拉黑"
	}
	return ""
}

func (user *User) autoHeartbeat() {
	if user.heartbeatStop {
		log.Debug("userID: %v 心跳停止", user.data.userData.UserID)
		user.Close()
		return
	}
	user.heartbeatStop = true
	user.WriteMsg(&msg.S2C_Heartbeat{})
	// 服务端发送心跳包间隔120秒
	user.heartbeatTimer = skeleton.AfterFunc(120*time.Second, func() {
		user.autoHeartbeat()
	})
}

func (user *User) getAllPlayers(r interface{}) {
	switch r.(type) {
	case *HNZZRoom:
		hnzzRoom := r.(*HNZZRoom)
		hnzzRoom.GetAllPlayers(user)
	}
}

func (user *User) completeDailyShare() {
	ok := false
	nowTime := time.Now()
	if user.data.userData.CompleteDailyShareAt == 0 {
		ok = true
	} else {
		completeTime := time.Unix(user.data.userData.CompleteDailyShareAt, 0)
		completeMidnightTime := time.Date(completeTime.Year(), completeTime.Month(), completeTime.Day(), 0, 0, 0, 0, time.Local)
		nowMidnightTime := time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day(), 0, 0, 0, 0, time.Local)
		if nowMidnightTime.Unix() > completeMidnightTime.Unix() {
			ok = true
		}
	}
	if ok {
		user.data.userData.CompleteDailyShareAt = nowTime.Unix()
		rndRoomCards := rand.Intn(2) + 1
		if user.data.userData.RoomCards < 30 {
			rndRoomCards = rand.Intn(5) + 1
		}
		user.data.userData.RoomCards += rndRoomCards
		user.WriteMsg(&msg.S2C_UpdateRoomCards{
			RoomCards: user.data.userData.RoomCards,
		})
		user.WriteMsg(&msg.S2C_CompleteDailyShare{
			RoomCards: rndRoomCards,
		})

		user.data.userData.GiftRoomCards += rndRoomCards
		user.saveShareRoomcardData(user.data.userData.CompleteDailyShareAt, rndRoomCards)
	} else {
		user.WriteMsg(&msg.S2C_CompleteDailyShare{
			RoomCards: 0,
		})
	}
}

func (user *User) sendTextMessage(r interface{}, message string) {
	switch r.(type) {
	case *HNZZRoom:
		hnzzRoom := r.(*HNZZRoom)
		playerData := hnzzRoom.userIDPlayerDatas[user.data.userData.UserID]
		broadcast(&msg.S2C_TextMessage{
			Position: playerData.position,
			Message:  message,
		}, hnzzRoom.positionUserIDs, -1)
	}
}

func (user *User) sendExpressionMessage(r interface{}, expression int) {
	switch r.(type) {
	case *HNZZRoom:
		hnzzRoom := r.(*HNZZRoom)
		playerData := hnzzRoom.userIDPlayerDatas[user.data.userData.UserID]
		broadcast(&msg.S2C_ExpressionMessage{
			Position:   playerData.position,
			Expression: expression,
		}, hnzzRoom.positionUserIDs, -1)
	}
}

func (user *User) sendGCloudVoiceMessage(r interface{}, fileID string) {
	switch r.(type) {
	case *HNZZRoom:
		hnzzRoom := r.(*HNZZRoom)
		playerData := hnzzRoom.userIDPlayerDatas[user.data.userData.UserID]
		broadcast(&msg.S2C_GCloudVoiceMessage{
			Position: playerData.position,
			FileID:   fileID,
		}, hnzzRoom.positionUserIDs, -1)
	}
}

func (user *User) verifyTestEnvironmentReceipt(receiptData string) {
	body, err := common.HttpPost("https://sandbox.itunes.apple.com/verifyReceipt", receiptData)
	if err == nil {
		m := map[string]interface{}{}
		json.Unmarshal(body, &m)
		if status, ok := m["status"].(float64); ok {
			log.Debug("test status: %v", status)
			switch int(status) {
			case 0:
				receipt := m["receipt"].(map[string]interface{})
				inApp := receipt["in_app"].([]interface{})
				product := inApp[0].(map[string]interface{})
				productID := product["product_id"]
				log.Debug("product_id: %v", productID)
			}
		}
	} else {
		log.Debug("userID %v 请求测试环境URL出错 %v", user.data.userData.UserID, err)
	}
}

func (user *User) verifyProductionEnvironmentReceipt(receiptData string) {
	body, err := common.HttpPost("https://buy.itunes.apple.com/verifyReceipt", receiptData)
	if err == nil {
		m := map[string]interface{}{}
		json.Unmarshal(body, &m)
		if status, ok := m["status"].(float64); ok {
			log.Debug("production status: %v, %v", status, reflect.TypeOf(status))
			switch int(status) {
			case 0:

			case 21007:
				user.verifyTestEnvironmentReceipt(receiptData)
			}
		}
	} else {
		log.Debug("userID %v 请求生产环境URL出错 %v", user.data.userData.UserID, err)
	}
}

func (user *User) isRobot() bool {
	return user.data.userData.Role == roleRobot
}

func (user *User) sendRedPacketMatchOnlineNumber() {
	if !checkRedPacketMatchingTime() {
		redPacketMatchOnlineNumber[0] = 0
		redPacketMatchOnlineNumber[1] = 0
	}
	user.WriteMsg(&msg.S2C_UpdateRedPacketMatchOnlineNumber{
		Numbers: redPacketMatchOnlineNumber,
	})
}

func (user *User) sendUntakenRedPacketMatchPrizeNumber() {
	count := 0
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		count, _ = db.DB(DB).C("redpacketmatchresult").
			Find(bson.M{"userid": user.data.userData.UserID, "redpacket": bson.M{"$gt": 0}, "taken": false}).Count()
	}, func() {
		user.WriteMsg(&msg.S2C_UpdateUntakenRedPacketMatchPrizeNumber{
			Number: count,
		})
	})
}

func (user *User) sendRedPacketMatchRecord(pageNumber int, pageSize int) {
	resultData := new(RedPacketMatchResultData)
	var items []msg.RedPacketMatchRecordItem
	count := 0
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		count, _ = db.DB(DB).C("redpacketmatchresult").Find(bson.M{"userid": user.data.userData.UserID}).Count()

		iter := db.DB(DB).C("redpacketmatchresult").Find(bson.M{"userid": user.data.userData.UserID}).
			Sort("-createdat").Skip((pageNumber - 1) * pageSize).Limit(pageSize).Iter()
		if err := iter.Close(); err != nil {
			log.Error("iter close error: %v", err)
		}
		for iter.Next(&resultData) {
			items = append(items, msg.RedPacketMatchRecordItem{
				ID:            resultData.ID,
				RedPacketType: resultData.RedPacketType,
				RedPacket:     resultData.RedPacket,
				Taken:         resultData.Taken,
				Date:          time.Unix(resultData.CreatedAt, 0).Format("2006/01/02 15:04:05"),
			})
		}
	}, func() {
		user.WriteMsg(&msg.S2C_RedPacketMatchRecord{
			Items:      items,
			Total:      count,
			PageNumber: pageNumber,
			PageSize:   pageSize,
		})
	})
}

func (user *User) takeRedPacketMatchPrize(id bson.ObjectId) {
	if user.data.userData.CircleID < 1 {
		user.WriteMsg(&msg.S2C_TakeRedPacketMatchPrize{
			Error: msg.S2C_TakeRedPacketMatchPrize_CircleIDInvalid,
		})
		user.requestCircleID()
		return
	}
	userID := user.data.userData.UserID

	resultData := new(RedPacketMatchResultData)
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)

		err := db.DB(DB).C("redpacketmatchresult").FindId(id).One(&resultData)
		if err != nil {
			resultData = nil
			log.Debug("find redpacketmatchresult: %v error: %v", id, err)
			return
		}
	}, func() {
		if resultData == nil {
			user.WriteMsg(&msg.S2C_TakeRedPacketMatchPrize{
				Error: msg.S2C_TakeRedPacketMatchPrize_IDInvalid,
				ID:    id,
			})
			return
		}
		if resultData.UserID != user.data.userData.UserID || resultData.RedPacket <= 0 {
			user.WriteMsg(&msg.S2C_TakeRedPacketMatchPrize{
				Error: msg.S2C_TakeRedPacketMatchPrize_NotYetWon,
				ID:    id,
			})
			return
		}
		if resultData.Taken {
			user.WriteMsg(&msg.S2C_TakeRedPacketMatchPrize{
				Error:     msg.S2C_TakeRedPacketMatchPrize_TakeRepeated,
				ID:        id,
				RedPacket: resultData.RedPacket,
			})
			return
		}
		if resultData.Handling {
			return
		}
		updateRedPacketMatchResultData(id, bson.M{"$set": bson.M{"handling": true}}, func() {
			// 请求生成一个圈圈红包
			desc := strconv.Itoa(resultData.RedPacketType) + "元红包比赛奖励"
			user.requestCircleRedPacket(resultData.RedPacket, desc, func() {
				takeRedPacketMatchPrizeSuccess(userID, id, resultData.RedPacket)
			}, func() {
				takeRedPacketMatchPrizeFail(userID, id)
			})
		})
	})
}

func takeRedPacketMatchPrizeSuccess(userID int, id bson.ObjectId, redPacket float64) {
	var cb func()
	if user, ok := userIDUsers[userID]; ok {
		user.WriteMsg(&msg.S2C_TakeRedPacketMatchPrize{
			Error:     msg.S2C_TakeRedPacketMatchPrize_OK,
			ID:        id,
			RedPacket: redPacket,
		})
		cb = func() {
			if theUser, ok := userIDUsers[userID]; ok {
				theUser.sendUntakenRedPacketMatchPrizeNumber()
			}
		}
	} else {
		cb = nil
	}
	updateRedPacketMatchResultData(id, bson.M{"$set": bson.M{"taken": true, "handling": false, "updatedat": time.Now().Unix()}}, cb)
}

func takeRedPacketMatchPrizeFail(userID int, id bson.ObjectId) {
	if user, ok := userIDUsers[userID]; ok {
		user.WriteMsg(&msg.S2C_TakeRedPacketMatchPrize{
			Error: msg.S2C_TakeRedPacketMatchPrize_Error,
			ID:    id,
		})
	}
	updateRedPacketMatchResultData(id, bson.M{"$set": bson.M{"handling": false}}, nil)
}

func (user *User) FakeWXPay(totalFee int) {
	if common.InArray([]int{1000, 3000, 5000}, totalFee) || user.isRobot() {
		outTradeNo := common.GetOutTradeNo()
		startWXPayOrder(outTradeNo, user.data.userData.AccountID, totalFee, func() {
			finishWXPayOrder(outTradeNo, totalFee, false)
		})
	}
}

func (user *User) setRobotRoomCards(roomCards int) {
	robotData := new(UserData)
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)

		iter := db.DB(DB).C("users").Find(bson.M{"role": roleRobot}).Iter()
		if err := iter.Close(); err != nil {
			log.Error("iter close error: %v", err)
		}
		for iter.Next(&robotData) {
			if robot, ok := userIDUsers[robotData.UserID]; ok {
				robot.data.userData.RoomCards += roomCards
			} else {
				updateUserData(robotData.UserID, bson.M{"$inc": bson.M{"roomcards": roomCards}})
			}
		}
	}, nil)
}
