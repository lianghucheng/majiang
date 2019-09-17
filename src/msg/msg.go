package msg

import (
	"github.com/name5566/leaf/network/json"
)

var Processor = json.NewProcessor()
var msgs []interface{}

func MsgRegisterInit() {
	for key := range msgs {
		Processor.Register(msgs[key])
	}
}
func MsgRegister(msg interface{}) {
	msgs = append(msgs, msg)
}

//代理信息
