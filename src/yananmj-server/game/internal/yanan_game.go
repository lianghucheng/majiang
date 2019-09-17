package internal

import (
	"github.com/name5566/leaf/log"
	"time"
	"yananmj-server/common"
	"yananmj-server/game/mahjong"
	"yananmj-server/msg"
)

//玩家一次摸牌、出牌
func (yananRoom *YananRoom) drawAndDiscard(userID int) {
	if yananRoom.kongCount%2 == 0 && !yananRoom.rule.RedDragonJoker {
		if len(yananRoom.rests) == 20 {
			yananRoom.EndGame()
			return
		}
	} else if yananRoom.kongCount%2 == 1 && !yananRoom.rule.RedDragonJoker {
		if len(yananRoom.rests) == 19 {
			yananRoom.EndGame()
			return
		}
	} else {
		if len(yananRoom.rests) == 0 {
			yananRoom.EndGame()
			return
		}
	}

	playerData := yananRoom.userIDPlayerDatas[userID]
	//摸牌
	tile := yananRoom.draw(userID)

	if yananRoom.rule.RoomType == roomPrivate {
		if yananRoom.rule.MaxPlayers == 4 && playerData.dealer { // 跟庄
			discardsLen := len(yananRoom.discards)
			if discardsLen == 4 && mahjong.Quadruplet(yananRoom.discards[discardsLen-4:]) {
				dealerPlayerData := yananRoom.userIDPlayerDatas[yananRoom.dealerUserID]
				for i := 1; i < yananRoom.rule.MaxPlayers; i++ {
					otherUserID := yananRoom.positionUserIDs[(dealerPlayerData.position+i)%yananRoom.rule.MaxPlayers]
					otherPlayerData := yananRoom.userIDPlayerDatas[otherUserID]
					otherPlayerResult := otherPlayerData.roundResult
					otherPlayerResult.FollowDealerScore += yananRoom.rule.BaseScore
					dealerPlayerData.roundResult.FollowDealerScore -= yananRoom.rule.BaseScore
				}
			}
		}

	}

	playerData.claimActionCode = 0
	playerData.quadruplets = [][]int{}
	playerData.quadruplet = []int{}
	playerData.triplet = []int{}
	playerData.sequences = [][]int{}
	playerData.sequence = []int{}
	playerData.claim = -1
	yananRoom.claimUserID = -1
	log.Debug("检测 userID: %v 自摸: %v", userID, mahjong.ToTileString([]int{tile}))

	win, winType := playerData.analyzer.Win(playerData.hands, tile, true)
	if win { //判断玩家自摸
		playerData.claim = tile
		playerData.winType = winType
		playerData.claimActionCode |= mahjong.ActionWin
		yananRoom.actionWinUsers[userID] = 1
	}

	pongKong, quadruplet := playerData.analyzer.PongKong(playerData.claims, tile)
	if pongKong {
		playerData.claim = tile
		playerData.kongType = mahjong.PongKong
		playerData.quadruplets = [][]int{quadruplet}

		playerData.claimActionCode |= mahjong.ActionKong

		yananRoom.actionKongUsers[userID] = 1
	} else {

		temp := append([]int{}, playerData.hands...)
		temp = append(temp, tile)
		hiddenKong := false
		hiddenKong, playerData.quadruplets = playerData.analyzer.HiddenKong(temp, playerData.quadruplets)
		if hiddenKong { //判断玩家暗杠
			playerData.kongType = mahjong.HiddenKong
			playerData.claimActionCode |= mahjong.ActionKong
			yananRoom.actionKongUsers[userID] = 1
		}
	}

	if playerData.claimActionCode < 1 {
		yananRoom.discard(userID)
		return
	}

	playerData.state = yananActionClaim

	if user, ok := userIDUsers[userID]; ok {
		user.WriteMsg(&data_struct.S2C_ActionMahjongClaim{
			Position:    playerData.position,
			ActionCode:  playerData.claimActionCode,
			Countdown:   cd_yananClaim,
			Quadruplets: playerData.quadruplets,
		})
	}

	if playerData.managed {
		if user, ok := userIDUsers[userID]; ok {
			user.WriteMsg(&data_struct.S2C_ManagedMahjongPass{})
		}
	}

	playerData.actionTimestamp = time.Now().Unix()
	log.Debug("等待 userID %v 要牌", userID)
	yananRoom.claimTimer = skeleton.AfterFunc((cd_yananClaim+2)*time.Second, func() {
		switch yananRoom.rule.RoomType {
		case roomPractice, roomRoomCardMatch, roomPrivate, roomRedPacketMatching, roomRedPacketPrivate:
			yananRoom.discard(userID)
		}
	})
}

//玩家摸起一张牌(摸起的那张牌不会马上加入到手上)
func (yananRoom *YananRoom) draw(userID int) int {
	yananRoom.drawerUserID = userID

	tile := yananRoom.rests[0]
	yananRoom.rests = yananRoom.rests[1:]
	playerData := yananRoom.userIDPlayerDatas[userID]
	playerData.draw = tile
	if user, ok := userIDUsers[userID]; ok {
		user.WriteMsg(&data_struct.S2C_MahjongDraw{
			Position:      playerData.position,
			Tile:          tile,
			NumberOfHands: len(playerData.hands),
		})
	}
	broadcast(&data_struct.S2C_MahjongDraw{
		Position:      playerData.position,
		Tile:          -1,
		NumberOfHands: len(playerData.hands),
	}, yananRoom.positionUserIDs, playerData.position)

	broadcast(&data_struct.S2C_UpdateMahjongRestsNumber{
		NumberOfRests: len(yananRoom.rests),
	}, yananRoom.positionUserIDs, -1)
	log.Debug("剩余牌数: %v", len(yananRoom.rests))
	return tile
}

//玩家出一张牌
func (yananRoom *YananRoom) discard(userID int) {
	yananRoom.drawerUserID = -1

	playerData := yananRoom.userIDPlayerDatas[userID]
	playerData.state = yananActionDiscard

	broadcast(&data_struct.S2C_ActionMahjongDiscard{
		Position:  playerData.position,
		Countdown: cd_yananDiscard,
	}, yananRoom.positionUserIDs, -1)
	playerData.actionTimestamp = time.Now().Unix()

	if playerData.managed {
		yananRoom.doDiscard(userID, playerData.draw)
	} else {
		log.Debug("等待: %v 出牌", userID)
		yananRoom.discardTimer = skeleton.AfterFunc((cd_yananDiscard+2)*time.Second, func() {
			log.Debug("userID %v 自动出牌 %v", userID, mahjong.ToTileString([]int{playerData.draw}))
			yananRoom.doDiscard(userID, playerData.draw)
			playerData.discardsCount++
			if playerData.discardsCount == 2 {
				playerData.managed = true
				if user, ok := userIDUsers[userID]; ok {
					user.WriteMsg(&data_struct.S2C_MahjongManaged{
						Managed: true,
					})
				}
			}
		})
	}
}

//玩家要牌
func (yananRoom *YananRoom) claimOrDiscard(tile int, pos int) {
	nextUserID := yananRoom.positionUserIDs[(pos+1)%yananRoom.rule.MaxPlayers]
	if common.InArray(yananRoom.jokers, tile) { // 癞子打出来 其他人都不能要
		yananRoom.drawAndDiscard(nextUserID)
		return
	}
	doClaim := false
	yananRoom.resetActionClaimUsers()
	yananRoom.claimUserID = -1

	for i := 1; i < yananRoom.rule.MaxPlayers; i++ {
		userID := yananRoom.positionUserIDs[(pos+i)%yananRoom.rule.MaxPlayers]
		playerData := yananRoom.userIDPlayerDatas[userID]
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
		if !yananRoom.rule.MustSelfDraw {
			log.Debug("检测 userID %v 平胡 %v", userID, mahjong.ToTileString([]int{tile}))
			win, winType := playerData.analyzer.Win(playerData.hands, tile, false)
			if win {
				playerData.claim = tile
				playerData.winType = winType
				playerData.claimActionCode |= mahjong.ActionWin
				yananRoom.actionWinUsers[userID] = 1
			}
		}

		kong, quadruplet := playerData.analyzer.ExposedKong(playerData.hands, tile)
		if kong {
			playerData.claim = tile
			playerData.kongType = mahjong.ExposedKong
			playerData.quadruplets = [][]int{quadruplet}
			playerData.triplet = quadruplet[:3]

			playerData.claimActionCode |= mahjong.ActionKong
			playerData.claimActionCode |= mahjong.ActionPong

			yananRoom.actionKongUsers[userID] = 1
			yananRoom.actionPongUsers[userID] = 1
		} else {
			pong, triplet := playerData.analyzer.Pong(playerData.hands, tile)
			if pong {
				playerData.claim = tile
				playerData.triplet = triplet
				playerData.claimActionCode |= mahjong.ActionPong
				yananRoom.actionPongUsers[userID] = 1
			}
		}
	}

	for i := 1; i < yananRoom.rule.MaxPlayers; i++ {
		userID := yananRoom.positionUserIDs[(pos+i)%yananRoom.rule.MaxPlayers]
		playerData := yananRoom.userIDPlayerDatas[userID]
		if playerData.claimActionCode < 1 {
			continue
		}
		doClaim = true
		playerData.state = yananActionClaim

		if user, ok := userIDUsers[userID]; ok {
			user.WriteMsg(&data_struct.S2C_ActionMahjongClaim{
				Position:    playerData.position,
				ActionCode:  playerData.claimActionCode,
				Countdown:   cd_yananClaim,
				Quadruplets: playerData.quadruplets,
			})
		}

		playerData.actionTimestamp = time.Now().Unix()
		log.Debug("等待 userID: %v 要牌", userID)
	}

	if doClaim {
		yananRoom.claimTimer = skeleton.AfterFunc((cd_yananClaim+2)*time.Second, func() {
			yananRoom.resetActionClaimUsers()
			yananRoom.doClaim()
		})
	} else {
		yananRoom.drawAndDiscard(nextUserID)
	}

}

//是否轮到我胡牌
func (yananRoom *YananRoom) prepareWin(userID int) bool {
	playerData := yananRoom.userIDPlayerDatas[userID]
	if playerData.state != yananActionClaim || yananRoom.actionWinUsers[userID] == 0 {
		return false
	}
	yananRoom.winnerUserIDs = append(yananRoom.winnerUserIDs, userID)

	playerData.state = yananWin
	playerData.claimActionCode = 0
	if playerData.winType == mahjong.YananWinBySelfDraw { // 自摸
		yananRoom.claimUserID = userID
		yananRoom.resetActionClaimUsers()
		return true
	}

	yananRoom.deleteActionClaimUsers(userID)
	if yananRoom.actionClaimUsersEmpty() || len(yananRoom.actionWinUsers) == 0 {
		yananRoom.claimUserID = userID
		yananRoom.resetActionClaimUsers()
		return true
	}

	if yananRoom.claimUserID < 1 {
		yananRoom.claimUserID = userID
	} else {
		discarderPlayerData := yananRoom.userIDPlayerDatas[yananRoom.discarderUserID]
		playerRelativePos := toRelativePosition(playerData.position, discarderPlayerData.position, yananRoom.rule.MaxPlayers)
		claimerPlayerData := yananRoom.userIDPlayerDatas[yananRoom.claimUserID]
		if claimerPlayerData.state == yananWin {
			claimerRelativePos := toRelativePosition(claimerPlayerData.position, discarderPlayerData.position, yananRoom.rule.MaxPlayers)
			if playerRelativePos < claimerRelativePos {
				yananRoom.claimUserID = userID
			}
		} else {
			yananRoom.claimUserID = userID
		}
	}
	return false
}

//是否轮到扛牌
func (yananRoom *YananRoom) prepareKong(userID int, meld []int) bool {
	playerData := yananRoom.userIDPlayerDatas[userID]
	if playerData.state != yananActionClaim || yananRoom.actionKongUsers[userID] == 0 {
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
	playerData.state = yananKong
	playerData.claimActionCode = 0
	playerData.quadruplet = meld

	yananRoom.deleteActionClaimUsers(userID)
	if yananRoom.actionClaimUsersEmpty() || len(yananRoom.actionWinUsers) == 0 {
		yananRoom.claimUserID = userID
		yananRoom.resetActionClaimUsers()
		return true
	}

	if yananRoom.claimUserID < 1 {
		yananRoom.claimUserID = userID
	}
	return false
}

//是否轮到我碰牌
func (yananRoom *YananRoom) preparePong(userID int) bool {
	playerData := yananRoom.userIDPlayerDatas[userID]
	if yananRoom.actionPongUsers[userID] == 0 || playerData.state != yananActionClaim {
		return false
	}
	playerData.state = yananPong
	playerData.claimActionCode = 0
	yananRoom.deleteActionClaimUsers(userID)
	if yananRoom.actionClaimUsersEmpty() || len(yananRoom.actionWinUsers) == 0 {
		yananRoom.claimUserID = userID
		yananRoom.resetActionClaimUsers()
		return true
	}

	if yananRoom.claimUserID < 1 {
		yananRoom.claimUserID = userID
	}
	return false
}

// 玩家准备
func (yananRoom *YananRoom) doPrepare(userID int) {
	playerData := yananRoom.userIDPlayerDatas[userID]
	playerData.state = yananReady
	playerData.gun = 0

	yananRoom.refusedDisbandRoom(userID)

	broadcast(&data_struct.S2C_Prepare{
		Position: playerData.position,
		Ready:    true,
	}, yananRoom.positionUserIDs, -1)

	if yananRoom.allReady() {
		switch yananRoom.rule.RoomType {
		case roomPractice:
			delete(yananPracticeRooms, yananRoom.creatorUserID)
		case roomRoomCardMatch, roomRedPacketMatching:
			delete(yananRoomCardMatchRooms, yananRoom.creatorUserID)
		}
		if yananRoom.rule.Gun {
			yananRoom.state = roomGame
			for _, userID := range yananRoom.positionUserIDs {
				if playerData, ok := yananRoom.userIDPlayerDatas[userID]; ok {
					playerData.actionTimestamp = time.Now().Unix()
					playerData.state = yananActionSetGun
				}
			}
			broadcast(&data_struct.S2C_ActionSetGun{
				Countdown: cd_yananGun,
			}, yananRoom.positionUserIDs, -1)
			yananRoom.setGunTimer = skeleton.AfterFunc((cd_yananGun+2)*time.Second, func() {
				yananRoom.setGunTimer = nil
				yananRoom.autoSetGun()
			})
		} else {
			yananRoom.StartGame()
		}
	}
}

func (yananRoom *YananRoom) doClaim() {

	if yananRoom.claimUserID < 1 { //无人碰、杠、胡牌
		yananRoom.resetActionClaimUsers()
		if yananRoom.drawerUserID < 1 { //无人摸牌
			//下家摸牌、出牌
			discarderPlayerData := yananRoom.userIDPlayerDatas[yananRoom.discarderUserID]
			nextUserID := yananRoom.positionUserIDs[(discarderPlayerData.position+1)%yananRoom.rule.MaxPlayers]
			yananRoom.drawAndDiscard(nextUserID)
		} else {
			yananRoom.discard(yananRoom.drawerUserID)
		}
		return
	}

	claimPlayerData := yananRoom.userIDPlayerDatas[yananRoom.claimUserID]
	switch claimPlayerData.state {
	case yananWin:
		yananRoom.doWin()
	case yananPong:
		yananRoom.doPong()
	case yananKong:
		yananRoom.doKong()
	}

}

func (yananRoom *YananRoom) doDiscard(userID int, tile int) {
	playerData := yananRoom.userIDPlayerDatas[userID]
	if playerData.state != yananActionDiscard || playerData.draw < 0 {
		return
	}

	tiles := append(playerData.hands, playerData.draw)
	if common.Index(tiles, tile) == -1 { // tile 无效
		return
	}
	log.Debug("userID %v 出牌: %v", userID, mahjong.ToTileString([]int{tile}))
	if yananRoom.discardTimer != nil {
		yananRoom.discardTimer.Stop()
		yananRoom.discardTimer = nil
	}
	playerData.state = yananWaiting
	// 手牌增加一张
	playerData.hands = tiles
	playerData.draw = -1

	yananRoom.discarderUserID = userID
	//记录打出的牌
	yananRoom.discards = append(yananRoom.discards, tile)
	if common.Count(yananRoom.discards, tile) > 4 {
		log.Debug("%v 超过4张", mahjong.ToTileString([]int{tile}))
	}
	playerData.discards = append(playerData.discards, tile)
	//减少一张手牌
	playerData.hands = common.RemoveOnce(playerData.hands, tile)
	//排序
	playerData.analyzer.Analyze(playerData.hands, yananRoom.jokers)
	playerData.hands = playerData.analyzer.Sort()
	playerData.winTiles = playerData.analyzer.GetWinTiles(playerData.hands)
	if len(playerData.winTiles) > 0 {
		log.Debug("胡牌提示: %v", mahjong.ToTileString(playerData.winTiles))
	}
	broadcast(&data_struct.S2C_UpdateMahjongDiscads{
		Position: playerData.position,
		Discards: playerData.discards,
	}, yananRoom.positionUserIDs, -1)

	broadcast(&data_struct.S2C_UpdateMahjongDiscardCusor{
		Position: playerData.position,
		Index:    len(playerData.discards) - 1,
	}, yananRoom.positionUserIDs, -1)

	broadcast(&data_struct.S2C_MahjongDiscard{
		Position: playerData.position,
		Tile:     tile,
	}, yananRoom.positionUserIDs, -1)

	if user, ok := userIDUsers[userID]; ok {
		user.WriteMsg(&data_struct.S2C_UpdateMahjongHands{
			Position:      playerData.position,
			Hands:         playerData.hands,
			NumberOfHands: len(playerData.hands),
		})
		user.WriteMsg(&data_struct.S2C_UpdateWinTiles{
			Tiles: playerData.winTiles,
		})
	}

	broadcast(&data_struct.S2C_UpdateMahjongHands{
		Position:      playerData.position,
		NumberOfHands: len(playerData.hands),
	}, yananRoom.positionUserIDs, playerData.position)
	yananRoom.claimOrDiscard(tile, playerData.position)
}

//玩家胡牌
func (yananRoom *YananRoom) doWin() {
	numberOfWinner := len(yananRoom.winnerUserIDs)

	if numberOfWinner == 0 {
		return
	}

	yananRoom.resetActionClaimUsers()

	log.Debug("userID %v 胡", yananRoom.winnerUserIDs[0])
	playerData := yananRoom.userIDPlayerDatas[yananRoom.winnerUserIDs[0]]
	playerData.roundResult.WinType = playerData.winType

	switch playerData.roundResult.WinType {
	case mahjong.YananWinByDiscard: //平胡
		discarderPlayerData := yananRoom.userIDPlayerDatas[yananRoom.discarderUserID]
		discarderPlayerData.roundResult.WinType = mahjong.YananDiscard

		discarderPlayerData.discards = discarderPlayerData.discards[:len(discarderPlayerData.discards)-1]
		yananRoom.updateDiscards(yananRoom.discarderUserID)

		broadcast(&data_struct.S2C_UpdateMahjongDiscardCusor{
			Position: discarderPlayerData.position,
			Index:    -1,
		}, yananRoom.positionUserIDs, -1)

		//计算平胡得分1番
		if yananRoom.rule.Gun {
			yananRoom.calculateGunScore(yananRoom.winnerUserIDs[0])
		}
		yananRoom.calculateWinScore(yananRoom.winnerUserIDs[0])
	case mahjong.YananWinBySelfDraw: //自摸 2番 庄家再翻
		if yananRoom.rule.Gun {
			yananRoom.calculateGunScore(yananRoom.winnerUserIDs[0])
		}
		yananRoom.calculateWinScore(yananRoom.winnerUserIDs[0])
	}
	broadcast(&data_struct.S2C_MahjongWin{
		Position: playerData.position,
		WinType:  playerData.roundResult.WinType,
	}, yananRoom.positionUserIDs, -1)

	switch yananRoom.rule.RoomType {
	case roomPrivate, roomPractice, roomRoomCardMatch:
		//延时2秒发送比赛结果
		skeleton.AfterFunc(2*time.Second, func() {
			yananRoom.EndGame()
		})
	case roomRedPacketMatching, roomRedPacketPrivate:
		skeleton.AfterFunc(1*time.Second, func() {
			yananRoom.calculateRedPacket(yananRoom.winnerUserIDs[0], yananRoom.rule.RedPacketType)
			skeleton.AfterFunc(2*time.Second, func() {
				yananRoom.EndGame()
			})
		})
	}

}

//玩家杠牌
func (yananRoom *YananRoom) doKong() {
	playerData := yananRoom.userIDPlayerDatas[yananRoom.claimUserID]
	if playerData.state != yananKong {
		return
	}

	log.Debug("userID %v 杠 %v", yananRoom.claimUserID, mahjong.ToTileString(playerData.quadruplet))

	broadcast(&data_struct.S2C_MahjongKong{
		Position: playerData.position,
	}, yananRoom.positionUserIDs, -1)

	log.Debug("kongType:%v", playerData.kongType)
	switch playerData.kongType {
	case mahjong.ExposedKong: //明杠
		discarderPlayerData := yananRoom.userIDPlayerDatas[yananRoom.discarderUserID]
		discarderPlayerData.discards = discarderPlayerData.discards[:len(discarderPlayerData.discards)-1]
		yananRoom.updateDiscards(yananRoom.discarderUserID)

		broadcast(&data_struct.S2C_UpdateMahjongDiscardCusor{
			Position: discarderPlayerData.position,
			Index:    -1,
		}, yananRoom.positionUserIDs, -1)
		playerData.hands = append(playerData.hands, playerData.claim)
		playerData.hands = common.Remove(playerData.hands, playerData.quadruplet)
		//计算明杠得分
		playerData.roundResult.ExposedKongScore += 3 * yananRoom.rule.BaseScore
		discarderPlayerData.roundResult.ExposedKongScore -= 3 * yananRoom.rule.BaseScore

		yananRoom.kongCount += 1
	case mahjong.PongKong:
		//手牌增加一张
		playerData.hands = append(playerData.hands, playerData.draw)
		log.Debug("userID: %v 碰杠", yananRoom.claimUserID)
		playerData.claims = mahjong.RemoveTriplet(playerData.claims, playerData.claim)
		playerData.hands = common.RemoveOnce(playerData.hands, playerData.claim)
		//计算碰杠得分
		playerData.roundResult.PongKongScore += (yananRoom.rule.MaxPlayers - 1) * 1 * yananRoom.rule.BaseScore
		for i := 1; i < yananRoom.rule.MaxPlayers; i++ {
			otherUserId := yananRoom.positionUserIDs[(playerData.position+i)%yananRoom.rule.MaxPlayers]
			otherPlayerData := yananRoom.userIDPlayerDatas[otherUserId]
			otherPlayerData.roundResult.PongKongScore += -1 * yananRoom.rule.BaseScore
		}
		yananRoom.kongCount += 1
	case mahjong.HiddenKong: //暗杠
		//手牌增加一张
		playerData.hands = append(playerData.hands, playerData.draw)
		log.Debug("userID: %v 暗杠", yananRoom.claimUserID)
		playerData.hands = common.Remove(playerData.hands, playerData.quadruplet)

		//计算暗杠得分
		playerData.roundResult.HiddenKongScore += (yananRoom.rule.MaxPlayers - 1) * 2 * yananRoom.rule.BaseScore
		for i := 1; i < yananRoom.rule.MaxPlayers; i++ {
			otherUserId := yananRoom.positionUserIDs[(playerData.position+i)%yananRoom.rule.MaxPlayers]
			otherPlayerData := yananRoom.userIDPlayerDatas[otherUserId]
			otherPlayerData.roundResult.HiddenKongScore += -2 * yananRoom.rule.BaseScore
		}
		yananRoom.kongCount += 1
	}
	log.Debug("kongCount : %v", yananRoom.kongCount)

	//排序
	playerData.analyzer.Analyze(playerData.hands, yananRoom.jokers)
	playerData.hands = playerData.analyzer.Sort()

	playerData.winTiles = []int{}

	if user, ok := userIDUsers[yananRoom.claimUserID]; ok {
		user.WriteMsg(&data_struct.S2C_UpdateMahjongHands{
			Position:      playerData.position,
			Hands:         playerData.hands,
			NumberOfHands: len(playerData.hands),
		})

		user.WriteMsg(&data_struct.S2C_UpdateWinTiles{
			Tiles: playerData.winTiles,
		})
	}
	broadcast(&data_struct.S2C_UpdateMahjongHands{
		Position:      playerData.position,
		NumberOfHands: len(playerData.hands),
	}, yananRoom.positionUserIDs, playerData.position)
	if playerData.kongType == mahjong.HiddenKong {
		playerData.claims = append(playerData.claims, []int{-1, -1, -1, playerData.quadruplet[0]})
	} else {
		playerData.claims = append(playerData.claims, playerData.quadruplet)
	}
	yananRoom.updateClaims(yananRoom.claimUserID)

	playerData.claim = -1
	playerData.quadruplet = []int{}
	yananRoom.drawAndDiscard(yananRoom.claimUserID)
}

//玩家碰牌
func (yananRoom *YananRoom) doPong() {
	playerData := yananRoom.userIDPlayerDatas[yananRoom.claimUserID]
	if playerData.state != yananPong {
		return
	}

	broadcast(&data_struct.S2C_MahjongPong{
		Position: playerData.position,
	}, yananRoom.positionUserIDs, -1)

	discarderPlayerData := yananRoom.userIDPlayerDatas[yananRoom.discarderUserID]
	discarderPlayerData.discards = discarderPlayerData.discards[:len(discarderPlayerData.discards)-1]
	yananRoom.updateDiscards(yananRoom.discarderUserID)

	broadcast(&data_struct.S2C_UpdateMahjongDiscardCusor{
		Position: discarderPlayerData.position,
		Index:    -1,
	}, yananRoom.positionUserIDs, -1)

	playerData.hands = append(playerData.hands, playerData.claim)
	playerData.hands = common.Remove(playerData.hands, playerData.triplet)

	//排序
	playerData.analyzer.Analyze(playerData.hands, yananRoom.jokers)
	playerData.hands = playerData.analyzer.Sort()
	//排序完之后取最后一张作为自摸的牌
	playerData.draw = playerData.hands[len(playerData.hands)-1]
	playerData.hands = playerData.hands[:len(playerData.hands)-1]
	//排序
	playerData.analyzer.Analyze(playerData.hands, yananRoom.jokers)
	playerData.hands = playerData.analyzer.Sort()

	playerData.winTiles = []int{}

	playerData.claims = append(playerData.claims, playerData.triplet)
	yananRoom.updateClaims(yananRoom.claimUserID)

	playerData.triplet = []int{}
	playerData.claim = -1
	broadcast(&data_struct.S2C_UpdateMahjongHands{
		Position:      playerData.position,
		NumberOfHands: len(playerData.hands),
	}, yananRoom.positionUserIDs, playerData.position)

	broadcast(&data_struct.S2C_MahjongDraw{
		Position:      playerData.position,
		Tile:          -1,
		NumberOfHands: len(playerData.hands),
	}, yananRoom.positionUserIDs, playerData.position)

	if user, ok := userIDUsers[yananRoom.claimUserID]; ok {
		user.WriteMsg(&data_struct.S2C_UpdateMahjongHands{
			Position:      playerData.position,
			Hands:         playerData.hands,
			NumberOfHands: len(playerData.hands),
		})

		user.WriteMsg(&data_struct.S2C_MahjongDraw{
			Position:      playerData.position,
			Tile:          playerData.draw,
			NumberOfHands: len(playerData.hands),
		})

		user.WriteMsg(&data_struct.S2C_UpdateWinTiles{
			Tiles: playerData.winTiles,
		})
	}
	yananRoom.discard(yananRoom.claimUserID)
}

//玩家过牌
func (yananRoom *YananRoom) doPass(userID int) {
	playerData := yananRoom.userIDPlayerDatas[userID]
	if yananRoom.actionWinUsers[userID] == 0 && yananRoom.actionPongUsers[userID] == 0 &&
		yananRoom.actionKongUsers[userID] == 0 {
		playerData.state = yananWaiting
		playerData.claimActionCode = 0
		return
	}

	playerData.state = yananWaiting
	playerData.claimActionCode = 0

	yananRoom.deleteActionClaimUsers(userID)
	if yananRoom.actionClaimUsersEmpty() {
		yananRoom.resetActionClaimUsers()
		yananRoom.doClaim()
		return
	}

	if len(yananRoom.actionWinUsers) == 0 && yananRoom.claimUserID > 0 {
		claimerPlayerData := yananRoom.userIDPlayerDatas[yananRoom.claimUserID]
		switch claimerPlayerData.state {
		case yananWin:
			yananRoom.resetActionClaimUsers()
			yananRoom.doWin()
		case yananKong:
			yananRoom.resetActionClaimUsers()
			yananRoom.doKong()
		case yananPong:
			yananRoom.resetActionClaimUsers()
			yananRoom.doPong()
		}
	}
}

//玩家下炮子
func (yananRoom *YananRoom) doSetGun(userID int, gun int) {
	playerData := yananRoom.userIDPlayerDatas[userID]
	if playerData.state != yananActionSetGun {
		return
	}
	playerData.state = yananWaiting
	playerData.gun = gun
	broadcast(&data_struct.S2C_SetGun{
		Position: playerData.position,
		Gun:      playerData.gun,
	}, yananRoom.positionUserIDs, -1)
	if yananRoom.allSetGun() {
		if yananRoom.setGunTimer != nil {
			yananRoom.setGunTimer.Stop()
			yananRoom.setGunTimer = nil
		}
		yananRoom.StartGame()
	}
}

func (yananRoom *YananRoom) autoSetGun() {
	for _, userID := range yananRoom.positionUserIDs {
		if playerData, ok := yananRoom.userIDPlayerDatas[userID]; ok {
			if playerData.state == yananActionSetGun {
				playerData.state = yananWaiting
				playerData.gun = 0
				broadcast(&data_struct.S2C_SetGun{
					Position: playerData.position,
					Gun:      playerData.gun,
				}, yananRoom.positionUserIDs, -1)
			}
		}
	}
	yananRoom.StartGame()
}

//玩家取消托管
func (yananRoom *YananRoom) doCancelTrusteeship(userID int, managed bool) {
	playerData := yananRoom.userIDPlayerDatas[userID]
	playerData.discardsCount = 0
	playerData.managed = managed
	if user, ok := userIDUsers[userID]; ok {
		user.WriteMsg(&data_struct.S2C_MahjongManaged{
			Managed: playerData.managed,
		})
	}
}
