package gate

import (
	"config"
	"game"
	"msg"
	. "msg/agent"
	. "msg/card"
	. "msg/circle"
	. "msg/communicate"
	. "msg/config"
	. "msg/login"
	. "msg/pay"
	. "msg/redpacket"
	. "msg/room/mahjong"
	. "msg/system"
	. "msg/user"

	"github.com/name5566/leaf/gate"
)

type Module struct {
	*gate.Gate
}

func (m *Module) OnInit() {
	m.Gate = &gate.Gate{
		MaxConnNum:      conf.Server.MaxConnNum,
		PendingWriteNum: conf.PendingWriteNum,
		MaxMsgLen:       conf.MaxMsgLen,
		WSAddr:          conf.Server.WSAddr,
		HTTPTimeout:     conf.HTTPTimeout,
		CertFile:        conf.Server.CertFile,
		KeyFile:         conf.Server.KeyFile,
		TCPAddr:         conf.Server.TCPAddr,
		LenMsgLen:       conf.LenMsgLen,
		LittleEndian:    conf.LittleEndian,
		Processor:       msg.Processor,
		AgentChanRPC:    game.ChanRPC,
	}
}

var (
	ModuleGate = new(Module)
)

func GateInit() {
	msg.Processor.SetRouter(&C2S_WeChatLogin{}, game.ChanRPC)
	msg.Processor.SetRouter(&C2S_TokenLogin{}, game.ChanRPC)
	msg.Processor.SetRouter(&C2S_UsernamePasswordLogin{}, game.ChanRPC)

	msg.Processor.SetRouter(&C2S_Heartbeat{}, game.ChanRPC)
	msg.Processor.SetRouter(&C2S_SetGDConfig{}, game.ChanRPC)
	msg.Processor.SetRouter(&C2S_SetSystemOn{}, game.ChanRPC)
	msg.Processor.SetRouter(&C2S_SetUsernamePassword{}, game.ChanRPC)
	msg.Processor.SetRouter(&C2S_SetUserRole{}, game.ChanRPC)
	msg.Processor.SetRouter(&C2S_CreateGDRoom{}, game.ChanRPC)
	msg.Processor.SetRouter(&C2S_StartGDMatching{}, game.ChanRPC)
	msg.Processor.SetRouter(&C2S_EnterRoom{}, game.ChanRPC)
	msg.Processor.SetRouter(&C2S_GetAllPlayers{}, game.ChanRPC)
	msg.Processor.SetRouter(&C2S_ExitOrDisbandRoom{}, game.ChanRPC)
	msg.Processor.SetRouter(&C2S_Prepare{}, game.ChanRPC)
	msg.Processor.SetRouter(&C2S_MahjongDiscard{}, game.ChanRPC)
	msg.Processor.SetRouter(&C2S_MahjongWin{}, game.ChanRPC)
	msg.Processor.SetRouter(&C2S_MahjongKong{}, game.ChanRPC)
	msg.Processor.SetRouter(&C2S_MahjongPong{}, game.ChanRPC)
	msg.Processor.SetRouter(&C2S_MahjongChow{}, game.ChanRPC)
	msg.Processor.SetRouter(&C2S_MahjongPass{}, game.ChanRPC)
	msg.Processor.SetRouter(&C2S_AgreeDisbandRoom{}, game.ChanRPC)
	msg.Processor.SetRouter(&C2S_RefuseDisbandRoom{}, game.ChanRPC)
	msg.Processor.SetRouter(&C2S_GetRoomCards{}, game.ChanRPC)
	msg.Processor.SetRouter(&C2S_TransferRoomCard{}, game.ChanRPC)
	msg.Processor.SetRouter(&C2S_GetTotalResults{}, game.ChanRPC)
	msg.Processor.SetRouter(&C2S_GetRoundResults{}, game.ChanRPC)
	msg.Processor.SetRouter(&C2S_CompleteDailyShare{}, game.ChanRPC)
	msg.Processor.SetRouter(&C2S_TextMessage{}, game.ChanRPC)
	msg.Processor.SetRouter(&C2S_ExpressionMessage{}, game.ChanRPC)
	msg.Processor.SetRouter(&C2S_GCloudVoiceMessage{}, game.ChanRPC)
	msg.Processor.SetRouter(&C2S_GetGDAndroidProductList{}, game.ChanRPC)
	msg.Processor.SetRouter(&C2S_GetGDIOSProductList{}, game.ChanRPC)
	msg.Processor.SetRouter(&C2S_IAPReceiptData{}, game.ChanRPC)
	msg.Processor.SetRouter(&C2S_GetUserInfo{}, game.ChanRPC)
	msg.Processor.SetRouter(&C2S_GetTransferRoomCardRecord{}, game.ChanRPC)
	msg.Processor.SetRouter(&C2S_GetAllTransferRoomCardRecord{}, game.ChanRPC)
	msg.Processor.SetRouter(&C2S_GetAllAgentInfo{}, game.ChanRPC)
	msg.Processor.SetRouter(&C2S_GetAllUserInfo{}, game.ChanRPC)
	msg.Processor.SetRouter(&C2S_GetBlackList{}, game.ChanRPC)
	msg.Processor.SetRouter(&C2S_GetRedPacketMatchRecord{}, game.ChanRPC)
	msg.Processor.SetRouter(&C2S_TakeRedPacketMatchPrize{}, game.ChanRPC)
	msg.Processor.SetRouter(&C2S_MahjongManaged{}, game.ChanRPC)
	msg.Processor.SetRouter(&C2S_FakeWXPay{}, game.ChanRPC)
	msg.Processor.SetRouter(&C2S_GetCircleLoginCode{}, game.ChanRPC)

	msg.Processor.SetRouter(&C2S_SetRobotData{}, game.ChanRPC)
}
