package room

import (
	"algorithm"
	//"msg/room/mahjong"
	"time"
	"util"

	"github.com/name5566/leaf/log"
)

func (r *GDRoom) Ready() {
	// 洗牌
	if r.Rule.WithHonors {
		r.Tiles = util.Shuffle(algorithm.GDAllTiles)
	} else {
		r.Tiles = util.Shuffle(algorithm.GDAllTilesWithoutHonors)
	}

	// 确定庄家
	if r.Currentround == 1 {
		r.Dealeruserid = r.PositionUserIDs[0]
		switch r.Rule.RoomType {
		case RoomPrivate, RoomRoomCardMatch, RoomRedPacketMatching, RoomRedPacketPrivate:
			r.StartTimestamp = time.Now().Unix()
			r.EachRoundStartTimestamp = r.StartTimestamp
			r.initTotalResultData()
		}
	} else {
		switch r.Rule.RoomType {
		case RoomPrivate, RoomRoomCardMatch, RoomRedPacketMatching, RoomRedPacketPrivate:
			r.EachRoundStartTimestamp = time.Now().Unix()
		}
		if len(r.Winneruserids) > 0 {
			r.Dealeruserid = r.Winneruserids[0]
		}
	}
	// 庄家
	dealerPlayerData := r.Useridplayerdatas[r.Dealeruserid]
	dealerPlayerData.Dealer = true
	// 闲家
	dealerPosition := dealerPlayerData.Position
	for i := 1; i < r.Rule.MaxPlayers; i++ {
		playerPos := (dealerPosition + i) % r.Rule.MaxPlayers
		playerUserID := r.PositionUserIDs[playerPos]
		playerData := r.Useridplayerdatas[playerUserID]
		playerData.Dealer = false
	}
	if r.Rule.NeedJoker {
		r.Wildcard = r.Tiles[0]                      // 确定混儿
		r.Jokers = algorithm.GetGDJokers(r.Wildcard) // 确定宝牌
		log.Debug("混儿: %v, 宝牌: %v", algorithm.ToTileString([]int{r.Wildcard}), algorithm.ToTileString(r.Jokers))
		// 剩余的牌
		r.Rests = r.Tiles[1:]
	} else {
		// 剩余的牌
		r.Rests = append([]int{}, r.Tiles...)
	}
	r.Discards = []int{}

	r.Discarderuserid = -1
	r.Draweruserid = -1
	//r.resetActionClaimUsers()
	r.Claimuserid = -1
	r.Disbandapplicantuserid = -1
	r.Winneruserids = []int{}

	for _, userID := range r.PositionUserIDs {
		playerData := r.Useridplayerdatas[userID]
		initPlayer(playerData)

	}
}
