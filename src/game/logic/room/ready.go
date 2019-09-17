package room

import (
	"game"
	"game/room"
	msg "msg/room/mahjong"
	"time"
)

type ReadyControl struct {
	CreateControl
}

func init() {
	game.HandleRegister(msgReflect(&msg.C2S_Prepare{}), handlePrepare)
}
func handlePrepare(args []interface{}) {
	ctx := new(ReadyControl)
	lable, person := ctx.userLegal(args[1])
	if !lable {
		return
	}
	ctx.uid = person.UserData.UserID
	if ri := room.GetRoomMgr().GetRoom(ctx.uid); ri != nil {
		ctx.ready(ri.(*room.GDRoom))
	}

}

func (ctx *ReadyControl) ready(r *room.GDRoom) {
	//断线重连
	if r.State == room.RoomGame {
		reconnect(ctx.uid)
		return

	}
	ctx.doReady(r)
}

func (ctx *ReadyControl) doReady(r *room.GDRoom) {
	playerData := r.Useridplayerdatas[ctx.uid]
	playerData.State = room.GdReady

	//gdRoom.refuseDisbandRoom(userID)

	r.BroadcastAll(&msg.S2C_Prepare{
		Position: playerData.Position,
		Ready:    true,
	})
	if r.AllReady() {
		r.State = room.RoomGame
		game.Skeleton.AfterFunc(1*time.Second, func() {
			StartGame(r)
		})
	}
}
