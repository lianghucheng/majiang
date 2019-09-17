package room

import (
	"game"
	msg "msg/communicate"
)

type HeartControl struct {
	CreateControl
}

func init() {
	game.HandleRegister(msgReflect(&msg.C2S_Heartbeat{}), handleHeartbeat)
}
func handleHeartbeat(args []interface{}) {
	ctx := new(CreateControl)
	label, person := ctx.userLegal(args[1])
	if !label {
		return
	}
	person.HeartbeatStop = false
}
