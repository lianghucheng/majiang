package internal

import (
	"fmt"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/timer"
	"gopkg.in/mgo.v2/bson"
	"time"
	"yananmj-server/common"
	"yananmj-server/game/mahjong"
	"yananmj-server/msg"
)

//玩家状态
const (
	_                  = iota
	yananReady         // 准备
	yananActionSetGun  // 下炮子
	yananWaiting       // 等待
	yananActionDiscard // 前端显示出牌动作
	yananActionClaim   // 前端显示 胡、杠、碰的动作
	yananWin           // 胡
	yananKong          // 杠
	yananPong          // 碰
)

//倒计时
const (
	cd_yananDiscard = 20
	cd_yananClaim   = 20
	cd_yananGun     = 10
)

//房间信息
type YananRoom struct {
	room
	rule              *mahjong.YananRule
	userIDPlayerDatas map[int]*YananPlayerData
	tiles             []int // 洗好的牌
	currentRound      int   // 第几局
	dealerUserID      int   // 庄家ID
	jokers            []int // 宝牌
	discards          []int // 玩家打的牌
	rests             []int // 剩余的牌
	kongCount         int   // 扛牌的次数
	drawerUserID      int   // 最近一次摸牌的人ime
	discarderUserID   int   // 最近一次出牌的人

	actionWinUsers  map[int]int // 可以胡的玩家 key:userID value:0、1
	actionKongUsers map[int]int // 可以杠的玩家 key:userID value:0、1
	actionPongUsers map[int]int // 可以碰的玩家 key:userID value:0、1

	discardTimer *timer.Timer
	claimTimer   *timer.Timer
	disbandTimer *timer.Timer
	setGunTimer  *timer.Timer
	claimUserID  int

	disbandApplicantUserID int //申请解散者房间
	winnerUserIDs          []int
}

//玩家数据
type YananPlayerData struct {
	user            *User
	totalResultData *TotalResultData
	roundResultData *RoundResultData
	state           int
	position        int  // 用户在桌子上的位置，从 0 开始
	owner           bool // 房主
	hands           []int
	analyzer        *mahjong.YananAnalyzer
	dealer          bool  // 庄家
	draw            int   // 摸的一张牌
	discards        []int // 打出的牌
	managed         bool  //是否托管
	discardsCount   int   // 记录自动出牌的次数
	claimActionCode int
	gun             int     //下的炮子数
	claims          [][]int // 吃、碰、杠到的牌
	actionTimestamp int64   // 记录操作时间戳

	claim       int   // 待吃、碰、杠的牌
	quadruplet  []int //四连对
	quadruplets [][]int
	triplet     []int // 刻子
	sequence    []int
	sequences   [][]int

	kongType    int
	winType     int
	roundResult *mahjong.YananPlayerRoundResult
	totalResult *mahjong.YananPlayerTotalResult

	disbandActionCode int
	winTiles          []int
}

func newYananRoom(yananRule *mahjong.YananRule) *YananRoom {
	yananRoom := new(YananRoom)
	yananRoom.state = roomIdle
	yananRoom.loginIPs = make(map[string]bool)
	yananRoom.positionUserIDs = make(map[int]int)
	yananRoom.userIDPlayerDatas = make(map[int]*YananPlayerData)
	yananRoom.actionWinUsers = make(map[int]int)
	yananRoom.actionPongUsers = make(map[int]int)
	yananRoom.actionKongUsers = make(map[int]int)

	yananRoom.currentRound = 1
	yananRoom.rule = yananRule

	if yananRule.RedDragonJoker {
		yananRoom.jokers = []int{31}
	}
	win := "点炮胡" // 胡法
	if yananRule.MustSelfDraw {
		win = "自摸胡"
	}
	redDragonJoker := ""
	if yananRule.RedDragonJoker {
		redDragonJoker = "红中癞子"
	}
	gun := ""
	if yananRule.Gun {
		gun = "下炮子"
	}
	switch yananRoom.rule.RoomType {
	case roomPrivate:
		yananRoom.desc = fmt.Sprintf("%v %v %v局 %v人 底分%v分 %v", win, redDragonJoker, yananRule.MaxRounds, yananRule.MaxPlayers, yananRule.BaseScore, gun)
	case roomRoomCardMatch:
		yananRoom.desc = fmt.Sprintf("%v人 底注%v房卡 %v %v", yananRule.MaxPlayers, yananRule.RoomCards, redDragonJoker, win)
	case roomRedPacketMatching, roomRedPacketPrivate:
		yananRoom.desc = fmt.Sprintf("%v人 %v元红包 %v %v", yananRule.MaxPlayers, yananRule.RedPacketType, redDragonJoker, win)
	}
	if yananRoom.rule.IPAntiCheat {
		yananRoom.desc += " IP防作弊"
	}
	if yananRoom.rule.GPSAntiCheat {
		yananRoom.desc += " GPS防作弊"
	}
	return yananRoom
}

//玩家全部准备
func (yananRoom *YananRoom) allReady() bool {
	count := 0
	if yananRoom.full() {
		for _, userID := range yananRoom.positionUserIDs {
			playerData := yananRoom.userIDPlayerDatas[userID]
			if playerData.state == yananReady {
				count++
			}
		}
		if count == yananRoom.rule.MaxPlayers {
			return true
		}
	}
	return false
}

//玩家全部同意
func (yananRoom *YananRoom) allAgree() bool {
	count := 0
	for _, userID := range yananRoom.positionUserIDs {
		playerData := yananRoom.userIDPlayerDatas[userID]
		if playerData.disbandActionCode == actionAgreeDisband {
			count++
		}
	}
	if count == yananRoom.rule.MaxPlayers {
		return true
	}
	return false
}

//房间空
func (yananRoom *YananRoom) empty() bool {
	return len(yananRoom.positionUserIDs) == 0
}

//房间满
func (yananRoom *YananRoom) full() bool {
	return len(yananRoom.positionUserIDs) == yananRoom.rule.MaxPlayers
}

//清空玩家数据
func (yananRoom *YananRoom) clean() {
	for _, userId := range yananRoom.positionUserIDs {
		delete(userIDRooms, userId)
	}
	for pos := range yananRoom.positionUserIDs {
		delete(yananRoom.positionUserIDs, pos)
	}
	for userID, playerData := range yananRoom.userIDPlayerDatas {
		playerData.user.Location = []float64{}
		delete(yananRoom.userIDPlayerDatas, userID)
	}
	if yananRoom.claimTimer != nil {
		yananRoom.claimTimer.Stop()
		yananRoom.claimTimer = nil
	}
	if yananRoom.discardTimer != nil {
		yananRoom.discardTimer.Stop()
		yananRoom.discardTimer = nil
	}
}

func (yananRoom *YananRoom) allSetGun() bool {
	for _, userID := range yananRoom.positionUserIDs {
		playerData := yananRoom.userIDPlayerDatas[userID]
		if playerData.state == yananActionSetGun {
			return false
		}
	}
	return true
}

// 玩家进入房间
func (yananRoom *YananRoom) Enter(user *User) bool {
	roomCards := 0
	switch yananRoom.rule.RoomType {
	case roomRoomCardMatch, roomRedPacketMatching, roomRedPacketPrivate:
		roomCards = yananRoom.rule.RoomCards
	}
	if playerData, ok := yananRoom.userIDPlayerDatas[user.data.userData.UserID]; ok { // 断线重连
		playerData.user = user
		user.WriteMsg(&data_struct.S2C_EnterRoom{
			Error:          data_struct.S2C_EnterRoom_Ok,
			RoomType:       yananRoom.rule.RoomType,
			RoomNumber:     yananRoom.number,
			Position:       playerData.position,
			RoomDesc:       yananRoom.desc,
			MaxPlayers:     yananRoom.rule.MaxPlayers,
			MaxRounds:      yananRoom.rule.MaxRounds,
			RoomCards:      roomCards,
			RedDragonJoker: yananRoom.rule.RedDragonJoker,
			RedPacketType:  yananRoom.rule.RedPacketType,
			GamePlaying:    yananRoom.state == roomGame,
		})
		log.Debug("userID: %v 重连进入房间, 房间类型: %v", user.data.userData.UserID, yananRoom.rule.RoomType)
		return true
	}
	if yananRoom.full() {
		user.WriteMsg(&data_struct.S2C_EnterRoom{
			Error:      data_struct.S2C_EnterRoom_NotAllowBystander,
			RoomNumber: yananRoom.number,
		})
		return false
	}

	switch yananRoom.rule.RoomType {
	case roomRoomCardMatch, roomRedPacketMatching, roomRedPacketPrivate:
		if !user.checkEnterRoomCards(yananRoom.rule.RoomCards) {
			return false
		}
	}

	if yananRoom.rule.IPAntiCheat {
		if _, ok := yananRoom.loginIPs[user.data.userData.LoginIP]; ok {
			user.WriteMsg(&data_struct.S2C_EnterRoom{
				Error: data_struct.S2C_EnterRoom_IPConflict,
			})
			return false
		}
		yananRoom.loginIPs[user.data.userData.LoginIP] = true
	}
	for pos := 0; pos < yananRoom.rule.MaxPlayers; pos++ {
		log.Debug("roomcards: %v", roomCards)
		if _, ok := yananRoom.positionUserIDs[pos]; !ok {
			yananRoom.SitDown(user, pos)
			user.WriteMsg(&data_struct.S2C_EnterRoom{
				Error:          data_struct.S2C_EnterRoom_Ok,
				RoomType:       yananRoom.rule.RoomType,
				RoomNumber:     yananRoom.number,
				Position:       pos,
				RoomDesc:       yananRoom.desc,
				MaxPlayers:     yananRoom.rule.MaxPlayers,
				MaxRounds:      yananRoom.rule.MaxRounds,
				RoomCards:      roomCards,
				RedPacketType:  yananRoom.rule.RedPacketType,
				RedDragonJoker: yananRoom.rule.RedDragonJoker,
				GamePlaying:    yananRoom.state == roomGame,
			})
			log.Debug("userID: %v 进入房间, 房间类型: %v", user.data.userData.UserID, yananRoom.rule.RoomType)
			switch yananRoom.rule.RoomType {
			case roomRoomCardMatch:
				calculateRoomCardMatchOnlineNumber(yananRoom.rule.RoomCards, false)
			case roomRedPacketMatching, roomRedPacketPrivate:
				calculateRedPacketMatchOnlineNumber(yananRoom.rule.RedPacketType)
			}
			return true
		}
	}
	user.WriteMsg(&data_struct.S2C_EnterRoom{
		Error:      data_struct.S2C_EnterRoom_Unknow,
		RoomNumber: yananRoom.number,
	})
	return false
}

//退出房间
func (yananRoom *YananRoom) Exit(user *User) {
	log.Debug("userID: %v 退出房间", user.data.userData.UserID)
	playerData := yananRoom.userIDPlayerDatas[user.data.userData.UserID]
	if playerData == nil {
		return
	}
	broadcast(&data_struct.S2C_StandUp{
		Position: playerData.position,
	}, yananRoom.positionUserIDs, -1)

	broadcast(&data_struct.S2C_ExitRoom{
		Error:    data_struct.S2C_ExitRoom_OK,
		Position: playerData.position,
	}, yananRoom.positionUserIDs, -1)

	//站起
	yananRoom.StandUp(user, playerData.position)
	//退出
	delete(userIDRooms, user.data.userData.UserID)
	user.Location = []float64{}
	// 删除玩家登录IP
	delete(yananRoom.loginIPs, user.data.userData.LoginIP)

	switch yananRoom.rule.RoomType {
	case roomRoomCardMatch:
		calculateRoomCardMatchOnlineNumber(yananRoom.rule.RoomCards, true)
	}

	if yananRoom.empty() { //玩家为空 解散房间
		switch yananRoom.rule.RoomType {
		case roomPractice:
			delete(yananPracticeRooms, yananRoom.creatorUserID)
		case roomRoomCardMatch, roomRedPacketMatching:
			delete(yananRoomCardMatchRooms, yananRoom.creatorUserID)
		case roomPrivate, roomRedPacketPrivate:
			delete(roomNumberRooms, yananRoom.number)
		}
	}
}

//玩家坐下
func (yananRoom *YananRoom) SitDown(user *User, pos int) {
	yananRoom.positionUserIDs[pos] = user.data.userData.UserID
	playerData := yananRoom.userIDPlayerDatas[user.data.userData.UserID]
	if playerData == nil {
		playerData = new(YananPlayerData)
		playerData.user = user
		playerData.position = pos
		playerData.owner = user.data.userData.UserID == yananRoom.ownerUserID
		playerData.analyzer = new(mahjong.YananAnalyzer)
		playerData.roundResult = new(mahjong.YananPlayerRoundResult)
		playerData.totalResult = new(mahjong.YananPlayerTotalResult)

		yananRoom.userIDPlayerDatas[user.data.userData.UserID] = playerData
	}
	message := &data_struct.S2C_SitDown{
		Position:   pos,
		Owner:      playerData.owner,
		AccountID:  playerData.user.data.userData.AccountID,
		LoginIP:    playerData.user.data.userData.LoginIP,
		Nickname:   playerData.user.data.userData.Nickname,
		Headimgurl: playerData.user.data.userData.Headimgurl,
		Sex:        playerData.user.data.userData.Sex,
		Ready:      playerData.state == yananReady,
	}
	if yananRoom.rule.GPSAntiCheat {
		message.Location = playerData.user.Location
	}
	broadcast(message, yananRoom.positionUserIDs, pos)
}

//玩家站起
func (yananRoom *YananRoom) StandUp(user *User, pos int) {
	delete(yananRoom.positionUserIDs, pos)
	delete(yananRoom.userIDPlayerDatas, user.data.userData.UserID)
}

//获取玩家信息
func (yananRoom *YananRoom) GetAllPlayers(user *User) {
	for pos := 0; pos < yananRoom.rule.MaxPlayers; pos++ {
		userID := yananRoom.positionUserIDs[pos]
		playerData := yananRoom.userIDPlayerDatas[userID]
		if playerData == nil {
			user.WriteMsg(&data_struct.S2C_StandUp{
				Position: pos,
			})
		} else {
			if playerData.user.isRobot() {
				skeleton.AfterFunc(time.Duration(pos+1)*time.Second, func() {
					user.WriteMsg(&data_struct.S2C_SitDown{
						Position:   playerData.position,
						Owner:      playerData.owner,
						AccountID:  playerData.user.data.userData.AccountID,
						LoginIP:    playerData.user.data.userData.LoginIP,
						Nickname:   playerData.user.data.userData.Nickname,
						Headimgurl: playerData.user.data.userData.Headimgurl,
						Sex:        playerData.user.data.userData.Sex,
						Ready:      playerData.state == yananReady,
						Location:   playerData.user.Location,
					})
				})
			} else {
				user.WriteMsg(&data_struct.S2C_SitDown{
					Position:   pos,
					Owner:      playerData.owner,
					AccountID:  playerData.user.data.userData.AccountID,
					LoginIP:    playerData.user.data.userData.LoginIP,
					Nickname:   playerData.user.data.userData.Nickname,
					Headimgurl: playerData.user.data.userData.Headimgurl,
					Sex:        playerData.user.data.userData.Sex,
					Ready:      playerData.state == yananReady,
					Location:   playerData.user.Location,
				})
			}
		}
	}
}

//解散房间
func (yananRoom *YananRoom) Disband(disbander *User) {
	//等待开局
	if yananRoom.state == roomIdle {
		log.Debug("userID: %v 解散房间", disbander.data.userData.UserID)
		broadcast(&data_struct.S2C_DisbandRoom{
			Error:         data_struct.S2C_DisbandRoom_OK,
			RoomNumber:    yananRoom.number,
			OwnerNickName: disbander.data.userData.Nickname,
		}, yananRoom.positionUserIDs, -1)
		//清空玩家数据
		yananRoom.clean()

		if yananRoom.rule.RoomType == roomPrivate {
			delete(roomNumberRooms, yananRoom.number)
		}
		return
	}
	log.Debug("userID: %v 申请解散房间", disbander.data.userData.UserID)
	yananRoom.disbandApplicantUserID = disbander.data.userData.UserID
	applicantPlayerData := yananRoom.userIDPlayerDatas[yananRoom.disbandApplicantUserID]
	applicantPlayerData.disbandActionCode = actionAgreeDisband
	for i := 1; i < yananRoom.rule.MaxPlayers; i++ {
		otherUserId := yananRoom.positionUserIDs[(applicantPlayerData.position+i)%yananRoom.rule.MaxPlayers]
		otherPlayerData := yananRoom.userIDPlayerDatas[otherUserId]
		otherPlayerData.disbandActionCode = actionWaitingDisband
	}
	playerDisbandInfo := []mahjong.YananPlayerDisbandInfo{}
	for i := 0; i < yananRoom.rule.MaxPlayers; i++ {
		userID := yananRoom.positionUserIDs[i]
		playerData := yananRoom.userIDPlayerDatas[userID]
		playerDisbandInfo = append(playerDisbandInfo, mahjong.YananPlayerDisbandInfo{
			Nickname:   playerData.user.data.userData.Nickname,
			ActionCode: playerData.disbandActionCode,
		})
	}
	for _, userID := range yananRoom.positionUserIDs {
		playerData := yananRoom.userIDPlayerDatas[userID]
		if user, ok := userIDUsers[userID]; ok {
			user.WriteMsg(&data_struct.S2C_ActionDisbandRoom{
				ApplicantNickname:  disbander.data.userData.Nickname,
				PlayerDisbandInfos: playerDisbandInfo,
				Enable:             playerData.disbandActionCode == actionWaitingDisband,
				WaitingTime:        120,
			})
		}
	}
	//120s后自动同意
	yananRoom.disbandTimer = skeleton.AfterFunc(122*time.Second, func() {
		for _, userID := range yananRoom.positionUserIDs {
			playerData := yananRoom.userIDPlayerDatas[userID]
			if playerData.disbandActionCode == actionWaitingDisband {
				log.Debug("userID: %v 自动同意", playerData.user.data.userData.UserID)
				yananRoom.agreeDisbandRoom(playerData.user.data.userData.UserID)
			}
		}
	})
}

//开始游戏
func (yananRoom *YananRoom) StartGame() {
	yananRoom.state = roomGame
	yananRoom.prepare()

	broadcast(&data_struct.S2C_GameStart{},
		yananRoom.positionUserIDs, -1)

	broadcast(&data_struct.S2C_UpdateMahjongCurrentRound{
		CurrentRound: yananRoom.currentRound,
	}, yananRoom.positionUserIDs, -1)

	dealerPlayerData := yananRoom.userIDPlayerDatas[yananRoom.dealerUserID]
	broadcast(&data_struct.S2C_DecideDealer{
		Position: dealerPlayerData.position,
	}, yananRoom.positionUserIDs, -1)

	if yananRoom.rule.RedDragonJoker {
		broadcast(&data_struct.S2C_DecideYananJoker{
			Jokers: yananRoom.jokers,
		}, yananRoom.positionUserIDs, -1)
	}

	//每个玩家发十三张牌
	for _, userID := range yananRoom.positionUserIDs {
		playerData := yananRoom.userIDPlayerDatas[userID]
		playerData.state = yananWaiting
		//手中有13张牌
		playerData.hands = append(playerData.hands, yananRoom.rests[:13]...)

		//排序
		playerData.analyzer.Analyze(playerData.hands, yananRoom.jokers)
		playerData.hands = playerData.analyzer.Sort()
		log.Debug("userID: %v 手牌: %v", userID, mahjong.ToTileString(playerData.hands))
		//获取可以胡的牌
		playerData.winTiles = playerData.analyzer.GetWinTiles(playerData.hands)
		if len(playerData.winTiles) > 0 {
			//log.Debug("胡牌提示: %v", mahjong.ToTileString(playerData.winTiles))
		}
		// 剩余的牌
		yananRoom.rests = yananRoom.rests[13:]
		if user, ok := userIDUsers[userID]; ok {
			user.WriteMsg(&data_struct.S2C_UpdateMahjongHands{
				Position:      playerData.position,
				Hands:         playerData.hands,
				NumberOfHands: len(playerData.hands),
			})
			user.WriteMsg(&data_struct.S2C_UpdateWinTiles{
				Tiles: playerData.winTiles,
			})
		}
		broadcast(&data_struct.S2C_UpdateMahjongHands{
			Position:      playerData.position,
			NumberOfHands: len(playerData.hands),
		}, yananRoom.positionUserIDs, playerData.position)
	}

	switch yananRoom.rule.RoomType {
	case roomRedPacketMatching, roomRedPacketPrivate:
		yananRoom.deductRoomCard()
	}
	//庄家摸牌、出牌
	yananRoom.drawAndDiscard(yananRoom.dealerUserID)
}

//游戏结束
func (yananRoom *YananRoom) EndGame() {
	log.Debug("游戏结束, 剩余: %v", mahjong.ToTileString(yananRoom.rests))
	yananRoom.endTimestamp = time.Now().Unix()
	yananRoom.kongCount = 0
	yananRoom.state = roomGameEnd

	for _, userID := range yananRoom.positionUserIDs {
		playerData := yananRoom.userIDPlayerDatas[userID]
		playerData.winTiles = []int{}
		if user, ok := userIDUsers[userID]; ok {
			user.WriteMsg(&data_struct.S2C_UpdateWinTiles{
				Tiles: playerData.winTiles,
			})
		}
	}
	if yananRoom.currentRound == 1 {
		switch yananRoom.rule.RoomType {
		case roomPrivate, roomRoomCardMatch:
			yananRoom.deductRoomCard()
		}
	}
	totalResults, roundResults := []PlayerResultData{}, []PlayerResultData{}

	for pos := 0; pos < yananRoom.rule.MaxPlayers; pos++ {
		userID := yananRoom.positionUserIDs[pos]
		playerData := yananRoom.userIDPlayerDatas[userID]
		//计算总分
		roundResult := playerData.roundResult
		roundResult.TotalScore = roundResult.WinScore + roundResult.ExposedKongScore + roundResult.PongKongScore + roundResult.HiddenKongScore + roundResult.GunScore + roundResult.FollowDealerScore
		if len(yananRoom.winnerUserIDs) == 0 {
			roundResult.LastTile = -1
			roundResult.RoomCards = 0
		} else {
			if common.InArray(yananRoom.winnerUserIDs, userID) {
				roundResult.LastTile = playerData.claim
				if yananRoom.rule.RoomType == roomRoomCardMatch {
					roundResult.RoomCards = yananRoom.rule.RoomCards * (yananRoom.rule.MaxPlayers - 1)
				}
			} else {
				roundResult.LastTile = -1
				if yananRoom.rule.RoomType == roomRoomCardMatch {
					roundResult.RoomCards -= yananRoom.rule.RoomCards
				}
			}
		}
		totalResult := playerData.totalResult
		totalResult.Scores = append(totalResult.Scores, roundResult.TotalScore)
		totalResult.TotalScore += roundResult.TotalScore

		broadcast(&data_struct.S2C_UpdateYananToTalScore{
			Position:   pos,
			TotalScore: totalResult.TotalScore,
		}, yananRoom.positionUserIDs, -1)

		switch yananRoom.rule.RoomType {
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
	//保存成绩
	switch yananRoom.rule.RoomType {
	case roomRedPacketPrivate, roomRedPacketMatching:
		for _, userID := range yananRoom.positionUserIDs {
			playerData := yananRoom.userIDPlayerDatas[userID]
			saveRedPacketMatchResultData(&RedPacketMatchResultData{
				UserID:        userID,
				RedPacketType: yananRoom.rule.RedPacketType,
				RedPacket:     playerData.roundResult.RedPacket,
				Taken:         false,
				CreatedAt:     time.Now().Unix(),
			})
		}
	case roomPrivate, roomRoomCardMatch:
		yananRoom.saveUserTotalResultData(totalResults)
		yananRoom.saveUserRoundResultData(yananRoom.currentRound, roundResults)
	}
	for _, userID := range yananRoom.positionUserIDs {
		var roundResults []mahjong.YananPlayerRoundResult

		playerData := yananRoom.userIDPlayerDatas[userID]
		roundResults = append(roundResults, mahjong.YananPlayerRoundResult{
			Nickname:          playerData.user.data.userData.Nickname,
			Headimgurl:        playerData.user.data.userData.Headimgurl,
			Dealer:            playerData.dealer,
			Hands:             playerData.hands,
			Claims:            playerData.claims,
			LastTile:          playerData.roundResult.LastTile,
			WinType:           playerData.roundResult.WinType,
			Gun:               playerData.gun,
			WinScore:          playerData.roundResult.WinScore,
			ExposedKongScore:  playerData.roundResult.ExposedKongScore,
			PongKongScore:     playerData.roundResult.PongKongScore,
			HiddenKongScore:   playerData.roundResult.HiddenKongScore,
			GunScore:          playerData.roundResult.GunScore,
			FollowDealerScore: playerData.roundResult.FollowDealerScore,
			TotalScore:        playerData.roundResult.TotalScore,
			RoomCards:         playerData.roundResult.RoomCards,
			RedPacket:         playerData.roundResult.RedPacket,
		})

		for i := 1; i < yananRoom.rule.MaxPlayers; i++ {
			otherUserID := yananRoom.positionUserIDs[(playerData.position+i)%yananRoom.rule.MaxPlayers]
			otherPlayerData := yananRoom.userIDPlayerDatas[otherUserID]
			roundResults = append(roundResults, mahjong.YananPlayerRoundResult{
				Nickname:          otherPlayerData.user.data.userData.Nickname,
				Headimgurl:        otherPlayerData.user.data.userData.Headimgurl,
				Dealer:            otherPlayerData.dealer,
				Hands:             otherPlayerData.hands,
				Claims:            otherPlayerData.claims,
				LastTile:          otherPlayerData.roundResult.LastTile,
				Gun:               otherPlayerData.gun,
				WinType:           otherPlayerData.roundResult.WinType,
				WinScore:          otherPlayerData.roundResult.WinScore,
				ExposedKongScore:  otherPlayerData.roundResult.ExposedKongScore,
				PongKongScore:     otherPlayerData.roundResult.PongKongScore,
				HiddenKongScore:   otherPlayerData.roundResult.HiddenKongScore,
				GunScore:          otherPlayerData.roundResult.GunScore,
				FollowDealerScore: otherPlayerData.roundResult.FollowDealerScore,
				TotalScore:        otherPlayerData.roundResult.TotalScore,
				RoomCards:         otherPlayerData.roundResult.RoomCards,
				RedPacket:         otherPlayerData.roundResult.RedPacket,
			})
		}

		if user, ok := userIDUsers[userID]; ok {
			result := mahjong.ResultLose
			if len(yananRoom.winnerUserIDs) == 0 { //无人胡牌
				result = mahjong.ResultDraw
			} else {
				if common.InArray(yananRoom.winnerUserIDs, userID) {
					result = mahjong.ResultWin
				}
			}
			continueGame := true
			switch yananRoom.rule.RoomType {
			case roomRedPacketPrivate, roomRedPacketMatching:
				continueGame = false
			case roomPrivate:
				continueGame = !(yananRoom.currentRound == yananRoom.rule.MaxRounds)
			}
			user.WriteMsg(&data_struct.S2C_YananRoundResult{
				Result:       result,
				RoomDesc:     yananRoom.desc,
				Jokers:       yananRoom.jokers,
				RoundResults: roundResults,
				ContinueGame: continueGame,
			})
		}
	}
	if yananRoom.currentRound < yananRoom.rule.MaxRounds {
		yananRoom.currentRound++
		return
	}

	switch yananRoom.rule.RoomType {
	case roomRoomCardMatch:
		yananRoom.calculateRoomCard()
	case roomPrivate:
		for _, userID := range yananRoom.positionUserIDs {
			var playerTotalResults []mahjong.YananPlayerTotalResult
			playerData := yananRoom.userIDPlayerDatas[userID]
			playerTotalResults = append(playerTotalResults, mahjong.YananPlayerTotalResult{
				Nickname:   playerData.user.data.userData.Nickname,
				Headimgurl: playerData.user.data.userData.Headimgurl,
				Owner:      playerData.owner,
				AccountID:  playerData.user.data.userData.AccountID,
				Scores:     playerData.totalResult.Scores,
				TotalScore: playerData.totalResult.TotalScore,
			})

			for i := 1; i < yananRoom.rule.MaxPlayers; i++ {
				otherUserID := yananRoom.positionUserIDs[(playerData.position+i)%yananRoom.rule.MaxPlayers]
				othererPlayerData := yananRoom.userIDPlayerDatas[otherUserID]
				playerTotalResults = append(playerTotalResults, mahjong.YananPlayerTotalResult{
					Nickname:   othererPlayerData.user.data.userData.Nickname,
					Headimgurl: othererPlayerData.user.data.userData.Headimgurl,
					Owner:      othererPlayerData.owner,
					AccountID:  othererPlayerData.user.data.userData.AccountID,
					Scores:     othererPlayerData.totalResult.Scores,
					TotalScore: othererPlayerData.totalResult.TotalScore,
				})
			}
			if user, ok := userIDUsers[userID]; ok {
				user.WriteMsg(&data_struct.S2C_YananTotalResult{
					TotalResults: playerTotalResults,
				})
				//保存游戏积分
				user.data.userData.GameScore += playerData.totalResult.TotalScore
			} else {
				playerData.user.data.userData.GameScore += playerData.totalResult.TotalScore
				updateUserData(userID, bson.M{"$set": bson.M{"gamescore": playerData.user.data.userData.GameScore}})
			}
		}
	}
	//清空玩家数据
	yananRoom.clean()
	//删除房间
	switch yananRoom.rule.RoomType {
	case roomPrivate, roomRedPacketPrivate:
		delete(roomNumberRooms, yananRoom.number) // 解散房间
	}
}

//同意解散房间
func (yananRoom *YananRoom) agreeDisbandRoom(userID int) {
	if yananRoom.state == roomIdle {
		return
	}

	playerData := yananRoom.userIDPlayerDatas[userID]
	if playerData.disbandActionCode != actionWaitingDisband {
		return
	}
	playerData.disbandActionCode = actionAgreeDisband
	if yananRoom.allAgree() {
		if yananRoom.currentRound == 1 && yananRoom.rule.RoomType == roomPrivate {
			yananRoom.deductRoomCard()
		}

		broadcast(&data_struct.S2C_DisbandRoom{
			Error:      data_struct.S2C_DisbandRoom_OK,
			RoomNumber: yananRoom.number,
		}, yananRoom.positionUserIDs, -1)
		yananRoom.clean()

		if yananRoom.rule.RoomType == roomPrivate {
			delete(roomNumberRooms, yananRoom.number)
		}
	} else {
		broadcast(&data_struct.S2C_AgreeDisbandRoom{
			Position: playerData.position,
			Nickname: playerData.user.data.userData.Nickname,
		}, yananRoom.positionUserIDs, -1)
	}
}

//拒绝解散房间
func (yananRoom *YananRoom) refusedDisbandRoom(userID int) {
	playerData := yananRoom.userIDPlayerDatas[userID]
	if yananRoom.state == roomIdle || yananRoom.disbandTimer == nil || playerData.disbandActionCode != actionWaitingDisband {
		return
	}
	log.Debug("userID: %v 拒绝解散房间", userID)
	if yananRoom.disbandTimer != nil {
		yananRoom.disbandTimer.Stop()
		yananRoom.disbandTimer = nil
	}
	yananRoom.disbandApplicantUserID = -1

	broadcast(&data_struct.S2C_DisbandRoom{
		Error:            data_struct.S2C_DisbandRoom_PlayerRefuse,
		RoomNumber:       yananRoom.number,
		RejecterNickName: playerData.user.data.userData.Nickname,
	}, yananRoom.positionUserIDs, -1)
}

//断线重连
func (yananRoom *YananRoom) reconnect(user *User) {
	log.Debug("userID %v 断线重连", user.data.userData.UserID)
	playerData := yananRoom.userIDPlayerDatas[user.data.userData.UserID]
	if playerData == nil {
		return
	}

	if yananRoom.setGunTimer == nil {
		user.WriteMsg(&data_struct.S2C_GameStart{})
		user.WriteMsg(&data_struct.S2C_UpdateWinTiles{
			Tiles: playerData.winTiles,
		})
	}
	if yananRoom.dealerUserID > 0 {
		dealerPlayerData := yananRoom.userIDPlayerDatas[yananRoom.dealerUserID]
		// log.Debug("胡牌提示: %v", mahjong.ToTileString(playerData.winTiles))
		user.WriteMsg(&data_struct.S2C_DecideDealer{
			Position: dealerPlayerData.position,
		})
	}

	if yananRoom.rule.RedDragonJoker {
		user.WriteMsg(&data_struct.S2C_DecideYananJoker{
			Jokers: yananRoom.jokers,
		})
	}

	user.WriteMsg(&data_struct.S2C_UpdateMahjongRestsNumber{
		NumberOfRests: len(yananRoom.rests),
	})
	user.WriteMsg(&data_struct.S2C_UpdateMahjongCurrentRound{
		CurrentRound: yananRoom.currentRound,
	})
	if yananRoom.claimUserID < 1 && yananRoom.discarderUserID > 0 {
		discarderPlayerData := yananRoom.userIDPlayerDatas[yananRoom.discarderUserID]
		user.WriteMsg(&data_struct.S2C_UpdateMahjongDiscardCusor{
			Position: discarderPlayerData.position,
			Index:    len(discarderPlayerData.discards) - 1,
		})
	}
	if yananRoom.disbandApplicantUserID > 0 {
		applicantPlayerData := yananRoom.userIDPlayerDatas[yananRoom.disbandApplicantUserID]
		playerDisbandInfos := []mahjong.YananPlayerDisbandInfo{}
		for i := 0; i < yananRoom.rule.MaxPlayers; i++ {
			userID := yananRoom.positionUserIDs[i]
			playerData := yananRoom.userIDPlayerDatas[userID]
			playerDisbandInfos = append(playerDisbandInfos, mahjong.YananPlayerDisbandInfo{
				Nickname:   playerData.user.data.userData.Nickname,
				ActionCode: playerData.disbandActionCode,
			})
		}
		user.WriteMsg(&data_struct.S2C_ActionDisbandRoom{
			ApplicantNickname:  applicantPlayerData.user.data.userData.Nickname,
			PlayerDisbandInfos: playerDisbandInfos,
			Enable:             playerData.disbandActionCode == actionWaitingDisband,
			WaitingTime:        300,
		})
	}
	yananRoom.getPlayerData(user, playerData, false)

	for i := 1; i < yananRoom.rule.MaxPlayers; i++ {
		otherUserID := yananRoom.positionUserIDs[(playerData.position+i)%yananRoom.rule.MaxPlayers]
		otherPlayerData := yananRoom.userIDPlayerDatas[otherUserID]
		yananRoom.getPlayerData(user, otherPlayerData, true)
	}
}

//获取玩家数据
func (yananRoom *YananRoom) getPlayerData(user *User, playerData *YananPlayerData, other bool) {
	if yananRoom.rule.Gun && playerData.gun > 0 {
		user.WriteMsg(&data_struct.S2C_SetGun{
			Position: playerData.position,
			Gun:      playerData.gun,
		})
	}
	if yananRoom.setGunTimer != nil {
		switch playerData.state {
		case yananActionSetGun:
			if other {
				return
			}
			after := int(time.Now().Unix() - playerData.actionTimestamp)
			countdown := cd_yananGun - after
			if countdown > 1 {
				user.WriteMsg(&data_struct.S2C_ActionSetGun{
					Countdown: countdown - 1,
				})
			}
		}
		return
	}
	user.WriteMsg(&data_struct.S2C_UpdateMahjongDiscads{
		Position: playerData.position,
		Discards: playerData.discards,
	})
	user.WriteMsg(&data_struct.S2C_UpdateMahjongClaims{
		Position: playerData.position,
		Claims:   playerData.claims,
	})
	hands := playerData.hands
	if other {
		hands = []int{}
	}
	user.WriteMsg(&data_struct.S2C_UpdateMahjongHands{
		Position:      playerData.position,
		Hands:         hands,
		NumberOfHands: len(playerData.hands),
	})
	if playerData.draw > -1 {
		draw := playerData.draw
		if other {
			draw = -1
		}
		user.WriteMsg(&data_struct.S2C_MahjongDraw{
			Position:      playerData.position,
			Tile:          draw,
			NumberOfHands: len(playerData.hands),
		})
	}
	user.WriteMsg(&data_struct.S2C_UpdateYananToTalScore{
		Position:   playerData.position,
		TotalScore: playerData.totalResult.TotalScore,
	})
	switch playerData.state {
	case yananActionDiscard:
		after := int(time.Now().Unix() - playerData.actionTimestamp)
		countdown := cd_yananDiscard - after
		if countdown > 1 {
			user.WriteMsg(&data_struct.S2C_ActionMahjongDiscard{
				Position:  playerData.position,
				Countdown: countdown - 1,
			})
		}
	case yananActionClaim:
		after := int(time.Now().Unix() - playerData.actionTimestamp)
		countdown := cd_yananClaim - after
		if countdown > 1 {
			user.WriteMsg(&data_struct.S2C_ActionMahjongClaim{
				Position:    playerData.position,
				ActionCode:  playerData.claimActionCode,
				Countdown:   countdown - 1,
				Quadruplets: playerData.quadruplets,
			})
		}
		if playerData.managed {
			user.WriteMsg(&data_struct.S2C_ManagedMahjongPass{})
		}
	}
}

//重置碰、杠、胡的玩家
func (yananRoom *YananRoom) resetActionClaimUsers() {
	if yananRoom.claimTimer != nil {
		yananRoom.claimTimer.Stop()
		yananRoom.claimTimer = nil
	}

	for _, userID := range yananRoom.positionUserIDs {
		yananRoom.deleteActionClaimUsers(userID)
	}
}

//删除碰、杠、胡的玩家
func (yananRoom *YananRoom) deleteActionClaimUsers(userID int) {
	playerData := yananRoom.userIDPlayerDatas[userID]
	if playerData.state == yananActionClaim {
		playerData.state = yananWaiting
		playerData.claimActionCode = 0
	}
	delete(yananRoom.actionWinUsers, userID)
	delete(yananRoom.actionKongUsers, userID)
	delete(yananRoom.actionPongUsers, userID)

}

func (yananRoom *YananRoom) actionClaimUsersEmpty() bool {
	if len(yananRoom.actionWinUsers) == 0 && len(yananRoom.actionKongUsers) == 0 && len(yananRoom.actionPongUsers) == 0 {
		return true
	}
	return false
}

//准备游戏
func (yananRoom *YananRoom) prepare() {
	// 洗牌
	if yananRoom.rule.WithHonors {
		yananRoom.tiles = common.Shuffle(mahjong.YananAllTiles)
	} else if !yananRoom.rule.WithHonors && yananRoom.rule.RedDragonJoker {
		yananRoom.tiles = common.Shuffle(mahjong.YananAllTilesWithoutHonorsWithRed)
	} else {
		yananRoom.tiles = common.Shuffle(mahjong.YananAllTilesWithoutHonors)
	}
	//确定庄家
	if yananRoom.currentRound == 1 {
		yananRoom.dealerUserID = yananRoom.positionUserIDs[0]
		switch yananRoom.rule.RoomType {
		case roomPrivate, roomRoomCardMatch, roomRedPacketMatching, roomRedPacketPrivate:
			yananRoom.startTimestamp = time.Now().Unix()
			yananRoom.eachRoundStartTimestamp = yananRoom.startTimestamp
			yananRoom.initTotalResultData()
		}
	} else {
		switch yananRoom.rule.RoomType {
		case roomPrivate, roomRoomCardMatch, roomRedPacketMatching, roomRedPacketPrivate:
			yananRoom.eachRoundStartTimestamp = time.Now().Unix()
		}

		if len(yananRoom.winnerUserIDs) > 0 {
			yananRoom.dealerUserID = yananRoom.winnerUserIDs[0]
		}
	}
	//确定庄家
	dealerPlayerData := yananRoom.userIDPlayerDatas[yananRoom.dealerUserID]
	dealerPlayerData.dealer = true
	//确定闲家(闲家：player)
	dealerPosition := dealerPlayerData.position
	for i := 1; i < yananRoom.rule.MaxPlayers; i++ {
		playerPos := (dealerPosition + i) % yananRoom.rule.MaxPlayers
		playerUserID := yananRoom.positionUserIDs[playerPos]
		playerData := yananRoom.userIDPlayerDatas[playerUserID]
		playerData.dealer = false
	}

	if yananRoom.rule.RedDragonJoker {
		log.Debug("癞子: %v", mahjong.ToTileString(yananRoom.jokers))
	}

	//玩家打的牌为空
	yananRoom.discards = []int{}
	//剩余的牌
	yananRoom.rests = append([]int{}, yananRoom.tiles...)

	yananRoom.discarderUserID = -1
	yananRoom.drawerUserID = -1
	yananRoom.resetActionClaimUsers()
	yananRoom.claimUserID = -1
	yananRoom.disbandApplicantUserID = -1
	yananRoom.winnerUserIDs = []int{}

	for _, userID := range yananRoom.positionUserIDs {
		playerData := yananRoom.userIDPlayerDatas[userID]
		playerData.draw = -1
		playerData.hands = []int{}
		playerData.discards = []int{}
		playerData.claims = [][]int{}
		playerData.managed = false
		playerData.discardsCount = 0
		playerData.actionTimestamp = 0

		roundresult := playerData.roundResult
		roundresult.WinType = 0
		roundresult.WinScore = 0
		roundresult.ExposedKongScore = 0
		roundresult.PongKongScore = 0
		roundresult.HiddenKongScore = 0
		roundresult.FollowDealerScore = 0
		roundresult.TotalScore = 0
		if yananRoom.rule.Gun {
			roundresult.GunScore = 0
		}
	}
}

//更新出牌
func (yananRoom *YananRoom) updateDiscards(userID int) {
	playerData := yananRoom.userIDPlayerDatas[userID]
	broadcast(&data_struct.S2C_UpdateMahjongDiscads{
		Position: playerData.position,
		Discards: playerData.discards,
	}, yananRoom.positionUserIDs, -1)
}

//更新碰、杠、胡的牌
func (shaanxRoom *YananRoom) updateClaims(userID int) {
	playerData := shaanxRoom.userIDPlayerDatas[userID]
	broadcast(&data_struct.S2C_UpdateMahjongClaims{
		Position: playerData.position,
		Claims:   playerData.claims,
	}, shaanxRoom.positionUserIDs, -1)
}

//计算胡牌得分
func (yananRoom *YananRoom) calculateWinScore(winnerUserID int) {
	if common.Index(yananRoom.winnerUserIDs, winnerUserID) == -1 {
		return
	}
	winnerPlayerData := yananRoom.userIDPlayerDatas[winnerUserID]
	winnerWinType := winnerPlayerData.roundResult.WinType

	switch winnerWinType {
	case mahjong.YananWinByDiscard:
		discarderPlayerData := yananRoom.userIDPlayerDatas[yananRoom.discarderUserID]
		if discarderPlayerData.dealer {
			discarderPlayerData.roundResult.WinScore = -4 * yananRoom.rule.BaseScore
		} else {
			if winnerPlayerData.dealer {
				discarderPlayerData.roundResult.WinScore = -6 * yananRoom.rule.BaseScore
			} else {
				discarderPlayerData.roundResult.WinScore = -3 * yananRoom.rule.BaseScore
			}
		}
	case mahjong.YananWinBySelfDraw:
		for i := 1; i < yananRoom.rule.MaxPlayers; i++ {
			otherUserID := yananRoom.positionUserIDs[(winnerPlayerData.position+i)%yananRoom.rule.MaxPlayers]
			otherPlayerData := yananRoom.userIDPlayerDatas[otherUserID]
			if winnerPlayerData.dealer {
				otherPlayerData.roundResult.WinScore = -4 * yananRoom.rule.BaseScore
			} else {
				if otherPlayerData.dealer {
					otherPlayerData.roundResult.WinScore = -4 * yananRoom.rule.BaseScore
				} else {
					otherPlayerData.roundResult.WinScore = -2 * yananRoom.rule.BaseScore
				}
			}
		}
	}

	// 计算赢家胡牌分
	loserWinScore := 0
	for i := 1; i < yananRoom.rule.MaxPlayers; i++ {
		otherUserID := yananRoom.positionUserIDs[(winnerPlayerData.position+i)%yananRoom.rule.MaxPlayers]
		otherPlayerData := yananRoom.userIDPlayerDatas[otherUserID]
		loserWinScore += otherPlayerData.roundResult.WinScore
	}
	winnerPlayerData.roundResult.WinScore = -1 * loserWinScore

}

//计算下炮子分数
func (yananRoom *YananRoom) calculateGunScore(winnerUserID int) {
	winnerPlayerData := yananRoom.userIDPlayerDatas[winnerUserID]
	winnerWinType := winnerPlayerData.roundResult.WinType

	switch winnerWinType {
	case mahjong.YananWinByDiscard, mahjong.YananWinBySelfDraw:
		for i := 1; i < yananRoom.rule.MaxPlayers; i++ {
			otherUserID := yananRoom.positionUserIDs[(winnerPlayerData.position+i)%yananRoom.rule.MaxPlayers]
			otherPlayerData := yananRoom.userIDPlayerDatas[otherUserID]
			otherPlayerData.roundResult.GunScore = -otherPlayerData.gun - winnerPlayerData.gun
		}
	}
	//计算赢家炮子分
	loserGunScore := 0
	for i := 1; i < yananRoom.rule.MaxPlayers; i++ {
		otherUserID := yananRoom.positionUserIDs[(winnerPlayerData.position+i)%yananRoom.rule.MaxPlayers]
		otherPlayerData := yananRoom.userIDPlayerDatas[otherUserID]
		loserGunScore += otherPlayerData.roundResult.GunScore
	}
	winnerPlayerData.roundResult.GunScore = -1 * loserGunScore
}

func (yananRoom *YananRoom) calculateRedPacket(userID int, redPacketType int) {
	if common.Index([]int{1, 10, 100, 999}, redPacketType) == -1 {
		return
	}
	playerData := yananRoom.userIDPlayerDatas[userID]
	roundResult := playerData.roundResult
	roundResult.RedPacket = float64(yananRoom.rule.RedPacketType)
}

//扣除房卡
func (yananRoom *YananRoom) deductRoomCard() {
	switch yananRoom.rule.RoomType {
	case roomRoomCardMatch, roomRedPacketMatching, roomRedPacketPrivate:
		for _, userID := range yananRoom.positionUserIDs {
			playerData := yananRoom.userIDPlayerDatas[userID]
			if user, ok := userIDUsers[userID]; ok {
				user.data.userData.RoomCards -= yananRoom.rule.RoomCards
				user.data.userData.ConsumedRoomCards += yananRoom.rule.RoomCards
			} else {
				playerData.user.data.userData.RoomCards -= yananRoom.rule.RoomCards
				playerData.user.data.userData.ConsumedRoomCards += yananRoom.rule.RoomCards
				updateUserData(userID, bson.M{"$set": bson.M{"roomcards": playerData.user.data.userData.RoomCards, "consumedroomcards": playerData.user.data.userData.ConsumedRoomCards}})
			}

			if playerData.user.isRobot() {
				cards := -yananRoom.rule.RoomCards
				switch yananRoom.rule.RoomType {
				case roomRoomCardMatch:
					upsertRobotData(time.Now().Format("20060102"), bson.M{"$inc": bson.M{"roomcardmatchbalance": cards}})
				case roomRedPacketMatching:
					upsertRobotData(time.Now().Format("20060102"), bson.M{"$inc": bson.M{"redpacketmatchbalance": cards}})
				}
			}
		}
	case roomPrivate:
		if owner, ok := userIDUsers[yananRoom.ownerUserID]; ok {
			owner.data.userData.RoomCards -= yananRoom.rule.RoomCards
			owner.data.userData.ConsumedRoomCards += yananRoom.rule.RoomCards
		} else {
			playerData := yananRoom.userIDPlayerDatas[yananRoom.ownerUserID]
			playerData.user.data.userData.RoomCards -= yananRoom.rule.RoomCards
			playerData.user.data.userData.ConsumedRoomCards += yananRoom.rule.RoomCards
			updateUserData(yananRoom.ownerUserID, bson.M{"$set": bson.M{"roomcards": playerData.user.data.userData.RoomCards, "consumedroomcards": playerData.user.data.userData.ConsumedRoomCards}})
		}
	}
}

// 结算房卡
func (yananRoom *YananRoom) calculateRoomCard() {
	if len(yananRoom.winnerUserIDs) == 0 { // 流局
		for _, userID := range yananRoom.positionUserIDs {
			playerData := yananRoom.userIDPlayerDatas[userID]
			if user, ok := userIDUsers[userID]; ok {
				user.data.userData.RoomCards += yananRoom.rule.RoomCards
				user.WriteMsg(&data_struct.S2C_UpdateRoomCards{
					RoomCards: user.data.userData.RoomCards,
				})
			} else {
				playerData.user.data.userData.RoomCards += yananRoom.rule.RoomCards
				updateUserData(userID, bson.M{"$set": bson.M{"roomcards": playerData.user.data.userData.RoomCards}})
			}
			if playerData.user.isRobot() {
				cards := yananRoom.rule.RoomCards
				upsertRobotData(time.Now().Format("20060102"), bson.M{"$inc": bson.M{"roomcardmatchbalance": cards}})
			}
		}
	} else {
		winnerUserID := yananRoom.winnerUserIDs[0]
		playerData := yananRoom.userIDPlayerDatas[winnerUserID]
		if user, ok := userIDUsers[winnerUserID]; ok {
			user.data.userData.RoomCards += yananRoom.rule.RoomCards * yananRoom.rule.MaxPlayers
			user.WriteMsg(&data_struct.S2C_UpdateRoomCards{
				RoomCards: user.data.userData.RoomCards,
			})
		} else {
			playerData.user.data.userData.RoomCards += yananRoom.rule.RoomCards * yananRoom.rule.MaxPlayers
			updateUserData(winnerUserID, bson.M{"$set": bson.M{"roomcards": playerData.user.data.userData.RoomCards}})
		}
		if playerData.user.isRobot() {
			cards := yananRoom.rule.RoomCards * yananRoom.rule.MaxPlayers
			upsertRobotData(time.Now().Format("20060102"), bson.M{"$inc": bson.M{"roomcardmatchbalance": cards}})
		}
	}
}

//初始数据
func (yananRoom *YananRoom) initTotalResultData() {
	for _, userID := range yananRoom.positionUserIDs {
		playerData := yananRoom.userIDPlayerDatas[userID]
		playerData.totalResultData = new(TotalResultData)
		skeleton.Go(func() {
			err := playerData.totalResultData.initValue(playerData.user.data.userData.UserID)
			if err != nil {
				log.Error("init userID: %v totalresultdata errr: %v", playerData.user.data.userData.UserID, err)
				playerData.totalResultData = nil
			}
		}, func() {
			if playerData.totalResultData != nil {
				playerData.totalResultData.UserID = playerData.user.data.userData.UserID
				playerData.totalResultData.RoomNumber = yananRoom.number
				playerData.totalResultData.RoomDesc = yananRoom.desc
				playerData.totalResultData.StartTimestamp = yananRoom.startTimestamp
				playerData.totalResultData.Position = playerData.position
			}
		})
	}
}

//保存玩家总成绩
func (yananRoom *YananRoom) saveUserTotalResultData(results []PlayerResultData) {
	for pos := 0; pos < yananRoom.rule.MaxPlayers; pos++ {
		userID := yananRoom.positionUserIDs[pos]
		playerData := yananRoom.userIDPlayerDatas[userID]
		if playerData.totalResultData != nil {
			playerData.totalResultData.RoomType = yananRoom.rule.RoomType
			playerData.totalResultData.EndTimestamp = yananRoom.endTimestamp
			playerData.totalResultData.Results = results
			playerData.totalResultData.UpdatedAt = time.Now().Unix()

			saveTotalResultData(playerData.totalResultData)
		}
	}
}

//保存玩家单局的总成绩
func (yananRoom *YananRoom) saveUserRoundResultData(round int, results []PlayerResultData) {
	for pos := 0; pos < yananRoom.rule.MaxPlayers; pos++ {
		userID := yananRoom.positionUserIDs[pos]
		playerData := yananRoom.userIDPlayerDatas[userID]
		if playerData.totalResultData != nil {
			playerData.roundResultData = new(RoundResultData)
			skeleton.Go(func() {
				err := playerData.roundResultData.initValue(playerData.totalResultData.ID)
				if err != nil {
					log.Error("init totalresult: %v roundresule error: %v", playerData.totalResultData.ID, err)
					playerData.roundResultData = nil
				}
			}, func() {
				if playerData.roundResultData != nil {
					playerData.roundResultData.Round = round
					playerData.roundResultData.StartTimestamp = yananRoom.eachRoundStartTimestamp
					playerData.roundResultData.EndTimestamp = yananRoom.endTimestamp
					playerData.roundResultData.Position = playerData.position
					playerData.roundResultData.Results = results
					playerData.roundResultData.UpdatedAt = time.Now().Unix()
					saveRoundResultData(playerData.roundResultData)
				}
			})
		}
	}
}
