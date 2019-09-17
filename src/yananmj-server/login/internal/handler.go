package internal

import (
	"github.com/name5566/leaf/gate"
	"reflect"
	"yananmj-server/game"
	"yananmj-server/msg"
)

func handleMsg(m interface{}, h interface{}) {
	skeleton.RegisterChanRPC(reflect.TypeOf(m), h)
}

func init() {
	handleMsg(&data_struct.C2S_WeChatLogin{}, handleC2SWeChatLogin)
	handleMsg(&data_struct.C2S_TokenLogin{}, handleC2STokenLogin)
	handleMsg(&data_struct.C2S_UsernamePasswordLogin{}, handleC2SUsernamePasswordLogin)
}

//微信登录
func handleC2SWeChatLogin(args []interface{}) {
	//获取发送的消息
	m := args[0].(*data_struct.C2S_WeChatLogin)
	//消息的发送者
	a := args[1].(gate.Agent)
	game.ChanRPC.Go("WeChatLogin", a, m)
}

//Token登录
func handleC2STokenLogin(args []interface{}) {
	m := args[0].(*data_struct.C2S_TokenLogin)
	a := args[1].(gate.Agent)
	game.ChanRPC.Go("TokenLogin", a, m)
}

func handleC2SUsernamePasswordLogin(args []interface{}) {
	m := args[0].(*data_struct.C2S_UsernamePasswordLogin)
	a := args[1].(gate.Agent)
	game.ChanRPC.Go("UsernamePasswordLogin", a, m)
}
