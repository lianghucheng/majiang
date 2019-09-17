package internal

import (
	"github.com/name5566/leaf/gate"
	"strings"
	"yananmj-server/msg"
)

type AgentInfo struct {
	user *User
}

func init() {
	skeleton.RegisterChanRPC("NewAgent", rpcNewAgent)
	skeleton.RegisterChanRPC("CloseAgent", rpcCloseAgent)
	skeleton.RegisterChanRPC("WeChatLogin", rpcWeChatLogin)
	skeleton.RegisterChanRPC("TokenLogin", rpcTokenLogin)
	skeleton.RegisterChanRPC("UsernamePasswordLogin", rpcUsernamePasswordLogin)
}

func rpcNewAgent(args []interface{}) {
	a := args[0].(gate.Agent)
	a.SetUserData(new(AgentInfo))
}

//微信登录
func rpcWeChatLogin(args []interface{}) {
	a := args[0].(gate.Agent)
	m := args[1].(*data_struct.C2S_WeChatLogin)

	if a.UserData() == nil || a.UserData().(*AgentInfo).user != nil {
		return
	}

	if strings.TrimSpace(m.Unionid) == "" {
		a.WriteMsg(&data_struct.S2C_Close{Error: data_struct.S2C_Close_UnionIDInvalid})
		a.Close()
		return
	}

	if !systemOn && m.Unionid != "o8c-nt6tO8aIBNPoxvXOQTVJUxY0" {
		a.WriteMsg(&data_struct.S2C_Close{Error: data_struct.S2C_Close_SystemOff})
		a.Close()
		return
	}

	newUser := newUser(a)
	a.UserData().(*AgentInfo).user = newUser
	newUser.weChatLogin(m)
}

//Token登录
func rpcTokenLogin(args []interface{}) {
	a := args[0].(gate.Agent)
	m := args[1].(*data_struct.C2S_TokenLogin)

	if a.UserData() == nil || a.UserData().(*AgentInfo).user != nil {
		return
	}

	if strings.TrimSpace(m.Token) == "" {
		a.WriteMsg(&data_struct.S2C_Close{Error: data_struct.S2C_Close_TokenInvalid})
		a.Close()
		return
	}

	if !systemOn {
		a.WriteMsg(&data_struct.S2C_Close{Error: data_struct.S2C_Close_SystemOff})
		a.Close()
		return
	}
	newUser := newUser(a)
	a.UserData().(*AgentInfo).user = newUser
	newUser.tokenLogin(m.Token)
}

func rpcUsernamePasswordLogin(args []interface{}) {
	a := args[0].(gate.Agent)
	m := args[1].(*data_struct.C2S_UsernamePasswordLogin)

	if a.UserData() == nil || a.UserData().(*AgentInfo).user != nil {
		return
	}

	if strings.TrimSpace(m.Username) == "" {
		a.WriteMsg(&data_struct.S2C_Close{Error: data_struct.S2C_Close_UsernameInvalid})
		a.Close()
		return
	}
	if !systemOn {
		a.WriteMsg(&data_struct.S2C_Close{Error: data_struct.S2C_Close_SystemOff})
		a.Close()
		return
	}
	newUser := newUser(a)
	a.UserData().(*AgentInfo).user = newUser
	newUser.usernamePasswordLogin(m.Username, m.Password)
}

func rpcCloseAgent(args []interface{}) {
	a := args[0].(gate.Agent)

	user := a.UserData().(*AgentInfo).user
	a.SetUserData(nil)
	if user == nil {
		return
	}

	if user.state == userLogin {
		user.state = userLogout
		user.logout()
	}
}
