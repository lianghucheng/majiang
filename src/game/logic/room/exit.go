package room

import (
	"game"
	"game/player"
	"game/room"
	msg "msg/room/mahjong"
	"time"

	"github.com/name5566/leaf/log"
)

type ExitRoom struct {
	CreateControl
}

func init() {
	game.HandleRegister(msgReflect(&msg.C2S_ExitOrDisbandRoom{}), handleExitOrDisbandRoom)
}
func handleExitOrDisbandRoom(args []interface{}) {
	ctx := new(ExitRoom)
	lable, person := ctx.userLegal(args[1])
	if !lable {
		return
	}
	ctx.uid = person.UserData.UserID
	if ri := room.GetRoomMgr().GetRoom(ctx.uid); ri != nil {
		ctx.exitOrDisRoom(ri.(*room.GDRoom), true)
	}
}

func (ctx *ExitRoom) exitOrDisRoom(r *room.GDRoom, forcible bool) {
	person := player.GetPersonMgr().GetPerson(ctx.uid)
	if person.IsRobot() {
		forcible = true
	}
	if r.State == room.RoomIdle {
		if forcible {
			if r.OwnerUserID == ctx.uid {
				ctx.dis(r)
			} else {
				ctx.exit(r)
			}
		}
		return
	}
	switch r.Rule.RoomType {
	case room.RoomRoomCardMatch:
		person.WriteMsg(&msg.S2C_ExitRoom{
			Error: msg.S2C_ExitRoom_GamePlaying,
		})
	case room.RoomPrivate:
		if forcible {
			ctx.dis(r)
		}
	}
}
func (ctx *ExitRoom) dis(r *room.GDRoom) {
	person := player.GetPersonMgr().GetPerson(ctx.uid)
	if r.State == room.RoomIdle {
		log.Debug("userID: %v 解散房间 %v", ctx.uid, r.Number)
		r.BroadcastAll(&msg.S2C_DisbandRoom{
			Error:         msg.S2C_DisbandRoom_OK,
			RoomNumber:    r.Number,
			OwnerNickName: person.UserData.Nickname,
		})
		// 清空玩家数据
		r.Clean()
		if r.Rule.RoomType == room.RoomPrivate {
			room.GetRoom().DelRoom(r.Number) // 解散房间
		}
		return
	}
	log.Debug("userID: %v 申请解散房间 %v", ctx.uid, r.Number)
	r.Disbandapplicantuserid = ctx.uid
	applicantPlayerData := r.Useridplayerdatas[r.Disbandapplicantuserid]
	applicantPlayerData.DisbandActionCode = room.ActionAgreeDisband
	for key, value := range r.Useridplayerdatas {
		if key != ctx.uid {
			value.DisbandActionCode = room.ActionWaitingDisband
		}
	}
	playerDisbandInfos := []msg.GDPlayerDisbandInfo{}
	for i := 0; i < r.Rule.MaxPlayers; i++ {
		userID := r.PositionUserIDs[i]
		playerData := r.Useridplayerdatas[userID]
		playerDisbandInfos = append(playerDisbandInfos, msg.GDPlayerDisbandInfo{
			Nickname:   playerData.User.UserData.Nickname,
			ActionCode: playerData.DisbandActionCode,
		})
	}
	for _, userID := range r.PositionUserIDs {
		playerData := r.Useridplayerdatas[userID]
		person := player.GetPersonMgr().GetPerson(userID)
		if person == nil {
			continue
		}
		person.WriteMsg(&msg.S2C_ActionDisbandRoom{
			ApplicantNickname:  playerData.User.UserData.Nickname,
			PlayerDisbandInfos: playerDisbandInfos,
			Enable:             playerData.DisbandActionCode == room.ActionAgreeDisband,
			WaitingTime:        120,
		})
	}
	if r.Discardtimer != nil {
		r.Discardtimer.Stop()
		r.Discardtimer = nil
	}
	r.Disbandtimer = game.Skeleton.AfterFunc(122*time.Second, func() {
		for _, userID := range r.PositionUserIDs {
			playerData := r.Useridplayerdatas[userID]
			if playerData.DisbandActionCode == room.ActionWaitingDisband {
				log.Debug("userID: %v 自动同意", playerData.User.UserData.UserID)
				ctx.agreeDisbandRoom(r)
			}
		}
	})
}

func (ctx *ExitRoom) agreeDisbandRoom(r *room.GDRoom) {
	if r.State == room.RoomIdle {
		return
	}
	playerData := r.Useridplayerdatas[ctx.uid]
	if playerData.DisbandActionCode != room.ActionWaitingDisband {
		return
	}
	playerData.DisbandActionCode = room.ActionAgreeDisband
	if r.AllAgree() {
		if r.Rule.RoomType == room.RoomPrivate && r.Currentround == 1 {
			CostCard(r)
		}
		r.BroadcastAll(&msg.S2C_DisbandRoom{
			Error:      msg.S2C_DisbandRoom_OK,
			RoomNumber: r.Number,
		})
		r.Clean()
		if r.Rule.RoomType == room.RoomPrivate {
			room.GetRoom().DelRoom(r.Number)
		}
	} else {
		r.BroadcastAll(&msg.S2C_AgreeDisbandRoom{
			Position: playerData.Position,
			Nickname: playerData.User.UserData.Nickname,
		})
	}
}

func (ctx *ExitRoom) exit(r *room.GDRoom) {
	playerData := r.Useridplayerdatas[ctx.uid]
	if playerData == nil {
		return
	}
	r.BroadcastAll(&msg.S2C_StandUp{
		Position: playerData.Position,
	})
	log.Debug("userID: %v 退出房间  %v", ctx.uid, r.Number)
	r.BroadcastAll(&msg.S2C_ExitRoom{
		Error:    msg.S2C_ExitRoom_OK,
		Position: playerData.Position,
	})
	// 站起
	delete(r.PositionUserIDs, playerData.Position)
	delete(r.Useridplayerdatas, ctx.uid)
	// 退出
	room.GetRoomMgr().DelPerson(ctx.uid)
	// 删除玩家登陆ip
	delete(r.LoginIPs, playerData.User.UserData.LoginIP)
	playerData.User.Location = []float64{}

	switch r.Rule.RoomType {
	case room.RoomRoomCardMatch:
		//calculateRoomCardMatchOnlineNumber(gdRoom.rule.RoomCards, true)
	}

	if r.Empty() {
		switch r.Rule.RoomType {
		case room.RoomPrivate, room.RoomRedPacketPrivate:
			room.GetRoom().DelRoom(r.Number)
		}
	}
}
