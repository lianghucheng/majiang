package room

import (
	"game"
	"game/player"
	"game/room"
	msg "msg/room/mahjong"
	"util"
)

type MatchControl struct {
	CreateControl
}

func init() {
	game.HandleRegister(msgReflect(&msg.C2S_StartGDMatching{}), handleStartGDMatching)
}
func handleStartGDMatching(args []interface{}) {

	m := args[0].(*msg.C2S_StartGDMatching)
	ctx := new(MatchControl)
	lable, user := ctx.userLegal(args[1])
	if !lable {
		return
	}
	ctx.uid = user.UserData.UserID

	person := player.GetPersonMgr().GetPerson(ctx.uid)
	if !player.SystemOn {
		person.Close()
		return
	}
	if r := room.GetRoomMgr().GetRoom(ctx.uid); r != nil {
		person.WriteMsg(&msg.S2C_CreateRoom{
			Error: msg.S2C_CreateRoom_InOtherRoom,
		})
		return
	}
	switch m.RoomType {
	case room.RoomPractice:
		ctx.matchPraticeRoom()
		return
	case room.RoomRoomCardMatch:
		ctx.matchCradRoom(m.RoomCards)
		return
	case room.RoomRedPacketMatching:
		ctx.matchRedPacketRoom(m.RedPacketType)
		return
	default:
		person.WriteMsg(&msg.S2C_CreateRoom{
			Error: msg.S2C_CreateRoom_InnerError,
		})
	}
}

func (ctx *MatchControl) matchPraticeRoom() {
	for _, r := range room.GetRoomMgr().MapPerson {
		if !r.(*room.GDRoom).Full() {
			if ctx.enter(r.(*room.GDRoom)) {
				room.GetRoomMgr().AddPerson(ctx.uid, r)
			}
		}
	}
	rule := &room.GDRule{
		RoomType:   room.RoomPractice,
		MaxRounds:  1,
		MaxPlayers: 4,
		BaseScore:  1,
		NeedJoker:  false,
	}
	r := ctx.room(rule)
	if ctx.enter(r) {
		room.GetRoomMgr().AddPerson(ctx.uid, r)
	}
}

func (ctx *MatchControl) matchCradRoom(roomCards int) {
	person := player.GetPersonMgr().GetPerson(ctx.uid)
	if util.Index([]int{1, 10, 50, 100}, roomCards) == -1 {
		person.WriteMsg(&msg.S2C_CreateRoom{
			Error:     msg.S2C_CreateRoom_RuleError,
			RoomCards: roomCards,
		})
		return
	}
	if person.UserData.RoomCards < roomCards {
		person.WriteMsg(&msg.S2C_CreateRoom{
			Error:     msg.S2C_CreateRoom_LackOfRoomCards,
			RoomCards: roomCards,
		})
		return
	}
	ipAntiCheat := true
	if person.IsRobot() {
		ipAntiCheat = false
	}
	if person.IsRobot() {
		r := ctx.MatchroomByCard(roomCards, 2)
		if r != nil {
			if ctx.enter(r) {
				room.GetRoomMgr().AddPerson(ctx.uid, r)
			}
			return
		}
		r = ctx.MatchroomByCard(roomCards, 1)
		if r != nil {
			if ctx.enter(r) {
				room.GetRoomMgr().AddPerson(ctx.uid, r)
			}
			return
		}
	} else {

		r := ctx.MatchroomByCard(roomCards, 3)
		if r != nil {
			if ctx.enter(r) {
				room.GetRoomMgr().AddPerson(ctx.uid, r)
			}
			return
		}
		r = ctx.MatchroomByCard(roomCards, 2)
		if r != nil {
			if ctx.enter(r) {
				room.GetRoomMgr().AddPerson(ctx.uid, r)
			}
			return
		}
		r = ctx.MatchroomByCard(roomCards, 1)
		if r != nil {
			if ctx.enter(r) {
				room.GetRoomMgr().AddPerson(ctx.uid, r)
			}
			return
		}
	}
	rule := &room.GDRule{
		RoomType:     room.RoomRoomCardMatch,
		MaxRounds:    1,
		MaxPlayers:   4,
		MustSelfDraw: true,
		BuyHorse:     0,
		WithHonors:   false,
		NeedJoker:    true,
		RoomCards:    roomCards,
		IPAntiCheat:  ipAntiCheat,
	}
	r := ctx.room(rule)
	if ctx.enter(r) {
		room.GetRoomMgr().AddPerson(ctx.uid, r)
	}
}

func (ctx *MatchControl) MatchroomByCard(roomcards, number int) *room.GDRoom {
	person := player.GetPersonMgr().GetPerson(ctx.uid)
	for _, ri := range room.GetRoomMgr().MapPerson {
		r := ri.(*room.GDRoom)
		if r.Rule.IPAntiCheat {
			if !r.Full() && !r.LoginIPs[person.UserData.LoginIP] && r.JoinNumber() == number && r.Rule.RoomCards == roomcards {
				return r
			}
		}
		if !r.Full() && r.JoinNumber() == number && r.Rule.RoomCards == roomcards {
			return r
		}
	}
	return nil
}

func (ctx *MatchControl) MatchroomByRedPacket(redPacketType, number int) *room.GDRoom {
	person := player.GetPersonMgr().GetPerson(ctx.uid)
	for _, ri := range room.GetRoomMgr().MapPerson {
		r := ri.(*room.GDRoom)
		if r.Rule.IPAntiCheat {
			if !r.Full() && !r.LoginIPs[person.UserData.LoginIP] && r.JoinNumber() == number && r.Rule.RedPacketType == redPacketType {
				return r
			}
		}
		if !r.Full() && r.JoinNumber() == number && r.Rule.RedPacketType == redPacketType {
			return r
		}
	}
	return nil
}
func (ctx *MatchControl) matchRedPacketRoom(redPacketType int) {
	person := player.GetPersonMgr().GetPerson(ctx.uid)
	roomCards := 0
	switch redPacketType {
	case 1:
		roomCards = 2
	case 10:
		roomCards = 15
	default:
		person.WriteMsg(&msg.S2C_CreateRoom{Error: msg.S2C_CreateRoom_RuleError})
		return
	}

	if !room.Start() {
		person.WriteMsg(&msg.S2C_EnterRoom{Error: msg.S2C_EnterRoom_NotRightNow})
		return
	}
	if roomCards > person.UserData.RoomCards {
		person.WriteMsg(&msg.S2C_CreateRoom{
			Error:     msg.S2C_CreateRoom_LackOfRoomCards,
			RoomCards: roomCards,
		})
	}
	if person.IsRobot() {
		r := ctx.MatchroomByRedPacket(redPacketType, 2)
		if r != nil {
			if ctx.enter(r) {
				room.GetRoomMgr().AddPerson(ctx.uid, r)
			}
			return
		}
		r = ctx.MatchroomByRedPacket(redPacketType, 1)
		if r != nil {
			if ctx.enter(r) {
				room.GetRoomMgr().AddPerson(ctx.uid, r)
			}
			return
		}
	} else {

		r := ctx.MatchroomByRedPacket(redPacketType, 3)
		if r != nil {
			if ctx.enter(r) {
				room.GetRoomMgr().AddPerson(ctx.uid, r)
			}
			return
		}
		r = ctx.MatchroomByRedPacket(redPacketType, 2)
		if r != nil {
			if ctx.enter(r) {
				room.GetRoomMgr().AddPerson(ctx.uid, r)
			}
			return
		}
		r = ctx.MatchroomByRedPacket(redPacketType, 1)
		if r != nil {
			if ctx.enter(r) {
				room.GetRoomMgr().AddPerson(ctx.uid, r)
			}
			return
		}
	}
	rule := &room.GDRule{
		RoomType:      room.RoomRedPacketMatching,
		MaxPlayers:    4,
		MaxRounds:     1,
		MustSelfDraw:  true,
		RoomCards:     roomCards,
		RedPacketType: redPacketType,
		IPAntiCheat:   true,
		NeedJoker:     true,
	}
	r := ctx.room(rule)
	if ctx.enter(r) {
		room.GetRoomMgr().AddPerson(ctx.uid, r)
	}
}
