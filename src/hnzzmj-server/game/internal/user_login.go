package internal

import (
	"github.com/name5566/leaf/log"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"hnzzmj-server/common"
	"hnzzmj-server/msg"
	"strings"
	"time"
)

func (user *User) wechatLogin(info *msg.C2S_WeChatLogin) {
	userData := new(UserData)
	firstLogin := false
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		// load
		err := db.DB(DB).C("users").Find(bson.M{"unionid": info.Unionid}).One(userData)
		if err == nil {
			return
		}
		// unknown error
		if err == mgo.ErrNotFound {
			firstLogin = true
		} else {
			log.Error("load unionid %v data error: %v", info.Unionid, err)
			userData = nil
			user.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_InnerError})
			user.Close()
			return
		}
		// new
		err = userData.initValue()
		if err != nil {
			log.Error("load unionid %v data error: %v", info.Unionid, err)
			userData = nil
			user.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_InnerError})
			user.Close()
			return
		}
	}, func() {
		if userData == nil || user.state == userLogout {
			return
		}
		if userData.Role == -1 {
			user.WriteMsg(&msg.S2C_Close{
				Error:        msg.S2C_Close_RoleBlack,
				WeChatNumber: hnzzConfigData.WeChatNumber,
			})
			user.Close()
			return
		}
		anotherLogin := false
		if oldUser, ok := userIDUsers[userData.UserID]; ok {
			if oldUser.data.userData.Serial != info.Serial {
				anotherLogin = true
			}
			oldUser.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_LoginRepeated})
			oldUser.Close()
			log.Debug("userID: %v 重复登录", userData.UserID)
			if oldUser == user {
				return
			}
			user.data = oldUser.data
			userData = oldUser.data.userData
		}
		userIDUsers[userData.UserID] = user
		userData.updateWeChatInfo(info)
		user.data.userData = userData
		user.onLogin(firstLogin, anotherLogin)
		if firstLogin {
			log.Debug("userID: %v WeChat首次登录 unionid: %v, 在线人数: %v", user.data.userData.UserID, user.data.userData.Unionid, len(userIDUsers))
		} else {
			log.Debug("userID: %v WeChat登录 unionid: %v, 在线人数: %v", user.data.userData.UserID, user.data.userData.Unionid, len(userIDUsers))
		}
	})
}

func (user *User) tokenLogin(token string) {
	userData := new(UserData)
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)

		err := db.DB(DB).C("users").Find(bson.M{"token": token}).One(userData)
		if err != nil {
			log.Debug("find token %v error: %v", token, err)
			userData = nil
			user.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_TokenInvalid})
			user.Close()
		}
	}, func() {
		if userData == nil || user.state == userLogout {
			return
		}
		if userData.Role == -1 {
			user.WriteMsg(&msg.S2C_Close{
				Error:        msg.S2C_Close_RoleBlack,
				WeChatNumber: hnzzConfigData.WeChatNumber,
			})
			user.Close()
			return
		}
		if oldUser, ok := userIDUsers[userData.UserID]; ok {
			oldUser.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_LoginRepeated})
			oldUser.Close()
			log.Debug("userID: %v 重复登录", userData.UserID)
			if oldUser == user {
				return
			}
			user.data = oldUser.data
			userData = oldUser.data.userData
		}
		userIDUsers[userData.UserID] = user
		user.data.userData = userData
		user.onLogin(false, false)
		log.Debug("userID: %v Token登录, 在线人数: %v", userData.UserID, len(userIDUsers))
	})
}

func (user *User) usernamePasswordLogin(username string, password string) {
	userData := new(UserData)
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		// load
		err := db.DB(DB).C("users").Find(bson.M{"username": username, "password": password}).One(userData)
		if err != nil {
			log.Error("用户名: %v, 密码不正确: %v", username, err)
			userData = nil
			user.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_UsernameInvalid})
			user.Close()
		}
	}, func() {
		if userData == nil || user.state == userLogout {
			return
		}
		if userData.Role == -1 {
			user.WriteMsg(&msg.S2C_Close{
				Error:        msg.S2C_Close_RoleBlack,
				WeChatNumber: hnzzConfigData.WeChatNumber,
			})
			user.Close()
			return
		}
		if oldUser, ok := userIDUsers[userData.UserID]; ok {
			oldUser.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_LoginRepeated})
			oldUser.Close()
			log.Debug("userID: %v 重复登录", userData.UserID)
			if oldUser == user {
				return
			}
			user.data = oldUser.data
			userData = oldUser.data.userData
		}
		userIDUsers[userData.UserID] = user
		user.data.userData = userData
		user.onLogin(false, false)
		log.Debug("用户名: %v 密码登录", username)
	})
}

func (user *User) logout() {
	if user.heartbeatTimer != nil {
		user.heartbeatTimer.Stop()
	}
	if user.data == nil {
		return
	}
	if existUser, ok := userIDUsers[user.data.userData.UserID]; ok {
		if existUser == user {
			log.Debug("userID: %v 登出", user.data.userData.UserID)
			user.onLogout()
			delete(userIDUsers, user.data.userData.UserID)
			saveUserData(user.data.userData)
		}
	}
}

func (user *User) onLogin(firstLogin bool, anotherLogin bool) {
	if !user.isRobot() {
		user.data.userData.LoginIP = strings.Split(user.RemoteAddr().String(), ":")[0]
		user.data.userData.Token = common.GetToken(32)
		user.data.userData.LastLoginAt = time.Now().Unix()
	}
	if firstLogin {
		saveUserData(user.data.userData)
	} else {
		updateUserData(user.data.userData.UserID, bson.M{"$set": bson.M{"token": user.data.userData.Token}})
	}
	joinAgencyTime := ""
	if user.data.userData.JoinAgencyAt > 1509465600 {
		joinAgencyTime = time.Unix(user.data.userData.JoinAgencyAt, 0).Format("2006/01/02 15:04:05")
	}
	user.autoHeartbeat()
	user.WriteMsg(&msg.S2C_Login{
		AccountID:          user.data.userData.AccountID,
		Nickname:           user.data.userData.Nickname,
		Headimgurl:         user.data.userData.Headimgurl,
		Sex:                user.data.userData.Sex,
		RoomCards:          user.data.userData.RoomCards,
		Role:               user.data.userData.Role,
		Token:              user.data.userData.Token,
		AnotherLogin:       anotherLogin,
		AnotherRoom:        userIDRooms[user.data.userData.UserID] != nil,
		Notice:             hnzzConfigData.Notice,
		Radio:              hnzzConfigData.Radio,
		WeChatNumber:       hnzzConfigData.WeChatNumber,
		SaleRoomCardNumber: user.data.userData.SaleRoomCards,
		JoinAgencyAT:       joinAgencyTime,
	})
	user.requestCircleID()
}

func (user *User) onLogout() {
	if r, ok := userIDRooms[user.data.userData.UserID]; ok {
		user.exitOrDisbandRoom(r, false)
	}
}
