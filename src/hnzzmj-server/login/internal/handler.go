package internal

import (
	"github.com/name5566/leaf/gate"
	"reflect"
	"hnzzmj-server/game"
	"hnzzmj-server/msg"
)

func init() {
	handler(&msg.C2S_WeChatLogin{}, handleC2SWeChatLogin)
	handler(&msg.C2S_TokenLogin{}, handleC2STokenLogin)
	handler(&msg.C2S_UsernamePasswordLogin{}, handleC2SUsernamePasswordLogin)
}

func handler(m interface{}, h interface{}) {
	skeleton.RegisterChanRPC(reflect.TypeOf(m), h)
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
	// 收到的 C2S_TokenLogin 消息
	m := args[0].(*msg.C2S_TokenLogin)
	// 消息的发送者
	a := args[1].(gate.Agent)
	// login
	game.ChanRPC.Go("TokenLogin", a, m)
}

func handleC2SUsernamePasswordLogin(args []interface{}) {
	m := args[0].(*msg.C2S_UsernamePasswordLogin)
	a := args[1].(gate.Agent)
	// login
	game.ChanRPC.Go("UsernamePasswordLogin", a, m)
}
