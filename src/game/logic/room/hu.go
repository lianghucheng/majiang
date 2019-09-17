package room

import (
	"algorithm"
	"game"
	"game/room"
	msg "msg/room/mahjong"
	"time"

	"github.com/name5566/leaf/log"
)

type HuControl struct {
	CreateControl
}

func init() {
	game.HandleRegister(msgReflect(&msg.C2S_MahjongWin{}), handleMahjongWin)
}
func handleMahjongWin(args []interface{}) {
	ctx := new(HuControl)
	lable, person := ctx.userLegal(args[1])
	if !lable {
		return
	}
	ctx.uid = person.UserData.UserID
	ri := room.GetRoomMgr().GetRoom(ctx.uid)
	if ri == nil {
		return
	}
	r := ri.(*room.GDRoom)
	if r.State != room.RoomGame {
		return
	}
	if ctx.prepareWin(r) {
		//gdRoom.doWin()
	}
}

func (ctx *HuControl) prepareWin(r *room.GDRoom) bool {
	playerData := r.Useridplayerdatas[ctx.uid]
	if playerData.State != room.GdActionClaim || r.Actionwinusers[ctx.uid] == 0 {
		return false
	}
	r.Winneruserids = append(r.Winneruserids, ctx.uid)
	playerData.State = room.GdWin
	playerData.ClaimActionCode = 0
	if r.Draweruserid > 0 {
		r.Claimuserid = ctx.uid
		r.ResetActionClaimUsers()
		return true
	}
	r.DeleteActionClaimUsers(ctx.uid)
	discarderPlayerData := r.Useridplayerdatas[r.Discarderuserid]
	index := distance(playerData.Position, discarderPlayerData.Position, r.Rule.MaxPlayers)
	lastIndex := (discarderPlayerData.Position + 1) % r.Rule.MaxPlayers
	/*
				1:只有一个人胡,直接胡
				2:多人胡,按顺序第一个人点胡，直接胡
		        3:多人胡,按顺序第一人胡不了,第二人点胡直接胡
	*/
	if len(r.Actionwinusers) == 0 || index == 1 || index == 2 && r.Actionwinusers[r.PositionUserIDs[lastIndex]] == 0 {
		r.Claimuserid = ctx.uid
		r.ResetActionClaimUsers()
		return true
	}
	if r.Claimuserid < 1 {
		r.Claimuserid = ctx.uid
	} else {
		discarderPlayerData := r.Useridplayerdatas[r.Discarderuserid]
		playerRelativePos := distance(playerData.Position, discarderPlayerData.Position, r.Rule.MaxPlayers)

		claimerPlayerData := r.Useridplayerdatas[r.Claimuserid]
		if claimerPlayerData.State == room.GdWin {
			claimerRelativePos := distance(claimerPlayerData.Position, discarderPlayerData.Position, r.Rule.MaxPlayers)
			if playerRelativePos < claimerRelativePos {
				r.Claimuserid = ctx.uid
			}
		} else {
			r.Claimuserid = ctx.uid
		}
	}
	return false
}

func distance(pos int, zeroPos int, maxPlayers int) int {
	return (maxPlayers - zeroPos + pos) % maxPlayers
}

func (ctx *HuControl) doWin(r *room.GDRoom) {
	winners := len(r.Winneruserids)
	if winners == 0 {
		return
	}
	log.Debug("userID: %v 胡", r.Winneruserids[0])

	playerData := r.Useridplayerdatas[r.Winneruserids[0]]
	playerData.RoundResult.WinType = playerData.WinType

	switch playerData.WinType {
	case algorithm.GDWinByDiscard:
		discarderPlayerData := r.Useridplayerdatas[r.Discarderuserid]
		discarderPlayerData.RoundResult.WinType = algorithm.GDDiscard

		discarderPlayerData.Discards = discarderPlayerData.Discards[:len(discarderPlayerData.Discards)-1]
		//gdRoom.updateDiscards(gdRoom.discarderUserID)
		r.BroadcastAll(&msg.S2C_UpdateMahjongDiscardCusor{
			Position: discarderPlayerData.Position,
			Index:    -1,
		})
	}
	//gdRoom.calculateWinScore(gdRoom.winnerUserIDs[0])
	r.BroadcastAll(&msg.S2C_MahjongWin{
		Position: playerData.Position,
		WinType:  playerData.WinType,
	})
	r.BroadcastAll(&msg.S2C_MahjongWin{
		Position: playerData.Position,
		WinType:  playerData.WinType,
	})
	switch r.Rule.RoomType {
	case room.RoomPrivate:
		// 延时开马
		game.Skeleton.AfterFunc(2*time.Second, func() {
			catchHorse(r)
			// 延迟发送比赛结果
			game.Skeleton.AfterFunc(3*time.Second, func() {
				//gdRoom.EndGame()
			})
		})
	case room.RoomPractice, room.RoomRoomCardMatch:
		// 延时发送比赛结果
		game.Skeleton.AfterFunc(2*time.Second, func() {
			//gdRoom.EndGame()
		})
	case room.RoomRedPacketPrivate, room.RoomRedPacketMatching:
		game.Skeleton.AfterFunc(1*time.Second, func() {
			//gdRoom.calculateRedPacket(gdRoom.winnerUserIDs[0], gdRoom.rule.RedPacketType)
			game.Skeleton.AfterFunc(2*time.Second, func() {
				//gdRoom.EndGame()
			})
		})
	}
}

func catchHorse(r *room.GDRoom) {
	for _, userID := range r.PositionUserIDs {
		playerData := r.Useridplayerdatas[userID]
		// 计算玩家买马分数
		horseTiles := algorithm.Position(r.Dealeruserid, r.Winneruserids[0], r.Rule.MaxPlayers)
		_, ok := algorithm.CatchHorse(horseTiles, playerData.HorseTile)

		if ok {
			//gdRoom.calculateHorseScore(userID, horseCount)
		}
		r.BroadcastAll(&msg.S2C_GDBuyHorse{
			Position: playerData.Position,
			Tiles:    playerData.HorseTile,
		})
	}
}
