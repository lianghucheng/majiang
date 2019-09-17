package internal

import (
	"github.com/name5566/leaf/gate"
	"hnzzmj-server/msg"
	"strings"
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

func rpcWeChatLogin(args []interface{}) {
	a := args[0].(gate.Agent)
	m := args[1].(*msg.C2S_WeChatLogin)
	// network closed
	if a.UserData() == nil || a.UserData().(*AgentInfo).user != nil {
		return
	}
	if strings.TrimSpace(m.Unionid) == "" {
		a.WriteMsg(&msg.S2C_Close{
			Error: msg.S2C_Close_UnionIDInvalid,
		})
		a.Close()
		return
	}
	if !systemOn && m.Unionid != "o8c-nt6tO8aIBNPoxvXOQTVJUxY0" {
		a.WriteMsg(&msg.S2C_Close{
			Error: msg.S2C_Close_SystemOff,
		})
		a.Close()
		return
	}
	newUser := newUser(a)
	a.UserData().(*AgentInfo).user = newUser
	newUser.wechatLogin(m)
}

func rpcTokenLogin(args []interface{}) {
	a := args[0].(gate.Agent)
	m := args[1].(*msg.C2S_TokenLogin)
	// network closed
	if a.UserData() == nil || a.UserData().(*AgentInfo).user != nil {
		return
	}
	if strings.TrimSpace(m.Token) == "" {
		a.WriteMsg(&msg.S2C_Close{
			Error: msg.S2C_Close_TokenInvalid,
		})
		a.Close()
		return
	}
	if !systemOn {
		a.WriteMsg(&msg.S2C_Close{
			Error: msg.S2C_Close_SystemOff,
		})
		a.Close()
		return
	}
	newUser := newUser(a)
	a.UserData().(*AgentInfo).user = newUser
	newUser.tokenLogin(m.Token)
}

func rpcUsernamePasswordLogin(args []interface{}) {
	a := args[0].(gate.Agent)
	m := args[1].(*msg.C2S_UsernamePasswordLogin)
	// network closed
	if a.UserData() == nil || a.UserData().(*AgentInfo).user != nil {
		return
	}
	if strings.TrimSpace(m.Username) == "" {
		a.WriteMsg(&msg.S2C_Close{
			Error: msg.S2C_Close_UsernameInvalid,
		})
		a.Close()
		return
	}
	if !systemOn {
		a.WriteMsg(&msg.S2C_Close{
			Error: msg.S2C_Close_SystemOff,
		})
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
