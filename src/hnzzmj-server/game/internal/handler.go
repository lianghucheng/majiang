package internal

import (
	"encoding/json"
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
	"hnzzmj-server/common"
	"hnzzmj-server/game/mahjong"
	"hnzzmj-server/msg"
	"reflect"
	"strings"
)

func init() {
	handler(&msg.C2S_SetSystemOn{}, handleSetSystemOn)
	handler(&msg.C2S_SetUsernamePassword{}, handleSetUsernamePassword)
	handler(&msg.C2S_SetHNZZConfig{}, handleSetHNZZConfig)
	handler(&msg.C2S_SetUserRole{}, handleSetUserRole)
	handler(&msg.C2S_GetUserInfo{}, handleGetUserInfo)
	handler(&msg.C2S_TransferRoomCard{}, handleTransferRoomCard)
	handler(&msg.C2S_GetTransferRoomCardRecord{}, handleGetTransferRoomCardRecord)
	handler(&msg.C2S_GetAllTransferRoomCardRecord{}, handleGetAllTransferRoomCardRecord)
	handler(&msg.C2S_GetAllAgentInfo{}, handleGetAllAgentInfo)
	handler(&msg.C2S_GetAllUserInfo{}, handleGetAllUserInfo)
	handler(&msg.C2S_GetBlackList{}, handleGetBlackList)

	handler(&msg.C2S_Heartbeat{}, handleHeartbeat)
	handler(&msg.C2S_CreateHNZZRoom{}, handleCreateHNZZRoom)
	handler(&msg.C2S_EnterRoom{}, handleEnterRoom)
	handler(&msg.C2S_GetAllPlayers{}, handleGetAllPlayers)
	handler(&msg.C2S_ExitOrDisbandRoom{}, handleExitOrDisbandRoom)
	handler(&msg.C2S_Prepare{}, handlePrepare)
	handler(&msg.C2S_MahjongDiscard{}, handleMahjongDiscard)
	handler(&msg.C2S_MahjongWin{}, handleMahjongWin)
	handler(&msg.C2S_MahjongKong{}, handleMahjongKong)
	handler(&msg.C2S_MahjongPong{}, handleMahjongPong)
	handler(&msg.C2S_MahjongChow{}, handleMahjongChow)
	handler(&msg.C2S_MahjongPass{}, handleMahjongPass)
	handler(&msg.C2S_AgreeDisbandRoom{}, handleAgreeDisbandRoom)
	handler(&msg.C2S_RefuseDisbandRoom{}, handleRefuseDisbandRoom)
	handler(&msg.C2S_GetRoomCards{}, handleGetRoomCards)
	handler(&msg.C2S_StartHNZZMatching{}, handleStartHNZZMatching)
	handler(&msg.C2S_GetTotalResults{}, handleGetTotalResults)
	handler(&msg.C2S_GetRoundResults{}, handleGetRoundResults)
	handler(&msg.C2S_CompleteDailyShare{}, handleCompleteDailyShare)
	handler(&msg.C2S_TextMessage{}, handleTextMessage)
	handler(&msg.C2S_ExpressionMessage{}, handleExpressionMessage)
	handler(&msg.C2S_GCloudVoiceMessage{}, handleGCloudVoiceMessage)
	handler(&msg.C2S_GetHNZZIOSProductList{}, handleGetHNZZIOSProductList)
	handler(&msg.C2S_GetHNZZAndroidProductList{}, handleGetHNZZAndroidProductList)
	handler(&msg.C2S_IAPReceiptData{}, handleIAPReceiptData)
	handler(&msg.C2S_GetRedPacketMatchRecord{}, handleGetRedPacketMatchRecord)
	handler(&msg.C2S_TakeRedPacketMatchPrize{}, handleTakeRedPacketMatchPrize)
	handler(&msg.C2S_FakeWXPay{}, handleFakeWXPay)
	handler(&msg.C2S_GetCircleLoginCode{}, handleGetCircleLoginCode)

	// 机器人
	handler(&msg.C2S_SetRobotData{}, handleSetRobotData)
}

func handler(m interface{}, h interface{}) {
	skeleton.RegisterChanRPC(reflect.TypeOf(m), h)
}

func handleSetSystemOn(args []interface{}) {
	m := args[0].(*msg.C2S_SetSystemOn)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if user.data.userData.Role < roleRoot {
		log.Debug("userID: %v 没有权限", user.data.userData.UserID)
		user.WriteMsg(&msg.S2C_SetSystemOn{
			Error: msg.S2C_SetSystemOn_PermissionDenied,
		})
		return
	}
	systemOn = m.On
	user.WriteMsg(&msg.S2C_SetSystemOn{
		Error: msg.S2C_SetSystemOn_OK,
		On:    systemOn,
	})
	if systemOn {
		log.Debug("userID: %v 设置系统开", user.data.userData.UserID)
	} else {
		log.Debug("userID: %v 设置系统关", user.data.userData.UserID)
	}
}

func handleSetUsernamePassword(args []interface{}) {
	m := args[0].(*msg.C2S_SetUsernamePassword)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if strings.TrimSpace(m.Username) == "" || strings.TrimSpace(m.Password) == "" {
		// 用户名或密码不能为空
		return
	}
	if user.data.userData.Username == "" || user.data.userData.Username == m.Username {
		user.setUsernamePassword(m.Username, m.Password)
	} else {
		log.Debug("userID: %v 用户名无需更改", user.data.userData.UserID)
	}
}

func handleSetHNZZConfig(args []interface{}) {
	m := args[0].(*msg.C2S_SetHNZZConfig)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if user.data.userData.Role < roleAdmin {
		log.Debug("userID: %v 没有权限", user.data.userData.UserID)
		user.WriteMsg(&msg.S2C_SetHNZZConfig{
			Error: msg.S2C_SetHNZZConfig_PermissionDenied,
		})
		return
	}
	if m.AndroidVersion > 0 {
		user.setHNZZAndroidVersion(m.AndroidVersion)
	}
	if len(m.AndroidDownloadUrl) > 0 {
		user.setHNZZAndroidDownloadUrl(m.AndroidDownloadUrl)
	}
	if m.IOSVersion > 0 {
		user.setHNZZIOSVersion(m.IOSVersion)
	}
	if len(m.IOSDownloadUrl) > 0 {
		user.setHNZZIOSDownloadUrl(m.IOSDownloadUrl)
	}
	if len(m.Notice) > 0 {
		user.setHNZZNotice(m.Notice)
	}
	if len(m.Radio) > 0 {
		user.setHNZZRadio(m.Radio)
	}
	if len(m.WeChatNumber) > 0 {
		user.setHNZZWeChatNumber(m.WeChatNumber)
	}
}

func handleSetUserRole(args []interface{}) {
	m := args[0].(*msg.C2S_SetUserRole)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if m.AccountID == 0 {
		log.Debug("账户ID为0")
		user.WriteMsg(&msg.S2C_SetUserRole{
			Error: msg.S2C_SetUserRole_AccountIDInvalid,
		})
		return
	}
	if user.data.userData.AccountID == m.AccountID {
		log.Debug("不能设置自己")
		user.WriteMsg(&msg.S2C_SetUserRole{
			Error: msg.S2C_SetUserRole_NotYourself,
		})
		return
	}
	if common.Index([]int{-1, 0, 1, 2, 3, 4}, m.Role) == -1 {
		log.Debug("角色 %v 无效", m.Role)
		user.WriteMsg(&msg.S2C_SetUserRole{
			Error: msg.S2C_SetUserRole_RoleInvalid,
			Role:  m.Role,
		})
		return
	}
	if user.data.userData.Role <= m.Role {
		log.Debug("userID: %v 没有权限", user.data.userData.UserID)
		user.WriteMsg(&msg.S2C_SetUserRole{
			Error: msg.S2C_SetUserRole_PermissionDenied,
		})
		return
	}
	user.setRole(m.AccountID, m.Role)
}

func handleGetUserInfo(args []interface{}) {
	m := args[0].(*msg.C2S_GetUserInfo)
	// 消息的发送者
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil || m.AccountID < 1 {
		return
	}
	user.getUserInfo(m.AccountID)
}

func handleTransferRoomCard(args []interface{}) {
	m := args[0].(*msg.C2S_TransferRoomCard)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if m.AccountID == 0 {
		log.Debug("账户ID为0")
		user.WriteMsg(&msg.S2C_TransferRoomCard{
			Error: msg.S2C_TransferRoomCard_AccountIDInvalid,
		})
		return
	}
	if user.data.userData.AccountID == m.AccountID {
		log.Debug("不能给自己转房卡")
		user.WriteMsg(&msg.S2C_TransferRoomCard{
			Error: msg.S2C_TransferRoomCard_NotYourself,
		})
		return
	}
	if m.RoomCards < 1 || m.RoomCards > user.data.userData.RoomCards {
		log.Debug("房卡数量: %v 无效", m.RoomCards)
		user.WriteMsg(&msg.S2C_TransferRoomCard{
			Error: msg.S2C_TransferRoomCard_RoomCardsInvalid,
		})
		return
	}
	if user.data.userData.Role < roleAgent {
		log.Debug("userID: %v 没有权限", user.data.userData.UserID)
		user.WriteMsg(&msg.S2C_TransferRoomCard{
			Error: msg.S2C_TransferRoomCard_PermissionDenied,
		})
		return
	}
	if systemOn {
		user.transferRoomCard(m.AccountID, m.RoomCards)
	} else {
		user.Close()
	}
}

func handleGetTransferRoomCardRecord(args []interface{}) {
	m := args[0].(*msg.C2S_GetTransferRoomCardRecord)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}

	if m.PageNumber > 0 && m.PageSize > 0 {
		user.getTransferRoomCardRecordByPage(m)
	} else {
		if m.AccountID < 1 || m.AccountID == user.data.userData.AccountID {
			user.getTransferRoomCardRecord(user.data.userData.AccountID)
		} else {
			if user.data.userData.Role < roleAdmin {
				log.Debug("userID: %v 没有权限", user.data.userData.UserID)
				user.WriteMsg(&msg.S2C_TransferRoomCardRecord{
					Error: msg.S2C_TransferRoomCardRecord_PermissionDenied,
				})
				return
			}
			user.getTransferRoomCardRecord(m.AccountID)
		}
	}
}

func handleGetAllTransferRoomCardRecord(args []interface{}) {
	m := args[0].(*msg.C2S_GetAllTransferRoomCardRecord)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil || m.PageNumber < 1 || m.PageSize < 1 ||
		common.Index([]int{10, 15, 20}, m.PageSize) == -1 {
		return
	}
	if m.StartTime > 0 || m.EndTime > 0 {
		if m.EndTime > m.StartTime {
			user.getAllTransferRoomCardRecordByTime(m)
		}
	} else {
		user.getAllTransferRoomCardRecord(m)
	}
}

func handleGetAllAgentInfo(args []interface{}) {
	m := args[0].(*msg.C2S_GetAllAgentInfo)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}

	user := a.UserData().(*AgentInfo).user
	if user == nil || m.PageNumber < 1 || m.PageSize < 1 ||
		common.Index([]int{10, 15, 20}, m.PageSize) == -1 {
		return
	}
	if m.StartTime > 0 || m.EndTime > 0 {
		if m.EndTime > m.StartTime {
			user.getAllAgentInfoByTime(m)
		}
	} else {
		user.getAllAgentInfo(m)
	}
}

func handleGetAllUserInfo(args []interface{}) {
	m := args[0].(*msg.C2S_GetAllUserInfo)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}

	user := a.UserData().(*AgentInfo).user
	if user == nil || m.PageNumber < 1 || m.PageSize < 1 ||
		common.Index([]int{10, 15, 20}, m.PageSize) == -1 {
		return
	}
	if len(m.Nickname) > 0 {
		user.getAllUserInfoByNickname(m)
	} else {
		user.getAllUserInfo(m)
	}
}

func handleGetBlackList(args []interface{}) {
	m := args[0].(*msg.C2S_GetBlackList)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}

	user := a.UserData().(*AgentInfo).user
	if user == nil || m.PageNumber < 1 || m.PageSize < 1 ||
		common.Index([]int{10, 15, 20}, m.PageSize) == -1 {
		return
	}
	user.getBlackList(m)
}

func handleHeartbeat(args []interface{}) {
	// 收到的 C2S_Heartbeat 消息
	m := args[0].(*msg.C2S_Heartbeat)
	_ = m
	// 消息的发送者
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	user.heartbeatStop = false
}

func handleCreateHNZZRoom(args []interface{}) {
	// 收到的 C2S_CreateHNZZRoom 消息
	m := args[0].(*msg.C2S_CreateHNZZRoom)
	// 消息的发送者
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if _, ok := userIDRooms[user.data.userData.UserID]; ok {
		user.WriteMsg(&msg.S2C_CreateRoom{
			Error: msg.S2C_CreateRoom_InOtherRoom,
		})
		return
	}
	if !systemOn {
		user.Close()
		return
	}
	switch m.RoomType {
	case roomPrivate:
		user.createPrivateRoom(m)
		return
	case roomRedPacketPrivate:
		user.createRedPacketPrivateRoom(m.RedPacketType)
		return
	}
}

func handleEnterRoom(args []interface{}) {
	// 收到的 C2S_EnterRoom 消息
	m := args[0].(*msg.C2S_EnterRoom)
	// 消息的发送者
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if strings.TrimSpace(m.RoomNumber) == "" {
		if r, ok := userIDRooms[user.data.userData.UserID]; ok {
			hnzzRoom := r.(*HNZZRoom)
			if hnzzRoom.rule.GPSAntiCheat {
				user.enterGPSRoom(r, m.GPS, m.Location)
			} else {
				user.enterRoom(r)
			}
		} else {
			user.WriteMsg(&msg.S2C_EnterRoom{
				Error: msg.S2C_EnterRoom_Unknown,
			})
		}
		return
	}

	if r, ok := roomNumberRooms[m.RoomNumber]; ok {
		if systemOn {
			hnzzRoom := r.(*HNZZRoom)
			if hnzzRoom.rule.GPSAntiCheat {
				user.enterGPSRoom(r, m.GPS, m.Location)
			} else {
				user.enterRoom(r)
			}
		} else {
			user.Close()
		}
		return
	} else {
		user.WriteMsg(&msg.S2C_EnterRoom{
			Error:      msg.S2C_EnterRoom_NotCreated,
			RoomNumber: m.RoomNumber,
		})
	}
}

func handleGetAllPlayers(args []interface{}) {
	// 收到的 C2S_GetAllPlayers 消息
	m := args[0].(*msg.C2S_GetAllPlayers)
	_ = m
	// 消息的发送者
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if r, ok := userIDRooms[user.data.userData.UserID]; ok {
		user.getAllPlayers(r)
	}
}

func handleExitOrDisbandRoom(args []interface{}) {
	// 收到的 C2S_DisbandRoom 消息
	m := args[0].(*msg.C2S_ExitOrDisbandRoom)
	_ = m
	// 消息的发送者
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if r, ok := userIDRooms[user.data.userData.UserID]; ok {
		user.exitOrDisbandRoom(r, true)
	}
}

func handlePrepare(args []interface{}) {
	m := args[0].(*msg.C2S_Prepare)
	_ = m
	// 消息的发送者
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if r, ok := userIDRooms[user.data.userData.UserID]; ok {
		user.doPrepare(r)
	}
}

func handleMahjongDiscard(args []interface{}) {
	m := args[0].(*msg.C2S_MahjongDiscard)
	// 消息的发送者
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if r, ok := userIDRooms[user.data.userData.UserID]; ok {
		user.doDiscard(r, m.Tile)
	}
}

func handleMahjongWin(args []interface{}) {
	m := args[0].(*msg.C2S_MahjongWin)
	_ = m
	// 消息的发送者
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if r, ok := userIDRooms[user.data.userData.UserID]; ok {
		user.doWin(r)
	}
}

func handleMahjongKong(args []interface{}) {
	m := args[0].(*msg.C2S_MahjongKong)
	_ = m
	// 消息的发送者
	a := args[1].(gate.Agent)
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

func handleMahjongPong(args []interface{}) {
	m := args[0].(*msg.C2S_MahjongPong)
	_ = m
	// 消息的发送者
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if r, ok := userIDRooms[user.data.userData.UserID]; ok {
		log.Debug("userID %v 碰", user.data.userData.UserID)
		user.doPong(r)
	}
}

func handleMahjongChow(args []interface{}) {
	m := args[0].(*msg.C2S_MahjongChow)
	// 消息的发送者
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if r, ok := userIDRooms[user.data.userData.UserID]; ok {
		log.Debug("userID %v 吃 %v", user.data.userData.UserID, mahjong.ToTileString(m.Meld))
		user.doChow(r, m.Meld)
	}
}

func handleMahjongPass(args []interface{}) {
	m := args[0].(*msg.C2S_MahjongPass)
	_ = m
	// 消息的发送者
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if r, ok := userIDRooms[user.data.userData.UserID]; ok {
		user.doPass(r)
	}
}

func handleAgreeDisbandRoom(args []interface{}) {
	m := args[0].(*msg.C2S_AgreeDisbandRoom)
	_ = m
	// 消息的发送者
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if r, ok := userIDRooms[user.data.userData.UserID]; ok {
		user.agreeDisbandRoom(r)
	}
}

func handleRefuseDisbandRoom(args []interface{}) {
	m := args[0].(*msg.C2S_RefuseDisbandRoom)
	_ = m
	// 消息的发送者
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if r, ok := userIDRooms[user.data.userData.UserID]; ok {
		user.refuseDisbandRoom(r)
	}
}

func handleGetRoomCards(args []interface{}) {
	m := args[0].(*msg.C2S_GetRoomCards)
	_ = m
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	user.WriteMsg(&msg.S2C_UpdateRoomCards{
		RoomCards: user.data.userData.RoomCards,
	})
	user.WriteMsg(&msg.S2C_UpdateRoomCardsMatchOnlineNumber{
		Numbers: roomCardMatchOnlineNumber,
	})
	user.sendRedPacketMatchOnlineNumber()
	user.sendUntakenRedPacketMatchPrizeNumber()
}

func handleStartHNZZMatching(args []interface{}) {
	m := args[0].(*msg.C2S_StartHNZZMatching)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if _, ok := userIDRooms[user.data.userData.UserID]; ok {
		user.WriteMsg(&msg.S2C_CreateRoom{
			Error: msg.S2C_CreateRoom_InOtherRoom,
		})
		return
	}
	if !systemOn {
		user.Close()
		return
	}
	switch m.RoomType {
	case roomPractice:
		user.createOrEnterPracticeRoom()
		return
	case roomRoomCardMatch:
		user.createOrEnterRoomCardMatchRoom(m.RoomCards)
		return
	case roomRedPacketMatching:
		user.createOrEnterRedPacketMatchingRoom(m.RedPacketType)
		return
	default:
		user.WriteMsg(&msg.S2C_CreateRoom{
			Error: msg.S2C_CreateRoom_RuleError,
		})
	}
}

func handleGetTotalResults(args []interface{}) {
	m := args[0].(*msg.C2S_GetTotalResults)
	_ = m
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if systemOn {
		user.getTotalResult()
	} else {
		user.Close()
	}
}

func handleGetRoundResults(args []interface{}) {
	m := args[0].(*msg.C2S_GetRoundResults)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if systemOn {
		user.getRoundResults(m.TotalResultID)
	} else {
		user.Close()
	}
}

func handleCompleteDailyShare(args []interface{}) {
	m := args[0].(*msg.C2S_CompleteDailyShare)
	_ = m
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if systemOn {
		user.completeDailyShare()
	} else {
		user.Close()
	}
}

func handleTextMessage(args []interface{}) {
	m := args[0].(*msg.C2S_TextMessage)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil || m.Message == "" {
		return
	}
	if r, ok := userIDRooms[user.data.userData.UserID]; ok {
		log.Debug("userID: %v 发送文字: %v", user.data.userData.UserID, m.Message)
		user.sendTextMessage(r, m.Message)
	}
}

func handleExpressionMessage(args []interface{}) {
	m := args[0].(*msg.C2S_ExpressionMessage)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil || m.Expression < 0 {
		return
	}
	if r, ok := userIDRooms[user.data.userData.UserID]; ok {
		log.Debug("userID: %v 发送表情: %v", user.data.userData.UserID, m.Expression)
		user.sendExpressionMessage(r, m.Expression)
	}
}

func handleGCloudVoiceMessage(args []interface{}) {
	m := args[0].(*msg.C2S_GCloudVoiceMessage)
	a := args[1].(gate.Agent)

	if a.UserData() == nil || m.FileID == "" {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if r, ok := userIDRooms[user.data.userData.UserID]; ok {
		log.Debug("userID: %v 发送语音: %v", user.data.userData.UserID, m.FileID)
		user.sendGCloudVoiceMessage(r, m.FileID)
	}
}

func handleGetHNZZIOSProductList(args []interface{}) {
	m := args[0].(*msg.C2S_GetHNZZIOSProductList)
	_ = m
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if systemOn {
		user.WriteMsg(&msg.S2C_HNZZIOSProductList{
			Infos: hnzzIOSProductInfos,
		})
	} else {
		user.Close()
	}
}

func handleGetHNZZAndroidProductList(args []interface{}) {
	m := args[0].(*msg.C2S_GetHNZZAndroidProductList)
	_ = m
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if systemOn {
		user.WriteMsg(&msg.S2C_HNZZAndroidProductList{
			Infos: hnzzAndroidProductInfos,
		})
	} else {
		user.Close()
	}
}

func handleIAPReceiptData(args []interface{}) {
	m := args[0].(*msg.C2S_IAPReceiptData)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if systemOn {
		if len(m.ReceiptData) > 0 {
			log.Debug("%v", m.ReceiptData)
			mm := map[string]interface{}{}
			mm["receipt-data"] = m.ReceiptData
			data, err := json.Marshal(mm)
			if err == nil {
				user.verifyProductionEnvironmentReceipt(string(data))
			} else {
				log.Error("marshal message %v error: %v", reflect.TypeOf(mm), err)
			}
		}
	} else {
		user.Close()
	}
}

func handleGetRedPacketMatchRecord(args []interface{}) {
	m := args[0].(*msg.C2S_GetRedPacketMatchRecord)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if m.PageNumber < 1 {
		m.PageNumber = 1
	}
	if m.PageSize < 1 {
		m.PageSize = 10
	}
	user.sendRedPacketMatchRecord(m.PageNumber, m.PageSize)
}

func handleTakeRedPacketMatchPrize(args []interface{}) {
	m := args[0].(*msg.C2S_TakeRedPacketMatchPrize)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	user.takeRedPacketMatchPrize(m.ID)
}

func handleFakeWXPay(args []interface{}) {
	m := args[0].(*msg.C2S_FakeWXPay)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	user.FakeWXPay(m.TotalFee)
}

func handleGetCircleLoginCode(args []interface{}) {
	m := args[0].(*msg.C2S_GetCircleLoginCode)
	_ = m
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	userID := user.data.userData.UserID
	user.requestCircleLoginCode(func(loginCode string) {
		if theUser, ok := userIDUsers[userID]; ok {
			theUser.WriteMsg(&msg.S2C_UpdateCircleLoginCode{
				Error:     msg.S2C_UpdateCircleLoginCode_OK,
				LoginCode: loginCode,
			})
		}
	}, func() {
		if theUser, ok := userIDUsers[userID]; ok {
			if theUser == user {
				theUser.WriteMsg(&msg.S2C_UpdateCircleLoginCode{
					Error: msg.S2C_UpdateCircleLoginCode_Error,
				})
			}
		}
	})

}

func handleSetRobotData(args []interface{}) {
	m := args[0].(*msg.C2S_SetRobotData)
	a := args[1].(gate.Agent)

	agentInfo := a.UserData().(*AgentInfo)
	if agentInfo == nil || agentInfo.user == nil {
		return
	}
	user := agentInfo.user
	if user.isRobot() {
		if m.RoomCards > 0 {
			user.setRobotRoomCards(m.RoomCards)
		}
	} else {
		user.data.userData.RoomCards = 1000
		user.data.userData.Role = roleRobot
		user.data.userData.LoginIP = m.LoginIP
	}
}
