package room

import (
	"game/player"
	"game/room"
	msg "msg/room/mahjong"

	"algorithm"

	"github.com/name5566/leaf/log"
)

func StartGame(r *room.GDRoom) {
	r.Ready()
	r.BroadcastAll(&msg.S2C_GameStart{})

	r.BroadcastAll(&msg.S2C_UpdateMahjongCurrentRound{
		CurrentRound: r.Currentround,
	})

	dealerPlayerData := r.Useridplayerdatas[r.Dealeruserid]
	r.BroadcastAll(&msg.S2C_DecideDealer{
		Position: dealerPlayerData.Position,
	})

	if r.Rule.NeedJoker {
		if len(r.Jokers) == 0 {
			log.Error("joker is nil: %v", r.Jokers)
		}
		r.BroadcastAll(&msg.S2C_DecideGDJoker{
			WildCard: r.Wildcard,
			Jokers:   r.Jokers,
		})
	}
	// 所有玩家发13张牌
	for _, userID := range r.PositionUserIDs {
		playerData := r.Useridplayerdatas[userID]
		playerData.State = room.GdWaiting
		// 手牌13张
		playerData.Hands = append(playerData.Hands, r.Rests[:13]...)
		// 排序
		playerData.Analyzer.Analyze(playerData.Hands, r.Jokers)
		playerData.Hands = playerData.Analyzer.Sort()
		log.Debug("userID %v 手牌: %v", userID, algorithm.ToTileString(playerData.Hands))
		// 获取可以胡的牌
		playerData.WinTiles = playerData.Analyzer.GetWinTiles(playerData.Hands)
		if len(playerData.WinTiles) > 0 {
			log.Debug("胡牌提示: %v", algorithm.ToTileString(playerData.WinTiles))
		}
		r.Rests = r.Rests[13:]
		// 所有玩家生成马牌
		if r.Rule.BuyHorse == 1 {
			playerData.HorseTile = append(playerData.HorseTile, r.Rests[:1]...)
			log.Debug("马牌: %v", algorithm.ToTileString(playerData.HorseTile))
			r.Rests = r.Rests[1:]

			r.BroadcastAll(&msg.S2C_GDBuyHorse{
				Position: playerData.Position,
				Tiles:    []int{-1},
			})
		} else if r.Rule.BuyHorse == 2 {
			playerData.HorseTile = append(playerData.HorseTile, r.Rests[:2]...)
			log.Debug("马牌: %v", algorithm.ToTileString(playerData.HorseTile))
			r.Rests = r.Rests[2:]

			r.BroadcastAll(&msg.S2C_GDBuyHorse{
				Position: playerData.Position,
				Tiles:    []int{-1, -1},
			})
		}
		person := player.GetPersonMgr().GetPerson(userID)
		if person == nil {
			continue
		}
		person.WriteMsg(&msg.S2C_UpdateMahjongHands{
			Position:      playerData.Position,
			Hands:         playerData.Hands, // 不包含摸到的那一张牌
			NumberOfHands: len(playerData.Hands),
		})
		person.WriteMsg(&msg.S2C_UpdateWinTiles{
			Tiles: playerData.WinTiles,
		})

		r.Broadcast(&msg.S2C_UpdateMahjongHands{
			Position:      playerData.Position,
			NumberOfHands: len(playerData.Hands),
		}, playerData.Position)
	}

	if r.Rule.NeedJoker {
		dealerPlayerData.Discards = append(dealerPlayerData.Discards, r.Wildcard)
		r.BroadcastAll(&msg.S2C_UpdateMahjongDiscads{
			Position: dealerPlayerData.Position,
			Discards: dealerPlayerData.Discards,
		})
		r.BroadcastAll(&msg.S2C_MahjongDiscard{
			Position: dealerPlayerData.Position,
			Tile:     r.Wildcard,
		})
	}

	// 庄家摸牌、出牌
	//r.drawAndDiscard(r.dealerUserID)

}
