package internal

import (
	"gdmj-server/common"
	"gdmj-server/game/mahjong"
	"gdmj-server/msg"
	"time"

	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
)

func handleMahjongDiscard(args []interface{}) {
	m := args[0].(*msg.C2S_MahjongDiscard)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if r, ok := userIDRooms[user.data.userData.UserID]; ok {
		user.doDiscard(r, m.Tile)
	}
}

func (user *User) doDiscard(r interface{}, tile int) {
	switch r.(type) {
	case *GDRoom:
		gdRoom := r.(*GDRoom)
		if gdRoom.state == roomGame {
			playerData := gdRoom.userIDPlayerDatas[user.data.userData.UserID]
			// 托管计数清0
			playerData.discardsCount = 0
			playerData.managed = false
			gdRoom.doDiscard(user.data.userData.UserID, tile)
		}
	}
}

func (gdRoom *GDRoom) doDiscard(userID int, tile int) {
	playerData := gdRoom.userIDPlayerDatas[userID]
	if playerData.state != gdActionDiscard || playerData.draw < 0 {
		return
	}
	gdRoom.drawerUserID = -1
	tiles := append(playerData.hands, playerData.draw)
	if common.Index(tiles, tile) == -1 { // tile无效
		return
	}
	log.Debug("userID %v 出牌: %v", userID, mahjong.ToTileString([]int{tile}))
	if gdRoom.discardTimer != nil {
		gdRoom.discardTimer.Stop()
		gdRoom.discardTimer = nil
	}
	playerData.state = gdWaiting
	// 手牌增加一张
	playerData.hands = tiles
	playerData.draw = -1

	gdRoom.discarderUserID = userID
	// 记录打出的牌
	gdRoom.discards = append(gdRoom.discards, tile)
	if common.Count(gdRoom.discards, tile) > 4 {
		log.Error("剩余的牌: %v", mahjong.ToTileString(gdRoom.rests))
		log.Error("打出去的牌: %v", mahjong.ToTileString(gdRoom.discards))
		for _, theUserID := range gdRoom.positionUserIDs {
			thePlayerData := gdRoom.userIDPlayerDatas[theUserID]
			log.Error("玩家手牌: %v", thePlayerData.hands)
		}
		log.Error("%v 超过四张", mahjong.ToTileString([]int{tile}))

	}
	playerData.discards = append(playerData.discards, tile)
	// 手牌减少一张
	playerData.hands = common.RemoveOnce(playerData.hands, tile)
	// 排序
	playerData.analyzer.Analyze(playerData.hands, gdRoom.jokers)
	//牌值自动把癞子,万,条,饼的顺序排放
	playerData.hands = playerData.analyzer.Sort()
	//听牌结果
	playerData.winTiles = playerData.analyzer.GetWinTiles(playerData.hands)
	//更新玩家的出牌的牌墙
	broadcast(&msg.S2C_UpdateMahjongDiscads{
		Position: playerData.position,
		Discards: playerData.discards,
	}, gdRoom.positionUserIDs, -1)
	broadcast(&msg.S2C_UpdateMahjongDiscardCusor{
		Position: playerData.position,
		Index:    len(playerData.discards) - 1,
	}, gdRoom.positionUserIDs, -1)

	broadcast(&msg.S2C_MahjongDiscard{
		Position: playerData.position,
		Tile:     tile,
	}, gdRoom.positionUserIDs, -1)

	if user, ok := userIDUsers[userID]; ok {
		user.WriteMsg(&msg.S2C_UpdateMahjongHands{
			Position:      playerData.position,
			Hands:         playerData.hands,
			NumberOfHands: len(playerData.hands),
		})
		user.WriteMsg(&msg.S2C_UpdateWinTiles{
			Tiles: playerData.winTiles,
		})
	}
	broadcast(&msg.S2C_UpdateMahjongHands{
		Position:      playerData.position,
		NumberOfHands: len(playerData.hands),
	}, gdRoom.positionUserIDs, playerData.position)

	gdRoom.claimOrDiscard(tile, playerData.position)
}

// 检测其他玩家是否要牌，如果没人要牌则下家出牌
func (gdRoom *GDRoom) claimOrDiscard(tile int, pos int) {
	nextUserID := gdRoom.positionUserIDs[(pos+1)%gdRoom.rule.MaxPlayers]
	if common.InArray(gdRoom.jokers, tile) { // 癞子打出来 其他人都不能要
		gdRoom.drawAndDiscard(nextUserID)
		return
	}
	doClaim := false
	gdRoom.resetActionClaimUsers()
	gdRoom.claimUserID = -1
	for i := 1; i < gdRoom.rule.MaxPlayers; i++ {
		userID := gdRoom.positionUserIDs[(pos+i)%gdRoom.rule.MaxPlayers]
		playerData := gdRoom.userIDPlayerDatas[userID]
		playerData.claimActionCode = 0

		playerData.claim = -1
		playerData.quadruplets = [][]int{}
		playerData.quadruplet = []int{}
		playerData.triplet = []int{}
		playerData.sequence = []int{}
		playerData.sequences = [][]int{}

		playerData.kongType = 0
		playerData.winType = 0

		if playerData.managed {
			continue
		}
		if !gdRoom.rule.MustSelfDraw {
			win, winType := playerData.analyzer.Win(playerData.hands, tile, false)
			if win {
				playerData.claim = tile
				playerData.winType = winType
				//动作掩码
				playerData.claimActionCode |= mahjong.ActionWin

				gdRoom.actionWinUsers[userID] = 1
			}
		}
		//是否存在明杠
		kong, quadruplet := playerData.analyzer.ExposedKong(playerData.hands, tile)
		if kong {
			playerData.claim = tile
			playerData.kongType = mahjong.ExposedKong
			//保存按顺序的可以杠碰的牌0 111 1 222 这种模式
			playerData.quadruplets = append(playerData.quadruplets, quadruplet)
			playerData.triplet = append(playerData.triplet, quadruplet[:3]...)

			playerData.claimActionCode |= mahjong.ActionKong
			playerData.claimActionCode |= mahjong.ActionPong

			gdRoom.actionKongUsers[userID] = 1
			gdRoom.actionPongUsers[userID] = 1
		} else {
			pong, triplet := playerData.analyzer.Pong(playerData.hands, tile)
			if pong {
				playerData.claim = tile
				playerData.triplet = append([]int{}, triplet...)

				playerData.claimActionCode |= mahjong.ActionPong

				gdRoom.actionPongUsers[userID] = 1
			}
		}
		if playerData.claimActionCode < 1 {
			continue
		}
		doClaim = true
		playerData.state = gdActionClaim

		if user, ok := userIDUsers[userID]; ok {
			user.WriteMsg(&msg.S2C_ActionMahjongClaim{
				Position:    playerData.position,
				ActionCode:  playerData.claimActionCode,
				Countdown:   cd_gdClaim,
				Sequences:   playerData.sequences,
				Quadruplets: playerData.quadruplets,
			})
		}
		playerData.actionTimestamp = time.Now().Unix()
		log.Debug("等待 userID: %v 要牌", userID)

	}
	if doClaim {
		gdRoom.claimTimer = skeleton.AfterFunc((cd_gdClaim+2)*time.Second, func() {
			gdRoom.resetActionClaimUsers()
			gdRoom.doClaim()
		})
	} else {
		gdRoom.drawAndDiscard(nextUserID)
	}

}

func (gdRoom *GDRoom) doClaim() {
	if gdRoom.claimUserID < 1 { // 无人吃、碰、杠、胡
		//最近一次摸牌得人,没有人摸牌则下家抓牌出牌
		if gdRoom.drawerUserID < 1 {
			discarderPlayerData := gdRoom.userIDPlayerDatas[gdRoom.discarderUserID]
			nextUserID := gdRoom.positionUserIDs[(discarderPlayerData.position+1)%gdRoom.rule.MaxPlayers]
			gdRoom.drawAndDiscard(nextUserID)
		} else {
			//选择过牌,该玩家出牌
			gdRoom.discard(gdRoom.drawerUserID)
		}
		return
	}
	claimerPlayerData := gdRoom.userIDPlayerDatas[gdRoom.claimUserID]
	switch claimerPlayerData.state {
	case gdWin:
		gdRoom.doWin()
	case gdKong:
		gdRoom.doKong()
	case gdPong:
		gdRoom.doPong()
	case gdChow:
		gdRoom.doChow()
	}
}
