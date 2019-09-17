package internal

import (
	"gdmj-server/msg"

	"gdmj-server/game/mahjong"

	"time"

	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
)

func handleMahjongWin(args []interface{}) {
	m := args[0].(*msg.C2S_MahjongWin)
	_ = m
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if r, ok := userIDRooms[user.data.userData.UserID]; ok {
		user.doWin(r)
	}
}

func (user *User) doWin(r interface{}) {
	switch r.(type) {
	case *GDRoom:
		gdRoom := r.(*GDRoom)
		if gdRoom.state == roomGame && gdRoom.prepareWin(user.data.userData.UserID) {
			gdRoom.doWin()
		}
	}
}
func (gdRoom *GDRoom) prepareWin(userID int) bool {
	playerData := gdRoom.userIDPlayerDatas[userID]
	if playerData.state != gdActionClaim || gdRoom.actionWinUsers[userID] == 0 {
		return false
	}
	gdRoom.winnerUserIDs = append(gdRoom.winnerUserIDs, userID)
	playerData.state = gdWin
	playerData.claimActionCode = 0
	if playerData.state == mahjong.GDWinBySelfDraw {
		gdRoom.claimUserID = userID
		gdRoom.resetActionClaimUsers()
		return true
	}
	gdRoom.deleteActionClaimUsers(userID)
	discarderPlayerData := gdRoom.userIDPlayerDatas[gdRoom.discarderUserID]
	index := toRelativePosition(playerData.position, discarderPlayerData.position, gdRoom.rule.MaxPlayers)
	lastIndex := (discarderPlayerData.position + 1) % gdRoom.rule.MaxPlayers
	/*
				1:只有一个人胡,直接胡
				2:多人胡,按顺序第一个人点胡，直接胡
		        3:多人胡,按顺序第一人胡不了,第二人点胡直接胡
	*/
	if len(gdRoom.actionWinUsers) == 0 || index == 1 || index == 2 && gdRoom.actionChowUsers[lastIndex] == 0 {
		gdRoom.claimUserID = userID
		gdRoom.resetActionClaimUsers()
		return true
	}
	if gdRoom.claimUserID < 1 {
		gdRoom.claimUserID = userID
	} else {
		discarderPlayerData := gdRoom.userIDPlayerDatas[gdRoom.discarderUserID]
		playerRelativePos := toRelativePosition(playerData.position, discarderPlayerData.position, gdRoom.rule.MaxPlayers)

		claimerPlayerData := gdRoom.userIDPlayerDatas[gdRoom.claimUserID]
		if claimerPlayerData.state == gdWin {
			claimerRelativePos := toRelativePosition(claimerPlayerData.position, discarderPlayerData.position, gdRoom.rule.MaxPlayers)
			if playerRelativePos < claimerRelativePos {
				gdRoom.claimUserID = userID
			}
		} else {
			gdRoom.claimUserID = userID
		}
	}
	return false
}

func (gdRoom *GDRoom) doWin() {
	numberOfWinner := len(gdRoom.winnerUserIDs)
	if numberOfWinner == 0 {
		return
	}

	if numberOfWinner == 1 {
		log.Debug("userID: %v 胡", gdRoom.winnerUserIDs[0])

		playerData := gdRoom.userIDPlayerDatas[gdRoom.winnerUserIDs[0]]
		playerData.roundResult.WinType = playerData.winType
		if gdRoom.gameType== 1 {
			switch playerData.roundResult.WinType {
			case mahjong.GDWinByDiscard:

				discarderPlayerData := gdRoom.userIDPlayerDatas[gdRoom.discarderUserID]
				discarderPlayerData.roundResult.WinType = mahjong.GDDiscard

				discarderPlayerData.discards = discarderPlayerData.discards[:len(discarderPlayerData.discards)-1]
				gdRoom.updateDiscards(gdRoom.discarderUserID)

				broadcast(&msg.S2C_UpdateMahjongDiscardCusor{
					Position: discarderPlayerData.position,
					Index:    -1,
				}, gdRoom.positionUserIDs, -1)
			}
			gdRoom.calculateWinScore(gdRoom.winnerUserIDs[0])
			broadcast(&msg.S2C_MahjongWin{
				Position: playerData.position,
				WinType:  playerData.winType,
			}, gdRoom.positionUserIDs, -1)
		}

		if gdRoom.gameType == 2 {
			discarderPlayerData := gdRoom.userIDPlayerDatas[gdRoom.discarderUserID]
			discarderPlayerData.roundResult.WinType = mahjong.GDDiscard

			discarderPlayerData.discards = discarderPlayerData.discards[:len(discarderPlayerData.discards)-1]
			gdRoom.updateDiscards(gdRoom.discarderUserID)

			broadcast(&msg.S2C_UpdateMahjongDiscardCusor{
				Position: discarderPlayerData.position,
				Index:    -1,
			}, gdRoom.positionUserIDs, -1)

			//计算平胡得分1番
			if gdRoom.rule.Gun {
				gdRoom.calculateGunScore(gdRoom.winnerUserIDs[0])
			}
			gdRoom.calculateYanAnWinScore(gdRoom.winnerUserIDs[0])
		}
		if gdRoom.gameType==3{
			switch playerData.roundResult.WinType {
			case mahjong.HNZZWinByDiscard, mahjong.HNZZWinByEarthlyHand:
				discarderPlayerData := gdRoom.userIDPlayerDatas[gdRoom.discarderUserID]
				discarderPlayerData.roundResult.WinType = mahjong.HNZZDiscard

				discarderPlayerData.discards = discarderPlayerData.discards[:len(discarderPlayerData.discards)-1]
				gdRoom.updateDiscards(gdRoom.discarderUserID)

				broadcast(&msg.S2C_UpdateMahjongDiscardCusor{
					Position: discarderPlayerData.position,
					Index:    -1,
				}, gdRoom.positionUserIDs, -1)
			}
			gdRoom.calculateWinScore(gdRoom.winnerUserIDs[0])
			broadcast(&msg.S2C_MahjongWin{
				Position: playerData.position,
				WinType:  playerData.roundResult.WinType,
			}, gdRoom.positionUserIDs, -1)
		}

	} else {
		log.Debug("userID: %v 胡", gdRoom.winnerUserIDs[0])

		discarderPlayerData := gdRoom.userIDPlayerDatas[gdRoom.discarderUserID]
		discarderPlayerData.roundResult.WinType = mahjong.GDDiscard

		discarderPlayerData.discards = discarderPlayerData.discards[:len(discarderPlayerData.discards)-1]
		gdRoom.updateDiscards(gdRoom.discarderUserID)

		broadcast(&msg.S2C_UpdateMahjongDiscardCusor{
			Position: discarderPlayerData.position,
			Index:    -1,
		}, gdRoom.positionUserIDs, -1)

		winnerWinScore := 0
		for _, winnerUserID := range gdRoom.winnerUserIDs {
			winnerPlayerData := gdRoom.userIDPlayerDatas[winnerUserID]
			winnerPlayerData.roundResult.WinType = winnerPlayerData.winType
			gdRoom.calculateWinScore(winnerUserID)
			winnerWinScore += winnerPlayerData.roundResult.WinScore

			broadcast(&msg.S2C_MahjongWin{
				Position: winnerPlayerData.position,
				WinType:  winnerPlayerData.roundResult.WinType,
			}, gdRoom.positionUserIDs, -1)
		}
		discarderPlayerData.roundResult.WinScore = -1 * winnerWinScore
	}
	switch gdRoom.rule.RoomType {
	case roomPrivate:
		if gdRoom.gameType==3{
			tiles, birds := mahjong.CatchBird(gdRoom.rests, gdRoom.rule.Birds)
			gdRoom.calculateCatchBirdScore(len(birds))
			log.Debug("%v, 鸟: %v", mahjong.ToTileString(tiles), mahjong.ToTileString(birds))
			if len(tiles) > 0 {
				// 延时抓鸟
				skeleton.AfterFunc(2*time.Second, func() {
					broadcast(&msg.S2C_HNZZCatchBird{
						Position: gdRoom.userIDPlayerDatas[gdRoom.catchBirdUserID].position,
						Tiles:    tiles,
						Birds:    birds,
					}, gdRoom.positionUserIDs, -1)
					// 延时发送比赛结果
					skeleton.AfterFunc(2*time.Second, func() {
						gdRoom.EndGame()
					})
				})
			} else {
				// 延时发送比赛结果
				skeleton.AfterFunc(2*time.Second, func() {
					gdRoom.EndGame()
				})
			}
		}else{
			// 延时开马
			skeleton.AfterFunc(2*time.Second, func() {
				if gdRoom.gameType== 1 {
					gdRoom.catchHorse()
				}
				// 延迟发送比赛结果
				skeleton.AfterFunc(3*time.Second, func() {
					gdRoom.EndGame()
				})
			})
		}
	case roomPractice, roomRoomCardMatch:
		// 延时发送比赛结果
		skeleton.AfterFunc(2*time.Second, func() {
			gdRoom.EndGame()
		})
	case roomRedPacketPrivate, roomRedPacketMatching:
		skeleton.AfterFunc(1*time.Second, func() {
			gdRoom.calculateRedPacket(gdRoom.winnerUserIDs[0], gdRoom.rule.RedPacketType)
			skeleton.AfterFunc(2*time.Second, func() {
				gdRoom.EndGame()
			})
		})
	}
}
