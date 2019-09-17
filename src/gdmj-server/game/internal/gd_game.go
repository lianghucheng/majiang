package internal

import (
	"gdmj-server/common"
	"gdmj-server/game/mahjong"
	"gdmj-server/msg"
	"time"

	"github.com/name5566/leaf/log"
)

// 玩家出一张牌
func (gdRoom *GDRoom) discard(userID int) {
	//gdRoom.drawerUserID = -1
	playerData := gdRoom.userIDPlayerDatas[userID]
	playerData.state = gdActionDiscard

	broadcast(&msg.S2C_ActionMahjongDiscard{
		Position:  playerData.position,
		Countdown: cd_gdDiscard,
	}, gdRoom.positionUserIDs, -1)

	playerData.actionTimestamp = time.Now().Unix()

	if playerData.managed {
		gdRoom.doDiscard(userID, playerData.draw)
	} else {
		log.Debug("等待 userID %v 出牌", userID)
		gdRoom.discardTimer = skeleton.AfterFunc((cd_gdDiscard+2)*time.Second, func() {
			log.Debug("userID %v 自动出牌 %v", userID, mahjong.ToTileString([]int{playerData.draw}))
			gdRoom.doDiscard(userID, playerData.draw)
			playerData.discardsCount++
			if playerData.discardsCount == 2 {
				playerData.managed = true
				if user, ok := userIDUsers[userID]; ok {
					user.WriteMsg(&msg.S2C_MahjongManaged{
						Managed: true,
					})
				}
			}
		})
	}
}

func (gdRoom *GDRoom) prepareChow(userID int, meld []int) bool {
	playerData := gdRoom.userIDPlayerDatas[userID]
	if playerData.state != gdActionClaim || gdRoom.actionChowUsers[userID] == 0 {
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

	playerData.state = gdChow
	playerData.claimActionCode = 0
	playerData.sequence = meld
	gdRoom.deleteActionClaimUsers(userID)
	if gdRoom.actionClaimUsersEmpty() {
		gdRoom.claimUserID = userID
		gdRoom.resetActionClaimUsers()
		return true
	}
	if gdRoom.claimUserID < 1 {
		gdRoom.claimUserID = userID
	}
	return false
}

/*




playerData := yananRoom.userIDPlayerDatas[userID]
	playerData.state = yananReady
	playerData.gun = 0

	yananRoom.refusedDisbandRoom(userID)

	broadcast(&msg.S2C_Prepare{
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
			broadcast(&msg.S2C_ActionSetGun{
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












*/
func (gdRoom *GDRoom) doPrepare(userID int) {
	playerData := gdRoom.userIDPlayerDatas[userID]
	playerData.state = gdReady
	playerData.gun = 0
	gdRoom.refuseDisbandRoom(userID)

	broadcast(&msg.S2C_Prepare{
		Position: playerData.position,
		Ready:    true,
	}, gdRoom.positionUserIDs, -1)
	if gdRoom.allReady() {
		switch gdRoom.rule.RoomType {
		case roomPractice:
			delete(gdPracticeRooms, gdRoom.creatorUserID)
		case roomRoomCardMatch, roomRedPacketMatching:
			delete(gdMatchRooms, gdRoom.creatorUserID)
		}
		//广东麻将
		if gdRoom.gameType== 1 {
			gdRoom.state = roomGame
			skeleton.AfterFunc(1*time.Second, func() {
				gdRoom.StartGame()
			})
		}
		//延安麻将
		if gdRoom.gameType == 2 {
			if gdRoom.rule.Gun {
				gdRoom.state = roomGame
				for _, userID := range gdRoom.positionUserIDs {
					if playerData, ok := gdRoom.userIDPlayerDatas[userID]; ok {
						playerData.actionTimestamp = time.Now().Unix()
						playerData.state = ActionSetGun
					}
				}
				broadcast(&msg.S2C_ActionSetGun{
					Countdown: cd_yananGun,
				}, gdRoom.positionUserIDs, -1)
				gdRoom.setGunTimer = skeleton.AfterFunc((cd_yananGun+2)*time.Second, func() {
					gdRoom.setGunTimer = nil
					gdRoom.autoSetGun()
				})
			} else {
				gdRoom.StartGame()
			}

		}
		if gdRoom.gameType==3{
			gdRoom.state = roomGame
			skeleton.AfterFunc(2*time.Second, func() {
				gdRoom.StartGame()
			})
		}
	}
}

func (gdRoom *GDRoom) doChow() {
	playerData := gdRoom.userIDPlayerDatas[gdRoom.claimUserID]
	if playerData.state != gdChow {
		return
	}
	broadcast(&msg.S2C_MahjongChow{
		Position: playerData.position,
	}, gdRoom.positionUserIDs, -1)

	discarderPlayerData := gdRoom.userIDPlayerDatas[gdRoom.discarderUserID]

	discarderPlayerData.discards = discarderPlayerData.discards[:len(discarderPlayerData.discards)-1]
	gdRoom.updateDiscards(gdRoom.discarderUserID)

	broadcast(&msg.S2C_UpdateMahjongDiscardCusor{
		Position: discarderPlayerData.position,
		Index:    -1,
	}, gdRoom.positionUserIDs, -1)

	playerData.hands = append(playerData.hands, playerData.claim)
	playerData.hands = common.Remove(playerData.hands, playerData.sequence)

	// 排序
	playerData.analyzer.Analyze(playerData.hands, gdRoom.jokers)
	playerData.hands = playerData.analyzer.Sort()
	// 排完序后取最后一张作为摸到的牌
	playerData.draw = playerData.hands[len(playerData.hands)-1]
	playerData.hands = playerData.hands[:len(playerData.hands)-1]
	// 排序
	playerData.analyzer.Analyze(playerData.hands, gdRoom.jokers)
	playerData.hands = playerData.analyzer.Sort()

	playerData.winTiles = []int{}

	playerData.claims = append(playerData.claims, playerData.sequence)
	gdRoom.updateClaims(gdRoom.claimUserID)

	playerData.sequence = []int{}
	playerData.claim = -1

	if user, ok := userIDUsers[gdRoom.claimUserID]; ok {
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
	}, gdRoom.positionUserIDs, playerData.position)

	broadcast(&msg.S2C_MahjongDraw{
		Position:      playerData.position,
		Tile:          -1,
		NumberOfHands: len(playerData.hands),
	}, gdRoom.positionUserIDs, playerData.position)
	gdRoom.discard(gdRoom.claimUserID)
}

func (gdRoom *GDRoom) doPass(userID int) {
	playerData := gdRoom.userIDPlayerDatas[userID]
	/*
		if playerData.state != gdActionClaim || gdRoom.actionWinUsers[userID] == 0 && gdRoom.actionKongUsers[userID] == 0 &&
			gdRoom.actionPongUsers[userID] == 0 {
			return
		}
	*/
	if playerData.state != gdActionClaim || gdRoom.existClaimUsers(userID) {
		return
	}
	passhu := false
	if _, ok := gdRoom.actionWinUsers[userID]; ok {
		passhu = ok
	}
	playerData.state = gdWaiting
	playerData.claimActionCode = 0
	gdRoom.deleteActionClaimUsers(userID)
	//过牌之后,没有别的玩家有吃,碰,杠,胡的状态了,他们都操作了
	if gdRoom.actionClaimUsersEmpty() {
		gdRoom.resetActionClaimUsers()
		gdRoom.doClaim()
		return
	}
	//如果出牌的下家过胡,则过胡人的下家可以直接胡牌
	if gdRoom.claimUserID > 0 && gdRoom.discarderUserID > 0 {
		claimerPlayerData := gdRoom.userIDPlayerDatas[gdRoom.claimUserID]
		discarderPlayerData := gdRoom.userIDPlayerDatas[gdRoom.discarderUserID]
		if claimerPlayerData.state == gdWin && passhu {
			index1 := toRelativePosition(playerData.position, discarderPlayerData.position, gdRoom.rule.MaxPlayers)
			index2 := toRelativePosition(claimerPlayerData.position, discarderPlayerData.position, gdRoom.rule.MaxPlayers)
			if index1 == 1 && index2 == 2 {
				gdRoom.resetActionClaimUsers()
				gdRoom.doClaim()
				return
			}

		}
	}
	if len(gdRoom.actionWinUsers) == 0 && gdRoom.claimUserID > 0 {
		claimerPlayerData := gdRoom.userIDPlayerDatas[gdRoom.claimUserID]
		switch claimerPlayerData.state {
		case gdWin:
			gdRoom.resetActionClaimUsers()
			gdRoom.doWin()
		case gdKong:
			gdRoom.resetActionClaimUsers()
			gdRoom.doKong()
		case gdPong:
			gdRoom.resetActionClaimUsers()
			gdRoom.doPong()
		}
	}
}

func (gdRoom *GDRoom) doCancelTrusteeship(userID int) {
	playerData := gdRoom.userIDPlayerDatas[userID]
	playerData.discardsCount = 0
	playerData.managed = false
	if user, ok := userIDUsers[userID]; ok {
		user.WriteMsg(&msg.S2C_MahjongManaged{
			Managed: false,
		})
	}
}

//玩家下炮子
func (gdRoom *GDRoom) doSetGun(userID int, gun int) {
	playerData := gdRoom.userIDPlayerDatas[userID]
	if playerData.state != ActionSetGun {
		return
	}
	playerData.state = gdWaiting
	playerData.gun = gun
	broadcast(&msg.S2C_SetGun{
		Position: playerData.position,
		Gun:      playerData.gun,
	}, gdRoom.positionUserIDs, -1)
	if gdRoom.allSetGun() {
		if gdRoom.setGunTimer != nil {
			gdRoom.setGunTimer.Stop()
			gdRoom.setGunTimer = nil
		}
		gdRoom.StartGame()
	}
}

func (gdRoom *GDRoom) autoSetGun() {
	for _, userID := range gdRoom.positionUserIDs {
		if playerData, ok := gdRoom.userIDPlayerDatas[userID]; ok {
			if playerData.state == ActionSetGun {
				playerData.state = gdWaiting
				playerData.gun = 0
				broadcast(&msg.S2C_SetGun{
					Position: playerData.position,
					Gun:      playerData.gun,
				}, gdRoom.positionUserIDs, -1)
			}
		}
	}
	gdRoom.StartGame()
}
