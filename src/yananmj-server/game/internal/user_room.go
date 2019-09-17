package internal

import (
	"github.com/name5566/leaf/log"
	"yananmj-server/common"
	"yananmj-server/game/mahjong"
	"yananmj-server/msg"
)

func (user *User) checkCreateRoomCards(roomCards int) bool {
	if roomCards > user.data.userData.RoomCards {
		user.WriteMsg(&data_struct.S2C_CreateRoom{
			Error:     data_struct.S2C_CreateRoom_LackOfRoomCards,
			RoomCards: roomCards,
		})
		return false
	}
	return true
}

func (user *User) checkEnterRoomCards(roomCards int) bool {
	if roomCards > user.data.userData.RoomCards {
		user.WriteMsg(&data_struct.S2C_EnterRoom{
			Error:     data_struct.S2C_EnterRoom_LackOfRoomCards,
			RoomCards: roomCards,
		})
		return false
	}
	return true
}

//创建房间
func (user *User) createYananRoom(yananRule *mahjong.YananRule) {
	roomNumber := ""
	switch yananRule.RoomType {
	case roomPrivate, roomRedPacketPrivate:
		roomNumber = getOneRoomNumber()
		if _, ok := roomNumberRooms[roomNumber]; ok {
			user.WriteMsg(&data_struct.S2C_CreateRoom{
				Error: data_struct.S2C_CreateRoom_InnerError,
			})
			return
		}
	}

	yananRoom := newYananRoom(yananRule)
	switch yananRule.RoomType {
	case roomPractice:
		log.Debug("userID: %v 创建Yanan练习房", user.data.userData.UserID)
		yananPracticeRooms[user.data.userData.UserID] = yananRoom
	case roomRoomCardMatch, roomRedPacketMatching:
		log.Debug("userID: %v 创建Yanan房卡匹配房", user.data.userData.UserID)
		yananRoomCardMatchRooms[user.data.userData.UserID] = yananRoom
	case roomPrivate, roomRedPacketPrivate:
		log.Debug("userID: %v 创建Yanan私人房间成功,红中癞子: %v, 局数: %v, 人数: %v ,底分: %v 是否下炮子: %v 是否带字牌: %v", user.data.userData.UserID, yananRule.RedDragonJoker, yananRule.MaxRounds, yananRule.MaxPlayers, yananRule.BaseScore, yananRule.Gun, yananRule.WithHonors)
		yananRoom.number = roomNumber
		yananRoom.ownerUserID = user.data.userData.UserID
		roomNumberRooms[roomNumber] = yananRoom
	}
	yananRoom.creatorUserID = user.data.userData.UserID
	user.enterRoom(yananRoom)
}

func (user *User) createPrivateRoom(info *data_struct.C2S_CreateYananRoom) {
	if common.Index([]int{4, 8, 16}, info.MaxRounds) == -1 || common.Index([]int{2, 3, 4}, info.MaxPlayers) == -1 {
		user.WriteMsg(&data_struct.S2C_CreateRoom{
			Error: data_struct.S2C_CreateRoom_RuleError,
		})
		return
	}

	roomCards := 2
	if info.MaxRounds == 8 {
		roomCards = 3
	} else if info.MaxRounds == 16 {
		roomCards = 5
	}
	if roomCards > user.data.userData.RoomCards {
		log.Debug("userID: %v 房卡不足", user.data.userData.UserID)
		user.WriteMsg(&data_struct.S2C_CreateRoom{
			Error:     data_struct.S2C_CreateRoom_LackOfRoomCards,
			RoomCards: roomCards,
		})
		return
	}

	yananRule := &mahjong.YananRule{
		RoomType:       roomPrivate,
		MaxRounds:      info.MaxRounds,
		MaxPlayers:     info.MaxPlayers,
		BaseScore:      1,
		RedDragonJoker: info.RedDragonJoker,
		MustSelfDraw:   info.MustSelfDraw,
		RoomCards:      roomCards,
		WithHonors:     info.WithHonors,
		Gun:            info.Gun,
		IPAntiCheat:    info.IPAntiCheat,
	}
	if info.GPSAntiCheat {
		if common.CheckLocation(info.Location) {
			user.Location = info.Location
			yananRule.GPSAntiCheat = true
		} else {
			user.WriteMsg(&data_struct.S2C_EnterRoom{
				Error: data_struct.S2C_CreateRoom_LocationError,
			})
			return
		}
	}
	user.createYananRoom(yananRule)
}

func (user *User) createRedPacketPrivateRoom(redPacketType int) {
	roomCards := 0
	switch redPacketType {
	case 100:
		roomCards = 128
	case 999:
		roomCards = 1198
	default:
		user.WriteMsg(&data_struct.S2C_CreateRoom{Error: data_struct.S2C_CreateRoom_RuleError})
		return
	}
	if !user.checkCreateRoomCards(roomCards) {
		return
	}
	rule := &mahjong.YananRule{
		RoomType:       roomRedPacketPrivate,
		MaxPlayers:     4,
		MaxRounds:      1,
		MustSelfDraw:   true,
		RedDragonJoker: true,
		RoomCards:      roomCards,
		RedPacketType:  redPacketType,
		IPAntiCheat:    true,
		GPSAntiCheat:   false,
	}
	user.createYananRoom(rule)
}

func (user *User) createOrEnterPracticeRoom() {
	for _, r := range yananPracticeRooms {
		yananRoom := r.(*YananRoom)
		if !yananRoom.full() {
			user.enterRoom(r)
			return
		}
	}
	yananRule := &mahjong.YananRule{
		RoomType:       roomPractice,
		MaxRounds:      1,
		MaxPlayers:     4,
		BaseScore:      1,
		RedDragonJoker: false,
		WithHonors:     true,
		Gun:            false,
	}
	user.createYananRoom(yananRule)
}

func (user *User) createOrEnterRedPacketMatchingRoom(redPacketType int) {
	roomCards := 0
	switch redPacketType {
	case 1:
		roomCards = 2
	case 10:
		roomCards = 15
	default:
		user.WriteMsg(&data_struct.S2C_CreateRoom{Error: data_struct.S2C_CreateRoom_RuleError})
		return
	}
	if !checkRedPacketMatchingTime() {
		user.WriteMsg(&data_struct.S2C_EnterRoom{Error: data_struct.S2C_EnterRoom_NotRightNow})
		return
	}
	if !user.checkCreateRoomCards(roomCards) {
		return
	}
	rule := &mahjong.YananRule{
		RoomType:       roomRedPacketMatching,
		MaxPlayers:     4,
		MaxRounds:      1,
		MustSelfDraw:   true,
		RedDragonJoker: true,
		RoomCards:      roomCards,
		RedPacketType:  redPacketType,
		IPAntiCheat:    true,
		GPSAntiCheat:   false,
	}
	if user.isRobot() {
		if user.enterRedPacketMatchingRoom(redPacketType, 2) {
			return
		}
		if user.enterRedPacketMatchingRoom(redPacketType, 1) {
			return
		}
	} else {
		if user.enterRedPacketMatchingRoom(redPacketType, 3) {
			return
		}
		if user.enterRedPacketMatchingRoom(redPacketType, 2) {
			return
		}
		if user.enterRedPacketMatchingRoom(redPacketType, 1) {
			return
		}
	}
	user.createYananRoom(rule)
}

func (user *User) enterRedPacketMatchingRoom(redPacketType int, playerNumber int) bool {
	for _, r := range yananRoomCardMatchRooms {
		yananRoom := r.(*YananRoom)
		if yananRoom.rule.IPAntiCheat {
			if !yananRoom.loginIPs[user.data.userData.LoginIP] && yananRoom.rule.RedPacketType == redPacketType && len(yananRoom.positionUserIDs) == playerNumber {
				user.enterRoom(r)
				return true
			}
		} else {
			if yananRoom.rule.RedPacketType == redPacketType && len(yananRoom.positionUserIDs) == playerNumber {
				user.enterRoom(r)
				return true
			}
		}
	}
	return false
}

func (user *User) createOrEnterRoomCardMatchRoom(roomCards int) {
	if common.Index([]int{1, 10, 50, 100}, roomCards) == -1 {
		user.WriteMsg(&data_struct.S2C_CreateRoom{
			Error:     data_struct.S2C_CreateRoom_RuleError,
			RoomCards: roomCards,
		})
		return
	}
	if user.data.userData.RoomCards < roomCards {
		user.WriteMsg(&data_struct.S2C_CreateRoom{
			Error:     data_struct.S2C_CreateRoom_LackOfRoomCards,
			RoomCards: roomCards,
		})
		return
	}
	ipAntiCheat := true
	if user.data.userData.Role == roleRoot {
		ipAntiCheat = false
	}
	yananRule := &mahjong.YananRule{
		RoomType:       roomRoomCardMatch,
		MaxRounds:      1,
		MaxPlayers:     4,
		BaseScore:      1,
		RedDragonJoker: true,
		MustSelfDraw:   true,
		WithHonors:     false,
		Gun:            false,
		RoomCards:      roomCards,
		IPAntiCheat:    ipAntiCheat,
		GPSAntiCheat:   false,
	}
	if user.isRobot() {
		if user.enterRoomCardMatchingRoom(roomCards, 2) {
			return
		}
		if user.enterRoomCardMatchingRoom(roomCards, 1) {
			return
		}
	} else {
		if user.enterRoomCardMatchingRoom(roomCards, 3) {
			return
		}
		if user.enterRoomCardMatchingRoom(roomCards, 2) {
			return
		}
		if user.enterRoomCardMatchingRoom(roomCards, 1) {
			return
		}
	}
	user.createYananRoom(yananRule)
}

func (user *User) enterRoomCardMatchingRoom(roomCards int, playerNumber int) bool {
	for _, r := range yananRoomCardMatchRooms {
		yananRoom := r.(*YananRoom)
		if yananRoom.rule.IPAntiCheat {
			if !yananRoom.loginIPs[user.data.userData.LoginIP] && yananRoom.rule.RoomCards == roomCards && len(yananRoom.positionUserIDs) == playerNumber {
				user.enterRoom(r)
				return true
			}
		} else {
			if yananRoom.rule.RoomCards == roomCards && len(yananRoom.positionUserIDs) == playerNumber {
				user.enterRoom(r)
				return true
			}
		}
	}
	return false
}

//进入房间
func (user *User) enterRoom(r interface{}) {
	var sitDown = false

	yananRoom := r.(*YananRoom)
	sitDown = yananRoom.Enter(user)

	if sitDown {
		userIDRooms[user.data.userData.UserID] = r
	}
}

func (user *User) enterGPSRoom(r interface{}, gps bool, location []float64) {
	if gps {
		if common.CheckLocation(location) {
			user.Location = location
		} else {
			user.WriteMsg(&data_struct.S2C_EnterRoom{
				Error: data_struct.S2C_EnterRoom_LocationError,
			})
			return
		}
	} else {
		user.WriteMsg(&data_struct.S2C_EnterRoom{
			Error: data_struct.S2C_EnterRoom_GPSNotOpen,
		})
		return
	}
	user.enterRoom(r)
}

//玩家解散或退出房间
func (user *User) exitOrDisbandRoom(r interface{}, forcible bool) {
	if user.isRobot() {
		forcible = true
	}
	yananRoom := r.(*YananRoom)
	if yananRoom.state == roomIdle {
		if forcible {
			if yananRoom.ownerUserID == user.data.userData.UserID {
				yananRoom.Disband(user)
			} else {
				yananRoom.Exit(user)
			}
		}
		return
	}
	switch yananRoom.rule.RoomType {
	case roomRoomCardMatch:
		user.WriteMsg(&data_struct.S2C_ExitRoom{
			Error: data_struct.S2C_ExitRoom_GamePlaying,
		})
	case roomPrivate:
		if forcible {
			yananRoom.Disband(user)
		}
	}
}

//同意解散房间
func (user *User) agreeDisbandRoom(r interface{}) {
	yananRoom := r.(*YananRoom)
	yananRoom.agreeDisbandRoom(user.data.userData.UserID)
}

//拒绝解散房间
func (user *User) refusedDisbandRoom(r interface{}) {
	yananRoom := r.(*YananRoom)
	yananRoom.refusedDisbandRoom(user.data.userData.UserID)
}
