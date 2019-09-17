package gate

import (
	"yananmj-server/game"
	"yananmj-server/login"
	"yananmj-server/msg"
)

func init() {
	data_struct.Processor.SetRouter(&data_struct.C2S_WeChatLogin{}, login.ChanRPC)
	data_struct.Processor.SetRouter(&data_struct.C2S_TokenLogin{}, login.ChanRPC)
	data_struct.Processor.SetRouter(&data_struct.C2S_UsernamePasswordLogin{}, login.ChanRPC)

	data_struct.Processor.SetRouter(&data_struct.C2S_Heartbeat{}, game.ChanRPC)
	data_struct.Processor.SetRouter(&data_struct.C2S_SetYananConfig{}, game.ChanRPC)
	data_struct.Processor.SetRouter(&data_struct.C2S_SetSystemOn{}, game.ChanRPC)
	data_struct.Processor.SetRouter(&data_struct.C2S_SetUserRole{}, game.ChanRPC)
	data_struct.Processor.SetRouter(&data_struct.C2S_SetUsernamePassword{}, game.ChanRPC)
	data_struct.Processor.SetRouter(&data_struct.C2S_CreateYananRoom{}, game.ChanRPC)
	data_struct.Processor.SetRouter(&data_struct.C2S_EnterRoom{}, game.ChanRPC)
	data_struct.Processor.SetRouter(&data_struct.C2S_GetAllPlayers{}, game.ChanRPC)
	data_struct.Processor.SetRouter(&data_struct.C2S_Prepare{}, game.ChanRPC)
	data_struct.Processor.SetRouter(&data_struct.C2S_SetGun{}, game.ChanRPC)
	data_struct.Processor.SetRouter(&data_struct.C2S_ExitOrDisbandRoom{}, game.ChanRPC)
	data_struct.Processor.SetRouter(&data_struct.C2S_AgreeDisbandRoom{}, game.ChanRPC)
	data_struct.Processor.SetRouter(&data_struct.C2S_RefuseDisbandRoom{}, game.ChanRPC)
	data_struct.Processor.SetRouter(&data_struct.C2S_MahjongDiscard{}, game.ChanRPC)
	data_struct.Processor.SetRouter(&data_struct.C2S_MahjongWin{}, game.ChanRPC)
	data_struct.Processor.SetRouter(&data_struct.C2S_MahjongKong{}, game.ChanRPC)
	data_struct.Processor.SetRouter(&data_struct.C2S_MahjongPong{}, game.ChanRPC)
	data_struct.Processor.SetRouter(&data_struct.C2S_MahjongPass{}, game.ChanRPC)
	data_struct.Processor.SetRouter(&data_struct.C2S_GetRoomCards{}, game.ChanRPC)
	data_struct.Processor.SetRouter(&data_struct.C2S_TransferRoomCard{}, game.ChanRPC)
	data_struct.Processor.SetRouter(&data_struct.C2S_GetTotalResults{}, game.ChanRPC)
	data_struct.Processor.SetRouter(&data_struct.C2S_GetRoundResults{}, game.ChanRPC)
	data_struct.Processor.SetRouter(&data_struct.C2S_MahjongManaged{}, game.ChanRPC)
	data_struct.Processor.SetRouter(&data_struct.C2S_TextMessage{}, game.ChanRPC)
	data_struct.Processor.SetRouter(&data_struct.C2S_ExpressionMessage{}, game.ChanRPC)
	data_struct.Processor.SetRouter(&data_struct.C2S_CompleteDailyShare{}, game.ChanRPC)
	data_struct.Processor.SetRouter(&data_struct.C2S_GCloudVoiceMessage{}, game.ChanRPC)
	data_struct.Processor.SetRouter(&data_struct.C2S_GetYananIOSProductList{}, game.ChanRPC)
	data_struct.Processor.SetRouter(&data_struct.C2S_GetYananAndriodProductList{}, game.ChanRPC)
	data_struct.Processor.SetRouter(&data_struct.C2S_IAPReceiptData{}, game.ChanRPC)
	data_struct.Processor.SetRouter(&data_struct.C2S_StartYananMatching{}, game.ChanRPC)

	data_struct.Processor.SetRouter(&data_struct.C2S_GetTransferRoomCardRecord{}, game.ChanRPC)
	data_struct.Processor.SetRouter(&data_struct.C2S_GetAllTransferRoomCardRecord{}, game.ChanRPC)
	data_struct.Processor.SetRouter(&data_struct.C2S_GetAllAgentInfo{}, game.ChanRPC)
	data_struct.Processor.SetRouter(&data_struct.C2S_GetAllUserInfo{}, game.ChanRPC)
	data_struct.Processor.SetRouter(&data_struct.C2S_GetUserInfo{}, game.ChanRPC)
	data_struct.Processor.SetRouter(&data_struct.C2S_GetBlackList{}, game.ChanRPC)
	data_struct.Processor.SetRouter(&data_struct.C2S_GetRedPacketMatchRecord{}, game.ChanRPC)
	data_struct.Processor.SetRouter(&data_struct.C2S_TakeRedPacketMatchPrize{}, game.ChanRPC)
	data_struct.Processor.SetRouter(&data_struct.C2S_GetCircleLoginCode{}, game.ChanRPC)
	data_struct.Processor.SetRouter(&data_struct.C2S_FakeWXPay{}, game.ChanRPC)

	data_struct.Processor.SetRouter(&data_struct.C2S_SetRobotData{}, game.ChanRPC)
}
