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

type DrawControl struct {
	CreateControl
}

// 玩家依次摸牌、出牌
func (ctx *DrawControl) drawAndDiscard() {
	person := player.GetPersonMgr().GetPerson(ctx.uid)
	if person == nil {
		return
	}
	ri := room.GetRoomMgr().GetRoom(ctx.uid)
	if ri == nil {
		return
	}
	r := ri.(*room.GDRoom)
	if len(r.Rests) == 0 {
		if r.Rule.RoomType == room.RoomPrivate {
			game.Skeleton.AfterFunc(2*time.Second, func() {
				//gdRoom.catchHorse()
				// 延迟发送比赛结果
				game.Skeleton.AfterFunc(3*time.Second, func() {
					//gdRoom.EndGame()
				})
			})
			return
		}
		//gdRoom.EndGame()
		return
	}
	playerData := r.Useridplayerdatas[ctx.uid]
	tile := ctx.draw(r)

	playerData.ClaimActionCode = 0
	playerData.Quadruplets = [][]int{}
	playerData.Quadruplet = []int{}
	playerData.Triplet = []int{}
	playerData.Sequence = []int{}
	playerData.Sequences = [][]int{}
	playerData.Claim = -1
	r.Claimuserid = -1

	win, winType := playerData.Analyzer.Win(playerData.Hands, tile, true)
	if win { // 判断玩家自摸
		playerData.Claim = tile
		playerData.WinType = winType
		playerData.ClaimActionCode |= algorithm.ActionWin
		r.Actionwinusers[ctx.uid] = 1
	}
	//是否存在插杠
	/*

	   代码为什么没有检查,是否存在以前就有的插杠

	*/
	pongKong, quadruplet := playerData.Analyzer.PongKong(playerData.Claims, tile)
	if pongKong {
		playerData.Claim = tile
		playerData.KongType = algorithm.PongKong
		playerData.Quadruplets = append(playerData.Quadruplets, quadruplet)

		playerData.ClaimActionCode |= algorithm.ActionKong
		r.Actionkongusers[ctx.uid] = 1
	} else {
		temp := append([]int{}, playerData.Hands...)
		temp = append(temp, tile)

		hiddenKong := false
		hiddenKong, playerData.Quadruplets = playerData.Analyzer.HiddenKong(temp, playerData.Quadruplets)
		if hiddenKong {
			playerData.KongType = algorithm.HiddenKong

			playerData.ClaimActionCode |= algorithm.ActionKong

			r.Actionkongusers[ctx.uid] = 1
		}
	}
	if playerData.ClaimActionCode < 1 {
		ctx.automaicDiscardTimer(r)

		return
	}
	playerData.State = room.GdActionClaim
	person.WriteMsg(&msg.S2C_ActionMahjongClaim{
		Position:    playerData.Position,
		ActionCode:  playerData.ClaimActionCode,
		Countdown:   room.Cd_gdClaim,
		Sequences:   playerData.Sequences,
		Quadruplets: playerData.Quadruplets,
	})
	playerData.ActionTimestamp = time.Now().Unix()
	log.Debug("等待 userID %v 要牌", ctx.uid)
	r.Claimtimer = game.Skeleton.AfterFunc((room.Cd_gdClaim+2)*time.Second, func() {
		ctx.automaicDiscardTimer(r)
	})
}

func (ctx *DrawControl) draw(r *room.GDRoom) int {
	r.Draweruserid = ctx.uid

	tile := r.Rests[0]
	// 剩余的牌
	r.Rests = r.Rests[1:]
	playerData := r.Useridplayerdatas[ctx.uid]
	playerData.Draw = tile
	//检测玩家听牌
	cards := append([]int{}, playerData.Hands...)
	cards = append(cards, tile)
	depulicateCards := util.Deduplicate(cards)
	tingCards := make([]msg.TingCard, 0)
	for i := 0; i < len(depulicateCards); i++ {
		result := algorithm.TingCards(cards, depulicateCards[i], r.Jokers)
		if len(result) > 0 {
			log.Debug(" userid %v 打 %v 胡 %v", ctx.uid, algorithm.ToTileString([]int{depulicateCards[i]}), algorithm.ToTileString(result))
			tingCards = append(tingCards, msg.TingCard{depulicateCards[i], result})
		}
	}
	person := player.GetPersonMgr().GetPerson(ctx.uid)
	person.WriteMsg(&msg.S2C_MahjongDraw{
		Position:      playerData.Position,
		Tile:          tile,
		NumberOfHands: len(playerData.Hands),
	})
	person.WriteMsg(&msg.S2C_GDDisCardTing{tingCards, len(tingCards)})
	r.Broadcast(&msg.S2C_MahjongDraw{
		Position:      playerData.Position,
		Tile:          -1,
		NumberOfHands: len(playerData.Hands),
	}, playerData.Position)
	r.BroadcastAll(&msg.S2C_UpdateMahjongRestsNumber{
		NumberOfRests: len(r.Rests)})

	return tile
}

// 玩家出一张牌
func (ctx *DrawControl) automaicDiscardTimer(r *room.GDRoom) {
	playerData := r.Useridplayerdatas[ctx.uid]
	playerData.State = room.GdActionDiscard
	r.BroadcastAll(&msg.S2C_ActionMahjongDiscard{
		Position:  playerData.Position,
		Countdown: room.Cd_gdDiscard,
	})

	playerData.ActionTimestamp = time.Now().Unix()

	if playerData.Managed {
		disCtx := new(DiscradControl)
		disCtx.uid = ctx.uid
		disCtx.discrad(r, playerData.Draw)
	} else {
		log.Debug("等待 userID %v 出牌", ctx.uid)
		r.Discardtimer = game.Skeleton.AfterFunc((room.Cd_gdDiscard+2)*time.Second, func() {
			log.Debug("userID %v 自动出牌 %v", ctx.uid, algorithm.ToTileString([]int{playerData.Draw}))
			//gdRoom.doDiscard(userID, playerData.draw)
			playerData.DiscardsCount++
			if playerData.DiscardsCount == 2 {
				playerData.Managed = true
				if person := player.GetPersonMgr().GetPerson(ctx.uid); person != nil {
					person.WriteMsg(&msg.S2C_MahjongManaged{
						Managed: true,
					})
				}
			}
		})
	}
}
