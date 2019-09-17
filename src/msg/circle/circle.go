package circle

import (
	"msg"
)

func init() {
	msg.MsgRegister(&C2S_GetCircleLoginCode{})
	msg.MsgRegister(&S2C_UpdateCircleLoginCode{})
}

type C2S_GetCircleLoginCode struct{}

const (
	S2C_UpdateCircleLoginCode_OK    = 0
	S2C_UpdateCircleLoginCode_Error = 1 // 圈圈授权出错，请稍后重试
)

type S2C_UpdateCircleLoginCode struct {
	Error     int
	LoginCode string
}
