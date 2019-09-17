package room

import (
	"algorithm"
	"game"
	"game/player"
	"game/room"
	msg "msg/room/mahjong"
	"reflect"
	"util"

	"github.com/name5566/leaf/log"
)

func init() {
	game.HandleRegister(msgReflect(&msg.C2S_MahjongKong{}), handleMahjongKong)
}

type KongControl struct {
	CreateControl
}

func handleMahjongKong(args []interface{}) {
	ctx := new(KongControl)
	label, person := ctx.userLegal(args[1])
	if !label {
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
}

func (ctx *KongControl) prepareKong(r *room.GDRoom, meld []int) bool {
	playerData := r.Useridplayerdatas[ctx.uid]
	if playerData.State != room.GdActionClaim || r.Actionkongusers[ctx.uid] == 0 {
		return false
	}
	contain := false
	for _, v := range playerData.Quadruplets {
		if reflect.DeepEqual(meld, v) {
			contain = true
			break
		}
	}
	if !contain {
		return false
	}
	playerData.State = room.GdKong
	playerData.ClaimActionCode = 0
	playerData.Quadruplet = meld
	r.DeleteActionClaimUsers(ctx.uid)
	if len(r.Actionwinusers) == 0 || r.ActionClaimUsersEmpty() {
		r.Claimuserid = ctx.uid
		r.ResetActionClaimUsers()
		return true
	}
	if r.Claimuserid < 1 {
		r.Claimuserid = ctx.uid
		return false
	}
	claimerPlayerData := r.Useridplayerdatas[r.Claimuserid]
	if claimerPlayerData.State == room.GdChow {
		r.Claimuserid = ctx.uid
		return false
	}
	return false
}

func (ctx *KongControl) doKong(r *room.GDRoom) {
	playerData := r.Useridplayerdatas[r.Claimuserid]
	log.Debug("userID %v 杠 %v", r.Claimuserid, algorithm.ToTileString(playerData.Quadruplet))
	//Todo 需添加明杠或者暗杠类型给前端
	r.BroadcastAll(&msg.S2C_MahjongKong{
		Position: playerData.Position})

	switch playerData.KongType {
	case algorithm.ExposedKong:
		discarderPlayerData := r.Useridplayerdatas[r.Discarderuserid]

		discarderPlayerData.Discards = discarderPlayerData.Discards[:len(discarderPlayerData.Discards)-1]
		//更新出牌人的牌面,剔除被杠的牌
		//gdRoom.updateDiscards(gdRoom.discarderUserID)
		r.BroadcastAll(&msg.S2C_UpdateMahjongDiscardCusor{
			Position: discarderPlayerData.Position,
			Index:    -1,
		})

		playerData.Hands = append(playerData.Hands, playerData.Claim)
		playerData.Hands = util.Remove(playerData.Hands, playerData.Quadruplet)
		// 计算明杠得分(点杠一家3番)
		playerData.RoundResult.ExposedKongScore += 3 * r.Rule.BaseScore
		discarderPlayerData.RoundResult.ExposedKongScore -= 3 * r.Rule.BaseScore
	case algorithm.PongKong:
		// 手牌加一张
		playerData.Hands = append(playerData.Hands, playerData.Draw)
		log.Debug("userID %v 碰杠", r.Claimuserid)
		playerData.Claims = algorithm.RemoveTriplet(playerData.Claims, playerData.Claim)
		playerData.Hands = util.RemoveOnce(playerData.Hands, playerData.Claim)
		// 计算碰杠得分
		playerData.RoundResult.PongKongScore += (r.Rule.MaxPlayers - 1) * r.Rule.BaseScore
		for i := 1; i < r.Rule.MaxPlayers; i++ {
			otherUserID := r.PositionUserIDs[(playerData.Position+i)%r.Rule.MaxPlayers]
			otherPlayerData := r.Useridplayerdatas[otherUserID]
			otherPlayerData.RoundResult.PongKongScore += -r.Rule.BaseScore
		}
	case algorithm.HiddenKong:
		// 手牌增加一张
		playerData.Hands = append(playerData.Hands, playerData.Draw)
		log.Debug("userID %v 暗杠", r.Claimuserid)
		playerData.Hands = util.Remove(playerData.Hands, playerData.Quadruplet)
		// 计算暗杠得分
		playerData.RoundResult.HiddenKongScore += (r.Rule.MaxPlayers - 1) * 2 * r.Rule.BaseScore
		for i := 1; i < r.Rule.MaxPlayers; i++ {
			otherUserID := r.PositionUserIDs[(playerData.Position+i)%r.Rule.MaxPlayers]
			otherPlayerData := r.Useridplayerdatas[otherUserID]
			otherPlayerData.RoundResult.HiddenKongScore += -2 * r.Rule.BaseScore

		}
	}
	// 排序
	playerData.Analyzer.Analyze(playerData.Hands, r.Jokers)
	playerData.Hands = playerData.Analyzer.Sort()

	playerData.WinTiles = []int{}
	person := player.GetPersonMgr().GetPerson(ctx.uid)
	if person != nil {
		person.WriteMsg(&msg.S2C_UpdateMahjongHands{
			Position:      playerData.Position,
			Hands:         playerData.Hands,
			NumberOfHands: len(playerData.Hands),
		})
		person.WriteMsg(&msg.S2C_UpdateWinTiles{
			Tiles: playerData.WinTiles,
		})
	}
	r.Broadcast(&msg.S2C_UpdateMahjongHands{
		Position:      playerData.Position,
		NumberOfHands: len(playerData.Hands)}, playerData.Position)

	if playerData.KongType == algorithm.HiddenKong {
		playerData.Claims = append(playerData.Claims, []int{-1, -1, -1, playerData.Quadruplet[0]})
	} else {
		playerData.Claims = append(playerData.Claims, playerData.Quadruplet)
	}
	r.BroadcastAll(&msg.S2C_UpdateMahjongClaims{
		Position: playerData.Position,
		Claims:   playerData.Claims,
	})

	playerData.Claim = -1
	playerData.Quadruplet = []int{}

	gdRoom.drawAndDiscard(gdRoom.claimUserID)
	drawCtx := new(DrawControl)
	drawCtx.uid = ctx.uid
	drawCtx.drawAndDiscard()
}
