package communicate

import (
	"msg"
)

func init() {
	msg.MsgRegister(&C2S_Heartbeat{})
	msg.MsgRegister(&S2C_Heartbeat{})
	msg.MsgRegister(&C2S_CompleteDailyShare{})
	msg.MsgRegister(&S2C_CompleteDailyShare{})
	msg.MsgRegister(&C2S_TextMessage{})
	msg.MsgRegister(&S2C_TextMessage{})
	msg.MsgRegister(&C2S_ExpressionMessage{})
	msg.MsgRegister(&S2C_ExpressionMessage{})
	msg.MsgRegister(&C2S_GCloudVoiceMessage{})
	msg.MsgRegister(&S2C_GCloudVoiceMessage{})
}

type C2S_Heartbeat struct{}

type S2C_Heartbeat struct{}

type C2S_CompleteDailyShare struct{}

type S2C_CompleteDailyShare struct {
	RoomCards int
}

type C2S_TextMessage struct {
	Message string
}

type S2C_TextMessage struct {
	Position int
	Message  string
}

type C2S_ExpressionMessage struct {
	Expression int
}

type S2C_ExpressionMessage struct {
	Position   int
	Expression int
}

type C2S_GCloudVoiceMessage struct {
	FileID string
}

type S2C_GCloudVoiceMessage struct {
	Position int
	FileID   string
}
