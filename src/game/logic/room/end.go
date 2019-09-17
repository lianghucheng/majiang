package room

import (
	"algorithm"
	"game/player"
	"game/room"
	msg "msg/room/mahjong"
	"time"
	"util"

	"github.com/name5566/leaf/log"
	"gopkg.in/mgo.v2/bson"
)

type EndControl struct {
	CreateControl
}

func EndGame(r *room.GDRoom) {
	log.Debug("游戏结束")
	r.EndTimestamp = time.Now().Unix()
	for _, userID := range r.PositionUserIDs {
		playerData := r.Useridplayerdatas[userID]
		playerData.WinTiles = make([]int, 0)
		person := player.GetPersonMgr().GetPerson(userID)
		if person == nil {
			continue
		}
		person.WriteMsg(&msg.S2C_UpdateWinTiles{
			Tiles: playerData.WinTiles,
		})
	}
	if r.Currentround == 1 && r.Rule.RoomType > 0 {
		CostCard(r)
	}
	record(r)
	roundResultsRecords := make([]msg.GDPlayerRoundResult, 0)
	settleMap := make(map[int]int, r.Rule.MaxPlayers)
	for _, userID := range r.PositionUserIDs {
		playerData := r.Useridplayerdatas[userID]
		roundResultsRecords = append(roundResultsRecords, msg.GDPlayerRoundResult{
			Nickname:         playerData.User.UserData.Nickname,
			Headimgurl:       playerData.User.UserData.Headimgurl,
			Dealer:           playerData.Dealer,
			Hands:            playerData.Hands,
			Claims:           playerData.Claims,
			LastTile:         playerData.RoundResult.LastTile,
			WinType:          playerData.RoundResult.WinType,
			WinScore:         playerData.RoundResult.WinScore,
			CatchHorseScore:  playerData.RoundResult.CatchHorseScore,
			ExposedKongScore: playerData.RoundResult.ExposedKongScore,
			PongKongScore:    playerData.RoundResult.PongKongScore,
			HiddenKongScore:  playerData.RoundResult.HiddenKongScore,
			TotalScore:       playerData.RoundResult.TotalScore,
			RoomCards:        playerData.RoundResult.RoomCards,
			RedPacket:        playerData.RoundResult.RedPacket,
		})
		settleMap[userID] = playerData.RoundResult.TotalScore
	}
	for _, userID := range r.PositionUserIDs {
		person := player.GetPersonMgr().GetPerson(userID)
		if person == nil {
			continue
		}
		result := 0
		if len(r.Winneruserids) == 0 {
			result = algorithm.ResultDraw
		} else {
			if util.InArray(r.Winneruserids, userID) {
				result = algorithm.ResultWin
			}
		}
		continueGame := true
		switch r.Rule.RoomType {
		case room.RoomRedPacketPrivate, room.RoomRedPacketMatching:
			continueGame = false
		case room.RoomPrivate:
			continueGame = !(r.Currentround == r.Rule.MaxRounds)
		}
		person.WriteMsg(&msg.S2C_GDRoundResult{
			Result:       result,
			RoomDesc:     r.Desc,
			Jokers:       r.Jokers,
			RoundResults: roundResultsRecords,
			ContinueGame: continueGame,
		})
	}
	if len(r.Winneruserids) > 0 {
		winnerPlayerData := r.Useridplayerdatas[r.Winneruserids[0]]
		winnerPlayerData.User.UserData.WinRounds += 1
		player.UpdateUserData(r.Winneruserids[0], bson.M{"$set": bson.M{"winrounds": winnerPlayerData.User.UserData.WinRounds}})

	}
	if r.Currentround < r.Rule.MaxRounds {
		r.Currentround++
		r.State = room.RoomGameEnd
		return
	}
	switch r.Rule.RoomType {
	case room.RoomRoomCardMatch:
		roomCardSettlement(r)
	case room.RoomPrivate:
		for _, userID := range r.PositionUserIDs {
			playerTotalResults := make([]msg.GDPlayerTotalResult, 0)
			playerData := r.Useridplayerdatas[userID]
			playerTotalResults = append(playerTotalResults, msg.GDPlayerTotalResult{
				Nickname:   playerData.User.UserData.Nickname,
				Headimgurl: playerData.User.UserData.Headimgurl,
				Owner:      playerData.Owner,
				AccountID:  playerData.User.UserData.AccountID,
				Scores:     playerData.TotalResult.Scores,
				TotalScore: playerData.TotalResult.TotalScore,
			})
			for i := 1; i < r.Rule.MaxPlayers; i++ {
				otherUserID := r.PositionUserIDs[(playerData.Position+i)%r.Rule.MaxPlayers]
				otherPlayerData := r.Useridplayerdatas[otherUserID]
				playerTotalResults = append(playerTotalResults, msg.GDPlayerTotalResult{
					Nickname:   otherPlayerData.User.UserData.Nickname,
					Headimgurl: otherPlayerData.User.UserData.Headimgurl,
					Owner:      otherPlayerData.Owner,
					AccountID:  otherPlayerData.User.UserData.AccountID,
					Scores:     otherPlayerData.TotalResult.Scores,
					TotalScore: otherPlayerData.TotalResult.TotalScore,
				})
			}
			person := player.GetPersonMgr().GetPerson(userID)
			if person != nil {
				person.WriteMsg(&msg.S2C_GDTotalResult{
					TotalResults: playerTotalResults,
				})
			}
			// 保存游戏积分
			playerData.User.UserData.GameScore += playerData.TotalResult.TotalScore
			// 用户玩的总局数
			playerData.User.UserData.TotalRounds++
			player.UpdateUserData(userID, bson.M{"$set": bson.M{"gamescore": playerData.User.UserData.GameScore}})
		}
	}
	room.GetRoom().DelRoom(r.Number)
}
