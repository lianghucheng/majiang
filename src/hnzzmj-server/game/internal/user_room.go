package internal

import (
	"github.com/name5566/leaf/log"
	"hnzzmj-server/common"
	"hnzzmj-server/game/mahjong"
	"hnzzmj-server/msg"
)

func (user *User) checkCreateRoomCards(roomCards int) bool {
	if roomCards > user.data.userData.RoomCards {
		user.WriteMsg(&msg.S2C_CreateRoom{
			Error:     msg.S2C_CreateRoom_LackOfRoomCards,
			RoomCards: roomCards,
		})
		return false
	}
	return true
}

func (user *User) checkEnterRoomCards(roomCards int) bool {
	if roomCards > user.data.userData.RoomCards {
		user.WriteMsg(&msg.S2C_EnterRoom{
			Error:     msg.S2C_EnterRoom_LackOfRoomCards,
			RoomCards: roomCards,
		})
		return false
	}
	return true
}

func (user *User) createHNZZRoom(hnzzRule *mahjong.HNZZRule) {
	roomNumber := ""
	switch hnzzRule.RoomType {
	case roomPrivate, roomRedPacketPrivate:
		roomNumber = getRoomNumber()
		if _, ok := roomNumberRooms[roomNumber]; ok {
			user.WriteMsg(&msg.S2C_CreateRoom{
				Error: msg.S2C_CreateRoom_InnerError,
			})
			return
		}
	}
	hnzzRoom := newHNZZRoom(hnzzRule)
	switch hnzzRule.RoomType {
	case roomRoomCardMatch, roomRedPacketMatching:
		log.Debug("userID: %v 创建HNZ匹配房", user.data.userData.UserID)
		hnzzRoomCardMatchRooms[user.data.userData.UserID] = hnzzRoom
	case roomPrivate, roomRedPacketPrivate:
		log.Debug("userID: %v 创建HNZZ私人房 房号: %v, 局数: %v, 人数: %v, 底分: %v", user.data.userData.UserID, roomNumber, hnzzRule.MaxRounds, hnzzRule.MaxPlayers, hnzzRule.BaseScore)
		hnzzRoom.number = roomNumber
		hnzzRoom.ownerUserID = user.data.userData.UserID
		roomNumberRooms[roomNumber] = hnzzRoom
	}
	hnzzRoom.creatorUserID = user.data.userData.UserID
	user.enterRoom(hnzzRoom)
}

func (user *User) createPrivateRoom(info *msg.C2S_CreateHNZZRoom) {
	if common.Index([]int{4, 8, 16}, info.MaxRounds) == -1 || common.Index([]int{2, 3, 4}, info.MaxPlayers) == -1 || common.Index([]int{2, 4, 6}, info.Birds) == -1 {
		user.WriteMsg(&msg.S2C_CreateRoom{
			Error: msg.S2C_CreateRoom_RuleError,
		})
		return
	}
	log.Debug("info: %v", info)
	needRoomCards := 2
	if info.MaxRounds == 8 {
		needRoomCards = 3
	} else if info.MaxRounds == 16 {
		needRoomCards = 5
	}
	if needRoomCards > user.data.userData.RoomCards {
		user.WriteMsg(&msg.S2C_CreateRoom{
			Error:     msg.S2C_CreateRoom_LackOfRoomCards,
			RoomCards: needRoomCards,
		})
		return
	}
	hnzzRule := &mahjong.HNZZRule{
		RoomType:          roomPrivate,
		MaxRounds:         info.MaxRounds,
		MaxPlayers:        info.MaxPlayers,
		MustSelfDraw:      info.MustSelfDraw,
		BaseScore:         1,
		DistinguishDealer: true,
		Birds:             info.Birds,
		RoomCards:         needRoomCards,
		IPAntiCheat:       info.IPAntiCheat,
	}
	if info.GPSAntiCheat {
		if common.CheckLocation(info.Location) {
			user.location = info.Location
			hnzzRule.GPSAntiCheat = true
		} else {
			user.WriteMsg(&msg.S2C_EnterRoom{
				Error: msg.S2C_CreateRoom_LocationError,
			})
			return
		}
	}
	user.createHNZZRoom(hnzzRule)
}

func (user *User) createRedPacketPrivateRoom(redPacketType int) {
	roomCards := 0
	switch redPacketType {
	case 100:
		roomCards = 128
	case 999:
		roomCards = 1198
	default:
		user.WriteMsg(&msg.S2C_CreateRoom{Error: msg.S2C_CreateRoom_RuleError})
		return
	}
	if !user.checkCreateRoomCards(roomCards) {
		return
	}
	rule := &mahjong.HNZZRule{
		RoomType:      roomRedPacketPrivate,
		MaxPlayers:    4,
		MaxRounds:     1,
		MustSelfDraw:  true,
		RoomCards:     roomCards,
		RedPacketType: redPacketType,
		IPAntiCheat:   true,
		GPSAntiCheat:  false,
	}
	user.createHNZZRoom(rule)
}

func (user *User) createOrEnterPracticeRoom() {
	for _, r := range hnzzPracticeRooms {
		hnzzRoom := r.(*HNZZRoom)
		if !hnzzRoom.full() {
			user.enterRoom(r)
			return
		}
	}
	hnzzRule := &mahjong.HNZZRule{
		RoomType:   roomPractice,
		MaxRounds:  1,
		MaxPlayers: 4,
		BaseScore:  1,
		Birds:      4,
	}
	user.createHNZZRoom(hnzzRule)
}

func (user *User) createOrEnterRoomCardMatchRoom(roomCards int) {
	if common.Index([]int{1, 10, 50, 100}, roomCards) == -1 {
		user.WriteMsg(&msg.S2C_CreateRoom{
			Error:     msg.S2C_CreateRoom_RuleError,
			RoomCards: roomCards,
		})
		return
	}
	if user.data.userData.RoomCards < roomCards {
		user.WriteMsg(&msg.S2C_CreateRoom{
			Error:     msg.S2C_CreateRoom_LackOfRoomCards,
			RoomCards: roomCards,
		})
		return
	}

	ipAntiCheat := true
	if user.data.userData.Role == roleRoot {
		ipAntiCheat = false
	}
	hnzzRule := &mahjong.HNZZRule{
		RoomType:     roomRoomCardMatch,
		MaxRounds:    1,
		MaxPlayers:   4,
		MustSelfDraw: true,
		BaseScore:    1,
		Birds:        4,
		RoomCards:    roomCards,
		IPAntiCheat:  ipAntiCheat,
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
	user.createHNZZRoom(hnzzRule)
}

func (user *User) enterRoomCardMatchingRoom(roomCards int, playerNumber int) bool {
	for _, r := range hnzzRoomCardMatchRooms {
		hnzzRoom := r.(*HNZZRoom)
		if hnzzRoom.rule.IPAntiCheat {
			if !hnzzRoom.loginIPs[user.data.userData.LoginIP] && hnzzRoom.rule.RoomCards == roomCards && len(hnzzRoom.positionUserIDs) == playerNumber {
				user.enterRoom(r)
				return true
			}
		} else {
			if hnzzRoom.rule.RoomCards == roomCards && len(hnzzRoom.positionUserIDs) == playerNumber {
				user.enterRoom(r)
				return true
			}
		}
	}
	return false
}

func (user *User) createOrEnterRedPacketMatchingRoom(redPacketType int) {
	roomCards := 0
	switch redPacketType {
	case 1:
		roomCards = 4
	case 10:
		roomCards = 28
	default:
		user.WriteMsg(&msg.S2C_CreateRoom{Error: msg.S2C_CreateRoom_RuleError})
		return
	}
	if !checkRedPacketMatchingTime() {
		user.WriteMsg(&msg.S2C_EnterRoom{Error: msg.S2C_EnterRoom_NotRightNow})
		return
	}
	if !user.checkCreateRoomCards(roomCards) {
		return
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

	rule := &mahjong.HNZZRule{
		RoomType:      roomRedPacketMatching,
		MaxPlayers:    4,
		MaxRounds:     1,
		MustSelfDraw:  true,
		RoomCards:     roomCards,
		RedPacketType: redPacketType,
		IPAntiCheat:   true,
	}
	user.createHNZZRoom(rule)
}

func (user *User) enterRedPacketMatchingRoom(redPacketType int, playerNumber int) bool {
	for _, r := range hnzzRoomCardMatchRooms {
		hnzzRoom := r.(*HNZZRoom)
		if hnzzRoom.rule.IPAntiCheat {
			if !hnzzRoom.loginIPs[user.data.userData.LoginIP] && hnzzRoom.rule.RedPacketType == redPacketType && len(hnzzRoom.positionUserIDs) == playerNumber {
				user.enterRoom(r)
				return true
			}
		} else {
			if hnzzRoom.rule.RedPacketType == redPacketType && len(hnzzRoom.positionUserIDs) == playerNumber {
				user.enterRoom(r)
				return true
			}
		}
	}
	return false
}

func (user *User) enterRoom(r interface{}) {
	var sitDown = false
	switch r.(type) {
	case *HNZZRoom:
		hnzzRoom := r.(*HNZZRoom)
		sitDown = hnzzRoom.Enter(user)
	}
	if sitDown {
		userIDRooms[user.data.userData.UserID] = r
	}
}

func (user *User) enterGPSRoom(r interface{}, gps bool, location []float64) {
	if gps {
		if common.CheckLocation(location) {
			user.location = location
		} else {
			user.WriteMsg(&msg.S2C_EnterRoom{
				Error: msg.S2C_EnterRoom_LocationError,
			})
			return
		}
	} else {
		user.WriteMsg(&msg.S2C_EnterRoom{
			Error: msg.S2C_EnterRoom_GPSNotOpen,
		})
		return
	}
	user.enterRoom(r)
}

func (user *User) exitOrDisbandRoom(r interface{}, forcible bool) {
	if user.robot {
		forcible = true
	}
	switch r.(type) {
	case *HNZZRoom:
		hnzzRoom := r.(*HNZZRoom)
		if hnzzRoom.state == roomIdle {
			if forcible {
				if hnzzRoom.ownerUserID == user.data.userData.UserID {
					hnzzRoom.Disband(user)
				} else {
					hnzzRoom.Exit(user)
				}
			}
			return
		}
		switch hnzzRoom.rule.RoomType {
		case roomRoomCardMatch:
			user.WriteMsg(&msg.S2C_ExitRoom{
				Error: msg.S2C_ExitRoom_GamePlaying,
			})
		case roomPrivate:
			if forcible {
				hnzzRoom.Disband(user)
			}
		}
	}
}

func (user *User) agreeDisbandRoom(r interface{}) {
	switch r.(type) {
	case *HNZZRoom:
		hnzzRoom := r.(*HNZZRoom)
		hnzzRoom.agreeDisbandRoom(user.data.userData.UserID)
	}
}

func (user *User) refuseDisbandRoom(r interface{}) {
	switch r.(type) {
	case *HNZZRoom:
		hnzzRoom := r.(*HNZZRoom)
		hnzzRoom.refuseDisbandRoom(user.data.userData.UserID)
	}
}
