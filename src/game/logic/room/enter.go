package room

import (
	"game"
	"game/player"
	"game/room"
	msg "msg/room/mahjong"
	"strings"
	"util"
)

type EnterRoom struct {
	CreateControl
}

func init() {
	game.HandleRegister(msgReflect(&msg.C2S_EnterRoom{}), handleEnterRoom)
}
func handleEnterRoom(args []interface{}) {
	ctx := new(EnterRoom)
	label, person := ctx.userLegal(args[1])
	if !label {
		return
	}
	ctx.uid = person.UserData.UserID
	m := args[0].(*msg.C2S_EnterRoom)
	if strings.TrimSpace(m.RoomNumber) == "" {
		if ri := room.GetRoomMgr().GetRoom(ctx.uid); ri != nil {
			r := ri.(*room.GDRoom)
			if r.Rule.GPSAntiCheat {
				if m.GPS {
					if util.CheckLocation(m.Location) {
						person.Location = m.Location
					} else {
						person.WriteMsg(&msg.S2C_EnterRoom{
							Error: msg.S2C_EnterRoom_LocationError,
						})
						return
					}
					ctx.enter(r)
					return
				} else {
					person.WriteMsg(&msg.S2C_EnterRoom{
						Error: msg.S2C_EnterRoom_GPSNotOpen,
					})
					return
				}
			}
			ctx.enter(r)
			return

		}
		person.WriteMsg(&msg.S2C_EnterRoom{
			Error:      msg.S2C_EnterRoom_Unknown,
			RoomNumber: m.RoomNumber,
		})
		return

	}

	if r, ok := room.GetRoom().RoomMap[m.RoomNumber]; ok {
		if !player.SystemOn {
			return
		}
		if r.Rule.GPSAntiCheat {
			if m.GPS {
				if util.CheckLocation(m.Location) {
					person.Location = m.Location
				} else {
					person.WriteMsg(&msg.S2C_EnterRoom{
						Error: msg.S2C_EnterRoom_LocationError,
					})
					return
				}
				ctx.enter(r)
				return
			} else {
				person.WriteMsg(&msg.S2C_EnterRoom{
					Error: msg.S2C_EnterRoom_GPSNotOpen,
				})
				return
			}
		}
		ctx.enter(r)
		return

	}
	person.WriteMsg(&msg.S2C_EnterRoom{
		Error:      msg.S2C_EnterRoom_NotCreated,
		RoomNumber: m.RoomNumber,
	})

}
