package room

import (
	"algorithm"
	"game"
	"game/player"
	"game/room"
	msg "msg/room/mahjong"
	"strconv"
	"time"
	"util"

	"github.com/name5566/leaf/log"

	"github.com/name5566/leaf/gate"
)

type CreateControl struct {
	uid int
}

func init() {
	game.HandleRegister(msgReflect(&msg.C2S_CreateGDRoom{}), handleCreateGDRoom)
}
func handleCreateGDRoom(args []interface{}) {
	m := args[0].(*msg.C2S_CreateGDRoom)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).User
	if user == nil {
		return
	}
	if r := room.GetRoomMgr().GetRoom(user.UserData.UserID); r != nil {
		user.WriteMsg(&msg.S2C_CreateRoom{
			Error: msg.S2C_CreateRoom_InOtherRoom,
		})
		//user.Close()
		return
	}
	if !player.SystemOn {
		user.Close()
		return
	}
	ctx := new(CreateControl)
	ctx.uid = user.UserData.UserID
	switch m.RoomType {
	case room.RoomPrivate:
		ctx.createPrivate(m)
		return
	case room.RoomRedPacketPrivate:
		//user.createRedPacketPrivateRoom(m.RedPacketType)
		return
	}
	user.WriteMsg(&msg.S2C_CreateRoom{Error: msg.S2C_CreateRoom_RuleError})
}

func (ctx *CreateControl) createPrivate(info *msg.C2S_CreateGDRoom) {
	person := player.GetPersonMgr().GetPerson(ctx.uid)
	if person == nil {
		log.Error("%v玩家已离线", ctx.uid)
		return
	}
	//房间参数是否合法
	if !ctx.paramLegal(info) {
		person.WriteMsg(&msg.S2C_CreateRoom{
			Error: msg.S2C_CreateRoom_RuleError,
		})
		return

	}
	//玩家房卡是否满足入场
	if person.UserData.RoomCards < ctx.enterCard(info.MaxRounds) {
		person.WriteMsg(&msg.S2C_EnterRoom{
			Error:     msg.S2C_EnterRoom_LackOfRoomCards,
			RoomCards: ctx.enterCard(info.MaxRounds),
		})
	}
	gdRule := room.NewRule(info, ctx.enterCard(info.MaxRounds))
	if info.GPSAntiCheat {
		if util.CheckLocation(info.Location) {
			person.Location = info.Location
			log.Debug("location: %v", person.Location)
			gdRule.GPSAntiCheat = true
		} else {
			person.WriteMsg(&msg.S2C_CreateRoom{
				Error: msg.S2C_CreateRoom_LocationError,
			})
			person.Close()
			return
		}
	}

	r := ctx.room(gdRule)
	if ctx.enter(r) {
		room.GetRoomMgr().AddPerson(ctx.uid, r)
	}
}

func (ctx *CreateControl) paramLegal(info *msg.C2S_CreateGDRoom) bool {
	if util.Index([]int{4, 8, 16}, info.MaxRounds) == -1 || util.Index([]int{2, 3, 4}, info.MaxPlayers) == -1 ||
		util.Index([]int{0, 1, 2}, info.BuyHorse) == -1 {

		return false
	}
	return true
}
func (ctx *CreateControl) userLegal(arg interface{}) (bool, *player.User) {
	a := arg.(gate.Agent)
	if a.UserData() == nil {
		return false, nil
	}
	user := a.UserData().(*AgentInfo).User
	if user == nil {
		return false, nil
	}
	return true, user
}
func (ctx *CreateControl) enterCard(maxrounds int) int {
	needRoomCards := 2
	if maxrounds == 8 {
		needRoomCards = 3
	} else if maxrounds == 16 {
		needRoomCards = 5
	}
	return needRoomCards
}

func (ctx *CreateControl) room(rule *room.GDRule) *room.GDRoom {
	roomNumber := 0
	if rule.RoomType == room.RoomPrivate || rule.RoomType == room.RoomRedPacketPrivate {
		roomNumber = room.GetRoom().GetID()
	}

	r := room.NewGDRoom(rule)
	if roomNumber > 0 {
		r.Number = strconv.Itoa(roomNumber)
		room.GetRoom().SetRoom(r.Number, r)
	}
	r.CreatorUserID = ctx.uid
	return r

}

func (ctx *CreateControl) enter(r *room.GDRoom) bool {
	roomCards := 0
	switch r.Rule.RoomType {
	case room.RoomRoomCardMatch, room.RoomRedPacketMatching, room.RoomRedPacketPrivate:
		roomCards = r.Rule.RoomCards
	}
	person := player.GetPersonMgr().GetPerson(ctx.uid)
	if playerData, ok := r.Useridplayerdatas[ctx.uid]; ok {
		playerData.User = person
		person.WriteMsg(&msg.S2C_EnterRoom{
			Error:         msg.S2C_EnterRoom_OK,
			RoomType:      r.Rule.RoomType,
			RedPacketType: r.Rule.RedPacketType,
			RoomNumber:    r.Number,
			Position:      playerData.Position,
			RoomDesc:      r.Desc,
			MaxPlayers:    r.Rule.MaxPlayers,
			MaxRounds:     r.Rule.MaxRounds,
			NeedJoker:     r.Rule.NeedJoker,
			BuyHorse:      r.Rule.BuyHorse,
			RoomCards:     roomCards,
			GamePlaying:   r.State == room.RoomGame,
		})
		time.Sleep(200 * time.Millisecond)
		ctx.getInfo(r)
		log.Debug("userID: %v 重连进入房间， 房间类型: %v", ctx.uid, r.Rule.RoomType)
		return true
	}
	// 玩家已满
	if r.Full() {
		person.WriteMsg(&msg.S2C_EnterRoom{
			Error:      msg.S2C_EnterRoom_Full,
			RoomNumber: r.Number,
		})
		return false
	}
	switch r.Rule.RoomType {
	case room.RoomRoomCardMatch, room.RoomRedPacketMatching, room.RoomRedPacketPrivate:
		if person.UserData.RoomCards < ctx.enterCard(r.Rule.MaxRounds) {
			person.WriteMsg(&msg.S2C_EnterRoom{
				Error:     msg.S2C_EnterRoom_LackOfRoomCards,
				RoomCards: ctx.enterCard(r.Rule.MaxRounds),
			})
			return false
		}
	}

	if r.Rule.IPAntiCheat {
		if _, ok := r.LoginIPs[person.UserData.LoginIP]; ok {
			person.WriteMsg(&msg.S2C_EnterRoom{
				Error: msg.S2C_EnterRoom_IPConflict,
			})
			return false
		}
		r.LoginIPs[person.UserData.LoginIP] = true
	}
	for pos := 0; pos < r.Rule.MaxPlayers; pos++ {
		if _, ok := r.PositionUserIDs[pos]; !ok {
			ctx.sitDown(r, pos)
			person.WriteMsg(&msg.S2C_EnterRoom{
				Error:         msg.S2C_EnterRoom_OK,
				RoomType:      r.Rule.RoomType,
				RedPacketType: r.Rule.RedPacketType,
				RoomNumber:    r.Number,
				Position:      pos,
				RoomDesc:      r.Desc,
				MaxPlayers:    r.Rule.MaxPlayers,
				MaxRounds:     r.Rule.MaxRounds,
				NeedJoker:     r.Rule.NeedJoker,
				BuyHorse:      r.Rule.BuyHorse,
				RoomCards:     roomCards,
				GamePlaying:   r.State == room.RoomGame,
			})
			time.Sleep(50 * time.Millisecond)
			ctx.getInfo(r)
			log.Debug("userID: %v 进入房间, 房间类型: %v", ctx.uid, r.Rule.RoomType)
			switch r.Rule.RoomType {
			case room.RoomRoomCardMatch:
				//calculateRoomCardMatchOnlineNumber(gdRoom.rule.RoomCards, false)
			case room.RoomRedPacketMatching, room.RoomRedPacketPrivate:
				//calculateRedPacketMatchOnlineNumber(gdRoom.rule.RedPacketType)
			}
			return true
		}
	}
	person.WriteMsg(&msg.S2C_EnterRoom{
		Error:      msg.S2C_EnterRoom_Unknown,
		RoomNumber: r.Number,
	})
	return false
}

func (ctx *CreateControl) getInfo(r *room.GDRoom) {
	person := player.GetPersonMgr().GetPerson(ctx.uid)
	for pos := 0; pos < r.Rule.MaxPlayers; pos++ {
		userID := r.PositionUserIDs[pos]
		playerData := r.Useridplayerdatas[userID]
		if playerData == nil {
			person.WriteMsg(&msg.S2C_StandUp{
				Position: pos,
			})
		} else {
			if playerData.User.UserData.Role == player.RoleRobot {
				game.Skeleton.AfterFunc(time.Duration(pos+1)*time.Second, func() {
					person.WriteMsg(&msg.S2C_SitDown{
						Position:   playerData.Position,
						Owner:      playerData.Owner,
						AccountID:  playerData.User.UserData.AccountID,
						LoginIP:    playerData.User.UserData.LoginIP,
						Nickname:   playerData.User.UserData.Nickname,
						Headimgurl: playerData.User.UserData.Headimgurl,
						Sex:        playerData.User.UserData.Sex,
						Ready:      playerData.State == room.GdReady,
						Location:   playerData.User.Location,
					})
				})
			} else {
				person.WriteMsg(&msg.S2C_SitDown{
					Position:   pos,
					Owner:      playerData.Owner,
					AccountID:  playerData.User.UserData.AccountID,
					LoginIP:    playerData.User.UserData.LoginIP,
					Nickname:   playerData.User.UserData.Nickname,
					Headimgurl: playerData.User.UserData.Headimgurl,
					Sex:        playerData.User.UserData.Sex,
					Ready:      playerData.State == room.GdReady,
					Location:   playerData.User.Location,
				})
			}
		}
	}
}

func (ctx *CreateControl) sitDown(r *room.GDRoom, pos int) {
	person := player.GetPersonMgr().GetPerson(ctx.uid)
	r.PositionUserIDs[pos] = ctx.uid

	playerData := r.Useridplayerdatas[ctx.uid]
	if playerData == nil {
		playerData = new(room.GDPlayerData)
		playerData.User = person
		playerData.Position = pos
		playerData.Owner = (ctx.uid == r.OwnerUserID)
		playerData.Analyzer = new(algorithm.GDAnalyzer)
		playerData.RoundResult = new(msg.GDPlayerRoundResult)
		playerData.TotalResult = new(msg.GDPlayerTotalResult)

		r.Useridplayerdatas[ctx.uid] = playerData
	}
	message := &msg.S2C_SitDown{
		Position:   pos,
		Owner:      playerData.Owner,
		AccountID:  playerData.User.UserData.AccountID,
		LoginIP:    playerData.User.UserData.LoginIP,
		Nickname:   playerData.User.UserData.Nickname,
		Headimgurl: playerData.User.UserData.Headimgurl,
		Sex:        playerData.User.UserData.Sex,
		Ready:      playerData.State == room.GdReady,
	}
	if r.Rule.GPSAntiCheat {
		message.Location = playerData.User.Location
	}
	r.Broadcast(message, pos)
}
