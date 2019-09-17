package internal

import (
	"github.com/name5566/leaf/log"
	"hnzzmj-server/common"
	"hnzzmj-server/game/mahjong"
	"hnzzmj-server/msg"
	"time"
)

// 玩家依次摸牌、出牌
func (hnzzRoom *HNZZRoom) drawAndDiscard(userID int) {
	if len(hnzzRoom.rests) == 0 {
		hnzzRoom.EndGame()
		return
	}
	playerData := hnzzRoom.userIDPlayerDatas[userID]
	tile := hnzzRoom.draw(userID)
	playerData.claimActionCode = 0
	playerData.quadruplets = [][]int{}
	playerData.quadruplet = []int{}
	playerData.triplet = []int{}
	playerData.sequence = []int{}
	playerData.sequences = [][]int{}
	playerData.claim = -1
	hnzzRoom.claimUserID = -1

	// log.Debug("检测 userID %v 自摸 %v", userID, mahjong.ToTileString([]int{tile}))
	win, winType := playerData.analyzer.Win(playerData.hands, tile, true)
	if win { // 判断玩家自摸
		if playerData.dealer && len(hnzzRoom.discards) == 0 { // 判断庄家天胡
			winType = mahjong.HNZZWinByHeavenlyHand
		}
		playerData.claim = tile
		playerData.winType = winType

		playerData.claimActionCode |= mahjong.ActionWin

		hnzzRoom.actionWinUsers[userID] = 1
	}

	pongKong, quadruplet := playerData.analyzer.PongKong(playerData.claims, tile)
	if pongKong {
		playerData.claim = tile
		playerData.kongType = mahjong.PongKong
		playerData.quadruplets = append(playerData.quadruplets, quadruplet)

		playerData.claimActionCode |= mahjong.ActionKong

		hnzzRoom.actionKongUsers[userID] = 1
	} else {
		temp := append([]int{}, playerData.hands...)
		temp = append(temp, tile)
		hiddenKong := false
		hiddenKong, playerData.quadruplets = playerData.analyzer.HiddenKong(temp, playerData.quadruplets)
		if hiddenKong { // 判断玩家暗杠
			playerData.kongType = mahjong.HiddenKong

			playerData.claimActionCode |= mahjong.ActionKong

			hnzzRoom.actionKongUsers[userID] = 1
		}
	}
	if playerData.claimActionCode < 1 {
		hnzzRoom.discard(userID)
		return
	}
	playerData.state = hnzzActionClaim
	if user, ok := userIDUsers[userID]; ok {
		user.WriteMsg(&msg.S2C_ActionMahjongClaim{
			Position:    playerData.position,
			ActionCode:  playerData.claimActionCode,
			Countdown:   cd_hnzzClaim,
			Quadruplets: playerData.quadruplets,
			Sequences:   playerData.sequences,
		})
	}
	playerData.actionTimestamp = time.Now().Unix()
	log.Debug("等待 userID %v 要牌", userID)
	hnzzRoom.claimTimer = skeleton.AfterFunc((cd_hnzzClaim+2)*time.Second, func() {
		hnzzRoom.discard(userID)
	})
}

// 玩家摸一张牌（摸起来的那张牌并不会马上加入到手牌中）
func (hnzzRoom *HNZZRoom) draw(userID int) int {
	hnzzRoom.drawerUserID = userID

	tile := hnzzRoom.rests[0]
	// 剩余的牌
	hnzzRoom.rests = hnzzRoom.rests[1:]

	playerData := hnzzRoom.userIDPlayerDatas[userID]
	playerData.draw = tile

	if user, ok := userIDUsers[userID]; ok {
		user.WriteMsg(&msg.S2C_MahjongDraw{
			Position:      playerData.position,
			Tile:          tile,
			NumberOfHands: len(playerData.hands),
		})
	}
	broadcast(&msg.S2C_MahjongDraw{
		Position:      playerData.position,
		Tile:          -1,
		NumberOfHands: len(playerData.hands),
	}, hnzzRoom.positionUserIDs, playerData.position)

	broadcast(&msg.S2C_UpdateMahjongRestsNumber{
		NumberOfRests: len(hnzzRoom.rests),
	}, hnzzRoom.positionUserIDs, -1)
	// log.Debug("剩余牌数: %v", len(hnzzRoom.rests))
	return tile
}

// 玩家出一张牌
func (hnzzRoom *HNZZRoom) discard(userID int) {
	hnzzRoom.drawerUserID = -1

	playerData := hnzzRoom.userIDPlayerDatas[userID]
	playerData.state = hnzzActionDiscard

	broadcast(&msg.S2C_ActionMahjongDiscard{
		Position:  playerData.position,
		Countdown: cd_hnzzDiscard,
	}, hnzzRoom.positionUserIDs, -1)

	playerData.actionTimestamp = time.Now().Unix()
	log.Debug("等待 userID %v 出牌", userID)
	hnzzRoom.discardTimer = skeleton.AfterFunc((cd_hnzzDiscard+2)*time.Second, func() {
		log.Debug("userID %v 自动出牌 %v", userID, mahjong.ToTileString([]int{playerData.draw}))
		hnzzRoom.doDiscard(userID, playerData.draw)
	})
}

// 检测其他玩家是否要牌，如果没人要牌则下家出牌
func (hnzzRoom *HNZZRoom) claimOrDiscard(tile int, pos int) {
	nextUserID := hnzzRoom.positionUserIDs[(pos+1)%hnzzRoom.rule.MaxPlayers]
	if common.InArray(hnzzRoom.jokers, tile) { // 癞子打出来都不能要
		hnzzRoom.drawAndDiscard(nextUserID)
		return
	}
	doClaim := false
	hnzzRoom.resetActionClaimUsers()
	hnzzRoom.claimUserID = -1

	for i := 1; i < hnzzRoom.rule.MaxPlayers; i++ {
		userID := hnzzRoom.positionUserIDs[(pos+i)%hnzzRoom.rule.MaxPlayers]
		playerData := hnzzRoom.userIDPlayerDatas[userID]
		playerData.claimActionCode = 0

		playerData.claim = -1
		playerData.quadruplets = [][]int{}
		playerData.quadruplet = []int{}
		playerData.triplet = []int{}
		playerData.sequence = []int{}
		playerData.sequences = [][]int{}
		playerData.kongType = 0
		playerData.winType = 0

		if !hnzzRoom.rule.MustSelfDraw {
			// log.Debug("检测 userID %v 平胡 %v", userID, mahjong.ToTileString([]int{tile}))
			win, winType := playerData.analyzer.Win(playerData.hands, tile, false)
			if win {
				if len(hnzzRoom.discards) == 1 { // 判断闲家地胡
					winType = mahjong.HNZZWinByEarthlyHand
				}
				playerData.claim = tile
				playerData.winType = winType

				playerData.claimActionCode |= mahjong.ActionWin

				hnzzRoom.actionWinUsers[userID] = 1
			}
		}
		kong, quadruplet := playerData.analyzer.ExposedKong(playerData.hands, tile)
		if kong {
			playerData.claim = tile
			playerData.kongType = mahjong.ExposedKong
			playerData.quadruplets = append(playerData.quadruplets, quadruplet)
			playerData.triplet = append(playerData.triplet, quadruplet[:3]...)

			playerData.claimActionCode |= mahjong.ActionKong
			playerData.claimActionCode |= mahjong.ActionPong

			hnzzRoom.actionKongUsers[userID] = 1
			hnzzRoom.actionPongUsers[userID] = 1
		} else {
			pong, triplet := playerData.analyzer.Pong(playerData.hands, tile)
			if pong {
				playerData.claim = tile
				playerData.triplet = triplet

				playerData.claimActionCode |= mahjong.ActionPong

				hnzzRoom.actionPongUsers[userID] = 1
			}
		}
	}
	for i := 1; i < hnzzRoom.rule.MaxPlayers; i++ {
		userID := hnzzRoom.positionUserIDs[(pos+i)%hnzzRoom.rule.MaxPlayers]
		playerData := hnzzRoom.userIDPlayerDatas[userID]
		if playerData.claimActionCode < 1 {
			continue
		}
		doClaim = true
		playerData.state = hnzzActionClaim

		if user, ok := userIDUsers[userID]; ok {
			user.WriteMsg(&msg.S2C_ActionMahjongClaim{
				Position:    playerData.position,
				ActionCode:  playerData.claimActionCode,
				Countdown:   cd_hnzzClaim,
				Sequences:   playerData.sequences,
				Quadruplets: playerData.quadruplets,
			})
		}
		playerData.actionTimestamp = time.Now().Unix()
		log.Debug("等待 userID %v 要牌", userID)
	}
	if doClaim {
		hnzzRoom.claimTimer = skeleton.AfterFunc((cd_hnzzClaim+2)*time.Second, func() {
			hnzzRoom.resetActionClaimUsers()
			hnzzRoom.doClaim()
		})
	} else {
		hnzzRoom.drawAndDiscard(nextUserID)
	}
}

func (hnzzRoom *HNZZRoom) prepareWin(userID int) bool {
	playerData := hnzzRoom.userIDPlayerDatas[userID]
	if playerData.state != hnzzActionClaim || hnzzRoom.actionWinUsers[userID] == 0 {
		return false
	}
	hnzzRoom.winnerUserIDs = append(hnzzRoom.winnerUserIDs, userID)

	playerData.state = hnzzWin
	playerData.claimActionCode = 0
	if playerData.winType == mahjong.HNZZWinBySelfDraw { // 自摸
		hnzzRoom.claimUserID = userID
		hnzzRoom.resetActionClaimUsers()
		return true
	}
	hnzzRoom.deleteActionClaimUsers(userID)
	if hnzzRoom.actionClaimUsersEmpty() || len(hnzzRoom.actionWinUsers) == 0 {
		hnzzRoom.claimUserID = userID
		hnzzRoom.resetActionClaimUsers()
		return true
	}
	if hnzzRoom.claimUserID < 1 {
		hnzzRoom.claimUserID = userID
	} else {
		discarderPlayerData := hnzzRoom.userIDPlayerDatas[hnzzRoom.discarderUserID]
		playerRelativePos := toRelativePosition(playerData.position, discarderPlayerData.position, hnzzRoom.rule.MaxPlayers)

		claimerPlayerData := hnzzRoom.userIDPlayerDatas[hnzzRoom.claimUserID]
		if claimerPlayerData.state == hnzzWin {
			claimerRelativePos := toRelativePosition(claimerPlayerData.position, discarderPlayerData.position, hnzzRoom.rule.MaxPlayers)
			if playerRelativePos < claimerRelativePos {
				hnzzRoom.claimUserID = userID
			}
		} else {
			hnzzRoom.claimUserID = userID
		}
	}
	return false
}

func (hnzzRoom *HNZZRoom) prepareKong(userID int, meld []int) bool {
	playerData := hnzzRoom.userIDPlayerDatas[userID]
	if playerData.state != hnzzActionClaim || hnzzRoom.actionKongUsers[userID] == 0 {
		return false
	}
	contain := false
	for _, v := range playerData.quadruplets {
		if common.Equal(meld, v) {
			contain = true
			break
		}
	}
	if !contain {
		return false
	}
	playerData.state = hnzzKong
	playerData.claimActionCode = 0
	playerData.quadruplet = meld

	hnzzRoom.deleteActionClaimUsers(userID)
	if hnzzRoom.actionClaimUsersEmpty() || len(hnzzRoom.actionWinUsers) == 0 {
		hnzzRoom.claimUserID = userID
		hnzzRoom.resetActionClaimUsers()
		return true
	}
	if hnzzRoom.claimUserID < 1 {
		hnzzRoom.claimUserID = userID
	} else {
		claimerPlayerData := hnzzRoom.userIDPlayerDatas[hnzzRoom.claimUserID]
		if claimerPlayerData.state == hnzzChow {
			hnzzRoom.claimUserID = userID
		}
	}
	return false
}

func (hnzzRoom *HNZZRoom) preparePong(userID int) bool {
	playerData := hnzzRoom.userIDPlayerDatas[userID]
	if playerData.state != hnzzActionClaim || hnzzRoom.actionPongUsers[userID] == 0 {
		return false
	}
	playerData.state = hnzzPong
	playerData.claimActionCode = 0

	hnzzRoom.deleteActionClaimUsers(userID)
	if hnzzRoom.actionClaimUsersEmpty() || len(hnzzRoom.actionWinUsers) == 0 {
		hnzzRoom.claimUserID = userID
		hnzzRoom.resetActionClaimUsers()
		return true
	}
	if hnzzRoom.claimUserID < 1 {
		hnzzRoom.claimUserID = userID
	} else {
		claimerPlayerData := hnzzRoom.userIDPlayerDatas[hnzzRoom.claimUserID]
		if claimerPlayerData.state == hnzzChow {
			hnzzRoom.claimUserID = userID
		}
	}
	return false
}

func (hnzzRoom *HNZZRoom) prepareChow(userID int, meld []int) bool {
	playerData := hnzzRoom.userIDPlayerDatas[userID]
	if playerData.state != hnzzActionClaim || hnzzRoom.actionChowUsers[userID] == 0 {
		return false
	}
	contain := false
	for _, v := range playerData.sequences {
		if common.Equal(meld, v) {
			contain = true
			break
		}
	}
	if !contain {
		return false
	}
	playerData.state = hnzzChow
	playerData.claimActionCode = 0
	playerData.sequence = meld

	hnzzRoom.deleteActionClaimUsers(userID)
	if hnzzRoom.actionClaimUsersEmpty() {
		hnzzRoom.claimUserID = userID
		hnzzRoom.resetActionClaimUsers()
		return true
	}
	if hnzzRoom.claimUserID < 1 {
		hnzzRoom.claimUserID = userID
	}
	return false
}

func (hnzzRoom *HNZZRoom) doPrepare(userID int) {
	playerData := hnzzRoom.userIDPlayerDatas[userID]
	playerData.state = hnzzReady

	hnzzRoom.refuseDisbandRoom(userID)

	broadcast(&msg.S2C_Prepare{
		Position: playerData.position,
		Ready:    true,
	}, hnzzRoom.positionUserIDs, -1)

	if hnzzRoom.allReady() {
		switch hnzzRoom.rule.RoomType {
		case roomRoomCardMatch, roomRedPacketMatching:
			delete(hnzzRoomCardMatchRooms, hnzzRoom.creatorUserID)
		}
		hnzzRoom.state = roomGame
		skeleton.AfterFunc(2*time.Second, func() {
			hnzzRoom.StartGame()
		})
	}
}

func (hnzzRoom *HNZZRoom) doClaim() {
	if hnzzRoom.claimUserID < 1 { // 无人吃、碰、杠、胡
		if hnzzRoom.drawerUserID < 1 { // 无人摸牌
			// 下家摸牌、出牌
			discarderPlayerData := hnzzRoom.userIDPlayerDatas[hnzzRoom.discarderUserID]
			nextUserID := hnzzRoom.positionUserIDs[(discarderPlayerData.position+1)%hnzzRoom.rule.MaxPlayers]
			hnzzRoom.drawAndDiscard(nextUserID)
		} else {
			hnzzRoom.discard(hnzzRoom.drawerUserID)
		}
		return
	}
	claimerPlayerData := hnzzRoom.userIDPlayerDatas[hnzzRoom.claimUserID]
	switch claimerPlayerData.state {
	case hnzzWin:
		hnzzRoom.doWin()
	case hnzzKong:
		hnzzRoom.doKong()
	case hnzzPong:
		hnzzRoom.doPong()
	case hnzzChow:
		hnzzRoom.doChow()
	}
}

func (hnzzRoom *HNZZRoom) doDiscard(userID int, tile int) {
	playerData := hnzzRoom.userIDPlayerDatas[userID]
	if playerData.state != hnzzActionDiscard || playerData.draw < 0 {
		return
	}
	tiles := append(playerData.hands, playerData.draw)
	if common.Index(tiles, tile) == -1 { // tile 无效
		return
	}
	log.Debug("userID %v 出牌: %v", userID, mahjong.ToTileString([]int{tile}))
	if hnzzRoom.discardTimer != nil {
		hnzzRoom.discardTimer.Stop()
		hnzzRoom.discardTimer = nil
	}
	playerData.state = hnzzWaiting
	// 手牌增加一张
	playerData.hands = tiles
	playerData.draw = -1

	hnzzRoom.discarderUserID = userID
	// 记录打出的牌
	hnzzRoom.discards = append(hnzzRoom.discards, tile)
	if common.Count(hnzzRoom.discards, tile) > 4 {
		log.Debug("%v 超过四张", mahjong.ToTileString([]int{tile}))
	}
	playerData.discards = append(playerData.discards, tile)
	// 手牌减少一张
	playerData.hands = common.RemoveOnce(playerData.hands, tile)
	// 排序
	playerData.analyzer.Analyze(playerData.hands)
	playerData.hands = playerData.analyzer.Sort(playerData.hands)
	playerData.winTiles = playerData.analyzer.GetWinTiles(playerData.hands)
	//if len(playerData.winTiles) > 0 {
	//	log.Debug("胡牌提示: %v", mahjong.ToTileString(playerData.winTiles))
	//}
	broadcast(&msg.S2C_UpdateMahjongDiscads{
		Position: playerData.position,
		Discards: playerData.discards,
	}, hnzzRoom.positionUserIDs, -1)

	broadcast(&msg.S2C_UpdateMahjongDiscardCusor{
		Position: playerData.position,
		Index:    len(playerData.discards) - 1,
	}, hnzzRoom.positionUserIDs, -1)

	broadcast(&msg.S2C_MahjongDiscard{
		Position: playerData.position,
		Tile:     tile,
	}, hnzzRoom.positionUserIDs, -1)

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
	}, hnzzRoom.positionUserIDs, playerData.position)

	hnzzRoom.claimOrDiscard(tile, playerData.position)
}

func (hnzzRoom *HNZZRoom) doWin() {
	numberOfWinner := len(hnzzRoom.winnerUserIDs)
	if numberOfWinner == 0 {
		return
	}
	if numberOfWinner == 1 {
		log.Debug("userID %v 胡", hnzzRoom.winnerUserIDs[0])
		hnzzRoom.catchBirdUserID = hnzzRoom.winnerUserIDs[0]

		playerData := hnzzRoom.userIDPlayerDatas[hnzzRoom.winnerUserIDs[0]]
		playerData.roundResult.WinType = playerData.winType
		switch playerData.roundResult.WinType {
		case mahjong.HNZZWinByDiscard, mahjong.HNZZWinByEarthlyHand:
			discarderPlayerData := hnzzRoom.userIDPlayerDatas[hnzzRoom.discarderUserID]
			discarderPlayerData.roundResult.WinType = mahjong.HNZZDiscard

			discarderPlayerData.discards = discarderPlayerData.discards[:len(discarderPlayerData.discards)-1]
			hnzzRoom.updateDiscards(hnzzRoom.discarderUserID)

			broadcast(&msg.S2C_UpdateMahjongDiscardCusor{
				Position: discarderPlayerData.position,
				Index:    -1,
			}, hnzzRoom.positionUserIDs, -1)
		}
		hnzzRoom.calculateWinScore(hnzzRoom.winnerUserIDs[0])
		broadcast(&msg.S2C_MahjongWin{
			Position: playerData.position,
			WinType:  playerData.roundResult.WinType,
		}, hnzzRoom.positionUserIDs, -1)
	} else {
		log.Debug("userIDs %v 胡", hnzzRoom.winnerUserIDs)
		hnzzRoom.catchBirdUserID = hnzzRoom.discarderUserID

		discarderPlayerData := hnzzRoom.userIDPlayerDatas[hnzzRoom.discarderUserID]
		discarderPlayerData.roundResult.WinType = mahjong.HNZZDiscard

		discarderPlayerData.discards = discarderPlayerData.discards[:len(discarderPlayerData.discards)-1]
		hnzzRoom.updateDiscards(hnzzRoom.discarderUserID)

		broadcast(&msg.S2C_UpdateMahjongDiscardCusor{
			Position: discarderPlayerData.position,
			Index:    -1,
		}, hnzzRoom.positionUserIDs, -1)

		winnerWinScore := 0
		for _, winnerUserID := range hnzzRoom.winnerUserIDs {
			winnerPlayerData := hnzzRoom.userIDPlayerDatas[winnerUserID]
			winnerPlayerData.roundResult.WinType = winnerPlayerData.winType
			hnzzRoom.calculateWinScore(winnerUserID)
			winnerWinScore += winnerPlayerData.roundResult.WinScore

			broadcast(&msg.S2C_MahjongWin{
				Position: winnerPlayerData.position,
				WinType:  winnerPlayerData.roundResult.WinType,
			}, hnzzRoom.positionUserIDs, -1)
		}
		discarderPlayerData.roundResult.WinScore = -1 * winnerWinScore
	}
	switch hnzzRoom.rule.RoomType {
	case roomRoomCardMatch:
		// 延时发送比赛结果
		skeleton.AfterFunc(2*time.Second, func() {
			hnzzRoom.EndGame()
		})
	case roomPrivate:
		tiles, birds := mahjong.CatchBird(hnzzRoom.rests, hnzzRoom.rule.Birds)
		hnzzRoom.calculateCatchBirdScore(len(birds))
		log.Debug("%v, 鸟: %v", mahjong.ToTileString(tiles), mahjong.ToTileString(birds))
		if len(tiles) > 0 {
			// 延时抓鸟
			skeleton.AfterFunc(2*time.Second, func() {
				broadcast(&msg.S2C_HNZZCatchBird{
					Position: hnzzRoom.userIDPlayerDatas[hnzzRoom.catchBirdUserID].position,
					Tiles:    tiles,
					Birds:    birds,
				}, hnzzRoom.positionUserIDs, -1)
				// 延时发送比赛结果
				skeleton.AfterFunc(2*time.Second, func() {
					hnzzRoom.EndGame()
				})
			})
		} else {
			// 延时发送比赛结果
			skeleton.AfterFunc(2*time.Second, func() {
				hnzzRoom.EndGame()
			})
		}
	case roomRedPacketMatching, roomRedPacketPrivate:
		skeleton.AfterFunc(1*time.Second, func() {
			hnzzRoom.calculateRedPacket(hnzzRoom.winnerUserIDs[0], hnzzRoom.rule.RedPacketType)
			skeleton.AfterFunc(2*time.Second, func() {
				hnzzRoom.EndGame()
			})
		})
	}
}

func (hnzzRoom *HNZZRoom) doKong() {
	playerData := hnzzRoom.userIDPlayerDatas[hnzzRoom.claimUserID]
	if playerData.state != hnzzKong {
		return
	}
	log.Debug("userID %v 杠 %v", hnzzRoom.claimUserID, mahjong.ToTileString(playerData.quadruplet))
	broadcast(&msg.S2C_MahjongKong{
		Position: playerData.position,
	}, hnzzRoom.positionUserIDs, -1)

	switch playerData.kongType {
	case mahjong.ExposedKong: // 明杠
		discarderPlayerData := hnzzRoom.userIDPlayerDatas[hnzzRoom.discarderUserID]

		discarderPlayerData.discards = discarderPlayerData.discards[:len(discarderPlayerData.discards)-1]
		hnzzRoom.updateDiscards(hnzzRoom.discarderUserID)

		broadcast(&msg.S2C_UpdateMahjongDiscardCusor{
			Position: discarderPlayerData.position,
			Index:    -1,
		}, hnzzRoom.positionUserIDs, -1)

		playerData.hands = append(playerData.hands, playerData.claim)
		playerData.hands = common.Remove(playerData.hands, playerData.quadruplet)
		// 计算明杠得分(点杠一家3番)
		playerData.roundResult.ExposedKongScore += 3 * hnzzRoom.rule.BaseScore
		discarderPlayerData.roundResult.ExposedKongScore += -3 * hnzzRoom.rule.BaseScore
	case mahjong.PongKong:
		// 手牌增加一张
		playerData.hands = append(playerData.hands, playerData.draw)
		log.Debug("userID %v 碰杠", hnzzRoom.claimUserID)
		playerData.claims = mahjong.RemoveTriplet(playerData.claims, playerData.claim)
		playerData.hands = common.RemoveOnce(playerData.hands, playerData.claim)
		// 计算碰杠得分(其他几家1番)
		playerData.roundResult.PongKongScore += (hnzzRoom.rule.MaxPlayers - 1) * hnzzRoom.rule.BaseScore
		for i := 1; i < hnzzRoom.rule.MaxPlayers; i++ {
			otherUserID := hnzzRoom.positionUserIDs[(playerData.position+i)%hnzzRoom.rule.MaxPlayers]
			otherPlayerData := hnzzRoom.userIDPlayerDatas[otherUserID]
			otherPlayerData.roundResult.PongKongScore += -1 * hnzzRoom.rule.BaseScore
		}
	case mahjong.HiddenKong:
		// 手牌增加一张
		playerData.hands = append(playerData.hands, playerData.draw)
		log.Debug("userID %v 暗杠", hnzzRoom.claimUserID)
		playerData.hands = common.Remove(playerData.hands, playerData.quadruplet)
		// 计算暗杠得分(其他几家2番)
		playerData.roundResult.HiddenKongScore += (hnzzRoom.rule.MaxPlayers - 1) * 2 * hnzzRoom.rule.BaseScore
		for i := 1; i < hnzzRoom.rule.MaxPlayers; i++ {
			otherUserID := hnzzRoom.positionUserIDs[(playerData.position+i)%hnzzRoom.rule.MaxPlayers]
			otherPlayerData := hnzzRoom.userIDPlayerDatas[otherUserID]
			otherPlayerData.roundResult.HiddenKongScore += -2 * hnzzRoom.rule.BaseScore
		}
	}
	// 排序
	playerData.analyzer.Analyze(playerData.hands)
	playerData.hands = playerData.analyzer.Sort(playerData.hands)

	playerData.winTiles = []int{}

	if user, ok := userIDUsers[hnzzRoom.claimUserID]; ok {
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
	}, hnzzRoom.positionUserIDs, playerData.position)

	if playerData.kongType == mahjong.HiddenKong {
		playerData.claims = append(playerData.claims, []int{-1, -1, -1, playerData.quadruplet[0]})
	} else {
		playerData.claims = append(playerData.claims, playerData.quadruplet)
	}
	hnzzRoom.updateClaims(hnzzRoom.claimUserID)

	playerData.claim = -1
	playerData.quadruplet = []int{}

	hnzzRoom.drawAndDiscard(hnzzRoom.claimUserID)
}

func (hnzzRoom *HNZZRoom) doPong() {
	playerData := hnzzRoom.userIDPlayerDatas[hnzzRoom.claimUserID]
	if playerData.state != hnzzPong {
		return
	}
	broadcast(&msg.S2C_MahjongPong{
		Position: playerData.position,
	}, hnzzRoom.positionUserIDs, -1)

	discarderPlayerData := hnzzRoom.userIDPlayerDatas[hnzzRoom.discarderUserID]

	discarderPlayerData.discards = discarderPlayerData.discards[:len(discarderPlayerData.discards)-1]
	hnzzRoom.updateDiscards(hnzzRoom.discarderUserID)

	broadcast(&msg.S2C_UpdateMahjongDiscardCusor{
		Position: discarderPlayerData.position,
		Index:    -1,
	}, hnzzRoom.positionUserIDs, -1)

	playerData.hands = append(playerData.hands, playerData.claim)
	playerData.hands = common.Remove(playerData.hands, playerData.triplet)
	// 排序
	playerData.analyzer.Analyze(playerData.hands)
	playerData.hands = playerData.analyzer.Sort(playerData.hands)
	// 排完序后取最后一张作为摸到的牌
	playerData.draw = playerData.hands[len(playerData.hands)-1]
	playerData.hands = playerData.hands[:len(playerData.hands)-1]
	// 排序
	playerData.analyzer.Analyze(playerData.hands)
	playerData.hands = playerData.analyzer.Sort(playerData.hands)

	playerData.winTiles = []int{}

	playerData.claims = append(playerData.claims, playerData.triplet)
	hnzzRoom.updateClaims(hnzzRoom.claimUserID)

	playerData.triplet = []int{}
	playerData.claim = -1

	broadcast(&msg.S2C_UpdateMahjongHands{
		Position:      playerData.position,
		NumberOfHands: len(playerData.hands),
	}, hnzzRoom.positionUserIDs, playerData.position)

	broadcast(&msg.S2C_MahjongDraw{
		Position:      playerData.position,
		Tile:          -1,
		NumberOfHands: len(playerData.hands),
	}, hnzzRoom.positionUserIDs, playerData.position)

	if user, ok := userIDUsers[hnzzRoom.claimUserID]; ok {
		user.WriteMsg(&msg.S2C_UpdateMahjongHands{
			Position:      playerData.position,
			Hands:         playerData.hands,
			NumberOfHands: len(playerData.hands),
		})
		user.WriteMsg(&msg.S2C_MahjongDraw{
			Position:      playerData.position,
			Tile:          playerData.draw,
			NumberOfHands: len(playerData.hands),
		})
		user.WriteMsg(&msg.S2C_UpdateWinTiles{
			Tiles: playerData.winTiles,
		})
	}

	hnzzRoom.discard(hnzzRoom.claimUserID)
}

func (hnzzRoom *HNZZRoom) doChow() {
	playerData := hnzzRoom.userIDPlayerDatas[hnzzRoom.claimUserID]
	if playerData.state != hnzzChow {
		return
	}
	broadcast(&msg.S2C_MahjongChow{
		Position: playerData.position,
	}, hnzzRoom.positionUserIDs, -1)

	discarderPlayerData := hnzzRoom.userIDPlayerDatas[hnzzRoom.discarderUserID]

	discarderPlayerData.discards = discarderPlayerData.discards[:len(discarderPlayerData.discards)-1]
	hnzzRoom.updateDiscards(hnzzRoom.discarderUserID)

	broadcast(&msg.S2C_UpdateMahjongDiscardCusor{
		Position: discarderPlayerData.position,
		Index:    -1,
	}, hnzzRoom.positionUserIDs, -1)

	playerData.hands = append(playerData.hands, playerData.claim)
	playerData.hands = common.Remove(playerData.hands, playerData.sequence)
	// 排序
	playerData.analyzer.Analyze(playerData.hands)
	playerData.hands = playerData.analyzer.Sort(playerData.hands)
	// 排完序后取最后一张作为摸到的牌
	playerData.draw = playerData.hands[len(playerData.hands)-1]
	playerData.hands = playerData.hands[:len(playerData.hands)-1]
	// 排序
	playerData.analyzer.Analyze(playerData.hands)
	playerData.hands = playerData.analyzer.Sort(playerData.hands)

	playerData.winTiles = []int{}

	playerData.claims = append(playerData.claims, playerData.sequence)
	hnzzRoom.updateClaims(hnzzRoom.claimUserID)

	playerData.sequence = []int{}
	playerData.claim = -1

	if user, ok := userIDUsers[hnzzRoom.claimUserID]; ok {
		user.WriteMsg(&msg.S2C_UpdateMahjongHands{
			Position:      playerData.position,
			Hands:         playerData.hands,
			NumberOfHands: len(playerData.hands),
		})
		user.WriteMsg(&msg.S2C_MahjongDraw{
			Position:      playerData.position,
			Tile:          playerData.draw,
			NumberOfHands: len(playerData.hands),
		})
		user.WriteMsg(&msg.S2C_UpdateWinTiles{
			Tiles: playerData.winTiles,
		})
	}
	broadcast(&msg.S2C_UpdateMahjongHands{
		Position:      playerData.position,
		NumberOfHands: len(playerData.hands),
	}, hnzzRoom.positionUserIDs, playerData.position)

	broadcast(&msg.S2C_MahjongDraw{
		Position:      playerData.position,
		Tile:          -1,
		NumberOfHands: len(playerData.hands),
	}, hnzzRoom.positionUserIDs, playerData.position)

	hnzzRoom.discard(hnzzRoom.claimUserID)
}

func (hnzzRoom *HNZZRoom) doPass(userID int) {
	playerData := hnzzRoom.userIDPlayerDatas[userID]
	if playerData.state != hnzzActionClaim || hnzzRoom.actionWinUsers[userID] == 0 && hnzzRoom.actionKongUsers[userID] == 0 &&
		hnzzRoom.actionPongUsers[userID] == 0 && hnzzRoom.actionChowUsers[userID] == 0 {
		return
	}
	playerData.state = hnzzWaiting
	playerData.claimActionCode = 0

	hnzzRoom.deleteActionClaimUsers(userID)
	if hnzzRoom.actionClaimUsersEmpty() {
		hnzzRoom.resetActionClaimUsers()
		hnzzRoom.doClaim()
		return
	}
	if len(hnzzRoom.actionWinUsers) == 0 && hnzzRoom.claimUserID > 0 {
		claimerPlayerData := hnzzRoom.userIDPlayerDatas[hnzzRoom.claimUserID]
		switch claimerPlayerData.state {
		case hnzzWin:
			hnzzRoom.resetActionClaimUsers()
			hnzzRoom.doWin()
		case hnzzKong:
			hnzzRoom.resetActionClaimUsers()
			hnzzRoom.doKong()
		case hnzzPong:
			hnzzRoom.resetActionClaimUsers()
			hnzzRoom.doPong()
		}
	}
}
