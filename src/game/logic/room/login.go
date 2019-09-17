package room

import (
	. "db"
	"game"
	"game/config"
	"game/player"
	"game/room"
	"msg/communicate"
	msg "msg/login"
	"reflect"
	"strings"
	"time"
	"util"

	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type AgentInfo struct {
	*player.User
}

type LoginControl struct {
	*player.User
}

func init() {
	game.HandleRegister("NewAgent", rpcNewAgent)
	game.HandleRegister(msgReflect(&msg.C2S_WeChatLogin{}), weChatLogin)
	game.HandleRegister(msgReflect(&msg.C2S_TokenLogin{}), tokenLogin)
	game.HandleRegister(msgReflect(&msg.C2S_UsernamePasswordLogin{}), usernamePasswordLogin)
}
func msgReflect(i interface{}) reflect.Type {
	return reflect.TypeOf(i)
}
func rpcNewAgent(args []interface{}) {
	a := args[0].(gate.Agent)
	a.SetUserData(new(AgentInfo))
	log.Debug("建立新的连接")
	game.Skeleton.AfterFunc(3*time.Second, func() {
		if a.UserData() != nil {
			agentInfo := a.UserData().(*AgentInfo)
			if agentInfo != nil && agentInfo.User == nil {
				a.Close()
			}
		}
	})
}

func weChatLogin(args []interface{}) {
	a := args[1].(gate.Agent)
	m := args[0].(*msg.C2S_WeChatLogin)

	if a.UserData() == nil || a.UserData().(*AgentInfo).User != nil {
		return
	}
	if strings.TrimSpace(m.Unionid) == "" {
		a.WriteMsg(&msg.S2C_Close{
			Error: msg.S2C_Close_UnionIDInvalid,
		})
		a.Close()
		return
	}
	if !player.SystemOn && m.Unionid != "o8c-nt6tO8aIBNPoxvXOQTVJUxY0" {
		a.WriteMsg(&msg.S2C_Close{
			Error: msg.S2C_Close_SystemOff,
		})
		a.Close()
		return
	}
	newUser := player.InitUser(a)
	a.UserData().(*AgentInfo).User = newUser
	ctx := new(LoginControl)
	ctx.User = newUser
	ctx.weChatLogin(m)
}

func (ctx *LoginControl) weChatLogin(info *msg.C2S_WeChatLogin) {
	userData := new(player.UserData)
	firstLogin := false
	game.Skeleton.Go(func() {
		db := MongoDB.Ref()
		defer MongoDB.UnRef(db)

		err := db.DB(DB).C("users").Find(bson.M{"unionid": info.Unionid}).One(userData)
		if err == nil {
			return
		}
		// unknow error
		if err == mgo.ErrNotFound {
			firstLogin = true
		} else {
			log.Error("load unionid %v data error: %v", info.Unionid, err)
			userData = nil
			ctx.WriteMsg(&msg.S2C_Close{
				Error: msg.S2C_Close_InnerError,
			})
			ctx.Close()
			return
		}
		// new
		err = userData.InitValue()
		if err != nil {
			log.Error("load unionid %v data error: %v", info.Unionid, err)
			userData = nil
			ctx.WriteMsg(&msg.S2C_Close{
				Error: msg.S2C_Close_InnerError,
			})
			ctx.Close()
			return
		}
	}, func() {
		if userData == nil || ctx.State == player.UserLogout {
			return
		}
		if userData.Role == player.RoleBlack {
			ctx.WriteMsg(&msg.S2C_Close{
				Error:        msg.S2C_Close_RoleBlack,
				WeChatNumber: conf.GdConfigData.WeChatNumber,
			})
			ctx.Close()
			return
		}
		anotherLogin := false
		if person := player.GetPersonMgr().GetPerson(userData.UserID); person != nil {
			log.Debug("**************************:", "已经登录过")
			if person.UserData.Serial != info.Serial {
				anotherLogin = true
			}
			person.WriteMsg(&msg.S2C_Close{
				Error: msg.S2C_Close_LoginRepeated,
			})
			person.Close()
			if person == ctx.User {
				return
			}
			ctx.UserData = person.UserData
			userData = person.UserData
		}
		player.GetPersonMgr().AddPerson(userData.UserID, ctx.User)
		userData.UpdateWeChatInfo(info)
		ctx.UserData = userData
		ctx.onLogin(firstLogin, anotherLogin)
		if firstLogin {
			log.Debug("userID: %v WeChat首次登录 unionid: %v, 在线人数: %v", userData.UserID, userData.Unionid, len(player.GetPersonMgr().MapPerson))
		} else {
			log.Debug("userID: %v WeChat登录 unionid: %v, 在线人数: %v", userData.UserID, userData.Unionid, len(player.GetPersonMgr().MapPerson))
		}
	})
}

func tokenLogin(args []interface{}) {
	a := args[1].(gate.Agent)
	m := args[0].(*msg.C2S_TokenLogin)
	if a.UserData() == nil || a.UserData().(*AgentInfo).User != nil {
		return
	}
	if strings.TrimSpace(m.Token) == "" {
		a.WriteMsg(&msg.S2C_Close{
			Error: msg.S2C_Close_TokenInvalid,
		})
		a.Close()
		return
	}
	if !player.SystemOn {
		a.WriteMsg(&msg.S2C_Close{
			Error: msg.S2C_Close_InnerError,
		})
		a.Close()
		return
	}
	newUser := player.InitUser(a)
	a.UserData().(*AgentInfo).User = newUser
	ctx := new(LoginControl)
	ctx.User = newUser
	ctx.tokenLogin(m.Token)
}
func (ctx *LoginControl) tokenLogin(token string) {
	userData := new(player.UserData)
	game.Skeleton.Go(func() {
		db := MongoDB.Ref()
		defer MongoDB.UnRef(db)

		err := db.DB(DB).C("users").Find(bson.M{"token": token, "expireat": bson.M{"$gt": time.Now().Unix()}}).One(userData)
		log.Debug("%v:", userData)
		if err != nil {
			log.Error("find token %v error: %v", token, err)
			ctx.WriteMsg(&msg.S2C_Close{
				Error: msg.S2C_Close_TokenInvalid,
			})
			userData = nil
			ctx.Close()
		}
	}, func() {
		if userData == nil || ctx.State == player.UserLogout {
			return
		}
		if userData.Role == player.RoleBlack {
			ctx.WriteMsg(&msg.S2C_Close{
				Error:        msg.S2C_Close_RoleBlack,
				WeChatNumber: conf.GdConfigData.WeChatNumber,
			})
			ctx.Close()
			return
		}

		if person := player.GetPersonMgr().GetPerson(userData.UserID); person != nil {
			person.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_LoginRepeated})
			person.Close()
			log.Debug("userID: %v 重复登录", userData.UserID)
			if person == ctx.User {
				return
			}
			userData = person.UserData
		}
		ctx.UserData = userData
		player.GetPersonMgr().AddPerson(userData.UserID, ctx.User)
		ctx.onLogin(false, false)
		log.Debug("userID: %v Token登录, 在线人数: %v", userData.UserID, len(player.GetPersonMgr().MapPerson))
	})
}

func usernamePasswordLogin(args []interface{}) {
	a := args[1].(gate.Agent)
	m := args[0].(*msg.C2S_UsernamePasswordLogin)
	if a.UserData() == nil || a.UserData().(*AgentInfo).User != nil {
		return
	}
	if strings.TrimSpace(m.Username) == "" {
		a.WriteMsg(&msg.S2C_Close{
			Error: msg.S2C_Close_UsernameInvalid,
		})
		a.Close()
		return
	}
	if !player.SystemOn {
		a.WriteMsg(&msg.S2C_Close{
			Error: msg.S2C_Close_InnerError,
		})
		a.Close()
		return
	}
	newUser := player.InitUser(a)
	a.UserData().(*AgentInfo).User = newUser

}
func (ctx *LoginControl) usernamePasswordLogin(username string, password string) {
	userData := new(player.UserData)
	game.Skeleton.Go(func() {
		db := MongoDB.Ref()
		defer MongoDB.UnRef(db)

		err := db.DB(DB).C("users").Find(bson.M{"username": username, "password": password}).One(userData)
		if err != nil {
			log.Error("用户名: %v 密码 不正确: %v", username, err)
			userData = nil
			ctx.WriteMsg(&msg.S2C_Close{
				Error: msg.S2C_Close_UsernameInvalid,
			})
			ctx.Close()
		}
	}, func() {
		if userData == nil || ctx.State == player.UserLogout {
			return
		}
		if ctx.UserData.Role == player.RoleBlack {
			ctx.WriteMsg(&msg.S2C_Close{
				Error:        msg.S2C_Close_RoleBlack,
				WeChatNumber: conf.GdConfigData.WeChatNumber,
			})
			ctx.Close()
			return
		}
		if person := player.GetPersonMgr().GetPerson(userData.UserID); person != nil {
			person.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_LoginRepeated})
			person.Close()
			log.Debug("userID: %v 重复登录", userData.UserID)
			if person == ctx.User {
				return
			}
			ctx.UserData = person.UserData
			userData = person.UserData
		}
		player.GetPersonMgr().AddPerson(userData.UserID, ctx.User)
		ctx.onLogin(false, false)
		log.Debug("用户名: %v 密码登录", username)
	})
}
func (ctx *LoginControl) onLogin(firstLogin bool, anotherLogin bool) {
	if ctx.UserData.Role != player.RoleRobot {
		ctx.UserData.LoginIP = strings.Split(ctx.RemoteAddr().String(), ":")[0]
		ctx.UserData.Token = util.GetToken(32)
		ctx.UserData.ExpireAt = time.Now().Add(2 * time.Hour).Unix()
	}
	ctx.UserData.LastLoginAt = time.Now().Unix()
	if firstLogin {
		player.SaveUserData(ctx.UserData)
	} else {
		player.UpdateUserData(ctx.UserData.UserID, bson.M{"$set": bson.M{"token": ctx.UserData.Token}})
	}
	joinAgencyTime := ""
	if ctx.UserData.JoinAgencyAt > 1509465600 {
		joinAgencyTime = time.Unix(ctx.UserData.JoinAgencyAt, 0).Format("2006/01/02 15:04:05")
	}
	ctx.autoHeartbeat()
	ctx.WriteMsg(&msg.S2C_Login{
		AccountID:     ctx.UserData.AccountID,
		Nickname:      ctx.UserData.Nickname,
		JoinAgencyAT:  joinAgencyTime,
		SaleRoomCards: ctx.UserData.SaleRoomCards,
		Headimgurl:    ctx.UserData.Headimgurl,
		Sex:           ctx.UserData.Sex,
		RoomCards:     ctx.UserData.RoomCards,
		Role:          ctx.UserData.Role,
		Token:         ctx.UserData.Token,
		AnotherLogin:  anotherLogin,
		AnotherRoom:   room.GetRoomMgr().GetRoom(ctx.UserData.UserID) != nil,
		Notice:        conf.GdConfigData.Notice,
		Radio:         conf.GdConfigData.Radio,
		WeChatNumber:  conf.GdConfigData.WeChatNumber,
	})
	//user.requestCircleID()
}

func (ctx *LoginControl) autoHeartbeat() {
	if ctx.HeartbeatStop {
		log.Debug("userID: %v 心跳停止", ctx.UserData.UserID)
		ctx.Close()
		return
	}
	ctx.HeartbeatStop = true
	ctx.WriteMsg(&communicate.S2C_Heartbeat{})
	// 服务端发送心跳包间隔120秒
	ctx.HeartbeatTimer = game.Skeleton.AfterFunc(120*time.Second, func() {
		ctx.autoHeartbeat()
	})
}
