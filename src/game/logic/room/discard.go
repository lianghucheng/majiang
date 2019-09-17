package room

import (
	"algorithm"
	"game"
	"game/player"
	"game/room"
	msg "msg/room/mahjong"
	"time"
	"util"

	"github.com/name5566/leaf/log"
)

type DiscradControl struct {
	CreateControl
}

func init() {
	game.HandleRegister(msgReflect(&msg.C2S_MahjongDiscard{}), handleMahjongDiscard)
}
func handleMahjongDiscard(args []interface{}) {
	ctx := new(DiscradControl)
	label, person := ctx.userLegal(args[1])
	if !label {
		return
	}
	m := args[0].(*msg.C2S_MahjongDiscard)
	ctx.uid = person.UserData.UserID
	ri := room.GetRoomMgr().GetRoom(ctx.uid)
	if ri == nil {
		return
	}
	r := ri.(*room.GDRoom)
	if r.State != room.RoomGame {
		return
	}
	ctx.discrad(r, m.Tile)
}
func (ctx *DiscradControl) roomSet(r *room.GDRoom, tile int) {
	r.Discarderuserid = -1
	if r.Discardtimer != nil {
		r.Discardtimer.Stop()
		r.Discardtimer = nil
	}
	r.Discarderuserid = ctx.uid
	// 记录打出的牌
	r.Discards = append(r.Discards, tile)
}
func (ctx *DiscradControl) discrad(r *room.GDRoom, tile int) {
	person := player.GetPersonMgr().GetPerson(ctx.uid)
	if person == nil {
		return
	}
	playerData := r.Useridplayerdatas[ctx.uid]
	playerData.DiscardsCount = 0
	playerData.Managed = false
	if playerData.State != room.GdActionDiscard || playerData.Draw < 0 {
		return
	}
	tiles := append(playerData.Hands, playerData.Draw)
	if util.Index(tiles, tile) == -1 { // tile无效
		return
	}
	ctx.roomSet(r, tile)
	log.Debug("userID %v 出牌: %v", ctx.uid, algorithm.ToTileString([]int{tile}))
	playerData.State = room.GdWaiting
	// 手牌增加一张
	playerData.Draw = -1
	playerData.Discards = append(playerData.Discards, tile)
	// 手牌减少一张
	playerData.Hands = util.RemoveOnce(tiles, tile)
	// 排序
	playerData.Analyzer.Analyze(playerData.Hands, r.Jokers)
	//牌值自动把癞子,万,条,饼的顺序排放
	playerData.Hands = playerData.Analyzer.Sort()
	//听牌结果
	playerData.WinTiles = playerData.Analyzer.GetWinTiles(playerData.Hands)
	//更新玩家的出牌的牌墙
	r.BroadcastAll(&msg.S2C_UpdateMahjongDiscads{
		Position: playerData.Position,
		Discards: playerData.Discards})
	r.BroadcastAll(&msg.S2C_UpdateMahjongDiscardCusor{
		Position: playerData.Position,
		Index:    len(playerData.Discards) - 1})
	r.BroadcastAll(&msg.S2C_MahjongDiscard{
		Position: playerData.Position,
		Tile:     tile,
	})
	person.WriteMsg(&msg.S2C_UpdateMahjongHands{
		Position:      playerData.Position,
		Hands:         playerData.Hands,
		NumberOfHands: len(playerData.Hands),
	})
	person.WriteMsg(&msg.S2C_UpdateWinTiles{
		Tiles: playerData.WinTiles,
	})
	r.Broadcast(&msg.S2C_UpdateMahjongHands{
		Position:      playerData.Position,
		NumberOfHands: len(playerData.Hands),
	}, playerData.Position)

	//gdRoom.claimOrDiscard(tile, playerData.position)
}

//出完牌后，其他玩家是否有对应的操作
func (ctx *DiscradControl) broadOpeator(r *room.GDRoom, pos, tile int) {

	nextUserID := r.PositionUserIDs[(pos+1)%r.Rule.MaxPlayers]
	drawCtx := new(DrawControl)
	drawCtx.uid = nextUserID
	if util.InArray(r.Jokers, tile) { // 癞子打出来 其他人都不能要

		drawCtx.drawAndDiscard()
		return
	}
	doClaim := false
	r.ResetActionClaimUsers()
	r.Claimuserid = -1
	for i := 1; i < r.Rule.MaxPlayers; i++ {
		userID := r.PositionUserIDs[(pos+i)%r.Rule.MaxPlayers]
		playerData := r.Useridplayerdatas[userID]
		playerData.ClaimActionCode = 0

		playerData.Claim = -1
		playerData.Quadruplets = [][]int{}
		playerData.Quadruplet = []int{}
		playerData.Triplet = []int{}
		playerData.Sequence = []int{}
		playerData.Sequences = [][]int{}

		playerData.KongType = 0
		playerData.WinType = 0

		if playerData.Managed {
			continue
		}
		if !r.Rule.MustSelfDraw {
			win, winType := playerData.Analyzer.Win(playerData.Hands, tile, false)
			if win {
				playerData.Claim = tile
				playerData.WinType = winType
				//动作掩码
				playerData.ClaimActionCode |= algorithm.ActionWin

				r.Actionwinusers[userID] = 1
			}
		}
		//是否存在明杠
		kong, quadruplet := playerData.Analyzer.ExposedKong(playerData.Hands, tile)
		if kong {
			playerData.Claim = tile
			playerData.KongType = algorithm.ExposedKong
			//保存按顺序的可以杠碰的牌0 111 1 222 这种模式
			playerData.Quadruplets = append(playerData.Quadruplets, quadruplet)
			playerData.Triplet = append(playerData.Triplet, quadruplet[:3]...)

			playerData.ClaimActionCode |= algorithm.ActionKong
			playerData.ClaimActionCode |= algorithm.ActionPong

			r.Actionkongusers[userID] = 1
			r.Actionpongusers[userID] = 1
		} else {
			pong, triplet := playerData.Analyzer.Pong(playerData.Hands, tile)
			if pong {
				playerData.Claim = tile
				playerData.Triplet = append([]int{}, triplet...)

				playerData.ClaimActionCode |= algorithm.ActionPong

				r.Actionpongusers[userID] = 1
			}
		}
		if playerData.ClaimActionCode < 1 {
			continue
		}
		doClaim = true
		playerData.State = room.GdActionClaim
		person := player.GetPersonMgr().GetPerson(userID)
		if person != nil {
			person.WriteMsg(&msg.S2C_ActionMahjongClaim{
				Position:    playerData.Position,
				ActionCode:  playerData.ClaimActionCode,
				Countdown:   room.Cd_gdClaim,
				Sequences:   playerData.Sequences,
				Quadruplets: playerData.Quadruplets,
			})
		}
		playerData.ActionTimestamp = time.Now().Unix()
		log.Debug("等待 userID: %v 要牌", userID)

	}
	if doClaim {
		r.Claimtimer = game.Skeleton.AfterFunc((room.Cd_gdClaim+2)*time.Second, func() {
			r.ResetActionClaimUsers()
			ctx.cliamTimeout(r)
		})
	}
	drawCtx.drawAndDiscard()
}

func (ctx *DiscradControl) cliamTimeout(r *room.GDRoom) {
	if r.Claimuserid < 1 { // 无人吃、碰、杠、胡
		drawCtx := new(DrawControl)

		if r.Draweruserid < 1 {
			discarderPlayerData := r.Useridplayerdatas[r.Discarderuserid]
			nextUserID := r.PositionUserIDs[(discarderPlayerData.Position+1)%r.Rule.MaxPlayers]
			drawCtx.uid = nextUserID
			drawCtx.drawAndDiscard()
		} else {
			//选择过牌,该玩家出牌
			drawCtx.uid = r.Draweruserid
			drawCtx.automaicDiscardTimer(r)
		}
		return
	}
	claimerPlayerData := r.Useridplayerdatas[r.Claimuserid]
	switch claimerPlayerData.State {
	}
	/*
		case room.gdWin:
			gdRoom.doWin()
		case gdKong:
			gdRoom.doKong()
		case gdPong:
			gdRoom.doPong()
		case gdChow:
			gdRoom.doChow()
		}
	*/
}
