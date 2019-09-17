package gate

import (
	"hnzzmj-server/game"
	"hnzzmj-server/login"
	"hnzzmj-server/msg"
)

func init() {
	// login
	msg.Processor.SetRouter(&msg.C2S_WeChatLogin{}, login.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_TokenLogin{}, login.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_UsernamePasswordLogin{}, login.ChanRPC)
	// game
	msg.Processor.SetRouter(&msg.C2S_Heartbeat{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_CreateHNZZRoom{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_EnterRoom{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_GetAllPlayers{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_ExitOrDisbandRoom{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_Prepare{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_MahjongDiscard{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_MahjongWin{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_MahjongKong{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_MahjongPong{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_MahjongChow{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_MahjongPass{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_AgreeDisbandRoom{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_RefuseDisbandRoom{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_SetUsernamePassword{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_SetUserRole{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_GetRoomCards{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_TransferRoomCard{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_SetHNZZConfig{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_StartHNZZMatching{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_SetSystemOn{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_GetTotalResults{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_GetRoundResults{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_CompleteDailyShare{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_TextMessage{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_ExpressionMessage{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_GCloudVoiceMessage{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_GetHNZZIOSProductList{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_GetHNZZAndroidProductList{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_IAPReceiptData{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_GetUserInfo{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_GetTransferRoomCardRecord{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_GetAllTransferRoomCardRecord{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_GetAllAgentInfo{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_GetAllUserInfo{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_GetBlackList{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_GetRedPacketMatchRecord{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_TakeRedPacketMatchPrize{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_FakeWXPay{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_GetCircleLoginCode{}, game.ChanRPC)

	msg.Processor.SetRouter(&msg.C2S_SetRobotData{}, game.ChanRPC)
}
