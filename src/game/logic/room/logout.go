package room

import (
	"game"
	"game/player"
	"game/room"

	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
)

type LogoutControl struct {
	*player.User
}

func init() {
	game.HandleRegister("CloseAgent", closeAgent)
}
func closeAgent(args []interface{}) {
	a := args[0].(gate.Agent)
	user := a.UserData().(*AgentInfo).User
	a.SetUserData(nil)
	if user == nil {
		return
	}
	if user.State == player.UserLogin {
		user.State = player.UserLogout
		ctx := new(LogoutControl)
		ctx.User = user
		ctx.logout()
	}
}

func (ctx *LogoutControl) logout() {
	if ctx.HeartbeatTimer != nil {
		ctx.HeartbeatTimer.Stop()
	}
	if ctx.UserData == nil {
		return
	}
	if person := player.GetPersonMgr().GetPerson(ctx.UserData.UserID); person != nil {
		if person == ctx.User {
			log.Debug("userID: %v 登出", person.UserData.UserID)
			ctx.exitRoom()
			player.GetPersonMgr().DelPerson(person.UserData.UserID)
			player.SaveUserData(ctx.UserData)
		}
	}
}

func (ctx *LogoutControl) exitRoom() {
	ri := room.GetRoomMgr().GetRoom(ctx.UserData.UserID)
	if ri == nil {
		return
	}
	exitCtx := new(ExitRoom)
	exitCtx.uid = ctx.UserData.UserID
	exitCtx.exitOrDisRoom(ri.(*room.GDRoom), false)

}
