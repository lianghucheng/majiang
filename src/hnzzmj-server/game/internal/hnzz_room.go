package internal

import (
	"fmt"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/timer"
	"gopkg.in/mgo.v2/bson"
	"hnzzmj-server/common"
	"hnzzmj-server/game/mahjong"
	"hnzzmj-server/msg"
	"time"
)

// 玩家状态
const (
	_                 = iota
	hnzzReady         // 1 准备
	hnzzWaiting       // 2 等待
	hnzzActionDiscard // 3 前端显示出牌动作
	hnzzActionClaim   // 4 前端显示胡、杠、碰、吃动作
	hnzzWin           // 5 胡
	hnzzKong          // 6 杠
	hnzzPong          // 7 碰
	hnzzChow          // 8 吃
)

// 倒计时
const (
	cd_hnzzDiscard = 20
	cd_hnzzClaim   = 20
)

type HNZZRoom struct {
	room
	rule              *mahjong.HNZZRule
	userIDPlayerDatas map[int]*HNZZPlayerData // key: userID
	tiles             []int                   // 洗好的牌
	currentRound      int                     // 第几局
	dealerUserID      int                     // 庄家 userID
	catchBirdUserID   int                     // 上一局抓鸟的玩家 userID
	jokers            []int                   // 宝牌
	discards          []int                   // 玩家打的牌
	rests             []int                   // 剩余的牌
	drawerUserID      int                     // 最近一次摸牌的人 userID
	discarderUserID   int                     // 最近一次出牌的人 userID

	actionWinUsers  map[int]int // 可以胡的玩家 key: userID, value: 1
	actionKongUsers map[int]int // 可以杠的玩家 key: userID, value: 1
	actionPongUsers map[int]int // 可以碰的玩家 key: userID, value: 1
	actionChowUsers map[int]int // 可以吃的玩家 key: userID, value: 1

	discardTimer *timer.Timer
	claimTimer   *timer.Timer
	disbandTimer *timer.Timer
	claimUserID  int

	disbandApplicantUserID int // 申请解散房间者 userID
	winnerUserIDs          []int
}

// 玩家数据
type HNZZPlayerData struct {
	user            *User
	totalResultData *TotalResultData
	roundResultData *RoundResultData
	state           int
	position        int // 用户在桌子上的位置，从 0 开始

	owner           bool // 房主
	dealer          bool // 庄家
	draw            int  // 摸的一张牌
	hands           []int
	discards        []int // 打出的牌
	analyzer        *mahjong.HNZZAnalyzer
	claimActionCode int
	claims          [][]int // 吃、碰、杠到的牌
	actionTimestamp int64   // 记录操作时间戳

	claim       int // 待吃、碰、杠的牌
	quadruplet  []int
	quadruplets [][]int
	triplet     []int
	sequence    []int
	sequences   [][]int

	kongType    int
	winType     int
	roundResult *mahjong.HNZZPlayerRoundResult
	totalResult *mahjong.HNZZPlayerTotalResult

	disbandActionCode int
	winTiles          []int
}

func newHNZZRoom(rule *mahjong.HNZZRule) *HNZZRoom {
	hnzzRoom := new(HNZZRoom)
	hnzzRoom.state = roomIdle
	hnzzRoom.loginIPs = make(map[string]bool)
	hnzzRoom.positionUserIDs = make(map[int]int)
	hnzzRoom.userIDPlayerDatas = make(map[int]*HNZZPlayerData)
	hnzzRoom.actionWinUsers = make(map[int]int)
	hnzzRoom.actionKongUsers = make(map[int]int)
	hnzzRoom.actionPongUsers = make(map[int]int)
	hnzzRoom.actionChowUsers = make(map[int]int)

	hnzzRoom.currentRound = 1
	hnzzRoom.jokers = []int{31} // 红中
	hnzzRoom.rule = rule

	win := "点炮胡" // 胡法
	if rule.MustSelfDraw {
		win = "自摸胡"
	}
	distinguish := "通庄"
	if rule.DistinguishDealer {
		distinguish = "分庄闲"
	}
	_ = distinguish
	switch hnzzRoom.rule.RoomType {
	case roomPrivate:
		hnzzRoom.desc = fmt.Sprintf("%v 红中癞子 %v局 %v人 底分%v分 抓%v鸟", win, rule.MaxRounds, rule.MaxPlayers, rule.BaseScore, rule.Birds)
	case roomRoomCardMatch:
		hnzzRoom.desc = fmt.Sprintf("%v人 底注%v房卡 红中癞子 %v", rule.MaxPlayers, rule.RoomCards, win)
	case roomRedPacketMatching, roomRedPacketPrivate:
		hnzzRoom.desc = fmt.Sprintf("%v人 %v元红包 红中癞子 %v", rule.MaxPlayers, rule.RedPacketType, win)
	}
	if hnzzRoom.rule.IPAntiCheat {
		hnzzRoom.desc += " IP防作弊"
	}
	if hnzzRoom.rule.GPSAntiCheat {
		hnzzRoom.desc += " GPS防作弊"
	}
	return hnzzRoom
}

func (hnzzRoom *HNZZRoom) allReady() bool {
	count := 0
	if hnzzRoom.full() {
		for _, userID := range hnzzRoom.positionUserIDs {
			playerData := hnzzRoom.userIDPlayerDatas[userID]
			if playerData.state == hnzzReady {
				count++
			}
		}
		if count == hnzzRoom.rule.MaxPlayers {
			return true
		}
		return false
	}
	return false
}

func (hnzzRoom *HNZZRoom) allAgree() bool {
	count := 0
	for _, userID := range hnzzRoom.positionUserIDs {
		playerData := hnzzRoom.userIDPlayerDatas[userID]
		if playerData.disbandActionCode == actionAgreeDisband {
			count++
		}
	}
	if count == hnzzRoom.rule.MaxPlayers {
		return true
	}
	return false
}

func (hnzzRoom *HNZZRoom) empty() bool {
	return len(hnzzRoom.positionUserIDs) == 0
}

func (hnzzRoom *HNZZRoom) full() bool {
	return len(hnzzRoom.positionUserIDs) == hnzzRoom.rule.MaxPlayers
}

func (hnzzRoom *HNZZRoom) clean() {
	for _, userID := range hnzzRoom.positionUserIDs {
		delete(userIDRooms, userID)
	}
	for pos := range hnzzRoom.positionUserIDs {
		delete(hnzzRoom.positionUserIDs, pos)
	}
	for userID := range hnzzRoom.userIDPlayerDatas {
		delete(hnzzRoom.userIDPlayerDatas, userID)
	}
	if hnzzRoom.claimTimer != nil {
		hnzzRoom.claimTimer.Stop()
		hnzzRoom.claimTimer = nil
	}
	if hnzzRoom.discardTimer != nil {
		hnzzRoom.discardTimer.Stop()
		hnzzRoom.discardTimer = nil
	}
}

func (hnzzRoom *HNZZRoom) Enter(user *User) bool {
	roomCards := 0
	switch hnzzRoom.rule.RoomType {
	case roomRoomCardMatch, roomRedPacketMatching, roomRedPacketPrivate:
		roomCards = hnzzRoom.rule.RoomCards
	}
	if playerData, ok := hnzzRoom.userIDPlayerDatas[user.data.userData.UserID]; ok { // 断线重连
		playerData.user = user
		user.WriteMsg(&msg.S2C_EnterRoom{
			Error:         msg.S2C_EnterRoom_OK,
			RoomType:      hnzzRoom.rule.RoomType,
			RedPacketType: hnzzRoom.rule.RedPacketType,
			RoomNumber:    hnzzRoom.number,
			Position:      playerData.position,
			RoomDesc:      hnzzRoom.desc,
			MaxPlayers:    hnzzRoom.rule.MaxPlayers,
			MaxRounds:     hnzzRoom.rule.MaxRounds,
			RoomCards:     roomCards,
			GamePlaying:   hnzzRoom.state == roomGame,
		})
		log.Debug("userID: %v 重连进入房间, 房间类型: %v", user.data.userData.UserID, hnzzRoom.rule.RoomType)
		return true
	}
	// 玩家已满
	if hnzzRoom.full() {
		user.WriteMsg(&msg.S2C_EnterRoom{
			Error:      msg.S2C_EnterRoom_Full,
			RoomNumber: hnzzRoom.number,
		})
		return false
	}
	switch hnzzRoom.rule.RoomType {
	case roomRoomCardMatch, roomRedPacketMatching, roomRedPacketPrivate:
		if !user.checkEnterRoomCards(hnzzRoom.rule.RoomCards) {
			return false
		}
	}
	if hnzzRoom.rule.IPAntiCheat {
		if _, ok := hnzzRoom.loginIPs[user.data.userData.LoginIP]; ok {
			user.WriteMsg(&msg.S2C_EnterRoom{
				Error: msg.S2C_EnterRoom_IPConflict,
			})
			return false
		}
		hnzzRoom.loginIPs[user.data.userData.LoginIP] = true
	}
	for pos := 0; pos < hnzzRoom.rule.MaxPlayers; pos++ {
		if _, ok := hnzzRoom.positionUserIDs[pos]; !ok {
			hnzzRoom.SitDown(user, pos)
			user.WriteMsg(&msg.S2C_EnterRoom{
				Error:         msg.S2C_EnterRoom_OK,
				RoomType:      hnzzRoom.rule.RoomType,
				RedPacketType: hnzzRoom.rule.RedPacketType,
				RoomNumber:    hnzzRoom.number,
				Position:      pos,
				RoomDesc:      hnzzRoom.desc,
				MaxPlayers:    hnzzRoom.rule.MaxPlayers,
				MaxRounds:     hnzzRoom.rule.MaxRounds,
				RoomCards:     roomCards,
				GamePlaying:   hnzzRoom.state == roomGame,
			})
			log.Debug("userID: %v 进入房间, 房间类型: %v", user.data.userData.UserID, hnzzRoom.rule.RoomType)
			switch hnzzRoom.rule.RoomType {
			case roomRoomCardMatch:
				calculateRoomCardMatchOnlineNumber(hnzzRoom.rule.RoomCards, false)
			case roomRedPacketMatching, roomRedPacketPrivate:
				calculateRedPacketMatchOnlineNumber(hnzzRoom.rule.RedPacketType)
			}
			return true
		}
	}
	user.WriteMsg(&msg.S2C_EnterRoom{
		Error:      msg.S2C_EnterRoom_Unknown,
		RoomNumber: hnzzRoom.number,
	})
	return false
}

func (hnzzRoom *HNZZRoom) Exit(user *User) {
	playerData := hnzzRoom.userIDPlayerDatas[user.data.userData.UserID]
	if playerData == nil {
		return
	}
	broadcast(&msg.S2C_StandUp{
		Position: playerData.position,
	}, hnzzRoom.positionUserIDs, -1)
	log.Debug("userID: %v 退出房间", user.data.userData.UserID)
	broadcast(&msg.S2C_ExitRoom{
		Error:    msg.S2C_ExitRoom_OK,
		Position: playerData.position,
	}, hnzzRoom.positionUserIDs, -1)
	// 站起
	hnzzRoom.StandUp(user, playerData.position)
	// 退出
	delete(userIDRooms, user.data.userData.UserID)
	// 删除玩家登录IP
	delete(hnzzRoom.loginIPs, user.data.userData.LoginIP)

	switch hnzzRoom.rule.RoomType {
	case roomRoomCardMatch:
		calculateRoomCardMatchOnlineNumber(hnzzRoom.rule.RoomCards, true)
	}

	if hnzzRoom.empty() { // 玩家为空，解散房间
		switch hnzzRoom.rule.RoomType {
		case roomPractice:
			delete(hnzzPracticeRooms, hnzzRoom.creatorUserID)
		case roomRoomCardMatch, roomRedPacketMatching:
			delete(hnzzRoomCardMatchRooms, hnzzRoom.creatorUserID)
		case roomPrivate, roomRedPacketPrivate:
			delete(roomNumberRooms, hnzzRoom.number)
		}
	}
}

func (hnzzRoom *HNZZRoom) SitDown(user *User, pos int) {
	hnzzRoom.positionUserIDs[pos] = user.data.userData.UserID

	playerData := hnzzRoom.userIDPlayerDatas[user.data.userData.UserID]
	if playerData == nil {
		playerData = new(HNZZPlayerData)
		playerData.user = user
		playerData.position = pos
		playerData.owner = user.data.userData.UserID == hnzzRoom.ownerUserID
		playerData.analyzer = new(mahjong.HNZZAnalyzer)
		playerData.roundResult = new(mahjong.HNZZPlayerRoundResult)
		playerData.totalResult = new(mahjong.HNZZPlayerTotalResult)

		hnzzRoom.userIDPlayerDatas[user.data.userData.UserID] = playerData
	}
	message := &msg.S2C_SitDown{
		Position:   pos,
		Owner:      playerData.owner,
		AccountID:  playerData.user.data.userData.AccountID,
		LoginIP:    playerData.user.data.userData.LoginIP,
		Nickname:   playerData.user.data.userData.Nickname,
		Headimgurl: playerData.user.data.userData.Headimgurl,
		Sex:        playerData.user.data.userData.Sex,
		Ready:      playerData.state == hnzzReady,
	}
	if hnzzRoom.rule.GPSAntiCheat {
		message.Location = playerData.user.location
	}
	broadcast(message, hnzzRoom.positionUserIDs, pos)
}

func (hnzzRoom *HNZZRoom) StandUp(user *User, pos int) {
	delete(hnzzRoom.positionUserIDs, pos)
	delete(hnzzRoom.userIDPlayerDatas, user.data.userData.UserID)
}

func (hnzzRoom *HNZZRoom) GetAllPlayers(user *User) {
	for pos := 0; pos < hnzzRoom.rule.MaxPlayers; pos++ {
		userID := hnzzRoom.positionUserIDs[pos]
		playerData := hnzzRoom.userIDPlayerDatas[userID]
		if playerData == nil {
			user.WriteMsg(&msg.S2C_StandUp{
				Position: pos,
			})
		} else {
			message := &msg.S2C_SitDown{
				Position:   pos,
				Owner:      playerData.owner,
				AccountID:  playerData.user.data.userData.AccountID,
				LoginIP:    playerData.user.data.userData.LoginIP,
				Nickname:   playerData.user.data.userData.Nickname,
				Headimgurl: playerData.user.data.userData.Headimgurl,
				Sex:        playerData.user.data.userData.Sex,
				Ready:      playerData.state == hnzzReady}
			if hnzzRoom.rule.GPSAntiCheat {
				message.Location = playerData.user.location
			}
			user.WriteMsg(message)
		}
	}
}

func (hnzzRoom *HNZZRoom) Disband(disbander *User) {
	// 等待开局
	if hnzzRoom.state == roomIdle {
		log.Debug("userID: %v 解散房间", disbander.data.userData.UserID)
		broadcast(&msg.S2C_DisbandRoom{
			Error:         msg.S2C_DisbandRoom_OK,
			RoomNumber:    hnzzRoom.number,
			OwnerNickName: disbander.data.userData.Nickname,
		}, hnzzRoom.positionUserIDs, -1)
		// 清空玩家数据
		hnzzRoom.clean()
		if hnzzRoom.rule.RoomType == roomPrivate {
			delete(roomNumberRooms, hnzzRoom.number) // 解散房间
		}
		return
	}
	log.Debug("userID: %v 申请解散房间", disbander.data.userData.UserID)
	hnzzRoom.disbandApplicantUserID = disbander.data.userData.UserID
	applicantPlayerData := hnzzRoom.userIDPlayerDatas[hnzzRoom.disbandApplicantUserID]
	applicantPlayerData.disbandActionCode = actionAgreeDisband
	for i := 1; i < hnzzRoom.rule.MaxPlayers; i++ {
		otherUserID := hnzzRoom.positionUserIDs[(applicantPlayerData.position+i)%hnzzRoom.rule.MaxPlayers]
		otherPlayerData := hnzzRoom.userIDPlayerDatas[otherUserID]
		otherPlayerData.disbandActionCode = actionWaitingDisband
	}
	playerDisbandInfos := []mahjong.HNZZPlayerDisbandInfo{}
	for i := 0; i < hnzzRoom.rule.MaxPlayers; i++ {
		userID := hnzzRoom.positionUserIDs[i]
		playerData := hnzzRoom.userIDPlayerDatas[userID]
		playerDisbandInfos = append(playerDisbandInfos, mahjong.HNZZPlayerDisbandInfo{
			Nickname:   playerData.user.data.userData.Nickname,
			ActionCode: playerData.disbandActionCode,
		})
	}
	for _, userID := range hnzzRoom.positionUserIDs {
		playerData := hnzzRoom.userIDPlayerDatas[userID]
		if user, ok := userIDUsers[userID]; ok {
			user.WriteMsg(&msg.S2C_ActionDisbandRoom{
				ApplicantNickname:  disbander.data.userData.Nickname,
				PlayerDisbandInfos: playerDisbandInfos,
				Enable:             playerData.disbandActionCode == actionWaitingDisband,
				WaitingTime:        120,
			})
		}
	}

	hnzzRoom.disbandTimer = skeleton.AfterFunc(122*time.Second, func() {
		for _, userID := range hnzzRoom.positionUserIDs {
			playerData := hnzzRoom.userIDPlayerDatas[userID]
			if playerData.disbandActionCode == actionWaitingDisband {
				log.Debug("userID: %v 自动同意", playerData.user.data.userData.UserID)
				hnzzRoom.agreeDisbandRoom(playerData.user.data.userData.UserID)
			}
		}
	})
}

func (hnzzRoom *HNZZRoom) StartGame() {
	hnzzRoom.state = roomGame
	hnzzRoom.prepare()

	broadcast(&msg.S2C_GameStart{},
		hnzzRoom.positionUserIDs, -1)

	broadcast(&msg.S2C_UpdateMahjongCurrentRound{
		CurrentRound: hnzzRoom.currentRound,
	}, hnzzRoom.positionUserIDs, -1)

	dealerPlayerData := hnzzRoom.userIDPlayerDatas[hnzzRoom.dealerUserID]
	broadcast(&msg.S2C_DecideDealer{
		Position: dealerPlayerData.position,
	}, hnzzRoom.positionUserIDs, -1)

	broadcast(&msg.S2C_DecideHNZZJoker{
		Jokers: hnzzRoom.jokers,
	}, hnzzRoom.positionUserIDs, -1)
	// 所有玩家都发十三张牌
	for _, userID := range hnzzRoom.positionUserIDs {
		playerData := hnzzRoom.userIDPlayerDatas[userID]
		playerData.state = hnzzWaiting
		// 手牌有十三张
		playerData.hands = append(playerData.hands, hnzzRoom.rests[:13]...)
		// 排序
		playerData.analyzer.Analyze(playerData.hands)
		playerData.hands = playerData.analyzer.Sort(playerData.hands)
		log.Debug("userID %v 手牌: %v", userID, mahjong.ToTileString(playerData.hands))
		playerData.winTiles = playerData.analyzer.GetWinTiles(playerData.hands)
		if len(playerData.winTiles) > 0 {
			log.Debug("胡牌提示: %v", mahjong.ToTileString(playerData.winTiles))
		}
		// 剩余的牌
		hnzzRoom.rests = hnzzRoom.rests[13:]

		if user, ok := userIDUsers[userID]; ok {
			user.WriteMsg(&msg.S2C_UpdateMahjongHands{
				Position:      playerData.position,
				Hands:         playerData.hands, // 不包含摸到的那一张牌
				NumberOfHands: len(playerData.hands),
			})
			user.WriteMsg(&msg.S2C_UpdateWinTiles{
				Tiles: playerData.winTiles,
			})
		}
		broadcast(&msg.S2C_UpdateMahjongHands{
			Position:      playerData.position,
			NumberOfHands: len(playerData.hands),
		}, hnzzRoom.positionUserIDs, playerData.position)
	}
	// 庄家摸牌、出牌
	hnzzRoom.drawAndDiscard(hnzzRoom.dealerUserID)
}

func (hnzzRoom *HNZZRoom) EndGame() {
	log.Debug("游戏结束")
	hnzzRoom.state = roomGameEnd
	hnzzRoom.endTimestamp = time.Now().Unix()
	for _, userID := range hnzzRoom.positionUserIDs {
		playerData := hnzzRoom.userIDPlayerDatas[userID]
		playerData.winTiles = []int{}
		if user, ok := userIDUsers[userID]; ok {
			user.WriteMsg(&msg.S2C_UpdateWinTiles{
				Tiles: playerData.winTiles,
			})
		}
	}
	if hnzzRoom.currentRound == 1 {
		switch hnzzRoom.rule.RoomType {
		case roomPrivate, roomRoomCardMatch, roomRedPacketMatching, roomRedPacketPrivate:
			hnzzRoom.deductRoomCard()
		}
	}
	totalResults := []PlayerResultData{}
	roundResults := []PlayerResultData{}
	for pos := 0; pos < hnzzRoom.rule.MaxPlayers; pos++ {
		userID := hnzzRoom.positionUserIDs[pos]
		playerData := hnzzRoom.userIDPlayerDatas[userID]
		// 计算总分
		roundResult := playerData.roundResult
		roundResult.TotalScore = roundResult.WinScore + roundResult.ExposedKongScore + roundResult.PongKongScore + roundResult.HiddenKongScore + roundResult.CatchBirdScore
		if len(hnzzRoom.winnerUserIDs) == 0 { // 无人胡牌
			roundResult.LastTile = -1
			roundResult.RoomCards = 0
		} else {
			if common.InArray(hnzzRoom.winnerUserIDs, userID) {
				roundResult.LastTile = playerData.claim
				if hnzzRoom.rule.RoomType == roomRoomCardMatch {
					roundResult.RoomCards = hnzzRoom.rule.RoomCards * (hnzzRoom.rule.MaxPlayers - 1)
				}
			} else {
				roundResult.LastTile = -1
				if hnzzRoom.rule.RoomType == roomRoomCardMatch {
					roundResult.RoomCards = -hnzzRoom.rule.RoomCards
				}
			}
		}
		totalResult := playerData.totalResult
		totalResult.Scores = append(totalResult.Scores, roundResult.TotalScore)
		totalResult.TotalScore += roundResult.TotalScore

		broadcast(&msg.S2C_UpdateHNZZTotalScore{
			Position:   pos,
			TotalScore: totalResult.TotalScore,
		}, hnzzRoom.positionUserIDs, -1)

		switch hnzzRoom.rule.RoomType {
		case roomPrivate, roomRoomCardMatch:
			totalRoomCards := playerData.user.data.userData.RoomCards + roundResult.RoomCards
			totalResults = append(totalResults, PlayerResultData{
				UserID:         playerData.user.data.userData.UserID,
				Nickname:       playerData.user.data.userData.Nickname,
				Score:          totalResult.TotalScore,
				RoomCards:      roundResult.RoomCards,
				TotalRoomCards: totalRoomCards,
			})
			roundResults = append(roundResults, PlayerResultData{
				UserID:   playerData.user.data.userData.UserID,
				Nickname: playerData.user.data.userData.Nickname,
				Score:    roundResult.TotalScore,
			})
		}
	}
	// 保存总成绩
	switch hnzzRoom.rule.RoomType {
	case roomRedPacketPrivate, roomRedPacketMatching:
		for _, userID := range hnzzRoom.positionUserIDs {
			playerData := hnzzRoom.userIDPlayerDatas[userID]
			saveRedPacketMatchResultData(&RedPacketMatchResultData{
				UserID:        userID,
				RedPacketType: hnzzRoom.rule.RedPacketType,
				RedPacket:     playerData.roundResult.RedPacket,
				Taken:         false,
				CreatedAt:     time.Now().Unix(),
			})
		}
	case roomPrivate, roomRoomCardMatch:
		hnzzRoom.saveUserTotalResultData(totalResults)
		hnzzRoom.saveUserRoundResultData(hnzzRoom.currentRound, roundResults)
	}
	for _, userID := range hnzzRoom.positionUserIDs {
		var roundResults []mahjong.HNZZPlayerRoundResult

		playerData := hnzzRoom.userIDPlayerDatas[userID]
		roundResults = append(roundResults, mahjong.HNZZPlayerRoundResult{
			Nickname:         playerData.user.data.userData.Nickname,
			Headimgurl:       playerData.user.data.userData.Headimgurl,
			Dealer:           playerData.dealer,
			Hands:            playerData.hands,
			Claims:           playerData.claims,
			LastTile:         playerData.roundResult.LastTile,
			WinType:          playerData.roundResult.WinType,
			WinScore:         playerData.roundResult.WinScore,
			ExposedKongScore: playerData.roundResult.ExposedKongScore,
			PongKongScore:    playerData.roundResult.PongKongScore,
			HiddenKongScore:  playerData.roundResult.HiddenKongScore,
			CatchBirdScore:   playerData.roundResult.CatchBirdScore,
			TotalScore:       playerData.roundResult.TotalScore,
			RoomCards:        playerData.roundResult.RoomCards,
			RedPacket:        playerData.roundResult.RedPacket,
		})
		for i := 1; i < hnzzRoom.rule.MaxPlayers; i++ {
			otherUserID := hnzzRoom.positionUserIDs[(playerData.position+i)%hnzzRoom.rule.MaxPlayers]
			otherPlayerData := hnzzRoom.userIDPlayerDatas[otherUserID]
			roundResults = append(roundResults, mahjong.HNZZPlayerRoundResult{
				Nickname:         otherPlayerData.user.data.userData.Nickname,
				Headimgurl:       otherPlayerData.user.data.userData.Headimgurl,
				Dealer:           otherPlayerData.dealer,
				Hands:            otherPlayerData.hands,
				Claims:           otherPlayerData.claims,
				LastTile:         otherPlayerData.roundResult.LastTile,
				WinType:          otherPlayerData.roundResult.WinType,
				WinScore:         otherPlayerData.roundResult.WinScore,
				ExposedKongScore: otherPlayerData.roundResult.ExposedKongScore,
				PongKongScore:    otherPlayerData.roundResult.PongKongScore,
				HiddenKongScore:  otherPlayerData.roundResult.HiddenKongScore,
				CatchBirdScore:   otherPlayerData.roundResult.CatchBirdScore,
				TotalScore:       otherPlayerData.roundResult.TotalScore,
				RoomCards:        otherPlayerData.roundResult.RoomCards,
				RedPacket:        otherPlayerData.roundResult.RedPacket,
			})
		}
		if user, ok := userIDUsers[userID]; ok {
			result := mahjong.ResultLose
			if len(hnzzRoom.winnerUserIDs) == 0 { // 无人胡牌
				result = mahjong.ResultDraw
			} else {
				if common.InArray(hnzzRoom.winnerUserIDs, userID) {
					result = mahjong.ResultWin
				}
			}
			continueGame := true
			switch hnzzRoom.rule.RoomType {
			case roomRedPacketPrivate, roomRedPacketMatching:
				continueGame = false
			case roomPrivate:
				continueGame = !(hnzzRoom.currentRound == hnzzRoom.rule.MaxRounds)
			}
			if hnzzRoom.rule.RoomType == roomPrivate {
			}
			user.WriteMsg(&msg.S2C_HNZZRoundResult{
				Result:       result,
				RoomDesc:     hnzzRoom.desc,
				Jokers:       hnzzRoom.jokers,
				RoundResults: roundResults,
				ContinueGame: continueGame,
			})
		}
	}
	if hnzzRoom.currentRound < hnzzRoom.rule.MaxRounds {
		hnzzRoom.currentRound++
		return
	}
	switch hnzzRoom.rule.RoomType {
	case roomRoomCardMatch:
		hnzzRoom.calculateRoomCard()
	case roomPrivate:
		for _, userID := range hnzzRoom.positionUserIDs {
			var playerTotalResults []mahjong.HNZZPlayerTotalResult

			playerData := hnzzRoom.userIDPlayerDatas[userID]
			playerTotalResults = append(playerTotalResults, mahjong.HNZZPlayerTotalResult{
				Nickname:   playerData.user.data.userData.Nickname,
				Headimgurl: playerData.user.data.userData.Headimgurl,
				Owner:      playerData.owner,
				AccountID:  playerData.user.data.userData.AccountID,
				Scores:     playerData.totalResult.Scores,
				TotalScore: playerData.totalResult.TotalScore,
			})
			for i := 1; i < hnzzRoom.rule.MaxPlayers; i++ {
				otherUserID := hnzzRoom.positionUserIDs[(playerData.position+i)%hnzzRoom.rule.MaxPlayers]
				otherPlayerData := hnzzRoom.userIDPlayerDatas[otherUserID]
				playerTotalResults = append(playerTotalResults, mahjong.HNZZPlayerTotalResult{
					Nickname:   otherPlayerData.user.data.userData.Nickname,
					Headimgurl: otherPlayerData.user.data.userData.Headimgurl,
					Owner:      otherPlayerData.owner,
					AccountID:  otherPlayerData.user.data.userData.AccountID,
					Scores:     otherPlayerData.totalResult.Scores,
					TotalScore: otherPlayerData.totalResult.TotalScore,
				})
			}
			if user, ok := userIDUsers[userID]; ok {
				user.WriteMsg(&msg.S2C_HNZZTotalResult{
					TotalResults: playerTotalResults,
				})
				// 保存游戏积分
				user.data.userData.GameScore += playerData.totalResult.TotalScore
				// saveUserData(user.data)
			} else {
				playerData.user.data.userData.GameScore += playerData.totalResult.TotalScore
				updateUserData(userID, bson.M{"$set": bson.M{"gamescore": playerData.user.data.userData.GameScore}})
			}
		}
	}
	// 清空玩家数据
	hnzzRoom.clean()
	switch hnzzRoom.rule.RoomType {
	case roomPrivate, roomRedPacketPrivate:
		delete(roomNumberRooms, hnzzRoom.number) // 解散房间
	}
}

func (hnzzRoom *HNZZRoom) agreeDisbandRoom(userID int) {
	if hnzzRoom.state == roomIdle {
		return
	}
	playerData := hnzzRoom.userIDPlayerDatas[userID]
	if playerData.disbandActionCode != actionWaitingDisband {
		return
	}
	playerData.disbandActionCode = actionAgreeDisband
	if hnzzRoom.allAgree() {
		if hnzzRoom.rule.RoomType == roomPrivate && hnzzRoom.currentRound == 1 {
			hnzzRoom.deductRoomCard()
		}
		broadcast(&msg.S2C_DisbandRoom{
			Error:      msg.S2C_DisbandRoom_OK,
			RoomNumber: hnzzRoom.number,
		}, hnzzRoom.positionUserIDs, -1)
		// 清空玩家数据
		hnzzRoom.clean()
		if hnzzRoom.rule.RoomType == roomPrivate {
			delete(roomNumberRooms, hnzzRoom.number) // 解散房间
		}
	} else {
		broadcast(&msg.S2C_AgreeDisbandRoom{
			Position: playerData.position,
			Nickname: playerData.user.data.userData.Nickname,
		}, hnzzRoom.positionUserIDs, -1)
	}
}

func (hnzzRoom *HNZZRoom) refuseDisbandRoom(userID int) {
	playerData := hnzzRoom.userIDPlayerDatas[userID]
	if hnzzRoom.state == roomIdle || hnzzRoom.disbandTimer == nil || playerData.disbandActionCode != actionWaitingDisband {
		return
	}
	log.Debug("userID: %v 拒绝解散房间", userID)
	if hnzzRoom.disbandTimer != nil {
		hnzzRoom.disbandTimer.Stop()
		hnzzRoom.disbandTimer = nil
	}
	hnzzRoom.disbandApplicantUserID = -1
	broadcast(&msg.S2C_DisbandRoom{
		Error:            msg.S2C_DisbandRoom_PlayerRefuse,
		RoomNumber:       hnzzRoom.number,
		RejecterNickName: playerData.user.data.userData.Nickname,
	}, hnzzRoom.positionUserIDs, -1)
}

// 断线重连
func (hnzzRoom *HNZZRoom) reconnect(user *User) {
	log.Debug("userID: %v 断线重连", user.data.userData.UserID)
	thePlayerData := hnzzRoom.userIDPlayerDatas[user.data.userData.UserID]
	if thePlayerData == nil {
		return
	}
	user.WriteMsg(&msg.S2C_GameStart{})

	dealerPlayerData := hnzzRoom.userIDPlayerDatas[hnzzRoom.dealerUserID]
	if dealerPlayerData != nil {
		user.WriteMsg(&msg.S2C_DecideDealer{
			Position: dealerPlayerData.position,
		})
	}
	user.WriteMsg(&msg.S2C_DecideHNZZJoker{
		Jokers: hnzzRoom.jokers,
	})
	user.WriteMsg(&msg.S2C_UpdateMahjongRestsNumber{
		NumberOfRests: len(hnzzRoom.rests),
	})
	user.WriteMsg(&msg.S2C_UpdateMahjongCurrentRound{
		CurrentRound: hnzzRoom.currentRound,
	})
	if len(thePlayerData.winTiles) > 0 {
		log.Debug("胡牌提示: %v", mahjong.ToTileString(thePlayerData.winTiles))
	}
	user.WriteMsg(&msg.S2C_UpdateWinTiles{
		Tiles: thePlayerData.winTiles,
	})
	if hnzzRoom.claimUserID < 1 && hnzzRoom.discarderUserID > 0 {
		discarderPlayerData := hnzzRoom.userIDPlayerDatas[hnzzRoom.discarderUserID]
		user.WriteMsg(&msg.S2C_UpdateMahjongDiscardCusor{
			Position: discarderPlayerData.position,
			Index:    len(discarderPlayerData.discards) - 1,
		})
	}
	if hnzzRoom.disbandApplicantUserID > 0 {
		applicantPlayerData := hnzzRoom.userIDPlayerDatas[hnzzRoom.disbandApplicantUserID]
		playerDisbandInfos := []mahjong.HNZZPlayerDisbandInfo{}
		for i := 0; i < hnzzRoom.rule.MaxPlayers; i++ {
			userID := hnzzRoom.positionUserIDs[i]
			playerData := hnzzRoom.userIDPlayerDatas[userID]
			playerDisbandInfos = append(playerDisbandInfos, mahjong.HNZZPlayerDisbandInfo{
				Nickname:   playerData.user.data.userData.Nickname,
				ActionCode: playerData.disbandActionCode,
			})
		}
		user.WriteMsg(&msg.S2C_ActionDisbandRoom{
			ApplicantNickname:  applicantPlayerData.user.data.userData.Nickname,
			PlayerDisbandInfos: playerDisbandInfos,
			Enable:             thePlayerData.disbandActionCode == actionWaitingDisband,
			WaitingTime:        300,
		})
	}
	hnzzRoom.getPlayerData(user, thePlayerData, false)

	for i := 1; i < hnzzRoom.rule.MaxPlayers; i++ {
		otherUserID := hnzzRoom.positionUserIDs[(thePlayerData.position+i)%hnzzRoom.rule.MaxPlayers]
		otherPlayerData := hnzzRoom.userIDPlayerDatas[otherUserID]

		hnzzRoom.getPlayerData(user, otherPlayerData, true)
	}
}

func (hnzzRoom *HNZZRoom) getPlayerData(user *User, playerData *HNZZPlayerData, other bool) {
	user.WriteMsg(&msg.S2C_UpdateMahjongDiscads{
		Position: playerData.position,
		Discards: playerData.discards,
	})
	user.WriteMsg(&msg.S2C_UpdateMahjongClaims{
		Position: playerData.position,
		Claims:   playerData.claims,
	})
	hands := playerData.hands
	if other {
		hands = []int{}
	}
	user.WriteMsg(&msg.S2C_UpdateMahjongHands{
		Position:      playerData.position,
		Hands:         hands,
		NumberOfHands: len(playerData.hands),
	})
	if playerData.draw > -1 {
		draw := playerData.draw
		if other {
			draw = -1
		}
		user.WriteMsg(&msg.S2C_MahjongDraw{
			Position:      playerData.position,
			Tile:          draw,
			NumberOfHands: len(playerData.hands),
		})
	}
	user.WriteMsg(&msg.S2C_UpdateHNZZTotalScore{
		Position:   playerData.position,
		TotalScore: playerData.totalResult.TotalScore,
	})
	switch playerData.state {
	case hnzzActionDiscard:
		after := int(time.Now().Unix() - playerData.actionTimestamp)
		countdown := cd_hnzzDiscard - after
		if countdown > 1 {
			user.WriteMsg(&msg.S2C_ActionMahjongDiscard{
				Position:  playerData.position,
				Countdown: countdown - 1,
			})
		}
	case hnzzActionClaim:
		after := int(time.Now().Unix() - playerData.actionTimestamp)
		countdown := cd_hnzzClaim - after
		if countdown > 1 {
			user.WriteMsg(&msg.S2C_ActionMahjongClaim{
				Position:    playerData.position,
				ActionCode:  playerData.claimActionCode,
				Countdown:   countdown - 1,
				Quadruplets: playerData.quadruplets,
				Sequences:   playerData.sequences,
			})
		}
	}
}

func (hnzzRoom *HNZZRoom) resetActionClaimUsers() {
	if hnzzRoom.claimTimer != nil {
		hnzzRoom.claimTimer.Stop()
		hnzzRoom.claimTimer = nil
	}
	for _, userID := range hnzzRoom.positionUserIDs {
		hnzzRoom.deleteActionClaimUsers(userID)
	}
}

func (hnzzRoom *HNZZRoom) deleteActionClaimUsers(userID int) {
	playerData := hnzzRoom.userIDPlayerDatas[userID]
	if playerData.state == hnzzActionClaim {
		playerData.state = hnzzWaiting
		playerData.claimActionCode = 0
	}
	delete(hnzzRoom.actionWinUsers, userID)
	delete(hnzzRoom.actionKongUsers, userID)
	delete(hnzzRoom.actionPongUsers, userID)
	delete(hnzzRoom.actionChowUsers, userID)
}

func (hnzzRoom *HNZZRoom) actionClaimUsersEmpty() bool {
	if len(hnzzRoom.actionWinUsers) == 0 && len(hnzzRoom.actionKongUsers) == 0 &&
		len(hnzzRoom.actionPongUsers) == 0 && len(hnzzRoom.actionChowUsers) == 0 {
		return true
	}
	return false
}

func (hnzzRoom *HNZZRoom) prepare() {
	// 洗牌
	hnzzRoom.tiles = common.Shuffle(mahjong.HNZZAllTiles)
	// 确定庄家
	if hnzzRoom.currentRound == 1 {
		hnzzRoom.dealerUserID = hnzzRoom.positionUserIDs[0]
		switch hnzzRoom.rule.RoomType {
		case roomPrivate, roomRoomCardMatch, roomRedPacketMatching, roomRedPacketPrivate:
			hnzzRoom.startTimestamp = time.Now().Unix()
			hnzzRoom.eachRoundStartTimestamp = hnzzRoom.startTimestamp
			hnzzRoom.initTotalResultData()
		}
	} else {
		switch hnzzRoom.rule.RoomType {
		case roomPrivate, roomRoomCardMatch, roomRedPacketMatching, roomRedPacketPrivate:
			hnzzRoom.eachRoundStartTimestamp = time.Now().Unix()
		}
		if hnzzRoom.catchBirdUserID > 0 {
			hnzzRoom.dealerUserID = hnzzRoom.catchBirdUserID
		}
	}
	dealerPlayerData := hnzzRoom.userIDPlayerDatas[hnzzRoom.dealerUserID]
	dealerPlayerData.dealer = true
	// 确定闲家(注：闲家的英文单词也为player)
	dealerPos := dealerPlayerData.position
	for i := 1; i < hnzzRoom.rule.MaxPlayers; i++ {
		playerPos := (dealerPos + i) % hnzzRoom.rule.MaxPlayers
		playerUserID := hnzzRoom.positionUserIDs[playerPos]
		playerPlayerData := hnzzRoom.userIDPlayerDatas[playerUserID]
		playerPlayerData.dealer = false
	}
	log.Debug("癞子: %v", mahjong.ToTileString(hnzzRoom.jokers))
	hnzzRoom.discards = []int{}
	// 剩余的牌
	hnzzRoom.rests = append([]int{}, hnzzRoom.tiles...)

	hnzzRoom.catchBirdUserID = -1
	hnzzRoom.discarderUserID = -1
	hnzzRoom.drawerUserID = -1
	hnzzRoom.resetActionClaimUsers()
	hnzzRoom.claimUserID = -1
	hnzzRoom.disbandApplicantUserID = -1
	hnzzRoom.winnerUserIDs = []int{}

	for _, userID := range hnzzRoom.positionUserIDs {
		playerData := hnzzRoom.userIDPlayerDatas[userID]
		playerData.draw = -1
		playerData.hands = []int{}
		playerData.discards = []int{}
		playerData.claims = [][]int{}
		playerData.actionTimestamp = 0

		roundResult := playerData.roundResult
		roundResult.WinType = 0
		roundResult.WinScore = 0
		roundResult.ExposedKongScore = 0
		roundResult.PongKongScore = 0
		roundResult.HiddenKongScore = 0
		roundResult.CatchBirdScore = 0
		roundResult.TotalScore = 0
	}
}

func (hnzzRoom *HNZZRoom) updateDiscards(userID int) {
	playerData := hnzzRoom.userIDPlayerDatas[userID]
	broadcast(&msg.S2C_UpdateMahjongDiscads{
		Position: playerData.position,
		Discards: playerData.discards,
	}, hnzzRoom.positionUserIDs, -1)
}

func (hnzzRoom *HNZZRoom) updateClaims(userID int) {
	playerData := hnzzRoom.userIDPlayerDatas[userID]
	broadcast(&msg.S2C_UpdateMahjongClaims{
		Position: playerData.position,
		Claims:   playerData.claims,
	}, hnzzRoom.positionUserIDs, -1)
}

// 计算胡牌分(通庄和分庄闲的胡牌分是一样的)
func (hnzzRoom *HNZZRoom) calculateWinScore(winnerUserID int) {
	numberOfWinner := len(hnzzRoom.winnerUserIDs)
	if numberOfWinner == 0 || common.Index(hnzzRoom.winnerUserIDs, winnerUserID) == -1 {
		return
	}
	winnerPlayerData := hnzzRoom.userIDPlayerDatas[winnerUserID]
	if numberOfWinner > 1 {
		discarderPlayerData := hnzzRoom.userIDPlayerDatas[hnzzRoom.discarderUserID]
		if discarderPlayerData.dealer {
			winnerPlayerData.roundResult.WinScore = 2 * hnzzRoom.rule.BaseScore
		} else {
			if winnerPlayerData.dealer {
				winnerPlayerData.roundResult.WinScore = 2 * hnzzRoom.rule.BaseScore
			} else {
				winnerPlayerData.roundResult.WinScore = 1 * hnzzRoom.rule.BaseScore
			}
		}
		return
	}
	winnerWinType := winnerPlayerData.roundResult.WinType
	switch winnerWinType {
	case mahjong.HNZZWinByDiscard, mahjong.HNZZWinByEarthlyHand:
		discarderPlayerData := hnzzRoom.userIDPlayerDatas[hnzzRoom.discarderUserID]
		if discarderPlayerData.dealer {
			discarderPlayerData.roundResult.WinScore = -2 * hnzzRoom.rule.BaseScore
		} else {
			if winnerPlayerData.dealer {
				discarderPlayerData.roundResult.WinScore = -2 * hnzzRoom.rule.BaseScore
			} else {
				discarderPlayerData.roundResult.WinScore = -1 * hnzzRoom.rule.BaseScore
			}
		}
	case mahjong.HNZZWinBySelfDraw, mahjong.HNZZWinByHeavenlyHand:
		for i := 1; i < hnzzRoom.rule.MaxPlayers; i++ {
			otherUserID := hnzzRoom.positionUserIDs[(winnerPlayerData.position+i)%hnzzRoom.rule.MaxPlayers]
			otherPlayerData := hnzzRoom.userIDPlayerDatas[otherUserID]
			if winnerPlayerData.dealer {
				otherPlayerData.roundResult.WinScore = -3 * hnzzRoom.rule.BaseScore
			} else {
				if otherPlayerData.dealer {
					otherPlayerData.roundResult.WinScore = -3 * hnzzRoom.rule.BaseScore
				} else {
					otherPlayerData.roundResult.WinScore = -2 * hnzzRoom.rule.BaseScore
				}
			}
		}
	}
	loserWinScore := 0
	for i := 1; i < hnzzRoom.rule.MaxPlayers; i++ {
		otherUserID := hnzzRoom.positionUserIDs[(winnerPlayerData.position+i)%hnzzRoom.rule.MaxPlayers]
		otherPlayerData := hnzzRoom.userIDPlayerDatas[otherUserID]
		loserWinScore += otherPlayerData.roundResult.WinScore
	}
	// 计算赢家胡牌分
	winnerPlayerData.roundResult.WinScore = -1 * loserWinScore
}

// 计算抓鸟分
func (hnzzRoom *HNZZRoom) calculateCatchBirdScore(numberOfBird int) {
	if numberOfBird < 1 {
		return
	}
	numberOfWinner := len(hnzzRoom.winnerUserIDs)
	if numberOfWinner > 1 {
		discarderPlayerData := hnzzRoom.userIDPlayerDatas[hnzzRoom.discarderUserID]
		discarderPlayerData.roundResult.CatchBirdScore = -1 * hnzzRoom.rule.BaseScore * numberOfBird * numberOfWinner
		for _, winnerUserID := range hnzzRoom.winnerUserIDs {
			winnerPlayerData := hnzzRoom.userIDPlayerDatas[winnerUserID]
			winnerPlayerData.roundResult.CatchBirdScore = 1 * hnzzRoom.rule.BaseScore * numberOfBird
		}
		return
	}
	winnerPlayerData := hnzzRoom.userIDPlayerDatas[hnzzRoom.winnerUserIDs[0]]
	winnerWinType := winnerPlayerData.roundResult.WinType
	switch winnerWinType {
	case mahjong.HNZZWinByDiscard, mahjong.HNZZWinByEarthlyHand:
		discarderPlayerData := hnzzRoom.userIDPlayerDatas[hnzzRoom.discarderUserID]
		discarderPlayerData.roundResult.CatchBirdScore = -1 * hnzzRoom.rule.BaseScore * numberOfBird
	case mahjong.HNZZWinBySelfDraw, mahjong.HNZZWinByHeavenlyHand:
		for i := 1; i < hnzzRoom.rule.MaxPlayers; i++ {
			otherUserID := hnzzRoom.positionUserIDs[(winnerPlayerData.position+i)%hnzzRoom.rule.MaxPlayers]
			otherPlayerData := hnzzRoom.userIDPlayerDatas[otherUserID]
			otherPlayerData.roundResult.CatchBirdScore = -1 * hnzzRoom.rule.BaseScore * numberOfBird
		}

	}
	loserCatchBirdScore := 0
	for i := 1; i < hnzzRoom.rule.MaxPlayers; i++ {
		otherUserID := hnzzRoom.positionUserIDs[(winnerPlayerData.position+i)%hnzzRoom.rule.MaxPlayers]
		otherPlayerData := hnzzRoom.userIDPlayerDatas[otherUserID]
		loserCatchBirdScore += otherPlayerData.roundResult.CatchBirdScore
	}
	// 计算赢家抓鸟分
	winnerPlayerData.roundResult.CatchBirdScore = -1 * loserCatchBirdScore
}

func (hnzzRoom *HNZZRoom) calculateRedPacket(userID int, redPacketType int) {
	if common.Index([]int{1, 10, 100, 999}, redPacketType) == -1 {
		return
	}
	playerData := hnzzRoom.userIDPlayerDatas[userID]
	roundResult := playerData.roundResult
	log.Debug("redPacket: %v", redPacketType)
	roundResult.RedPacket = float64(hnzzRoom.rule.RedPacketType)
}

// 扣除房卡
func (hnzzRoom *HNZZRoom) deductRoomCard() {
	switch hnzzRoom.rule.RoomType {
	case roomRoomCardMatch, roomRedPacketPrivate, roomRedPacketMatching:
		for _, userID := range hnzzRoom.positionUserIDs {
			playerData := hnzzRoom.userIDPlayerDatas[userID]
			if user, ok := userIDUsers[userID]; ok {
				user.data.userData.RoomCards -= hnzzRoom.rule.RoomCards
				user.data.userData.ConsumedRoomCards += hnzzRoom.rule.RoomCards
			} else {
				playerData.user.data.userData.RoomCards -= hnzzRoom.rule.RoomCards
				playerData.user.data.userData.ConsumedRoomCards += hnzzRoom.rule.RoomCards
				updateUserData(userID, bson.M{"$set": bson.M{"roomcards": playerData.user.data.userData.RoomCards, "consumedroomcards": playerData.user.data.userData.ConsumedRoomCards}})
			}
			if playerData.user.isRobot() {
				cards := -hnzzRoom.rule.RoomCards
				switch hnzzRoom.rule.RoomType {
				case roomRoomCardMatch:
					upsertRobotData(time.Now().Format("20060102"), bson.M{"$inc": bson.M{"roomcardmatchbalance": cards}})
				case roomRedPacketMatching:
					upsertRobotData(time.Now().Format("20060102"), bson.M{"$inc": bson.M{"redpacketmatchbalance": cards}})
				}
			}
		}
	case roomPrivate:
		if owner, ok := userIDUsers[hnzzRoom.ownerUserID]; ok {
			owner.data.userData.RoomCards -= hnzzRoom.rule.RoomCards
			owner.data.userData.ConsumedRoomCards += hnzzRoom.rule.RoomCards
		} else {
			playerData := hnzzRoom.userIDPlayerDatas[hnzzRoom.ownerUserID]
			playerData.user.data.userData.RoomCards -= hnzzRoom.rule.RoomCards
			playerData.user.data.userData.ConsumedRoomCards += hnzzRoom.rule.RoomCards
			updateUserData(hnzzRoom.ownerUserID, bson.M{"$set": bson.M{"roomcards": playerData.user.data.userData.RoomCards, "consumedroomcards": playerData.user.data.userData.ConsumedRoomCards}})
		}
	}
}

// 结算房卡
func (hnzzRoom *HNZZRoom) calculateRoomCard() {
	if len(hnzzRoom.winnerUserIDs) == 0 { // 流局
		for _, userID := range hnzzRoom.positionUserIDs {
			playerData := hnzzRoom.userIDPlayerDatas[userID]
			if user, ok := userIDUsers[userID]; ok {
				user.data.userData.RoomCards += hnzzRoom.rule.RoomCards
			} else {
				playerData.user.data.userData.RoomCards += hnzzRoom.rule.RoomCards
				updateUserData(userID, bson.M{"$set": bson.M{"roomcards": playerData.user.data.userData.RoomCards}})
			}
			if playerData.user.isRobot() {
				cards := hnzzRoom.rule.RoomCards
				upsertRobotData(time.Now().Format("20060102"), bson.M{"$inc": bson.M{"roomcardmatchbalance": cards}})
			}
		}
	} else {
		winnerUserID := hnzzRoom.winnerUserIDs[0]
		playerData := hnzzRoom.userIDPlayerDatas[winnerUserID]
		if user, ok := userIDUsers[winnerUserID]; ok {
			user.data.userData.RoomCards += hnzzRoom.rule.RoomCards * hnzzRoom.rule.MaxPlayers
		} else {
			playerData.user.data.userData.RoomCards += hnzzRoom.rule.RoomCards * hnzzRoom.rule.MaxPlayers
			updateUserData(winnerUserID, bson.M{"$set": bson.M{"roomcards": playerData.user.data.userData.RoomCards}})
		}
		if playerData.user.isRobot() {
			cards := hnzzRoom.rule.RoomCards * hnzzRoom.rule.MaxPlayers
			upsertRobotData(time.Now().Format("20060102"), bson.M{"$inc": bson.M{"roomcardmatchbalance": cards}})
		}
	}
}

func (hnzzRoom *HNZZRoom) initTotalResultData() {
	for _, userID := range hnzzRoom.positionUserIDs {
		playerData := hnzzRoom.userIDPlayerDatas[userID]
		playerData.totalResultData = new(TotalResultData)
		skeleton.Go(func() {
			err := playerData.totalResultData.initValue(playerData.user.data.userData.UserID)
			if err != nil {
				log.Error("init userID %v totalresult data error: %v", playerData.user.data.userData.UserID, err)
				playerData.totalResultData = nil
			}
		}, func() {
			if playerData.totalResultData != nil {
				playerData.totalResultData.UserID = playerData.user.data.userData.UserID
				playerData.totalResultData.RoomNumber = hnzzRoom.number
				playerData.totalResultData.RoomDesc = hnzzRoom.desc
				playerData.totalResultData.StartTimestamp = hnzzRoom.startTimestamp
				playerData.totalResultData.Position = playerData.position
			}
		})
	}
}

func (hnzzRoom *HNZZRoom) saveUserTotalResultData(results []PlayerResultData) {
	for pos := 0; pos < hnzzRoom.rule.MaxPlayers; pos++ {
		userID := hnzzRoom.positionUserIDs[pos]
		playerData := hnzzRoom.userIDPlayerDatas[userID]
		if playerData.totalResultData != nil {
			playerData.totalResultData.RoomType = hnzzRoom.rule.RoomType
			playerData.totalResultData.EndTimestamp = hnzzRoom.endTimestamp
			playerData.totalResultData.Results = results
			playerData.totalResultData.UpdatedAt = time.Now().Unix()

			saveTotalResultData(playerData.totalResultData)
		}
	}
}

func (hnzzRoom *HNZZRoom) saveUserRoundResultData(round int, results []PlayerResultData) {
	for pos := 0; pos < hnzzRoom.rule.MaxPlayers; pos++ {
		userID := hnzzRoom.positionUserIDs[pos]
		playerData := hnzzRoom.userIDPlayerDatas[userID]
		if playerData.totalResultData != nil {
			playerData.roundResultData = new(RoundResultData)
			skeleton.Go(func() {
				err := playerData.roundResultData.initValue(playerData.totalResultData.ID)
				if err != nil {
					log.Error("init totalresult %v round result data error: %v", playerData.totalResultData.ID, err)
					playerData.roundResultData = nil
				}
			}, func() {
				if playerData.roundResultData != nil {
					playerData.roundResultData.Round = round
					playerData.roundResultData.StartTimestamp = hnzzRoom.eachRoundStartTimestamp
					playerData.roundResultData.EndTimestamp = hnzzRoom.endTimestamp
					playerData.roundResultData.Position = playerData.position
					playerData.roundResultData.Results = results
					playerData.roundResultData.UpdatedAt = time.Now().Unix()

					saveRoundResultData(playerData.roundResultData)
				}
			})
		}
	}
}
