package internal

import (
	"gdmj-server/common"
	"gdmj-server/msg"

	"github.com/name5566/leaf/gate"
)

func handleMahjongPong(args []interface{}) {
	m := args[0].(*msg.C2S_MahjongPong)
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
		user.doPong(r)
	}
}

func (user *User) doPong(r interface{}) {
	switch r.(type) {
	case *GDRoom:
		gdRoom := r.(*GDRoom)
		if gdRoom.state == roomGame && gdRoom.preparePong(user.data.userData.UserID) {
			gdRoom.doPong()
		}
	}
}

func (gdRoom *GDRoom) preparePong(userID int) bool {
	playerData := gdRoom.userIDPlayerDatas[userID]
	if playerData.state != gdActionClaim || gdRoom.actionPongUsers[userID] == 0 {
		return false
	}
	playerData.state = gdPong
	playerData.claimActionCode = 0

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

func (gdRoom *GDRoom) doPong() {
	playerData := gdRoom.userIDPlayerDatas[gdRoom.claimUserID]
	if playerData.state != gdPong {
		return
	}
	broadcast(&msg.S2C_MahjongPong{
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
	playerData.hands = common.Remove(playerData.hands, playerData.triplet)
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

	playerData.claims = append(playerData.claims, playerData.triplet)
	gdRoom.updateClaims(gdRoom.claimUserID)

	playerData.triplet = []int{}
	playerData.claim = -1

	broadcast(&msg.S2C_UpdateMahjongHands{
		Position:      playerData.position,
		NumberOfHands: len(playerData.hands),
	}, gdRoom.positionUserIDs, playerData.position)

	broadcast(&msg.S2C_MahjongDraw{
		Position:      playerData.position,
		Tile:          -1,
		NumberOfHands: len(playerData.hands),
	}, gdRoom.positionUserIDs, playerData.position)

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

	gdRoom.discard(gdRoom.claimUserID)
}
