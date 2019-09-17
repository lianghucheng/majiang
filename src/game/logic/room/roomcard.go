package room

import (
	"game"
	msg "msg/card"
)

type CardControl struct {
	CreateControl
}

func init() {
	game.HandleRegister(msgReflect(&msg.C2S_GetRoomCards{}), handleGetRoomCards)
}
func handleGetRoomCards(args []interface{}) {
	ctx := new(CreateControl)
	label, person := ctx.userLegal(args[1])
	if !label {
		return
	}
	person.WriteMsg(&msg.S2C_UpdateRoomCards{
		RoomCards: person.UserData.RoomCards,
	})
	/*
		user.WriteMsg(&msg.S2C_UpdateRoomCardsMatchOnlineNumber{
			Numbers: roomCardMatchOnlineNumber,
		})
		user.sendRedPacketMatchOnlineNumber()
		user.sendUntakenRedPacketMatchPrizeNumber()
	*/
}
