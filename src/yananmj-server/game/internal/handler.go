package internal

import (
	"encoding/json"
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
	"reflect"
	"strings"
	"yananmj-server/common"
	"yananmj-server/game/mahjong"
	"yananmj-server/msg"
)

func init() {
	handler(&data_struct.C2S_SetSystemOn{}, handleSetSystemOn)
	handler(&data_struct.C2S_SetUsernamePassword{}, handleSetUsernamePassword)
	handler(&data_struct.C2S_SetYananConfig{}, handleSetYananConfig)
	handler(&data_struct.C2S_SetUserRole{}, handleSetUserRole)
	handler(&data_struct.C2S_TransferRoomCard{}, handleTansferRoomCard)
	handler(&data_struct.C2S_GetUserInfo{}, handleGetUserInfo)
	handler(&data_struct.C2S_GetTransferRoomCardRecord{}, handleGetTransferRoomCardRecord)
	handler(&data_struct.C2S_GetAllTransferRoomCardRecord{}, handleGetAllTransferRoomCardRecord)
	handler(&data_struct.C2S_GetAllAgentInfo{}, handleGetAllAgentInfo)
	handler(&data_struct.C2S_GetAllUserInfo{}, handleGetAllUserInfo)
	handler(&data_struct.C2S_GetBlackList{}, handleGetBlackList)

	handler(&data_struct.C2S_Heartbeat{}, handleHeartbeat)
	handler(&data_struct.C2S_CreateYananRoom{}, handleCreateYananRoom)
	handler(&data_struct.C2S_StartYananMatching{}, handleStartYananMatching)
	handler(&data_struct.C2S_EnterRoom{}, handleEnterRoom)
	handler(&data_struct.C2S_GetAllPlayers{}, handleGetAllPlayers)
	handler(&data_struct.C2S_Prepare{}, handlePrepare)
	handler(&data_struct.C2S_SetGun{}, handleSetGun)
	handler(&data_struct.C2S_ExitOrDisbandRoom{}, handleExitOrDisbandRoom)
	handler(&data_struct.C2S_AgreeDisbandRoom{}, handleAgreeDisbandRoom)
	handler(&data_struct.C2S_RefuseDisbandRoom{}, handleRefusedDisbandRoom)
	handler(&data_struct.C2S_MahjongDiscard{}, handleMahjongDiscard)
	handler(&data_struct.C2S_MahjongWin{}, handleMahjongWin)
	handler(&data_struct.C2S_MahjongKong{}, handleMahjongKong)
	handler(&data_struct.C2S_MahjongPong{}, handleMahjongPong)
	handler(&data_struct.C2S_MahjongPass{}, handleMahjongPass)
	handler(&data_struct.C2S_GetRoomCards{}, handleGetRoomCards)
	handler(&data_struct.C2S_GetTotalResults{}, handleGetTotalResults)
	handler(&data_struct.C2S_GetRoundResults{}, handleGetRoundResults)
	handler(&data_struct.C2S_MahjongManaged{}, handleMahjongManaged)
	handler(&data_struct.C2S_TextMessage{}, handleTextMessage)
	handler(&data_struct.C2S_ExpressionMessage{}, handleExpressionMessage)
	handler(&data_struct.C2S_CompleteDailyShare{}, handleCompleteDailyShare)
	handler(&data_struct.C2S_GCloudVoiceMessage{}, handleGCloudVoiceMessage)
	handler(&data_struct.C2S_GetYananIOSProductList{}, handleGetYananIOSProductList)
	handler(&data_struct.C2S_GetYananAndriodProductList{}, handleGetYananAndriodProductList)
	handler(&data_struct.C2S_IAPReceiptData{}, handleIAPReceiptData)
	handler(&data_struct.C2S_GetRedPacketMatchRecord{}, handleGetRedPacketMatchRecord)
	handler(&data_struct.C2S_TakeRedPacketMatchPrize{}, handleTakeRedPacketMatchPrize)
	handler(&data_struct.C2S_GetCircleLoginCode{}, handleGetCircleLoginCode)
	handler(&data_struct.C2S_FakeWXPay{}, handleFakeWXPay)

	handler(&data_struct.C2S_SetRobotData{}, handleSetRobotData)
}

func handler(m interface{}, h interface{}) {
	skeleton.RegisterChanRPC(reflect.TypeOf(m), h)
}

//心跳机制
func handleHeartbeat(args []interface{}) {
	m := args[0].(*data_struct.C2S_Heartbeat)
	_ = m
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

func handleSetYananConfig(args []interface{}) {
	m := args[0].(*data_struct.C2S_SetYananConfig)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		return
	}

	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}

	if user.data.userData.Role < roleAdmin {
		log.Debug("userId:%v 没有权限", user.data.userData.UserID)
		user.WriteMsg(&data_struct.S2C_SetYananConfig{
			Error: data_struct.S2C_SetYananConfig_PermissionDenied,
		})
		return
	}

	if m.AndriodVersion > 0 {
		user.setAndriodVersion(m.AndriodVersion)
	}
	if len(m.AndriodDownloadUrl) > 0 {
		user.setAndriodDownloadUrl(m.AndriodDownloadUrl)
	}
	if m.IOSVersion > 0 {
		user.setIOSVersion(m.IOSVersion)
	}
	if len(m.IOSDownloadUrl) > 0 {
		user.setIOSDownloadUrl(m.IOSDownloadUrl)
	}
	if len(m.Notice) > 0 {
		user.setNotice(m.Notice)
	}
	if len(m.Radio) > 0 {
		user.setRadio(m.Radio)
	}
	if len(m.WeChatNumber) > 0 {
		user.setWeChatNumber(m.WeChatNumber)
	}
}

//设置是否系统更新
func handleSetSystemOn(args []interface{}) {
	m := args[0].(*data_struct.C2S_SetSystemOn)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}

	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}

	if user.data.userData.Role < roleRoot {
		log.Error("userID：%v 没有权限", user.data.userData.UserID)
		user.WriteMsg(&data_struct.S2C_SetSystemOn{
			Error: data_struct.S2C_SetSystemOn_PermissionDenied,
		})
		return
	}

	systemOn = m.On
	user.WriteMsg(&data_struct.S2C_SetSystemOn{
		Error: data_struct.S2C_SetSystemOn_Ok,
		On:    m.On,
	})
	if systemOn {
		log.Debug("userID:%v  设置系统开成功", user.data.userData.UserID)
	} else {
		log.Debug("userID:%v 设置系统关成功", user.data.userData.UserID)
	}
}

//设置用户角色
func handleSetUserRole(args []interface{}) {
	m := args[0].(*data_struct.C2S_SetUserRole)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		return
	}

	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}

	if m.AccountID == 0 {
		log.Debug("账户ID为空")
		user.WriteMsg(&data_struct.S2C_SetUserRole{Error: data_struct.S2C_SetUserRole_AccountIDInvalid})
		return
	}

	if user.data.userData.AccountID == m.AccountID {
		log.Debug("不能设置自己")
		user.WriteMsg(&data_struct.S2C_SetUserRole{Error: data_struct.S2C_SetUserRole_NotYourself})
		return
	}

	if common.Index([]int{-1, 0, 1, 2, 3, 4}, m.Role) == -1 {
		log.Debug("设置角色: %v 错误", m.Role)
		user.WriteMsg(&data_struct.S2C_SetUserRole{Error: data_struct.S2C_SetUserRole_RoleInvalid})
		return
	}

	if user.data.userData.Role <= m.Role {
		log.Debug("没有权限")
		user.WriteMsg(&data_struct.S2C_SetUserRole{Error: data_struct.S2C_SetSystemOn_PermissionDenied})
		return
	}
	user.setRole(m.AccountID, m.Role)
}

//设置账密登录
func handleSetUsernamePassword(args []interface{}) {
	m := args[0].(*data_struct.C2S_SetUsernamePassword)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		return
	}

	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if strings.TrimSpace(m.Username) == "" || strings.TrimSpace(m.Password) == "" {
		//用户名密码不能为空
		return
	}

	if user.data.userData.Username == "" || user.data.userData.Username == m.Username {
		user.setUsernamePassword(m.Username, m.Password)
	} else {
		log.Debug("userID: %v 用户名无需更改", user.data.userData.UserID)
	}

}

//创建陕西麻将房间
func handleCreateYananRoom(args []interface{}) {
	m := args[0].(*data_struct.C2S_CreateYananRoom)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		return
	}

	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}

	if _, ok := userIDRooms[user.data.userData.UserID]; ok {
		user.WriteMsg(&data_struct.S2C_CreateRoom{
			Error: data_struct.S2C_CreateRoom_InOtherRoom,
		})
		return
	}
	if !systemOn {
		user.Close()
		return
	}

	log.Debug("roomType: %v", m.RoomType)
	switch m.RoomType {
	case roomPrivate:
		user.createPrivateRoom(m)
		return
	case roomRedPacketPrivate:
		user.createRedPacketPrivateRoom(m.RedPacketType)
		return
	}
}

func handleStartYananMatching(args []interface{}) {
	m := args[0].(*data_struct.C2S_StartYananMatching)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		return
	}

	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}

	if _, ok := userIDRooms[user.data.userData.UserID]; ok {
		user.WriteMsg(&data_struct.S2C_CreateRoom{
			Error: data_struct.S2C_CreateRoom_InOtherRoom,
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
		user.WriteMsg(&data_struct.S2C_CreateRoom{
			Error: data_struct.S2C_CreateRoom_InnerError,
		})
	}
}

//进入陕西延安麻将房间
func handleEnterRoom(args []interface{}) {
	m := args[0].(*data_struct.C2S_EnterRoom)
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
			yananRoom := r.(*YananRoom)
			if yananRoom.rule.GPSAntiCheat {
				user.enterGPSRoom(r, m.GPS, m.Location)
			} else {
				user.enterRoom(r)
			}
		} else {
			user.WriteMsg(&data_struct.S2C_EnterRoom{
				Error: data_struct.S2C_EnterRoom_Unknow,
			})
		}
		return
	}

	if r, ok := roomNumberRooms[m.RoomNumber]; ok {
		if systemOn {
			yananRoom := r.(*YananRoom)
			if yananRoom.rule.GPSAntiCheat {
				user.enterGPSRoom(r, m.GPS, m.Location)
			} else {
				user.enterRoom(r)
			}
		} else {
			user.Close()
		}
		return
	}

	user.WriteMsg(&data_struct.S2C_EnterRoom{
		Error:      data_struct.S2C_EnterRoom_NotCreated,
		RoomNumber: m.RoomNumber,
	})
	log.Debug("房间: %v 未创建", m.RoomNumber)
}

//获取所有玩家
func handleGetAllPlayers(args []interface{}) {
	m := args[0].(*data_struct.C2S_GetAllPlayers)
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
		user.getAllPlayers(r)
	}
}

//玩家准备
func handlePrepare(args []interface{}) {
	m := args[0].(*data_struct.C2S_Prepare)
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
		user.doPrepare(r)
	}
}

//下炮子
func handleSetGun(args []interface{}) {
	m := args[0].(*data_struct.C2S_SetGun)
	a := args[1].(gate.Agent)

	log.Debug("下的炮子数: %v", m.Gun)
	if a.UserData() == nil {
		return
	}

	user := a.UserData().(*AgentInfo).user
	if user == nil || m.Gun < 0 {
		return
	}

	if r, ok := userIDRooms[user.data.userData.UserID]; ok {
		user.doSetGun(r, m.Gun)
	}
}

//玩家出牌
func handleMahjongDiscard(args []interface{}) {
	m := args[0].(*data_struct.C2S_MahjongDiscard)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		return
	}

	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}

	if r, ok := userIDRooms[user.data.userData.UserID]; ok {
		log.Debug("userID: %v 出牌: %v", user.data.userData.UserID, mahjong.ToTileString([]int{m.Tile}))
		user.doDiscard(r, m.Tile)
	}
}

//胡牌
func handleMahjongWin(args []interface{}) {
	m := args[0].(*data_struct.C2S_MahjongWin)
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
		log.Debug("userID: %v 胡牌", user.data.userData.UserID)
		user.doWin(r)
	}
}

//杠牌
func handleMahjongKong(args []interface{}) {
	m := args[0].(*data_struct.C2S_MahjongKong)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}

	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}

	if r, ok := userIDRooms[user.data.userData.UserID]; ok {
		log.Debug("userID: %v 杠: %v", user.data.userData.UserID, mahjong.ToTileString(m.Meld))
		user.doKong(r, m.Meld)
	}
}

//碰牌
func handleMahjongPong(args []interface{}) {
	m := args[0].(*data_struct.C2S_MahjongPong)
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
		log.Debug("userID: %v 碰", user.data.userData.UserID)
		user.doPong(r)
	}

}

//过牌
func handleMahjongPass(args []interface{}) {
	m := args[0].(*data_struct.C2S_MahjongPass)
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
		user.doPass(r)
	}

}

//取消托管
func handleMahjongManaged(args []interface{}) {
	m := args[0].(*data_struct.C2S_MahjongManaged)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}

	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}

	if r, ok := userIDRooms[user.data.userData.UserID]; ok {
		user.doCancelTrusteeship(r, m.Managed)
	}
}

//玩家退出或解散房间
func handleExitOrDisbandRoom(args []interface{}) {
	m := args[0].(*data_struct.C2S_ExitOrDisbandRoom)
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
		user.exitOrDisbandRoom(r, true)
	}
}

//同意解散房间
func handleAgreeDisbandRoom(args []interface{}) {
	m := args[0].(*data_struct.C2S_AgreeDisbandRoom)
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
		user.agreeDisbandRoom(r)
	}
}

//拒绝解散房间
func handleRefusedDisbandRoom(args []interface{}) {
	m := args[0].(*data_struct.C2S_RefuseDisbandRoom)
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
		user.refusedDisbandRoom(r)
	}
}

//获取房卡数量
func handleGetRoomCards(args []interface{}) {
	m := args[0].(*data_struct.C2S_GetRoomCards)
	_ = m
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	user.WriteMsg(&data_struct.S2C_UpdateRoomCards{
		RoomCards: user.data.userData.RoomCards,
	})
	user.WriteMsg(&data_struct.S2C_UpdateRoomCardsMatchOnlineNumber{
		Numbers: roomCardMatchOnlineNumber,
	})
	user.sendRedPacketMatchOnlineNumber()
	user.sendUntakenRedPacketMatchPrizeNumber()
}

//转房卡
func handleTansferRoomCard(args []interface{}) {
	m := args[0].(*data_struct.C2S_TransferRoomCard)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		return
	}

	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}

	if m.AccountID == 0 {
		log.Debug("账户ID为空")
		user.WriteMsg(&data_struct.S2C_TransferRoomCard{
			Error: data_struct.S2C_TransferRoomCard_AccountIDInvalid,
		})
		return
	}

	if user.data.userData.AccountID == m.AccountID {
		log.Debug("不能给自己转房卡")
		user.WriteMsg(&data_struct.S2C_TransferRoomCard{
			Error: data_struct.S2C_TransferRoomCard_NotYourself,
		})
		return
	}

	if m.RoomCards < 1 || m.RoomCards > user.data.userData.RoomCards {
		log.Debug("房卡数量: %v 无效", m.RoomCards)
		user.WriteMsg(&data_struct.S2C_TransferRoomCard{
			Error: data_struct.S2C_TransferRoomCard_RoomCardsInvalid,
		})
		return
	}

	if user.data.userData.Role < roleAgent {
		log.Debug("userID: %v 没有权限", user.data.userData.UserID)
		user.WriteMsg(&data_struct.S2C_TransferRoomCard{
			Error: data_struct.S2C_TransferRoomCard_PermissionDenied,
		})
		return
	}

	if systemOn {
		user.transferRoomCard(m.AccountID, m.RoomCards)
	} else {
		user.Close()
	}
}

//获取历史总成绩
func handleGetTotalResults(args []interface{}) {
	m := args[0].(*data_struct.C2S_GetTotalResults)
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

//获取历史单局成绩
func handleGetRoundResults(args []interface{}) {
	m := args[0].(*data_struct.C2S_GetRoundResults)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		return
	}

	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}

	if systemOn {
		user.getRoundResult(m.TotalResultID)
	} else {
		user.Close()
	}
}

//发送文字消息
func handleTextMessage(args []interface{}) {
	m := args[0].(*data_struct.C2S_TextMessage)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		return
	}

	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}

	if r, ok := userIDRooms[user.data.userData.UserID]; ok {
		log.Debug("userID: %v 发送文字 %v", user.data.userData.UserID, m.Message)
		user.sendTextMessage(r, m.Message)
	}
}

//发送表情消息
func handleExpressionMessage(args []interface{}) {
	m := args[0].(*data_struct.C2S_ExpressionMessage)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		return
	}

	user := a.UserData().(*AgentInfo).user

	if user == nil {
		return
	}

	if r, ok := userIDRooms[user.data.userData.UserID]; ok {
		log.Debug("userID: %v 发送表情 %v", user.data.userData.UserID, m.Expression)
		user.sendExpressionMessage(r, m.Expression)
	}

}

//每日分享
func handleCompleteDailyShare(args []interface{}) {
	m := args[0].(*data_struct.C2S_CompleteDailyShare)
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

//发送语音
func handleGCloudVoiceMessage(args []interface{}) {
	m := args[0].(*data_struct.C2S_GCloudVoiceMessage)
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

//获取商品列表
func handleGetYananIOSProductList(args []interface{}) {
	m := args[0].(*data_struct.C2S_GetYananIOSProductList)
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
		user.WriteMsg(&data_struct.S2C_YananIOSProductList{
			Infos: yananIOSProductInfos,
		})
	} else {
		user.Close()
	}
}

func handleGetYananAndriodProductList(args []interface{}) {
	m := args[0].(*data_struct.C2S_GetYananAndriodProductList)
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
		user.WriteMsg(&data_struct.S2C_YananAndriodProductList{
			Infos: yananAndriodProductInfos,
		})
	} else {
		user.Close()
	}
}

//IOS内购验证
func handleIAPReceiptData(args []interface{}) {
	m := args[0].(*data_struct.C2S_IAPReceiptData)
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
			mm["receipt_data"] = m.ReceiptData
			data, err := json.Marshal(mm)
			if err == nil {
				user.verifyProductionEnvironmentReceipt(string(data))
			} else {
				log.Error("marshal message %v error %v ", reflect.TypeOf(mm), err)
			}
		}
	} else {
		user.Close()
	}
}

func handleGetTransferRoomCardRecord(args []interface{}) {
	m := args[0].(*data_struct.C2S_GetTransferRoomCardRecord)
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
				user.WriteMsg(&data_struct.S2C_TransferRoomCardRecord{
					Error: data_struct.S2C_TransferRoomCardRecord_PermissionDenied,
				})
				return
			}
			user.getTransferRoomCardRecord(m.AccountID)
		}
	}
}

//转卡记录
func handleGetAllTransferRoomCardRecord(args []interface{}) {
	m := args[0].(*data_struct.C2S_GetAllTransferRoomCardRecord)
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

//所有代理
func handleGetAllAgentInfo(args []interface{}) {
	m := args[0].(*data_struct.C2S_GetAllAgentInfo)
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
		user.getAllAgent(m)
	}
}

//所有玩家数据
func handleGetAllUserInfo(args []interface{}) {
	m := args[0].(*data_struct.C2S_GetAllUserInfo)
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

func handleGetUserInfo(args []interface{}) {
	m := args[0].(*data_struct.C2S_GetUserInfo)
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

//黑名单
func handleGetBlackList(args []interface{}) {
	m := args[0].(*data_struct.C2S_GetBlackList)
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

func handleGetRedPacketMatchRecord(args []interface{}) {
	m := args[0].(*data_struct.C2S_GetRedPacketMatchRecord)
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
	m := args[0].(*data_struct.C2S_TakeRedPacketMatchPrize)
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

func handleGetCircleLoginCode(args []interface{}) {
	m := args[0].(*data_struct.C2S_GetCircleLoginCode)
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
			theUser.WriteMsg(&data_struct.S2C_UpdateCircleLoginCode{
				Error:     data_struct.S2C_UpdateCircleLoginCode_OK,
				LoginCode: loginCode,
			})
		}
	}, func() {
		if theUser, ok := userIDUsers[userID]; ok {
			if theUser == user {
				theUser.WriteMsg(&data_struct.S2C_UpdateCircleLoginCode{
					Error: data_struct.S2C_UpdateCircleLoginCode_Error,
				})
			}
		}
	})
}

func handleFakeWXPay(args []interface{}) {
	m := args[0].(*data_struct.C2S_FakeWXPay)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	log.Debug("fee: %v", m.TotalFee)
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	user.FakeWXPay(m.TotalFee)
}

func handleSetRobotData(args []interface{}) {
	m := args[0].(*data_struct.C2S_SetRobotData)
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
