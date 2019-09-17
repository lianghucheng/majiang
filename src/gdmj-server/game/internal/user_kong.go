package internal

import (
	"gdmj-server/msg"
	"reflect"

	"gdmj-server/game/mahjong"

	"gdmj-server/common"

	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
)

func handleMahjongKong(args []interface{}) {
	m := args[0].(*msg.C2S_MahjongKong)
	a := args[1].(gate.Agent)
	log.Debug("meld:%v", m.Meld)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if r, ok := userIDRooms[user.data.userData.UserID]; ok {
		user.doKong(r, m.Meld)
	}
}

func (user *User) doKong(r interface{}, meld []int) {
	switch r.(type) {
	case *GDRoom:
		gdRoom := r.(*GDRoom)
		if gdRoom.state == roomGame && gdRoom.prepareKong(user.data.userData.UserID, meld) {
			gdRoom.doKong()
		}
	}
}

func (gdRoom *GDRoom) prepareKong(userID int, meld []int) bool {
	playerData := gdRoom.userIDPlayerDatas[userID]
	if playerData.state != gdActionClaim || gdRoom.actionKongUsers[userID] == 0 {
		return false
	}
	contain := false
	for _, v := range playerData.quadruplets {
		if reflect.DeepEqual(meld, v) {
			contain = true
			break
		}
	}
	if !contain {
		return false
	}
	playerData.state = gdKong
	playerData.claimActionCode = 0
	playerData.quadruplet = meld

	gdRoom.deleteActionClaimUsers(userID)
	if gdRoom.actionClaimUsersEmpty() || len(gdRoom.actionWinUsers) == 0 {
		gdRoom.claimUserID = userID
		gdRoom.resetActionClaimUsers()
		return true
	}
	if gdRoom.claimUserID < 1 {
		gdRoom.claimUserID = userID
	} else {
		claimerPlayerData := gdRoom.userIDPlayerDatas[gdRoom.claimUserID]
		if claimerPlayerData.state == gdChow {
			gdRoom.claimUserID = userID
		}
	}
	return false
}

func (gdRoom *GDRoom) doKong() {
	playerData := gdRoom.userIDPlayerDatas[gdRoom.claimUserID]
	if playerData.state != gdKong {
		return
	}
	log.Debug("userID %v 杠 %v", gdRoom.claimUserID, mahjong.ToTileString(playerData.quadruplet))
	//Todo 需添加明杠或者暗杠类型给前端
	broadcast(&msg.S2C_MahjongKong{
		Position: playerData.position,
	}, gdRoom.positionUserIDs, -1)

	switch playerData.kongType {
	case mahjong.ExposedKong:
		discarderPlayerData := gdRoom.userIDPlayerDatas[gdRoom.discarderUserID]

		discarderPlayerData.discards = discarderPlayerData.discards[:len(discarderPlayerData.discards)-1]
		//更新出牌人的牌面,剔除被杠的牌
		gdRoom.updateDiscards(gdRoom.discarderUserID)

		broadcast(&msg.S2C_UpdateMahjongDiscardCusor{
			Position: discarderPlayerData.position,
			Index:    -1,
		}, gdRoom.positionUserIDs, -1)

		playerData.hands = append(playerData.hands, playerData.claim)
		playerData.hands = common.Remove(playerData.hands, playerData.quadruplet)
		// 计算明杠得分(点杠一家3番)
		playerData.roundResult.ExposedKongScore += 3 * gdRoom.rule.BaseScore
		discarderPlayerData.roundResult.ExposedKongScore -= 3 * gdRoom.rule.BaseScore
	case mahjong.PongKong:
		// 手牌加一张
		playerData.hands = append(playerData.hands, playerData.draw)
		log.Debug("userID %v 碰杠", gdRoom.claimUserID)
		playerData.claims = mahjong.RemoveTriplet(playerData.claims, playerData.claim)
		playerData.hands = common.RemoveOnce(playerData.hands, playerData.claim)
		// 计算碰杠得分
		playerData.roundResult.PongKongScore += (gdRoom.rule.MaxPlayers - 1) * gdRoom.rule.BaseScore
		for i := 1; i < gdRoom.rule.MaxPlayers; i++ {
			otherUserID := gdRoom.positionUserIDs[(playerData.position+i)%gdRoom.rule.MaxPlayers]
			otherPlayerData := gdRoom.userIDPlayerDatas[otherUserID]
			otherPlayerData.roundResult.PongKongScore += -1 * gdRoom.rule.BaseScore
		}
	case mahjong.HiddenKong:
		// 手牌增加一张
		playerData.hands = append(playerData.hands, playerData.draw)
		log.Debug("userID %v 暗杠", gdRoom.claimUserID)
		playerData.hands = common.Remove(playerData.hands, playerData.quadruplet)
		// 计算暗杠得分
		playerData.roundResult.HiddenKongScore += (gdRoom.rule.MaxPlayers - 1) * 2 * gdRoom.rule.BaseScore
		for i := 1; i < gdRoom.rule.MaxPlayers; i++ {
			otherUserID := gdRoom.positionUserIDs[(playerData.position+i)%gdRoom.rule.MaxPlayers]
			otherPlayerData := gdRoom.userIDPlayerDatas[otherUserID]
			otherPlayerData.roundResult.HiddenKongScore += -2 * gdRoom.rule.BaseScore

		}
	}
	// 排序
	playerData.analyzer.Analyze(playerData.hands, gdRoom.jokers)
	playerData.hands = playerData.analyzer.Sort()

	playerData.winTiles = []int{}

	if user, ok := userIDUsers[gdRoom.claimUserID]; ok {
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
	if playerData.kongType == mahjong.HiddenKong {
		playerData.claims = append(playerData.claims, []int{-1, -1, -1, playerData.quadruplet[0]})
	} else {
		playerData.claims = append(playerData.claims, playerData.quadruplet)
	}
	gdRoom.updateClaims(gdRoom.claimUserID)

	playerData.claim = -1
	playerData.quadruplet = []int{}

	gdRoom.drawAndDiscard(gdRoom.claimUserID)
}
