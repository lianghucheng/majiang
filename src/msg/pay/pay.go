package pay

import (
	"msg"
)

func init() {
	msg.MsgRegister(&C2S_FakeWXPay{})
	msg.MsgRegister(&S2C_PayOK{})
}

type C2S_FakeWXPay struct {
	TotalFee int
}

// 购买S2C_PayOK.RoomCards 房卡成功
type S2C_PayOK struct {
	RoomCards int
}
