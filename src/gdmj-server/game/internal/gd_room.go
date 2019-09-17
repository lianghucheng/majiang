package internal

import (
	"fmt"
	"gdmj-server/common"
	"gdmj-server/game/mahjong"
	"gdmj-server/msg"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/timer"
	"gopkg.in/mgo.v2/bson"
	"time"
)

// 玩家状态
const (
	_               = iota
	gdReady         // 1 准备
	gdWaiting       // 2 等待
	gdActionDiscard // 3 前端显示出牌动作
	gdActionClaim   // 4 前端显示胡、杠、碰、吃动作
	gdWin           // 5 胡
	gdKong          // 6 杠
	gdPong          // 7 碰
	gdChow          // 8 吃
	ActionSetGun    // 9下炮子
)

// 倒计时
const (
	cd_gdDiscard = 20
	cd_gdClaim   = 20
	cd_yananGun  = 20
)

/*
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
*/
type GDRoom struct {
	gameType  int //游戏类型
	kongCount int // 扛牌的次数
	room
	rule              *mahjong.GDRule
	userIDPlayerDatas map[int]*GDPlayerData // key: userID
	tiles             []int                 // 洗好的牌
	currentRound      int                   // 第几局
	dealerUserID      int                   // 庄家 userID
	wildcard          int                   // 混儿，只有一张
	jokers            []int                 // 宝牌
	discards          []int                 // 玩家打的牌
	rests             []int                 // 剩余的牌
	drawerUserID      int                   // 最近一次摸牌的人 userID
	discarderUserID   int                   // 最近一次出牌的人 userID

	actionWinUsers  map[int]int // 可以胡的玩家 key: userID, value: 1
	actionKongUsers map[int]int // 可以杠的玩家 key: userID, value: 1
	actionPongUsers map[int]int // 可以碰的玩家 key: userID, value: 1
	actionChowUsers map[int]int // 可以吃的玩家 key: userID, value: 1

	discardTimer *timer.Timer // 出牌定时器
	claimTimer   *timer.Timer // 打牌后可以吃,碰,杠,胡得玩家操作
	disbandTimer *timer.Timer // 房间解散定时器
	setGunTimer  *timer.Timer // 下炮子倒计时
	claimUserID  int          // 当前谁可以操作

	countWinStreak         int // 连庄次数
	disbandApplicantUserID int // 申请解散房间者 userID
	winnerUserIDs          []int

	catchBirdUserID   int                     // 上一局抓鸟的玩家 userID---湖南转转麻将
}

// 玩家数据
type GDPlayerData struct {
	user            *User
	totalResultData *TotalResultData
	roundResultData *RoundResultData
	state           int
	position        int // 用户在桌子上的位置，从 0 开始

	owner           bool  // 房主
	dealer          bool  // 庄家
	draw            int   // 摸的一张牌
	hands           []int //手牌
	horseTile       []int // 马牌
	discards        []int // 打出的牌
	analyzer        *mahjong.GDAnalyzer
	claimActionCode int     //吃,碰,杠得状态码
	claims          [][]int // 吃、碰、杠到的牌
	actionTimestamp int64   // 记录操作时间戳
	managed         bool    //是否托管
	discardsCount   int     // 记录自动出牌的次数
	gun             int     //下的炮子数
	claim           int     // 待吃、碰、杠的牌
	quadruplet      []int
	quadruplets     [][]int //可以杠的牌型 一维表示长度 二维表示杠的牌
	triplet         []int
	sequence        []int
	sequences       [][]int

	kongType    int
	winType     int
	roundResult *mahjong.GDPlayerRoundResult
	totalResult *mahjong.GDPlayerTotalResult

	disbandActionCode int
	winTiles          []int
}

func newGDRoom(rule *mahjong.GDRule) *GDRoom {
	gdRoom := new(GDRoom)
	gdRoom.state = roomIdle
	gdRoom.loginIPs = make(map[string]bool)
	gdRoom.positionUserIDs = make(map[int]int)
	gdRoom.userIDPlayerDatas = make(map[int]*GDPlayerData)
	gdRoom.actionWinUsers = make(map[int]int)
	gdRoom.actionKongUsers = make(map[int]int)
	gdRoom.actionPongUsers = make(map[int]int)
	gdRoom.actionChowUsers = make(map[int]int)

	gdRoom.currentRound = 1
	gdRoom.rule = rule
	gdRoom.gameType = rule.GameType
	win := "点炮胡"
	if rule.MustSelfDraw {
		win = "自摸胡"
	}
	if gdRoom.gameType== 1 {
		redDragonJoker := ""
		if rule.NeedJoker {
			redDragonJoker = "癞子"
		}
		switch gdRoom.rule.RoomType {
		case roomPrivate:
			gdRoom.desc = fmt.Sprintf("%v %v %v局 %v人 底分%v分 %v匹马", win, redDragonJoker, rule.MaxRounds, rule.MaxPlayers, rule.BaseScore, rule.BuyHorse)
		case roomRoomCardMatch:
			gdRoom.desc = fmt.Sprintf("%v人 底注%v房卡 %v %v", rule.MaxPlayers, rule.RoomCards, redDragonJoker, win)
		case roomRedPacketMatching, roomRedPacketPrivate:
			gdRoom.desc = fmt.Sprintf("%v人 %v元红包 %v %v", rule.MaxPlayers, rule.RedPacketType, redDragonJoker, win)
		}
		if gdRoom.rule.IPAntiCheat {
			gdRoom.desc += " IP防作弊"
		}
		if gdRoom.rule.GPSAntiCheat {
			gdRoom.desc += " GPS防作弊"
		}
	}
	if gdRoom.gameType == 2 {
		redDragonJoker := ""
		if rule.RedDragonJoker {
			redDragonJoker = "红中癞子"
		}
		gun := ""
		if rule.Gun {
			gun = "下炮子"
		}
		switch gdRoom.rule.RoomType {
		case roomPrivate:
			gdRoom.desc = fmt.Sprintf("%v %v %v局 %v人 底分%v分 %v", win, redDragonJoker, rule.MaxRounds, rule.MaxPlayers, rule.BaseScore, gun)
		case roomRoomCardMatch:
			gdRoom.desc = fmt.Sprintf("%v人 底注%v房卡 %v %v", rule.MaxPlayers, rule.RoomCards, redDragonJoker, win)
		case roomRedPacketMatching, roomRedPacketPrivate:
			gdRoom.desc = fmt.Sprintf("%v人 %v元红包 %v %v", rule.MaxPlayers, rule.RedPacketType, redDragonJoker, win)
		}
		if gdRoom.rule.IPAntiCheat {
			gdRoom.desc += " IP防作弊"
		}
		if gdRoom.rule.GPSAntiCheat {
			gdRoom.desc += " GPS防作弊"
		}
	}
	if gdRoom.gameType == 3{
		gdRoom.jokers=[]int{31}
		distinguish := "通庄"
		if rule.DistinguishDealer {
			distinguish = "分庄闲"
		}
		_ = distinguish
		switch gdRoom.rule.RoomType {
		case roomPrivate:
			gdRoom.desc = fmt.Sprintf("%v 红中癞子 %v局 %v人 底分%v分 抓%v鸟", win, rule.MaxRounds, rule.MaxPlayers, rule.BaseScore, rule.Birds)
		case roomRoomCardMatch:
			gdRoom.desc = fmt.Sprintf("%v人 底注%v房卡 红中癞子 %v", rule.MaxPlayers, rule.RoomCards, win)
		case roomRedPacketMatching, roomRedPacketPrivate:
			gdRoom.desc = fmt.Sprintf("%v人 %v元红包 红中癞子 %v", rule.MaxPlayers, rule.RedPacketType, win)
		}
		if gdRoom.rule.IPAntiCheat {
			gdRoom.desc += " IP防作弊"
		}
		if gdRoom.rule.GPSAntiCheat {
			gdRoom.desc += " GPS防作弊"
		}
	}
	return gdRoom
}

func (gdRoom *GDRoom) allReady() bool {
	count := 0
	if gdRoom.full() {
		for _, userID := range gdRoom.positionUserIDs {
			playerData := gdRoom.userIDPlayerDatas[userID]
			if playerData.state == gdReady {
				count++
			}
		}
		if count == gdRoom.rule.MaxPlayers {
			return true
		}
		return false
	}
	return false
}

func (gdRoom *GDRoom) allAgree() bool {
	count := 0
	for _, userID := range gdRoom.positionUserIDs {
		playerData := gdRoom.userIDPlayerDatas[userID]
		if playerData.disbandActionCode == actionAgreeDisband {
			count++
		}
	}
	if count == gdRoom.rule.MaxPlayers {
		return true
	}
	return false
}

func (gdRoom *GDRoom) empty() bool {
	return len(gdRoom.positionUserIDs) == 0
}

func (gdRoom *GDRoom) full() bool {
	return len(gdRoom.positionUserIDs) == gdRoom.rule.MaxPlayers
}

func (gdRoom *GDRoom) clean() {
	for _, userID := range gdRoom.positionUserIDs {
		delete(userIDRooms, userID)
	}
	for pos := range gdRoom.positionUserIDs {
		delete(gdRoom.positionUserIDs, pos)
	}
	for userID, playerData := range gdRoom.userIDPlayerDatas {
		playerData.user.location = []float64{}
		delete(gdRoom.userIDPlayerDatas, userID)
	}
	if gdRoom.claimTimer != nil {
		gdRoom.claimTimer.Stop()
		gdRoom.claimTimer = nil
	}
	if gdRoom.discardTimer != nil {
		gdRoom.discardTimer.Stop()
		gdRoom.discardTimer = nil
	}
}

func (gdRoom *GDRoom) Enter(user *User) bool {
	roomCards := 0
	switch gdRoom.rule.RoomType {
	case roomRoomCardMatch, roomRedPacketMatching, roomRedPacketPrivate:
		roomCards = gdRoom.rule.RoomCards
	}
	if playerData, ok := gdRoom.userIDPlayerDatas[user.data.userData.UserID]; ok {
		playerData.user = user
		user.WriteMsg(&msg.S2C_EnterRoom{
			GameType:       gdRoom.gameType,
			Error:          msg.S2C_EnterRoom_OK,
			RoomType:       gdRoom.rule.RoomType,
			RedPacketType:  gdRoom.rule.RedPacketType,
			RoomNumber:     gdRoom.number,
			Position:       playerData.position,
			RoomDesc:       gdRoom.desc,
			MaxPlayers:     gdRoom.rule.MaxPlayers,
			MaxRounds:      gdRoom.rule.MaxRounds,
			NeedJoker:      gdRoom.rule.NeedJoker,
			BuyHorse:       gdRoom.rule.BuyHorse,
			RoomCards:      roomCards,
			GamePlaying:    gdRoom.state == roomGame,
			RedDragonJoker: gdRoom.rule.RedDragonJoker,
			Gun:            gdRoom.rule.Gun,
		})
		switch gdRoom.gameType {
		case mahjong.GD:
			time.Sleep(200 * time.Millisecond)
			user.getAllPlayers(gdRoom)
		case mahjong.YA:

		case mahjong.HNZZ:
		}

		log.Debug("userID: %v 重连进入房间， 房间类型: %v", user.data.userData.UserID, gdRoom.rule.RoomType)
		return true
	}
	// 玩家已满
	if gdRoom.full() {
		user.WriteMsg(&msg.S2C_EnterRoom{
			Error:      msg.S2C_EnterRoom_Full,
			RoomNumber: gdRoom.number,
		})
		return false
	}
	switch gdRoom.rule.RoomType {
	case roomRoomCardMatch, roomRedPacketMatching, roomRedPacketPrivate:
		if !user.checkEnterRoomCards(gdRoom.rule.RoomCards) {
			return false
		}
	}

	if gdRoom.rule.IPAntiCheat {
		if _, ok := gdRoom.loginIPs[user.data.userData.LoginIP]; ok {
			user.WriteMsg(&msg.S2C_EnterRoom{
				Error: msg.S2C_EnterRoom_IPConflict,
			})
			return false
		}
		gdRoom.loginIPs[user.data.userData.LoginIP] = true
	}
	for pos := 0; pos < gdRoom.rule.MaxPlayers; pos++ {
		if _, ok := gdRoom.positionUserIDs[pos]; !ok {
			gdRoom.SitDown(user, pos)
			user.WriteMsg(&msg.S2C_EnterRoom{
				GameType:       gdRoom.gameType,
				Error:          msg.S2C_EnterRoom_OK,
				RoomType:       gdRoom.rule.RoomType,
				RedPacketType:  gdRoom.rule.RedPacketType,
				RoomNumber:     gdRoom.number,
				Position:       pos,
				RoomDesc:       gdRoom.desc,
				MaxPlayers:     gdRoom.rule.MaxPlayers,
				MaxRounds:      gdRoom.rule.MaxRounds,
				NeedJoker:      gdRoom.rule.NeedJoker,
				BuyHorse:       gdRoom.rule.BuyHorse,
				RoomCards:      roomCards,
				GamePlaying:    gdRoom.state == roomGame,
				RedDragonJoker: gdRoom.rule.RedDragonJoker,
				Gun:            gdRoom.rule.Gun,
			})
			switch gdRoom.gameType {
			case mahjong.GD:
				time.Sleep(50 * time.Millisecond)
				user.getAllPlayers(gdRoom)
			case mahjong.YA:

			case mahjong.HNZZ:

			}
			log.Debug("userID: %v 进入房间, 房间类型: %v", user.data.userData.UserID, gdRoom.rule.RoomType)
			switch gdRoom.rule.RoomType {
			case roomRoomCardMatch:
				calculateRoomCardMatchOnlineNumber(gdRoom.rule.RoomCards, false)
			case roomRedPacketMatching, roomRedPacketPrivate:
				calculateRedPacketMatchOnlineNumber(gdRoom.rule.RedPacketType)
			}
			return true
		}
	}

	user.WriteMsg(&msg.S2C_EnterRoom{
		Error:      msg.S2C_EnterRoom_Unknown,
		RoomNumber: gdRoom.number,
	})
	return false
}

func (gdRoom *GDRoom) Exit(user *User) {
	playerData := gdRoom.userIDPlayerDatas[user.data.userData.UserID]
	if playerData == nil {
		return
	}
	broadcast(&msg.S2C_StandUp{
		Position: playerData.position,
	}, gdRoom.positionUserIDs, -1)
	log.Debug("userID: %v 退出房间", user.data.userData.UserID)
	broadcast(&msg.S2C_ExitRoom{
		Error:    msg.S2C_ExitRoom_OK,
		Position: playerData.position,
	}, gdRoom.positionUserIDs, -1)
	// 站起
	gdRoom.StandUp(user, playerData.position)
	// 退出
	delete(userIDRooms, user.data.userData.UserID)
	// 删除玩家登陆ip
	delete(gdRoom.loginIPs, user.data.userData.LoginIP)
	user.location = []float64{}

	switch gdRoom.rule.RoomType {
	case roomRoomCardMatch:
		calculateRoomCardMatchOnlineNumber(gdRoom.rule.RoomCards, true)
	}

	if gdRoom.empty() {
		switch gdRoom.gameType {
		case mahjong.GD:
			switch gdRoom.rule.RoomType {
			case roomPractice:
				delete(gdPracticeRooms, gdRoom.creatorUserID)
			case roomRoomCardMatch, roomRedPacketMatching:
				delete(gdMatchRooms, gdRoom.creatorUserID)
			case roomPrivate, roomRedPacketPrivate:
				delete(gdroomNumberRooms, gdRoom.number)
			}
		case mahjong.YA:

		case mahjong.HNZZ:
			switch gdRoom.rule.RoomType {
			case roomPractice:
				delete(hnzzPracticeRooms, gdRoom.creatorUserID)
			case roomRoomCardMatch, roomRedPacketMatching:
				delete(hnzzRoomCardMatchRooms, gdRoom.creatorUserID)
			case roomPrivate, roomRedPacketPrivate:
				delete(hnzzroomNumberRooms, gdRoom.number)
			}
		}
	}
}

func (gdRoom *GDRoom) SitDown(user *User, pos int) {
	gdRoom.positionUserIDs[pos] = user.data.userData.UserID

	playerData := gdRoom.userIDPlayerDatas[user.data.userData.UserID]
	if playerData == nil {
		playerData = new(GDPlayerData)
		playerData.user = user
		playerData.position = pos
		playerData.owner = user.data.userData.UserID == gdRoom.ownerUserID
		playerData.analyzer = new(mahjong.GDAnalyzer)
		playerData.roundResult = new(mahjong.GDPlayerRoundResult)
		playerData.totalResult = new(mahjong.GDPlayerTotalResult)

		gdRoom.userIDPlayerDatas[user.data.userData.UserID] = playerData
	}
	message := &msg.S2C_SitDown{
		Position:   pos,
		Owner:      playerData.owner,
		AccountID:  playerData.user.data.userData.AccountID,
		LoginIP:    playerData.user.data.userData.LoginIP,
		Nickname:   playerData.user.data.userData.Nickname,
		Headimgurl: playerData.user.data.userData.Headimgurl,
		Sex:        playerData.user.data.userData.Sex,
		Ready:      playerData.state == gdReady,
	}
	if gdRoom.rule.GPSAntiCheat {
		message.Location = playerData.user.location
	}
	broadcast(message, gdRoom.positionUserIDs, pos)
}

func (gdRoom *GDRoom) StandUp(user *User, pos int) {
	delete(gdRoom.positionUserIDs, pos)
	delete(gdRoom.userIDPlayerDatas, user.data.userData.UserID)
}

func (gdRoom *GDRoom) GetAllPlayers(user *User) {
	for pos := 0; pos < gdRoom.rule.MaxPlayers; pos++ {
		userID := gdRoom.positionUserIDs[pos]
		playerData := gdRoom.userIDPlayerDatas[userID]
		if playerData == nil {
			user.WriteMsg(&msg.S2C_StandUp{
				Position: pos,
			})
		} else {
			if playerData.user.isRobot() {
				skeleton.AfterFunc(time.Duration(pos+1)*time.Second, func() {
					user.WriteMsg(&msg.S2C_SitDown{
						Position:   playerData.position,
						Owner:      playerData.owner,
						AccountID:  playerData.user.data.userData.AccountID,
						LoginIP:    playerData.user.data.userData.LoginIP,
						Nickname:   playerData.user.data.userData.Nickname,
						Headimgurl: playerData.user.data.userData.Headimgurl,
						Sex:        playerData.user.data.userData.Sex,
						Ready:      playerData.state == gdReady,
						Location:   playerData.user.location,
					})
				})
			} else {
				user.WriteMsg(&msg.S2C_SitDown{
					Position:   pos,
					Owner:      playerData.owner,
					AccountID:  playerData.user.data.userData.AccountID,
					LoginIP:    playerData.user.data.userData.LoginIP,
					Nickname:   playerData.user.data.userData.Nickname,
					Headimgurl: playerData.user.data.userData.Headimgurl,
					Sex:        playerData.user.data.userData.Sex,
					Ready:      playerData.state == gdReady,
					Location:   playerData.user.location,
				})
			}
		}
	}
}

func (gdRoom *GDRoom) Disband(disbander *User) {
	if gdRoom.state == roomIdle {
		log.Debug("userID: %v 解散房间", disbander.data.userData.UserID)
		broadcast(&msg.S2C_DisbandRoom{
			Error:         msg.S2C_DisbandRoom_OK,
			RoomNumber:    gdRoom.number,
			OwnerNickName: disbander.data.userData.Nickname,
		}, gdRoom.positionUserIDs, -1)
		// 清空玩家数据
		gdRoom.clean()
		if gdRoom.rule.RoomType == roomPrivate {
			delete(gdroomNumberRooms, gdRoom.number) // 解散房间
		}
		return
	}
	log.Debug("userID: %v 申请解散房间", disbander.data.userData.UserID)
	gdRoom.disbandApplicantUserID = disbander.data.userData.UserID
	applicantPlayerData := gdRoom.userIDPlayerDatas[gdRoom.disbandApplicantUserID]
	applicantPlayerData.disbandActionCode = actionAgreeDisband
	for i := 1; i < gdRoom.rule.MaxPlayers; i++ {
		otherUserID := gdRoom.positionUserIDs[(applicantPlayerData.position+i)%gdRoom.rule.MaxPlayers]
		otherPlayerData := gdRoom.userIDPlayerDatas[otherUserID]
		otherPlayerData.disbandActionCode = actionWaitingDisband
	}
	playerDisbandInfos := []mahjong.GDPlayerDisbandInfo{}
	for i := 0; i < gdRoom.rule.MaxPlayers; i++ {
		userID := gdRoom.positionUserIDs[i]
		playerData := gdRoom.userIDPlayerDatas[userID]
		playerDisbandInfos = append(playerDisbandInfos, mahjong.GDPlayerDisbandInfo{
			Nickname:   playerData.user.data.userData.Nickname,
			ActionCode: playerData.disbandActionCode,
		})
	}
	for _, userID := range gdRoom.positionUserIDs {
		playerData := gdRoom.userIDPlayerDatas[userID]
		if user, ok := userIDUsers[userID]; ok {
			user.WriteMsg(&msg.S2C_ActionDisbandRoom{
				ApplicantNickname:  disbander.data.userData.Nickname,
				PlayerDisbandInfos: playerDisbandInfos,
				Enable:             playerData.disbandActionCode == actionWaitingDisband,
				WaitingTime:        120,
			})
		}
	}
	if gdRoom.discardTimer != nil {
		gdRoom.discardTimer.Stop()
		gdRoom.discardTimer = nil
	}
	gdRoom.disbandTimer = skeleton.AfterFunc(122*time.Second, func() {
		for _, userID := range gdRoom.positionUserIDs {
			playerData := gdRoom.userIDPlayerDatas[userID]
			if playerData.disbandActionCode == actionWaitingDisband {
				log.Debug("userID: %v 自动同意", playerData.user.data.userData.UserID)
				gdRoom.agreeDisbandRoom(playerData.user.data.userData.UserID)
			}
		}
	})
}

func (gdRoom *GDRoom) StartGame() {
	gdRoom.state = roomGame
	gdRoom.prepare()

	broadcast(&msg.S2C_GameStart{},
		gdRoom.positionUserIDs, -1)

	broadcast(&msg.S2C_UpdateMahjongCurrentRound{
		CurrentRound: gdRoom.currentRound,
	}, gdRoom.positionUserIDs, -1)

	dealerPlayerData := gdRoom.userIDPlayerDatas[gdRoom.dealerUserID]
	broadcast(&msg.S2C_DecideDealer{
		Position: dealerPlayerData.position,
	}, gdRoom.positionUserIDs, -1)

	if gdRoom.rule.NeedJoker {
		if gdRoom.jokers == nil {
			log.Error("joker is nil: %v", gdRoom.jokers)
		}
		broadcast(&msg.S2C_DecideGDJoker{
			WildCard: gdRoom.wildcard,
			Jokers:   gdRoom.jokers,
		}, gdRoom.positionUserIDs, -1)
	}
	// 所有玩家发13张牌
	for _, userID := range gdRoom.positionUserIDs {
		playerData := gdRoom.userIDPlayerDatas[userID]
		playerData.state = gdWaiting
		// 手牌13张
		playerData.hands = append(playerData.hands, gdRoom.rests[:13]...)
		// 排序
		playerData.analyzer.Analyze(playerData.hands, gdRoom.jokers)
		playerData.hands = playerData.analyzer.Sort()
		log.Debug("userID %v 手牌: %v", userID, mahjong.ToTileString(playerData.hands))
		// 获取可以胡的牌
		playerData.winTiles = playerData.analyzer.GetWinTiles(playerData.hands)
		if len(playerData.winTiles) > 0 {
			log.Debug("胡牌提示: %v", mahjong.ToTileString(playerData.winTiles))
		}
		gdRoom.rests = gdRoom.rests[13:]

		switch gdRoom.gameType {
		case mahjong.GD:
			// 所有玩家生成马牌
			if gdRoom.rule.BuyHorse == 1 {
				playerData.horseTile = append(playerData.horseTile, gdRoom.rests[:1]...)
				log.Debug("马牌: %v", mahjong.ToTileString(playerData.horseTile))
				gdRoom.rests = gdRoom.rests[1:]

				broadcast(&msg.S2C_GDBuyHorse{
					Position: playerData.position,
					Tiles:    []int{-1},
				}, gdRoom.positionUserIDs, -1)
			} else if gdRoom.rule.BuyHorse == 2 {
				playerData.horseTile = append(playerData.horseTile, gdRoom.rests[:2]...)
				log.Debug("马牌: %v", mahjong.ToTileString(playerData.horseTile))
				gdRoom.rests = gdRoom.rests[2:]

				broadcast(&msg.S2C_GDBuyHorse{
					Position: playerData.position,
					Tiles:    []int{-1, -1},
				}, gdRoom.positionUserIDs, -1)
			}

		case mahjong.YA:

		case mahjong.HNZZ:

		}

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
		}, gdRoom.positionUserIDs, playerData.position)
	}

	switch gdRoom.gameType {
	case mahjong.GD:
		if gdRoom.rule.NeedJoker {
			dealerPlayerData.discards = append(dealerPlayerData.discards, gdRoom.wildcard)
			broadcast(&msg.S2C_UpdateMahjongDiscads{
				Position: dealerPlayerData.position,
				Discards: dealerPlayerData.discards,
			}, gdRoom.positionUserIDs, -1)
			broadcast(&msg.S2C_MahjongDiscard{
				Position: dealerPlayerData.position,
				Tile:     gdRoom.wildcard,
			}, gdRoom.positionUserIDs, -1)
		}
	case mahjong.YA:

	case mahjong.HNZZ:

	}

	// 庄家摸牌、出牌
	gdRoom.drawAndDiscard(gdRoom.dealerUserID)

}

func (gdRoom *GDRoom) EndGame() {
	log.Debug("游戏结束")
	gdRoom.endTimestamp = time.Now().Unix()

	for _, userID := range gdRoom.positionUserIDs {
		playerData := gdRoom.userIDPlayerDatas[userID]
		playerData.winTiles = []int{}
		if user, ok := userIDUsers[userID]; ok {
			user.WriteMsg(&msg.S2C_UpdateWinTiles{
				Tiles: playerData.winTiles,
			})
		}
	}
	if gdRoom.currentRound == 1 {
		switch gdRoom.rule.RoomType {
		case roomPrivate, roomRoomCardMatch, roomRedPacketMatching, roomRedPacketPrivate:
			gdRoom.deductRoomCard()
		}
	}
	totalResults, roundResults := []PlayerResultData{}, []PlayerResultData{}
	for pos := 0; pos < gdRoom.rule.MaxPlayers; pos++ {
		userID := gdRoom.positionUserIDs[pos]
		playerData := gdRoom.userIDPlayerDatas[userID]
		// 计算总分
		roundResult := playerData.roundResult
		roundResult.TotalScore = roundResult.WinScore + roundResult.CatchHorseScore + roundResult.ExposedKongScore +
			roundResult.PongKongScore + roundResult.HiddenKongScore
		if len(gdRoom.winnerUserIDs) == 0 {
			roundResult.LastTile = -1
			roundResult.RoomCards = 0
		} else {
			if common.InArray(gdRoom.winnerUserIDs, userID) {
				roundResult.LastTile = playerData.claim
				if gdRoom.rule.RoomType == roomRoomCardMatch {
					roundResult.RoomCards = gdRoom.rule.RoomCards * (gdRoom.rule.MaxPlayers - 1)
				}
			} else {
				roundResult.LastTile = -1
				if gdRoom.rule.RoomType == roomRoomCardMatch {
					roundResult.RoomCards -= gdRoom.rule.RoomCards
				}
			}
		}
		totalResult := playerData.totalResult
		totalResult.Scores = append(totalResult.Scores, roundResult.TotalScore)
		totalResult.TotalScore += roundResult.TotalScore

		broadcast(&msg.S2C_UpdateGDTotalScore{
			Position:   pos,
			TotalScore: totalResult.TotalScore,
		}, gdRoom.positionUserIDs, -1)

		switch gdRoom.rule.RoomType {
		case roomPrivate, roomRoomCardMatch:
			totalRoomCards := playerData.user.data.userData.RoomCards + roundResult.RoomCards
			totalResults = append(totalResults, PlayerResultData{
				UserID:         playerData.user.data.userData.UserID,
				Nickname:       playerData.user.data.userData.Nickname,
				Score:          totalResult.TotalScore,
				RoomCards:      roundResult.RoomCards,
				RedPacketType:  gdRoom.rule.RedPacketType,
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
	switch gdRoom.rule.RoomType {
	case roomRedPacketPrivate, roomRedPacketMatching:
		for _, userID := range gdRoom.positionUserIDs {
			playerData := gdRoom.userIDPlayerDatas[userID]
			saveRedPacketMatchResultData(&RedPacketMatchResultData{
				UserID:        userID,
				RedPacketType: gdRoom.rule.RedPacketType,
				RedPacket:     playerData.roundResult.RedPacket,
				Taken:         false,
				CreatedAt:     time.Now().Unix(),
			})
		}
	case roomRoomCardMatch, roomPrivate:
		gdRoom.saveUserTotalResultData(totalResults)
		gdRoom.saveUserRoundResultData(gdRoom.currentRound, roundResults)
	}
	var roundResultsRecords []mahjong.GDPlayerRoundResult
	settleMap := make(map[int]int)
	for _, userID := range gdRoom.positionUserIDs {
		playerData := gdRoom.userIDPlayerDatas[userID]
		if gdRoom.gameType==3{
			roundResultsRecords = append(roundResultsRecords, mahjong.GDPlayerRoundResult{
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
			for i := 1; i < gdRoom.rule.MaxPlayers; i++ {
				otherUserID := gdRoom.positionUserIDs[(playerData.position+i)%gdRoom.rule.MaxPlayers]
				otherPlayerData := gdRoom.userIDPlayerDatas[otherUserID]
				roundResultsRecords = append(roundResultsRecords, mahjong.GDPlayerRoundResult{
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
		}else{
			if gdRoom.gameType== 1 {
				roundResultsRecords = append(roundResultsRecords, mahjong.GDPlayerRoundResult{
					Nickname:         playerData.user.data.userData.Nickname,
					Headimgurl:       playerData.user.data.userData.Headimgurl,
					Dealer:           playerData.dealer,
					Hands:            playerData.hands,
					Claims:           playerData.claims,
					LastTile:         playerData.roundResult.LastTile,
					WinType:          playerData.roundResult.WinType,
					WinScore:         playerData.roundResult.WinScore,
					CatchHorseScore:  playerData.roundResult.CatchHorseScore,
					ExposedKongScore: playerData.roundResult.ExposedKongScore,
					PongKongScore:    playerData.roundResult.PongKongScore,
					HiddenKongScore:  playerData.roundResult.HiddenKongScore,
					TotalScore:       playerData.roundResult.TotalScore,
					RoomCards:        playerData.roundResult.RoomCards,
					RedPacket:        playerData.roundResult.RedPacket,
				})
			}
			if gdRoom.gameType == 2 {
				roundResultsRecords = append(roundResultsRecords, mahjong.GDPlayerRoundResult{
					Nickname:   playerData.user.data.userData.Nickname,
					Headimgurl: playerData.user.data.userData.Headimgurl,
					Dealer:     playerData.dealer,
					Hands:      playerData.hands,
					Claims:     playerData.claims,
					LastTile:   playerData.roundResult.LastTile,
					WinType:    playerData.roundResult.WinType,
					//Gun:               playerData.gun,
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
			}
			settleMap[userID] = playerData.roundResult.TotalScore
		}
	}
	/*
		for _, userID := range gdRoom.positionUserIDs {
			var roundResults []mahjong.GDPlayerRoundResult

			playerData := gdRoom.userIDPlayerDatas[userID]
			roundResults = append(roundResults, mahjong.GDPlayerRoundResult{
				Nickname:         playerData.user.data.userData.Nickname,
				Headimgurl:       playerData.user.data.userData.Headimgurl,
				Dealer:           playerData.dealer,
				Hands:            playerData.hands,
				Claims:           playerData.claims,
				LastTile:         playerData.roundResult.LastTile,
				WinType:          playerData.roundResult.WinType,
				WinScore:         playerData.roundResult.WinScore,
				CatchHorseScore:  playerData.roundResult.CatchHorseScore,
				ExposedKongScore: playerData.roundResult.ExposedKongScore,
				PongKongScore:    playerData.roundResult.PongKongScore,
				HiddenKongScore:  playerData.roundResult.HiddenKongScore,
				TotalScore:       playerData.roundResult.TotalScore,
				RoomCards:        playerData.roundResult.RoomCards,
				RedPacket:        playerData.roundResult.RedPacket,
			})
			for i := 1; i < gdRoom.rule.MaxPlayers; i++ {
				otherUserID := gdRoom.positionUserIDs[(playerData.position+i)%gdRoom.rule.MaxPlayers]
				otherPlayerData := gdRoom.userIDPlayerDatas[otherUserID]
				roundResults = append(roundResults, mahjong.GDPlayerRoundResult{
					Nickname:         otherPlayerData.user.data.userData.Nickname,
					Headimgurl:       otherPlayerData.user.data.userData.Headimgurl,
					Dealer:           otherPlayerData.dealer,
					Hands:            otherPlayerData.hands,
					Claims:           otherPlayerData.claims,
					LastTile:         otherPlayerData.roundResult.LastTile,
					WinType:          otherPlayerData.roundResult.WinType,
					WinScore:         otherPlayerData.roundResult.WinScore,
					CatchHorseScore:  otherPlayerData.roundResult.CatchHorseScore,
					ExposedKongScore: otherPlayerData.roundResult.ExposedKongScore,
					PongKongScore:    otherPlayerData.roundResult.PongKongScore,
					HiddenKongScore:  otherPlayerData.roundResult.HiddenKongScore,
					TotalScore:       otherPlayerData.roundResult.TotalScore,
					RoomCards:        otherPlayerData.roundResult.RoomCards,
					RedPacket:        otherPlayerData.roundResult.RedPacket,
				})
			}
	*/
	for _, userID := range gdRoom.positionUserIDs {
		if user, ok := userIDUsers[userID]; ok {
			result := mahjong.ResultLose
			if len(gdRoom.winnerUserIDs) == 0 { // 无人胡牌
				result = mahjong.ResultDraw
			} else {
				if common.InArray(gdRoom.winnerUserIDs, userID) {
					result = mahjong.ResultWin
					// 计算用户赢得总局数
					//user.data.userData.WinRounds += 1
				}
			}
			continueGame := true
			switch gdRoom.rule.RoomType {
			case roomRedPacketPrivate, roomRedPacketMatching:
				continueGame = false
			case roomPrivate:
				continueGame = !(gdRoom.currentRound == gdRoom.rule.MaxRounds)
			}
			user.WriteMsg(&msg.S2C_GDRoundResult{
				Result:       result,
				RoomDesc:     gdRoom.desc,
				Jokers:       gdRoom.jokers,
				RoundResults: roundResultsRecords,
				ContinueGame: continueGame,
			})
		}
	}
	switch gdRoom.gameType {
	case mahjong.GD:
		if len(gdRoom.winnerUserIDs) > 0 {
			winnerPlayerData := gdRoom.userIDPlayerDatas[gdRoom.winnerUserIDs[0]]
			winnerPlayerData.user.data.userData.WinRounds += 1
			updateUserData(gdRoom.winnerUserIDs[0], bson.M{"$set": bson.M{"winrounds": winnerPlayerData.user.data.userData.WinRounds}})
		}
	case mahjong.YA:

	case mahjong.HNZZ:
	}

	if gdRoom.currentRound < gdRoom.rule.MaxRounds {
		gdRoom.currentRound++
		gdRoom.state = roomGameEnd
		return
	}
	switch gdRoom.rule.RoomType {
	case roomRoomCardMatch:
		gdRoom.calculateRoomCard()
	case roomPrivate:
		for _, userID := range gdRoom.positionUserIDs {
			var playerTotalResults []mahjong.GDPlayerTotalResult

			playerData := gdRoom.userIDPlayerDatas[userID]
			playerTotalResults = append(playerTotalResults, mahjong.GDPlayerTotalResult{
				Nickname:   playerData.user.data.userData.Nickname,
				Headimgurl: playerData.user.data.userData.Headimgurl,
				Owner:      playerData.owner,
				AccountID:  playerData.user.data.userData.AccountID,
				Scores:     playerData.totalResult.Scores,
				TotalScore: playerData.totalResult.TotalScore,
			})
			for i := 1; i < gdRoom.rule.MaxPlayers; i++ {
				otherUserID := gdRoom.positionUserIDs[(playerData.position+i)%gdRoom.rule.MaxPlayers]
				otherPlayerData := gdRoom.userIDPlayerDatas[otherUserID]
				playerTotalResults = append(playerTotalResults, mahjong.GDPlayerTotalResult{
					Nickname:   otherPlayerData.user.data.userData.Nickname,
					Headimgurl: otherPlayerData.user.data.userData.Headimgurl,
					Owner:      otherPlayerData.owner,
					AccountID:  otherPlayerData.user.data.userData.AccountID,
					Scores:     otherPlayerData.totalResult.Scores,
					TotalScore: otherPlayerData.totalResult.TotalScore,
				})
			}
			if user, ok := userIDUsers[userID]; ok {
				user.WriteMsg(&msg.S2C_GDTotalResult{
					TotalResults: playerTotalResults,
				})
				// 保存游戏积分
				user.data.userData.GameScore += playerData.totalResult.TotalScore
				// 用户玩的总局数
				user.data.userData.TotalRounds += gdRoom.currentRound
			} else {
				// 保存游戏积分
				playerData.user.data.userData.GameScore += playerData.totalResult.TotalScore
				// 用户玩的总局数
				playerData.user.data.userData.TotalRounds += gdRoom.currentRound
				updateUserData(userID, bson.M{"$set": bson.M{"gamescore": playerData.user.data.userData.GameScore}})
			}
		}
	}
	gdRoom.clean()
	switch gdRoom.rule.RoomType {
	case roomPrivate, roomRedPacketPrivate:
		delete(gdroomNumberRooms, gdRoom.number) // 解散房间
	}
	gdRoom.state = roomGameEnd
}

func (gdRoom *GDRoom) agreeDisbandRoom(userID int) {
	if gdRoom.state == roomIdle {
		return
	}
	playerData := gdRoom.userIDPlayerDatas[userID]
	if playerData.disbandActionCode != actionWaitingDisband {
		return
	}
	playerData.disbandActionCode = actionAgreeDisband
	if gdRoom.allAgree() {
		if gdRoom.rule.RoomType == roomPrivate && gdRoom.currentRound == 1 {
			gdRoom.deductRoomCard()
		}
		broadcast(&msg.S2C_DisbandRoom{
			Error:      msg.S2C_DisbandRoom_OK,
			RoomNumber: gdRoom.number,
		}, gdRoom.positionUserIDs, -1)
		gdRoom.clean()
		if gdRoom.rule.RoomType == roomPrivate {
			delete(gdroomNumberRooms, gdRoom.number)
		}
	} else {
		broadcast(&msg.S2C_AgreeDisbandRoom{
			Position: playerData.position,
			Nickname: playerData.user.data.userData.Nickname,
		}, gdRoom.positionUserIDs, -1)
	}
}

func (gdRoom *GDRoom) refuseDisbandRoom(userID int) {
	playerData := gdRoom.userIDPlayerDatas[userID]
	if gdRoom.state == roomIdle || gdRoom.disbandTimer == nil || playerData.disbandActionCode != actionWaitingDisband {
		return
	}
	log.Debug("userID: %v 拒绝解散房间", userID)
	if gdRoom.disbandTimer != nil {
		gdRoom.disbandTimer.Stop()
		gdRoom.disbandTimer = nil
	}
	gdRoom.disbandApplicantUserID = -1
	broadcast(&msg.S2C_DisbandRoom{
		Error:            msg.S2C_DisbandRoom_PlayerRefuse,
		RoomNumber:       gdRoom.number,
		RejecterNickName: playerData.user.data.userData.Nickname,
	}, gdRoom.positionUserIDs, -1)
	//log.Debug("%v,%v", gdRoom.drawerUserID, gdRoom.discarderUserID)
	/*
		if gdRoom.drawerUserID > 0 {
			playerData := gdRoom.userIDPlayerDatas[gdRoom.drawerUserID]
			if playerData.claimActionCode < 1 {
				gdRoom.discard(gdRoom.drawerUserID)
				return
			}
			playerData.state = gdActionClaim
			if user, ok := userIDUsers[gdRoom.drawerUserID]; ok {
				user.WriteMsg(&msg.S2C_ActionMahjongClaim{
					Position:    playerData.position,
					ActionCode:  playerData.claimActionCode,
					Countdown:   cd_gdClaim,
					Sequences:   playerData.sequences,
					Quadruplets: playerData.quadruplets,
				})
			}
			playerData.actionTimestamp = time.Now().Unix()
			log.Debug("等待 userID %v 要牌", gdRoom.drawerUserID)
			gdRoom.claimTimer = skeleton.AfterFunc((cd_gdClaim+2)*time.Second, func() {
				gdRoom.discard(gdRoom.drawerUserID)
			})
		}
	*/
}

// 断线重连
func (gdRoom *GDRoom) reconnect(user *User) {
	log.Debug("userID: %v 断线重连", user.data.userData.UserID)
	thePlayerData := gdRoom.userIDPlayerDatas[user.data.userData.UserID]
	if thePlayerData == nil {
		return
	}
	user.WriteMsg(&msg.S2C_GameStart{})

	dealerPlayerData := gdRoom.userIDPlayerDatas[gdRoom.dealerUserID]
	if dealerPlayerData != nil {
		user.WriteMsg(&msg.S2C_DecideDealer{
			Position: dealerPlayerData.position,
		})
	}
	switch gdRoom.gameType {
	case mahjong.GD:
		if gdRoom.rule.NeedJoker {
			if gdRoom.jokers == nil {
				log.Error("joker is nil: %v", gdRoom.jokers)
			}
			user.WriteMsg(&msg.S2C_DecideGDJoker{
				WildCard: gdRoom.wildcard,
				Jokers:   gdRoom.jokers,
			})
		}
		if gdRoom.rule.BuyHorse > 0 {
			horseTile := make([]int, 0)
			for i := 0; i < gdRoom.rule.BuyHorse; i++ {
				horseTile = append(horseTile, -1)
			}
			user.WriteMsg(&msg.S2C_GDBuyHorse{
				Position: thePlayerData.position,
				Tiles:    horseTile,
			})
			for i := 1; i < gdRoom.rule.MaxPlayers; i++ {
				otherUserID := gdRoom.positionUserIDs[(thePlayerData.position+i)%gdRoom.rule.MaxPlayers]
				otherPlayerData := gdRoom.userIDPlayerDatas[otherUserID]
				user.WriteMsg(&msg.S2C_GDBuyHorse{
					Position: otherPlayerData.position,
					Tiles:    horseTile,
				})
			}
		}
	case mahjong.YA:

	case mahjong.HNZZ:
		user.WriteMsg(&msg.S2C_DecideGDJoker{
			Jokers: gdRoom.jokers,
		})
	}
	user.WriteMsg(&msg.S2C_UpdateMahjongRestsNumber{
		NumberOfRests: len(gdRoom.rests),
	})
	user.WriteMsg(&msg.S2C_UpdateMahjongCurrentRound{
		CurrentRound: gdRoom.currentRound,
	})
	if len(thePlayerData.winTiles) > 0 {
		log.Debug("胡牌提示: %v", mahjong.ToTileString(thePlayerData.winTiles))
	}
	user.WriteMsg(&msg.S2C_UpdateWinTiles{
		Tiles: thePlayerData.winTiles,
	})
	if gdRoom.claimUserID < 1 && gdRoom.discarderUserID > 0 {
		discarderPlayerData := gdRoom.userIDPlayerDatas[gdRoom.discarderUserID]
		user.WriteMsg(&msg.S2C_UpdateMahjongDiscardCusor{
			Position: discarderPlayerData.position,
			Index:    len(discarderPlayerData.discards) - 1,
		})
	}
	if gdRoom.disbandApplicantUserID > 0 {
		applicantPlayerData := gdRoom.userIDPlayerDatas[gdRoom.disbandApplicantUserID]
		playerDisbandInfos := []mahjong.GDPlayerDisbandInfo{}
		for i := 0; i < gdRoom.rule.MaxPlayers; i++ {
			userID := gdRoom.positionUserIDs[i]
			playerData := gdRoom.userIDPlayerDatas[userID]
			playerDisbandInfos = append(playerDisbandInfos, mahjong.GDPlayerDisbandInfo{
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
	gdRoom.getPlayerData(user, thePlayerData, false)

	for i := 1; i < gdRoom.rule.MaxPlayers; i++ {
		otherUserID := gdRoom.positionUserIDs[(thePlayerData.position+i)%gdRoom.rule.MaxPlayers]
		otherPlayerData := gdRoom.userIDPlayerDatas[otherUserID]

		gdRoom.getPlayerData(user, otherPlayerData, true)
	}
}

func (gdRoom *GDRoom) getPlayerData(user *User, playerData *GDPlayerData, other bool) {
	user.WriteMsg(&msg.S2C_UpdateMahjongDiscads{
		Position: playerData.position,
		Discards: playerData.discards,
	})
	user.WriteMsg(&msg.S2C_UpdateMahjongClaims{
		Position: playerData.position,
		Claims:   playerData.claims,
	})
	if playerData.claims == nil {
		log.Error("claims is nil: %v", playerData.claims)
	}
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
	user.WriteMsg(&msg.S2C_UpdateGDTotalScore{
		Position:   playerData.position,
		TotalScore: playerData.totalResult.TotalScore,
	})
	switch playerData.state {
	case gdActionDiscard:
		after := int(time.Now().Unix() - playerData.actionTimestamp)
		countdown := cd_gdDiscard - after
		if countdown > 1 {
			user.WriteMsg(&msg.S2C_ActionMahjongDiscard{
				Position:  playerData.position,
				Countdown: countdown - 1,
			})
		}
	case gdActionClaim:
		after := int(time.Now().Unix() - playerData.actionTimestamp)
		countdown := cd_gdClaim - after
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

func (gdRoom *GDRoom) resetActionClaimUsers() {
	if gdRoom.claimTimer != nil {
		gdRoom.claimTimer.Stop()
		gdRoom.claimTimer = nil
	}
	for _, userID := range gdRoom.positionUserIDs {
		gdRoom.deleteActionClaimUsers(userID)
	}
}

func (gdRoom *GDRoom) deleteActionClaimUsers(userID int) {
	playerData := gdRoom.userIDPlayerDatas[userID]
	if playerData.state == gdActionClaim {
		playerData.state = gdWaiting
		playerData.claimActionCode = 0
	}
	delete(gdRoom.actionWinUsers, userID)
	delete(gdRoom.actionKongUsers, userID)
	delete(gdRoom.actionPongUsers, userID)
	delete(gdRoom.actionChowUsers, userID)
}

func (gdRoom *GDRoom) actionClaimUsersEmpty() bool {
	if len(gdRoom.actionWinUsers) == 0 && len(gdRoom.actionKongUsers) == 0 &&
		len(gdRoom.actionPongUsers) == 0 && len(gdRoom.actionChowUsers) == 0 {
		return true
	}
	return false
}

func (gdRoom *GDRoom) existClaimUsers(userId int) bool {
	_, chowState := gdRoom.actionChowUsers[userId]
	_, pengState := gdRoom.actionPongUsers[userId]
	_, kongState := gdRoom.actionKongUsers[userId]
	_, huState := gdRoom.actionWinUsers[userId]
	return chowState || pengState || kongState || huState
}
func (gdRoom *GDRoom) prepare() {
	// 洗牌
	switch gdRoom.gameType{
	case mahjong.GD:
		if gdRoom.rule.WithHonors {
			gdRoom.tiles = common.Shuffle(mahjong.GDAllTiles)
		} else {
			gdRoom.tiles = common.Shuffle(mahjong.GDAllTilesWithoutHonors)
		}
	case mahjong.YA:

	case mahjong.HNZZ:
		gdRoom.tiles = common.Shuffle(mahjong.GDAllTiles)
	}

	// 确定庄家
	if gdRoom.currentRound == 1 {
		gdRoom.dealerUserID = gdRoom.positionUserIDs[0]
		switch gdRoom.rule.RoomType {
		case roomPrivate, roomRoomCardMatch, roomRedPacketMatching, roomRedPacketPrivate:
			gdRoom.startTimestamp = time.Now().Unix()
			gdRoom.eachRoundStartTimestamp = gdRoom.startTimestamp
			gdRoom.initTotalResultData()
		}
	} else {
		switch gdRoom.rule.RoomType {
		case roomPrivate, roomRoomCardMatch, roomRedPacketMatching, roomRedPacketPrivate:
			gdRoom.eachRoundStartTimestamp = time.Now().Unix()
		}
		switch gdRoom.gameType {
		case mahjong.GD:
			if len(gdRoom.winnerUserIDs) > 0 {
				gdRoom.dealerUserID = gdRoom.winnerUserIDs[0]
			}
		case mahjong.YA:

		case mahjong.HNZZ:
			if gdRoom.catchBirdUserID > 0 {
				gdRoom.dealerUserID = gdRoom.catchBirdUserID
			}
		}
	}
	// 庄家
	dealerPlayerData := gdRoom.userIDPlayerDatas[gdRoom.dealerUserID]
	dealerPlayerData.dealer = true
	// 闲家
	dealerPosition := dealerPlayerData.position
	for i := 1; i < gdRoom.rule.MaxPlayers; i++ {
		playerPos := (dealerPosition + i) % gdRoom.rule.MaxPlayers
		playerUserID := gdRoom.positionUserIDs[playerPos]
		playerData := gdRoom.userIDPlayerDatas[playerUserID]
		if playerData == nil {
			continue
		}
		playerData.dealer = false
	}
	switch gdRoom.gameType {
	case mahjong.GD:
		if gdRoom.rule.NeedJoker {
			gdRoom.wildcard = gdRoom.tiles[0]                    // 确定混儿
			gdRoom.jokers = mahjong.GetGDJokers(gdRoom.wildcard) // 确定宝牌
			log.Debug("混儿: %v, 宝牌: %v", mahjong.ToTileString([]int{gdRoom.wildcard}), mahjong.ToTileString(gdRoom.jokers))
			// 剩余的牌
			gdRoom.rests = gdRoom.tiles[1:]
			if gdRoom.jokers == nil {
				log.Error("joker is nil: %v", gdRoom.jokers)
			}
		} else {
			// 剩余的牌
			gdRoom.rests = append([]int{}, gdRoom.tiles...)
		}
	case mahjong.YA:

	case mahjong.HNZZ:
		log.Debug("癞子: %v", mahjong.ToTileString(gdRoom.jokers))
		// 剩余的牌
		gdRoom.rests = append([]int{}, gdRoom.tiles...)

		gdRoom.catchBirdUserID = -1
	}

	gdRoom.discards = []int{}
	gdRoom.discarderUserID = -1
	gdRoom.drawerUserID = -1
	gdRoom.resetActionClaimUsers()
	gdRoom.claimUserID = -1
	gdRoom.disbandApplicantUserID = -1
	gdRoom.winnerUserIDs = []int{}

	for _, userID := range gdRoom.positionUserIDs {
		playerData := gdRoom.userIDPlayerDatas[userID]
		roundResult := playerData.roundResult
		playerData.draw = -1
		playerData.hands = []int{}
		playerData.discards = []int{}
		playerData.claims = [][]int{}
		playerData.actionTimestamp = 0

		switch gdRoom.gameType {
		case mahjong.GD:
			playerData.discardsCount = 0
			playerData.managed = false
			playerData.horseTile = []int{}

			roundResult.CatchHorseScore = 0
		case mahjong.YA:

		case mahjong.HNZZ:
			roundResult.CatchBirdScore = 0
		}

		roundResult.WinType = 0
		roundResult.WinScore = 0
		roundResult.ExposedKongScore = 0
		roundResult.PongKongScore = 0
		roundResult.HiddenKongScore = 0
		roundResult.TotalScore = 0
	}
}

func (gdRoom *GDRoom) updateDiscards(userID int) {
	playerData := gdRoom.userIDPlayerDatas[userID]
	broadcast(&msg.S2C_UpdateMahjongDiscads{
		Position: playerData.position,
		Discards: playerData.discards,
	}, gdRoom.positionUserIDs, -1)
}

func (gdRoom *GDRoom) updateClaims(userID int) {
	playerData := gdRoom.userIDPlayerDatas[userID]
	broadcast(&msg.S2C_UpdateMahjongClaims{
		Position: playerData.position,
		Claims:   playerData.claims,
	}, gdRoom.positionUserIDs, -1)
	if playerData.claims == nil {
		log.Error("claims is nil: %v", playerData.claims)
	}
}

// 计算胡牌分
func (gdRoom *GDRoom) calculateWinScore(winnerUserID int) {
	numberOfWinner := len(gdRoom.winnerUserIDs)
	if numberOfWinner == 0 || common.Index(gdRoom.winnerUserIDs, winnerUserID) == -1 {
		return
	}

	winnerPlayerData := gdRoom.userIDPlayerDatas[winnerUserID]
	if numberOfWinner > 1 {
		discarderPlayerData := gdRoom.userIDPlayerDatas[gdRoom.discarderUserID]
		if discarderPlayerData.dealer {
			winnerPlayerData.roundResult.WinScore = 2 * gdRoom.rule.BaseScore
		} else {
			if winnerPlayerData.dealer {
				winnerPlayerData.roundResult.WinScore = 2 * gdRoom.rule.BaseScore
			} else {
				winnerPlayerData.roundResult.WinScore = 1 * gdRoom.rule.BaseScore
			}
		}
		return
	}

	winnerWinType := winnerPlayerData.roundResult.WinType
	// 计算输家胡牌分
	switch winnerWinType {
	case mahjong.GDWinByDiscard:
		discarderPlayerData := gdRoom.userIDPlayerDatas[gdRoom.discarderUserID]
		if discarderPlayerData.dealer {
			discarderPlayerData.roundResult.WinScore = -(gdRoom.rule.MaxPlayers) * gdRoom.rule.BaseScore
		} else {
			if winnerPlayerData.dealer {
				discarderPlayerData.roundResult.WinScore = -(gdRoom.rule.MaxPlayers - 1) * 2 * gdRoom.rule.BaseScore
			} else {
				discarderPlayerData.roundResult.WinScore = -(gdRoom.rule.MaxPlayers - 1) * gdRoom.rule.BaseScore
			}
		}
	case mahjong.GDWinBySelfDraw:
		for i := 1; i < gdRoom.rule.MaxPlayers; i++ {
			otherUserID := gdRoom.positionUserIDs[(winnerPlayerData.position+i)%gdRoom.rule.MaxPlayers]
			otherPlayerData := gdRoom.userIDPlayerDatas[otherUserID]
			if winnerPlayerData.dealer {
				otherPlayerData.roundResult.WinScore = -4 * gdRoom.rule.BaseScore
			} else {
				if otherPlayerData.dealer {
					otherPlayerData.roundResult.WinScore = -4 * gdRoom.rule.BaseScore
				} else {
					otherPlayerData.roundResult.WinScore = -2 * gdRoom.rule.BaseScore
				}
			}
		}
	}

	// 计算赢家得分
	loserWinScore := 0
	for i := 1; i < gdRoom.rule.MaxPlayers; i++ {
		otherUserID := gdRoom.positionUserIDs[(winnerPlayerData.position+i)%gdRoom.rule.MaxPlayers]
		otherPlayerData := gdRoom.userIDPlayerDatas[otherUserID]
		loserWinScore += otherPlayerData.roundResult.WinScore
	}
	winnerPlayerData.roundResult.WinScore = -1 * loserWinScore
}

func (gdRoom *GDRoom) calculateHorseScore(userID int, horseCount int) {
	if horseCount < 1 || horseCount > 2 {
		return
	}
	playerData := gdRoom.userIDPlayerDatas[userID]
	if len(gdRoom.winnerUserIDs) > 0 && common.Index(gdRoom.winnerUserIDs, userID) != -1 {
		playerData.roundResult.CatchHorseScore += (gdRoom.rule.MaxPlayers - 1) * horseCount * 2 * gdRoom.rule.BaseScore
		for i := 1; i < gdRoom.rule.MaxPlayers; i++ {
			otherUserID := gdRoom.positionUserIDs[(playerData.position+i)%gdRoom.rule.MaxPlayers]
			otherPlayerData := gdRoom.userIDPlayerDatas[otherUserID]
			otherPlayerData.roundResult.CatchHorseScore += -1 * horseCount * 2 * gdRoom.rule.BaseScore
		}
	} else {
		playerData.roundResult.CatchHorseScore += (gdRoom.rule.MaxPlayers - 1) * horseCount * gdRoom.rule.BaseScore
		for i := 1; i < gdRoom.rule.MaxPlayers; i++ {
			otherUserID := gdRoom.positionUserIDs[(playerData.position+i)%gdRoom.rule.MaxPlayers]
			otherPlayerData := gdRoom.userIDPlayerDatas[otherUserID]
			otherPlayerData.roundResult.CatchHorseScore += -1 * horseCount * gdRoom.rule.BaseScore
		}
	}
}

func (gdRoom *GDRoom) calculateRedPacket(userID int, redPacketType int) {
	if common.Index([]int{1, 10, 100, 999}, redPacketType) == -1 {
		return
	}
	playerData := gdRoom.userIDPlayerDatas[userID]
	roundResult := playerData.roundResult
	roundResult.RedPacket = float64(gdRoom.rule.RedPacketType)
}

func (gdRoom *GDRoom) deductRoomCard() {
	switch gdRoom.rule.RoomType {
	case roomRoomCardMatch, roomRedPacketMatching, roomRedPacketPrivate:
		for _, userID := range gdRoom.positionUserIDs {
			playerData := gdRoom.userIDPlayerDatas[userID]
			if user, ok := userIDUsers[userID]; ok {
				user.data.userData.RoomCards -= gdRoom.rule.RoomCards
				user.data.userData.ConsumedRoomCards += gdRoom.rule.RoomCards
			} else {
				playerData.user.data.userData.RoomCards -= gdRoom.rule.RoomCards
				playerData.user.data.userData.ConsumedRoomCards += gdRoom.rule.RoomCards
				updateUserData(userID, bson.M{"$set": bson.M{"roomcards": playerData.user.data.userData.RoomCards, "consumedroomcards": playerData.user.data.userData.ConsumedRoomCards}})
			}
			if playerData.user.isRobot() {
				cards := -gdRoom.rule.RoomCards
				switch gdRoom.rule.RoomType {
				case roomRoomCardMatch:
					upsertRobotData(time.Now().Format("20060102"), bson.M{"$inc": bson.M{"roomcardmatchbalance": cards}})
				case roomRedPacketMatching:
					upsertRobotData(time.Now().Format("20060102"), bson.M{"$inc": bson.M{"redpacketmatchbalance": cards}})
				}
			}
		}

	case roomPrivate:
		if owner, ok := userIDUsers[gdRoom.ownerUserID]; ok {
			owner.data.userData.RoomCards -= gdRoom.rule.RoomCards
			owner.data.userData.ConsumedRoomCards += gdRoom.rule.RoomCards
		} else {
			playerData := gdRoom.userIDPlayerDatas[gdRoom.ownerUserID]
			playerData.user.data.userData.RoomCards -= gdRoom.rule.RoomCards
			playerData.user.data.userData.ConsumedRoomCards += gdRoom.rule.RoomCards
			updateUserData(gdRoom.ownerUserID, bson.M{"$set": bson.M{"roomcards": playerData.user.data.userData.RoomCards, "consumedroomcards": playerData.user.data.userData.ConsumedRoomCards}})
		}
	}
}

func (gdRoom *GDRoom) calculateRoomCard() {
	if len(gdRoom.winnerUserIDs) == 0 { // 流局
		for _, userID := range gdRoom.positionUserIDs {
			playerData := gdRoom.userIDPlayerDatas[userID]
			if user, ok := userIDUsers[userID]; ok {
				user.data.userData.RoomCards += gdRoom.rule.RoomCards
			} else {
				playerData.user.data.userData.RoomCards += gdRoom.rule.RoomCards
				updateUserData(userID, bson.M{"$set": bson.M{"roomcards": playerData.user.data.userData.RoomCards}})
			}
			if playerData.user.isRobot() {
				cards := gdRoom.rule.RoomCards
				upsertRobotData(time.Now().Format("20060102"), bson.M{"$inc": bson.M{"roomcardmatchbalance": cards}})
			}
		}
	} else {
		winnerUserID := gdRoom.winnerUserIDs[0]
		playerData := gdRoom.userIDPlayerDatas[winnerUserID]
		if user, ok := userIDUsers[winnerUserID]; ok {
			user.data.userData.RoomCards += gdRoom.rule.RoomCards * gdRoom.rule.MaxPlayers
		} else {
			playerData.user.data.userData.RoomCards += gdRoom.rule.RoomCards * gdRoom.rule.MaxPlayers
			updateUserData(winnerUserID, bson.M{"$set": bson.M{"roomcards": playerData.user.data.userData.RoomCards}})

		}
		if playerData.user.isRobot() {
			cards := gdRoom.rule.RoomCards * gdRoom.rule.MaxPlayers
			upsertRobotData(time.Now().Format("20060102"), bson.M{"$inc": bson.M{"roomcardmatchbalance": cards}})
		}
	}
}

func (gdRoom *GDRoom) initTotalResultData() {
	for _, userID := range gdRoom.positionUserIDs {
		playerData := gdRoom.userIDPlayerDatas[userID]
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
				playerData.totalResultData.RoomNumber = gdRoom.number
				playerData.totalResultData.RoomDesc = gdRoom.desc
				playerData.totalResultData.StartTimestamp = gdRoom.startTimestamp
				playerData.totalResultData.Position = playerData.position
			}
		})
	}
}

func (gdRoom *GDRoom) saveUserTotalResultData(results []PlayerResultData) {
	for pos := 0; pos < gdRoom.rule.MaxPlayers; pos++ {
		userID := gdRoom.positionUserIDs[pos]
		playerData := gdRoom.userIDPlayerDatas[userID]
		if playerData.totalResultData != nil {
			playerData.totalResultData.RoomType = gdRoom.rule.RoomType
			playerData.totalResultData.EndTimestamp = gdRoom.endTimestamp
			playerData.totalResultData.Results = results
			playerData.totalResultData.UpdatedAt = time.Now().Unix()

			saveTotalResultData(playerData.totalResultData)
		}
	}
}

func (gdRoom *GDRoom) saveUserRoundResultData(round int, results []PlayerResultData) {
	for pos := 0; pos < gdRoom.rule.MaxPlayers; pos++ {
		userID := gdRoom.positionUserIDs[pos]
		playerData := gdRoom.userIDPlayerDatas[userID]
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
					playerData.roundResultData.StartTimestamp = gdRoom.eachRoundStartTimestamp
					playerData.roundResultData.EndTimestamp = gdRoom.endTimestamp
					playerData.roundResultData.Position = playerData.position
					playerData.roundResultData.Results = results
					playerData.roundResultData.UpdatedAt = time.Now().Unix()

					saveRoundResultData(playerData.roundResultData)
				}
			})
		}
	}
}

func (gdRoom *GDRoom) catchHorse() {
	if len(gdRoom.winnerUserIDs) > 0 {
		for _, userID := range gdRoom.positionUserIDs {
			playerData := gdRoom.userIDPlayerDatas[userID]
			// 计算玩家买马分数
			//horseCount, ok := mahjong.CatchHorse(mahjong.GDHorseTile, playerData.horseTile)
			horseTile := mahjong.Position(gdRoom.dealerUserID, gdRoom.winnerUserIDs[0], gdRoom.rule.MaxPlayers)
			horseCount, ok := mahjong.CatchHorse(horseTile, playerData.horseTile)
			if ok {
				gdRoom.calculateHorseScore(userID, horseCount)
			}
			broadcast(&msg.S2C_GDBuyHorse{
				Position: playerData.position,
				Tiles:    playerData.horseTile,
			}, gdRoom.positionUserIDs, -1)
		}
	}
}

func (gdRoom *GDRoom) allSetGun() bool {
	for _, userID := range gdRoom.positionUserIDs {
		playerData := gdRoom.userIDPlayerDatas[userID]
		if playerData.state == ActionSetGun {
			return false
		}
	}
	return true
}

//计算下炮子分数
func (gdRoom *GDRoom) calculateGunScore(winnerUserID int) {
	winnerPlayerData := gdRoom.userIDPlayerDatas[winnerUserID]
	winnerWinType := winnerPlayerData.roundResult.WinType

	switch winnerWinType {
	case mahjong.GDWinByDiscard, mahjong.GDWinBySelfDraw:
		for i := 1; i < gdRoom.rule.MaxPlayers; i++ {
			otherUserID := gdRoom.positionUserIDs[(winnerPlayerData.position+i)%gdRoom.rule.MaxPlayers]
			otherPlayerData := gdRoom.userIDPlayerDatas[otherUserID]
			otherPlayerData.roundResult.GunScore = -otherPlayerData.gun - winnerPlayerData.gun
		}
	}
	//计算赢家炮子分
	loserGunScore := 0
	for i := 1; i < gdRoom.rule.MaxPlayers; i++ {
		otherUserID := gdRoom.positionUserIDs[(winnerPlayerData.position+i)%gdRoom.rule.MaxPlayers]
		otherPlayerData := gdRoom.userIDPlayerDatas[otherUserID]
		loserGunScore += otherPlayerData.roundResult.GunScore
	}
	winnerPlayerData.roundResult.GunScore = -1 * loserGunScore
}

//计算胡牌得分
func (gdRoom *GDRoom) calculateYanAnWinScore(winnerUserID int) {
	if common.Index(gdRoom.winnerUserIDs, winnerUserID) == -1 {
		return
	}
	winnerPlayerData := gdRoom.userIDPlayerDatas[winnerUserID]
	winnerWinType := winnerPlayerData.roundResult.WinType

	switch winnerWinType {
	case mahjong.GDWinByDiscard:
		discarderPlayerData := gdRoom.userIDPlayerDatas[gdRoom.discarderUserID]
		if discarderPlayerData.dealer {
			discarderPlayerData.roundResult.WinScore = -4 * gdRoom.rule.BaseScore
		} else {
			if winnerPlayerData.dealer {
				discarderPlayerData.roundResult.WinScore = -6 * gdRoom.rule.BaseScore
			} else {
				discarderPlayerData.roundResult.WinScore = -3 * gdRoom.rule.BaseScore
			}
		}
	case mahjong.GDWinBySelfDraw:
		for i := 1; i < gdRoom.rule.MaxPlayers; i++ {
			otherUserID := gdRoom.positionUserIDs[(winnerPlayerData.position+i)%gdRoom.rule.MaxPlayers]
			otherPlayerData := gdRoom.userIDPlayerDatas[otherUserID]
			if winnerPlayerData.dealer {
				otherPlayerData.roundResult.WinScore = -4 * gdRoom.rule.BaseScore
			} else {
				if otherPlayerData.dealer {
					otherPlayerData.roundResult.WinScore = -4 * gdRoom.rule.BaseScore
				} else {
					otherPlayerData.roundResult.WinScore = -2 * gdRoom.rule.BaseScore
				}
			}
		}
	}

	// 计算赢家胡牌分
	loserWinScore := 0
	for i := 1; i < gdRoom.rule.MaxPlayers; i++ {
		otherUserID := gdRoom.positionUserIDs[(winnerPlayerData.position+i)%gdRoom.rule.MaxPlayers]
		otherPlayerData := gdRoom.userIDPlayerDatas[otherUserID]
		loserWinScore += otherPlayerData.roundResult.WinScore
	}
	winnerPlayerData.roundResult.WinScore = -1 * loserWinScore

}

// 计算抓鸟分---湖南转转麻将
func (hnzzRoom *GDRoom) calculateCatchBirdScore(numberOfBird int) {
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