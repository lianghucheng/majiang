package room

import (
	"algorithm"
	"game/player"
	"game/room"
	msg "msg/room/mahjong"
	"time"

	"github.com/name5566/leaf/log"
)

// 断线重连
func reconnect(uid int) {
	log.Debug("userID: %v 断线重连", uid)
	person := player.GetPersonMgr().GetPerson(uid)
	ri := room.GetRoomMgr().GetRoom(uid)
	if ri == nil {
		return
	}
	r := ri.(*room.GDRoom)
	person.WriteMsg(&msg.S2C_GameStart{})
	dealerPlayerData := r.Useridplayerdatas[r.Dealeruserid]
	if dealerPlayerData != nil {
		person.WriteMsg(&msg.S2C_DecideDealer{
			Position: dealerPlayerData.Position,
		})
	}
	if r.Rule.NeedJoker {
		if r.Jokers == nil {
			log.Error("joker is nil: %v", r.Jokers)
		}
		person.WriteMsg(&msg.S2C_DecideGDJoker{
			WildCard: r.Wildcard,
			Jokers:   r.Jokers,
		})
	}
	playdata := r.Useridplayerdatas[uid]
	if r.Rule.BuyHorse > 0 {
		horseTile := getHorse(r.Rule.BuyHorse)
		for i := 0; i < r.Rule.MaxPlayers; i++ {
			person.WriteMsg(&msg.S2C_GDBuyHorse{
				Position: i,
				Tiles:    horseTile,
			})
		}
	}
	person.WriteMsg(&msg.S2C_UpdateMahjongRestsNumber{
		NumberOfRests: len(r.Rests),
	})
	person.WriteMsg(&msg.S2C_UpdateMahjongCurrentRound{
		CurrentRound: r.Currentround,
	})
	if len(playdata.WinTiles) > 0 {
		log.Debug("胡牌提示: %v", algorithm.ToTileString(playdata.WinTiles))
	}
	person.WriteMsg(&msg.S2C_UpdateWinTiles{
		Tiles: playdata.WinTiles,
	})
	if r.Claimuserid < 1 && r.Discarderuserid > 0 {
		discarderPlayerData := r.Useridplayerdatas[r.Discarderuserid]
		person.WriteMsg(&msg.S2C_UpdateMahjongDiscardCusor{
			Position: discarderPlayerData.Position,
			Index:    len(discarderPlayerData.Discards) - 1,
		})
	}
	if r.Disbandapplicantuserid > 0 {
		applicantPlayerData := r.Useridplayerdatas[r.Disbandapplicantuserid]
		playerDisbandInfos := []msg.GDPlayerDisbandInfo{}
		for i := 0; i < r.Rule.MaxPlayers; i++ {
			userID := r.PositionUserIDs[i]
			playerData := r.Useridplayerdatas[userID]
			playerDisbandInfos = append(playerDisbandInfos, msg.GDPlayerDisbandInfo{
				Nickname:   playerData.User.UserData.Nickname,
				ActionCode: playerData.DisbandActionCode,
			})
		}
		person.WriteMsg(&msg.S2C_ActionDisbandRoom{
			ApplicantNickname:  applicantPlayerData.User.UserData.Nickname,
			PlayerDisbandInfos: playerDisbandInfos,
			Enable:             playdata.DisbandActionCode == room.ActionWaitingDisband,
			WaitingTime:        300,
		})
	}
	for i := 0; i < r.Rule.MaxPlayers; i++ {
		oneplayData := r.Useridplayerdatas[i]
		person.WriteMsg(&msg.S2C_UpdateMahjongDiscads{
			Position: oneplayData.Position,
			Discards: oneplayData.Discards,
		})
		person.WriteMsg(&msg.S2C_UpdateMahjongClaims{
			Position: oneplayData.Position,
			Claims:   oneplayData.Claims,
		})
		hands := playdata.Hands
		if oneplayData.Position != playdata.Position {
			hands = []int{}
		}
		person.WriteMsg(&msg.S2C_UpdateMahjongHands{
			Position:      oneplayData.Position,
			Hands:         hands,
			NumberOfHands: len(oneplayData.Hands),
		})
		if oneplayData.Draw > -1 {
			draw := oneplayData.Draw
			if oneplayData.Position != playdata.Position {
				draw = -1
			}
			person.WriteMsg(&msg.S2C_MahjongDraw{
				Position:      oneplayData.Position,
				Tile:          draw,
				NumberOfHands: len(oneplayData.Hands),
			})
		}
		person.WriteMsg(&msg.S2C_UpdateGDTotalScore{
			Position:   oneplayData.Position,
			TotalScore: oneplayData.TotalResult.TotalScore,
		})
		switch oneplayData.State {
		case room.GdActionDiscard:
			after := int(time.Now().Unix() - oneplayData.ActionTimestamp)
			countdown := room.Cd_gdDiscard - after
			if countdown > 1 {
				person.WriteMsg(&msg.S2C_ActionMahjongDiscard{
					Position:  oneplayData.Position,
					Countdown: countdown - 1,
				})
			}
		case room.GdActionClaim:
			after := int(time.Now().Unix() - oneplayData.ActionTimestamp)
			countdown := room.Cd_gdClaim - after
			if countdown > 1 {
				person.WriteMsg(&msg.S2C_ActionMahjongClaim{
					Position:    oneplayData.Position,
					ActionCode:  oneplayData.ClaimActionCode,
					Countdown:   countdown - 1,
					Quadruplets: oneplayData.Quadruplets,
					Sequences:   oneplayData.Sequences,
				})
			}
		}
	}
}

func getHorse(number int) []int {
	if number == 1 {
		return []int{-1}
	}
	return []int{-1, -1}
}
