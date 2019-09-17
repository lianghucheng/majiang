package internal

import (
	"gdmj-server/common"
	"gdmj-server/game/mahjong"
	"gdmj-server/msg"
	"time"

	"github.com/name5566/leaf/log"
)

// 玩家依次摸牌、出牌
func (gdRoom *GDRoom) drawAndDiscard(userID int) {
	if len(gdRoom.rests) == 0 {
		if gdRoom.rule.RoomType == roomPrivate {
			skeleton.AfterFunc(2*time.Second, func() {
				if gdRoom.gameType== 1 {
					gdRoom.catchHorse()
				}
				// 延迟发送比赛结果
				skeleton.AfterFunc(3*time.Second, func() {
					gdRoom.EndGame()
				})
				//湖南转转麻将此处无特殊处理
			})
			return
		}
		gdRoom.EndGame()
		return
	}
	playerData := gdRoom.userIDPlayerDatas[userID]
	tile := gdRoom.draw(userID)

	playerData.claimActionCode = 0
	playerData.quadruplets = [][]int{}
	playerData.quadruplet = []int{}
	playerData.triplet = []int{}
	playerData.sequence = []int{}
	playerData.sequences = [][]int{}
	playerData.claim = -1
	gdRoom.claimUserID = -1

	win, winType := playerData.analyzer.Win(playerData.hands, tile, true)
	if win { // 判断玩家自摸
		if gdRoom.gameType==3{
			if playerData.dealer && len(gdRoom.discards) == 0 { // 判断庄家天胡
				winType = mahjong.HNZZWinByHeavenlyHand
			}
		}
		playerData.claim = tile
		playerData.winType = winType
		playerData.claimActionCode |= mahjong.ActionWin
		gdRoom.actionWinUsers[userID] = 1
	}
	//是否存在插杠
	/*

	   代码为什么没有检查,是否存在以前就有的插杠

	*/
	pongKong, quadruplet := playerData.analyzer.PongKong(playerData.claims, tile)
	if pongKong {
		playerData.claim = tile
		playerData.kongType = mahjong.PongKong
		playerData.quadruplets = append(playerData.quadruplets, quadruplet)

		playerData.claimActionCode |= mahjong.ActionKong
		gdRoom.actionKongUsers[userID] = 1
	} else {
		temp := append([]int{}, playerData.hands...)
		temp = append(temp, tile)

		hiddenKong := false
		hiddenKong, playerData.quadruplets = playerData.analyzer.HiddenKong(temp, playerData.quadruplets)
		if hiddenKong {
			playerData.kongType = mahjong.HiddenKong

			playerData.claimActionCode |= mahjong.ActionKong

			gdRoom.actionKongUsers[userID] = 1
		}
	}
	if playerData.claimActionCode < 1 {
		gdRoom.discard(userID)
		return
	}
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
	log.Debug("等待 userID %v 要牌", userID)
	gdRoom.claimTimer = skeleton.AfterFunc((cd_gdClaim+2)*time.Second, func() {
		gdRoom.discard(userID)
	})
}

// 玩家摸牌
func (gdRoom *GDRoom) draw(userID int) int {
	gdRoom.drawerUserID = userID

	tile := gdRoom.rests[0]
	// 剩余的牌
	gdRoom.rests = gdRoom.rests[1:]
	playerData := gdRoom.userIDPlayerDatas[userID]
	playerData.draw = tile
	//检测玩家听牌
	cards := append([]int{}, playerData.hands...)
	cards = append(cards, tile)
	depulicateCards := common.Deduplicate(cards)
	tingCards := make([]msg.TingCard, 0)
	for i := 0; i < len(depulicateCards); i++ {
		result := mahjong.TingCards(cards, depulicateCards[i], gdRoom.jokers)
		if len(result) > 0 {
			log.Debug(" userid %v 打 %v 胡 %v", userID, mahjong.ToTileString([]int{depulicateCards[i]}), mahjong.ToTileString(result))
			tingCards = append(tingCards, msg.TingCard{depulicateCards[i], result})
		}
	}
	if user, ok := userIDUsers[userID]; ok {
		user.WriteMsg(&msg.S2C_MahjongDraw{
			Position:      playerData.position,
			Tile:          tile,
			NumberOfHands: len(playerData.hands),
		})
		user.WriteMsg(&msg.S2C_GDDisCardTing{tingCards, len(tingCards)})
	}
	broadcast(&msg.S2C_MahjongDraw{
		Position:      playerData.position,
		Tile:          -1,
		NumberOfHands: len(playerData.hands),
	}, gdRoom.positionUserIDs, playerData.position)

	broadcast(&msg.S2C_UpdateMahjongRestsNumber{
		NumberOfRests: len(gdRoom.rests),
	}, gdRoom.positionUserIDs, -1)

	return tile
}
