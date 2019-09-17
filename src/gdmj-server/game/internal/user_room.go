package internal

import (
	"gdmj-server/common"
	"gdmj-server/game/mahjong"
	"gdmj-server/msg"

	"github.com/name5566/leaf/log"
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

func (user *User) createGDRoom(gdRule *mahjong.GDRule) {
	roomNumber := ""
	switch gdRule.RoomType {
	case roomPrivate, roomRedPacketPrivate:
		roomNumber = getRoomNumber()
		if _, ok := gdroomNumberRooms[roomNumber]; ok {
			user.WriteMsg(&msg.S2C_CreateRoom{
				Error: msg.S2C_CreateRoom_InnerError,
			})
			//user.Close()
			return
		}
	}
	gdRoom := newGDRoom(gdRule)
	if gdRoom.gameType== 1 {
		switch gdRule.RoomType {

		case roomPractice:
			log.Debug("userID: %v 创建广东练习房", user.data.userData.UserID)
			gdPracticeRooms[user.data.userData.UserID] = gdRoom
		case roomRoomCardMatch, roomRedPacketMatching:
			log.Debug("userID: %v 创建广东房卡比赛场", user.data.userData.UserID)
			gdMatchRooms[user.data.userData.UserID] = gdRoom
		case roomPrivate, roomRedPacketPrivate:
			log.Debug("userID: %v 创建广东私人房", user.data.userData.UserID)
			gdRoom.number = roomNumber
			gdRoom.ownerUserID = user.data.userData.UserID
			gdroomNumberRooms[roomNumber] = gdRoom
		}
	}
	if gdRoom.gameType == 2 {
		switch gdRule.RoomType {

		case roomPractice:
			log.Debug("userID: %v 创建延安练习房", user.data.userData.UserID)
			gdPracticeRooms[user.data.userData.UserID] = gdRoom
		case roomRoomCardMatch, roomRedPacketMatching:
			log.Debug("userID: %v 创建延安房卡比赛场", user.data.userData.UserID)
			gdMatchRooms[user.data.userData.UserID] = gdRoom
		case roomPrivate, roomRedPacketPrivate:
			log.Debug("userID: %v 创建延安私人房", user.data.userData.UserID)
			gdRoom.number = roomNumber
			gdRoom.ownerUserID = user.data.userData.UserID
			gdroomNumberRooms[roomNumber] = gdRoom
		}
	}
	if gdRoom.gameType==3{
		switch gdRule.RoomType {
		case roomRoomCardMatch, roomRedPacketMatching:
			log.Debug("userID: %v 创建HNZZ匹配房", user.data.userData.UserID)
			hnzzRoomCardMatchRooms[user.data.userData.UserID] = gdRoom
		case roomPrivate, roomRedPacketPrivate:
			log.Debug("userID: %v 创建HNZZ私人房 房号: %v, 局数: %v, 人数: %v, 底分: %v", user.data.userData.UserID, roomNumber, gdRule.MaxRounds, gdRule.MaxPlayers, gdRule.BaseScore)
			gdRoom.number = roomNumber
			gdRoom.ownerUserID = user.data.userData.UserID
			hnzzroomNumberRooms[roomNumber] = gdRoom
		}
	}
	gdRoom.creatorUserID = user.data.userData.UserID
	user.enterRoom(gdRoom)
}

func (user *User) createPrivateRoom(info *msg.C2S_CreateGDRoom) {
	switch info.GameType {
	case mahjong.GD:
		if common.Index([]int{4, 8, 16}, info.MaxRounds) == -1 ||
			common.Index([]int{2, 3, 4}, info.MaxPlayers) == -1 ||
			common.Index([]int{0, 1, 2}, info.BuyHorse) == -1 {
			user.WriteMsg(&msg.S2C_CreateRoom{
				Error: msg.S2C_CreateRoom_RuleError,
			})
			//user.Close()
			return
		}
	case mahjong.YA:

	case mahjong.HNZZ:
		if common.Index([]int{4, 8, 16}, info.MaxRounds) == -1 || common.Index([]int{2, 3, 4}, info.MaxPlayers) == -1 || common.Index([]int{2, 4, 6}, info.Birds) == -1 {
			user.WriteMsg(&msg.S2C_CreateRoom{
				Error: msg.S2C_CreateRoom_RuleError,
			})
			return
		}
	}

	needRoomCards := 2
	if info.MaxRounds == 8 {
		needRoomCards = 3
	} else if info.MaxRounds == 16 {
		needRoomCards = 5
	}
	if !user.checkCreateRoomCards(needRoomCards) {
		//user.Close()
		return
	}
	gdRule := &mahjong.GDRule{
		RoomType:       roomPrivate,
		MaxRounds:      info.MaxRounds,
		MaxPlayers:     info.MaxPlayers,
		MustSelfDraw:   info.MustSelfDraw,
		BaseScore:      1,
		BuyHorse:       info.BuyHorse,
		WithHonors:     info.WithHonors,
		NeedJoker:      info.NeedJoker,
		RoomCards:      needRoomCards,
		IPAntiCheat:    info.IPAntiCheat,
		Gun:            info.Gun,
		RedDragonJoker: info.RedDragonJoker,
		GameType:       info.GameType,
	}
	if info.GPSAntiCheat {
		if common.CheckLocation(info.Location) {
			user.location = info.Location
			log.Debug("location: %v", user.location)
			gdRule.GPSAntiCheat = true
		} else {
			user.WriteMsg(&msg.S2C_CreateRoom{
				Error: msg.S2C_CreateRoom_LocationError,
			})
			//user.Close()
			return
		}
	}
	user.createGDRoom(gdRule)
}

func (user *User) createOrEnterPracticeRoom() {
	for _, r := range gdPracticeRooms {
		gdRoom := r.(*GDRoom)
		if !gdRoom.full() {
			user.enterRoom(r)
			return
		}
	}
	gdRule := &mahjong.GDRule{
		RoomType:   roomPractice,
		MaxRounds:  1,
		MaxPlayers: 4,
		BaseScore:  1,
		NeedJoker:  false,
	}
	user.createGDRoom(gdRule)
}

func (user *User) createOrEnterRoomCardMatchRoom(roomCards int, gameType int) {
	if common.Index([]int{1, 10, 50, 100}, roomCards) == -1 {
		user.WriteMsg(&msg.S2C_CreateRoom{
			Error:     msg.S2C_CreateRoom_RuleError,
			RoomCards: roomCards,
		})
		//user.Close()
		return
	}
	if user.data.userData.RoomCards < roomCards {
		user.WriteMsg(&msg.S2C_CreateRoom{
			Error:     msg.S2C_CreateRoom_LackOfRoomCards,
			RoomCards: roomCards,
		})
		//user.Close()
		return
	}
	ipAntiCheat := true
	if user.data.userData.Role == roleRoot {
		ipAntiCheat = false
	}
	gdRule := &mahjong.GDRule{
		GameType:     gameType,
		RoomType:     roomRoomCardMatch,
		MaxRounds:    1,
		MaxPlayers:   4,
		MustSelfDraw: true,
		BuyHorse:     0,
		WithHonors:   false,
		NeedJoker:    true,
		RoomCards:    roomCards,
		IPAntiCheat:  ipAntiCheat,
	}
	/*
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
	*/
	if user.enterRoomCardMatchingRoom(roomCards, 3, gameType) {
		return
	}
	if user.enterRoomCardMatchingRoom(roomCards, 2, gameType) {
		return
	}
	if user.enterRoomCardMatchingRoom(roomCards, 1, gameType) {
		return
	}
	user.createGDRoom(gdRule)
}

func (user *User) enterRoomCardMatchingRoom(roomCards int, playerNumber int, gameType int) bool {
	for _, r := range gdMatchRooms {
		gdRoom := r.(*GDRoom)
		if gdRoom.rule.IPAntiCheat {
			if !gdRoom.loginIPs[user.data.userData.LoginIP] && gdRoom.rule.RoomCards == roomCards && len(gdRoom.positionUserIDs) == playerNumber && gdRoom.gameType == gameType {
				user.enterRoom(r)
				return true
			}
		} else {
			if gdRoom.rule.RoomCards == roomCards && len(gdRoom.positionUserIDs) == playerNumber && gdRoom.gameType == gameType {
				user.enterRoom(r)
				return true
			}
		}
	}
	return false
}

func (user *User) enterRedPacketMatchingRoom(redPacketType int, playerNumber int, gameType int) bool {
	for _, r := range gdMatchRooms {
		gdRoom := r.(*GDRoom)
		if gdRoom.rule.IPAntiCheat {
			if !gdRoom.loginIPs[user.data.userData.LoginIP] && gdRoom.rule.RedPacketType == redPacketType && len(gdRoom.positionUserIDs) == playerNumber && gdRoom.gameType == gameType {
				user.enterRoom(r)
				return true
			}
		} else {
			if gdRoom.rule.RedPacketType == redPacketType && len(gdRoom.positionUserIDs) == playerNumber && gdRoom.gameType == gameType {
				user.enterRoom(r)
				return true
			}
		}
	}
	return false
}

func (user *User) createOrEnterRedPacketMatchingRoom(redPacketType int, gameType int) {
	roomCards := 0
	switch redPacketType {
	case 1:
		roomCards = 2
	case 10:
		roomCards = 15
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
		if user.enterRedPacketMatchingRoom(redPacketType, 2, gameType) {
			return
		}
		if user.enterRedPacketMatchingRoom(redPacketType, 1, gameType) {
			return
		}
	} else {
		if user.enterRedPacketMatchingRoom(redPacketType, 3, gameType) {
			return
		}
		if user.enterRedPacketMatchingRoom(redPacketType, 2, gameType) {
			return
		}
		if user.enterRedPacketMatchingRoom(redPacketType, 1, gameType) {
			return
		}
	}

	rule := &mahjong.GDRule{
		GameType:      gameType,
		RoomType:      roomRedPacketMatching,
		MaxPlayers:    4,
		MaxRounds:     1,
		MustSelfDraw:  true,
		RoomCards:     roomCards,
		RedPacketType: redPacketType,
		IPAntiCheat:   true,
		NeedJoker:     true,
	}
	user.createGDRoom(rule)
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
	rule := &mahjong.GDRule{
		RoomType:      roomRedPacketPrivate,
		MaxPlayers:    4,
		MaxRounds:     1,
		MustSelfDraw:  true,
		NeedJoker:     true,
		RoomCards:     roomCards,
		RedPacketType: redPacketType,
		IPAntiCheat:   true,
	}
	user.createGDRoom(rule)
}

func (user *User) enterRoom(r interface{}) {
	var sitDown = false
	switch r.(type) {
	case *GDRoom:
		gdRoom := r.(*GDRoom)
		sitDown = gdRoom.Enter(user)
	}
	if sitDown {
		userIDRooms[user.data.userData.UserID] = r
		return
	}
	//user.Close()
}

func (user *User) enterGPSRoom(r interface{}, gps bool, location []float64) {
	if gps {
		if common.CheckLocation(location) {
			user.location = location
		} else {
			user.WriteMsg(&msg.S2C_EnterRoom{
				Error: msg.S2C_EnterRoom_LocationError,
			})
			//user.Close()
			return
		}
	} else {
		user.WriteMsg(&msg.S2C_EnterRoom{
			Error: msg.S2C_EnterRoom_GPSNotOpen,
		})
		//user.Close()
		return
	}
	user.enterRoom(r)
}

func (user *User) exitOrDisbandRoom(r interface{}, forcible bool) {
	if user.isRobot() {
		forcible = true
	}
	switch r.(type) {
	case *GDRoom:
		gdRoom := r.(*GDRoom)
		if gdRoom.state == roomIdle {
			if forcible {
				if gdRoom.ownerUserID == user.data.userData.UserID {
					gdRoom.Disband(user)
				} else {
					gdRoom.Exit(user)
				}
			}
			return
		}
		switch gdRoom.rule.RoomType {
		case roomRoomCardMatch:
			user.WriteMsg(&msg.S2C_ExitRoom{
				Error: msg.S2C_ExitRoom_GamePlaying,
			})
		case roomPrivate:
			if forcible {
				gdRoom.Disband(user)
			}
		}
	}
}

func (user *User) agreeDisbandRoom(r interface{}) {
	switch r.(type) {
	case *GDRoom:
		gdRoom := r.(*GDRoom)
		gdRoom.agreeDisbandRoom(user.data.userData.UserID)
	}
}

func (user *User) refuseDisbandRoom(r interface{}) {
	switch r.(type) {
	case *GDRoom:
		gdRoom := r.(*GDRoom)
		gdRoom.refuseDisbandRoom(user.data.userData.UserID)
	}
}
