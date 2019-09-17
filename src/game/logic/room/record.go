package room

import (
	"game/room"
	msg "msg/room/mahjong"
	"util"
)

func record(r *room.GDRoom) {
	totalResults := make([]room.PlayerResultData, 0)
	roundResults := make([]room.PlayerResultData, 0)
	for pos := 0; pos < r.Rule.MaxPlayers; pos++ {
		userID := r.PositionUserIDs[pos]
		playerData := r.Useridplayerdatas[userID]
		// 计算总分
		roundResult := playerData.RoundResult
		roundResult.TotalScore = roundResult.WinScore + roundResult.CatchHorseScore + roundResult.ExposedKongScore +
			roundResult.PongKongScore + roundResult.HiddenKongScore
		if len(r.Winneruserids) == 0 {
			roundResult.LastTile = -1
			roundResult.RoomCards = 0
		} else {
			if util.InArray(r.Winneruserids, userID) {
				roundResult.LastTile = playerData.Claim
				if r.Rule.RoomType == room.RoomRoomCardMatch {
					roundResult.RoomCards = r.Rule.RoomCards * (r.Rule.MaxPlayers - 1)
				}
			} else {
				roundResult.LastTile = -1
				if r.Rule.RoomType == room.RoomRoomCardMatch {
					roundResult.RoomCards -= r.Rule.RoomCards
				}
			}
		}
		totalResult := playerData.TotalResult
		totalResult.Scores = append(totalResult.Scores, roundResult.TotalScore)
		totalResult.TotalScore += roundResult.TotalScore
		r.BroadcastAll(&msg.S2C_UpdateGDTotalScore{
			Position:   pos,
			TotalScore: totalResult.TotalScore})

		switch r.Rule.RoomType {
		case room.RoomPrivate, room.RoomRoomCardMatch:
			totalRoomCards := playerData.User.UserData.RoomCards + roundResult.RoomCards
			totalResults = append(totalResults, room.PlayerResultData{
				UserID:         playerData.User.UserData.UserID,
				Nickname:       playerData.User.UserData.Nickname,
				Score:          totalResult.TotalScore,
				RoomCards:      roundResult.RoomCards,
				RedPacketType:  r.Rule.RedPacketType,
				TotalRoomCards: totalRoomCards,
			})
			roundResults = append(roundResults, room.PlayerResultData{
				UserID:   playerData.User.UserData.UserID,
				Nickname: playerData.User.UserData.Nickname,
				Score:    roundResult.TotalScore,
			})
		}
	}

	// 保存总成绩
	switch r.Rule.RoomType {
	case room.RoomRedPacketPrivate, room.RoomRedPacketMatching:
		/*
			for _, userID := range r.PositionUserIDs {
				playerData := r.Useridplayerdatas[userID]
				saveRedPacketMatchResultData(&RedPacketMatchResultData{
					UserID:        userID,
					RedPacketType: gdRoom.rule.RedPacketType,
					RedPacket:     playerData.roundResult.RedPacket,
					Taken:         false,
					CreatedAt:     time.Now().Unix(),
				})
			}
		*/
	case room.RoomRoomCardMatch, room.RoomPrivate:
		r.SaveUserTotalResultData(totalResults)
		r.SaveUserRoundResultData(r.Currentround, roundResults)
	}
}
