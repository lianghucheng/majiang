package internal

import (
	"gdmj-server/game"
	"gdmj-server/msg"
	"github.com/name5566/leaf/gate"
	"reflect"
)

func handleMsg(m interface{}, h interface{}) {
	skeleton.RegisterChanRPC(reflect.TypeOf(m), h)
}

func init() {
	handleMsg(&msg.C2S_WeChatLogin{}, handleC2SWeChatLogin)
	handleMsg(&msg.C2S_TokenLogin{}, handleC2STokenLogin)
	handleMsg(&msg.C2S_UsernamePasswordLogin{}, handleUsernamePasswordLogin)
}

func handleC2SWeChatLogin(args []interface{}) {
	// 收到的 C2S_WeChatLogin 消息
	m := args[0].(*msg.C2S_WeChatLogin)
	// 消息的发送者
	a := args[1].(gate.Agent)
	// login
	game.ChanRPC.Go("WeChatLogin", a, m)
}

func handleC2STokenLogin(args []interface{}) {
	m := args[0].(*msg.C2S_TokenLogin)
	a := args[1].(gate.Agent)
	game.ChanRPC.Go("TokenLogin", a, m)
}

func handleUsernamePasswordLogin(args []interface{}) {
	m := args[0].(*msg.C2S_UsernamePasswordLogin)
	a := args[1].(gate.Agent)
	game.ChanRPC.Go("UsernamePasswordLogin", a, m)
}
