package internal

import (
	"github.com/name5566/leaf/log"
	"gopkg.in/mgo.v2/bson"
	url2 "net/url"
	"time"
	"yananmj-server/msg"
)

var (
	yananIOSProductInfos = []data_struct.ProductInfo{
		{ID: "", Desc: "房卡48张", Price: 6},
		{ID: "", Desc: "房卡240张", Price: 30},
		{ID: "", Desc: "房卡500张", Price: 60},
	}

	yananAndriodProductInfos = []data_struct.ProductInfo{
		{Desc: "房卡48张", Price: 6},
		{Desc: "房卡240张", Price: 30},
		{Desc: "房卡500张", Price: 60},
	}
)

//设置安卓版本
func (user *User) setAndriodVersion(version int) {
	if version <= yananConfigData.AndriodVersion && user.data.userData.Role < roleRoot {
		log.Debug("设置陕西安卓版本无效 ：%v", version)
		user.WriteMsg(&data_struct.S2C_SetYananConfig{
			Error:          data_struct.S2C_SetYananConfig_VersionInvalid,
			AndriodVersion: version,
		})
		return
	}
	yananConfigData.AndriodVersion = version
	saveConfigData(yananConfigData)
	user.WriteMsg(&data_struct.S2C_SetYananConfig{
		Error:          data_struct.S2C_SetYananConfig_Ok,
		AndriodVersion: version,
	})
	log.Debug("userID:%v 设置银滩延安安卓版本:%v 成功", user.data.userData.UserID, version)
}

//设置安卓下载地址
func (user *User) setAndriodDownloadUrl(downloadUrl string) {
	url, err := url2.Parse(downloadUrl)
	if err == nil && url.Scheme == "" || err != nil || downloadUrl == yananConfigData.AndriodDownloadUrl {
		log.Debug("设置银滩延安安卓下载地址：%v 无效", downloadUrl)
		user.WriteMsg(&data_struct.S2C_SetYananConfig{
			Error:              data_struct.S2C_SetYananConfig_DownloadUrlInvalid,
			AndriodDownloadUrl: downloadUrl,
		})
		return
	}
	yananConfigData.AndriodDownloadUrl = downloadUrl
	saveConfigData(yananConfigData)
	user.WriteMsg(&data_struct.S2C_SetYananConfig{
		Error:              data_struct.S2C_SetYananConfig_Ok,
		AndriodDownloadUrl: downloadUrl,
	})
	log.Debug("设置银滩延安安卓下载地址:%v 成功", downloadUrl)
}

//设置ios版本
func (user *User) setIOSVersion(version int) {
	if version <= yananConfigData.IOSVersion && user.data.userData.Role < roleRoot {
		log.Debug("设置银滩延安iOS版本：%v 无效", version)
		user.WriteMsg(&data_struct.S2C_SetYananConfig{
			Error:      data_struct.S2C_SetYananConfig_VersionInvalid,
			IOSVersion: version,
		})
		return
	}
	yananConfigData.IOSVersion = version
	saveConfigData(yananConfigData)
	user.WriteMsg(&data_struct.S2C_SetYananConfig{
		Error:      data_struct.S2C_SetYananConfig_Ok,
		IOSVersion: version,
	})
	log.Debug("userID:%v 设置银滩延安iOS版本：%v 成功", user.data.userData.UserID, version)
}

//设置ios下载地址
func (user *User) setIOSDownloadUrl(downloadUrl string) {
	url, err := url2.Parse(downloadUrl)
	if err == nil && url.Scheme == "" || err != nil || downloadUrl == yananConfigData.IOSDownloadUrl {
		log.Debug("设置银滩延安iOS下载地址：%v 无效", downloadUrl)
		user.WriteMsg(&data_struct.S2C_SetYananConfig{
			Error:          data_struct.S2C_SetYananConfig_DownloadUrlInvalid,
			IOSDownloadUrl: downloadUrl,
		})
		return
	}

	yananConfigData.IOSDownloadUrl = downloadUrl
	saveConfigData(yananConfigData)
	user.WriteMsg(&data_struct.S2C_SetYananConfig{
		Error:          data_struct.S2C_SetYananConfig_Ok,
		IOSDownloadUrl: downloadUrl,
	})
	log.Debug("userID:%v 设置银滩延安iOS下载地址：%v 成功", user.data.userData.UserID, downloadUrl)
}

//设置公告
func (user *User) setNotice(notice string) {
	yananConfigData.Notice = notice
	saveConfigData(yananConfigData)
	user.WriteMsg(&data_struct.S2C_SetYananConfig{
		Error:  data_struct.S2C_SetYananConfig_Ok,
		Notice: notice,
	})
	broadcastAll(&data_struct.S2C_UpdateNotice{Notice: notice})
	log.Debug("设置银滩延安公告: %v 成功", notice)
}

//设置广播
func (user *User) setRadio(radio string) {
	yananConfigData.Radio = radio
	saveConfigData(yananConfigData)
	user.WriteMsg(&data_struct.S2C_SetYananConfig{
		Error: data_struct.S2C_SetYananConfig_Ok,
		Radio: radio,
	})
	broadcastAll(&data_struct.S2C_UpdateRadio{Radio: radio})
	log.Debug("设置银滩延安广播: %v 成功", radio)
}

//设置客服微信号
func (user *User) setWeChatNumber(wechatNumber string) {
	if wechatNumber == yananConfigData.WeChatNumber {
		log.Debug("设置的银滩延安客服微信号: %v 无效", wechatNumber)
		user.WriteMsg(&data_struct.S2C_SetYananConfig{
			Error:        data_struct.S2C_SetYananConfig_WeChatNumberInvalid,
			WeChatNumber: wechatNumber,
		})
		return
	}
	yananConfigData.WeChatNumber = wechatNumber
	saveConfigData(yananConfigData)
	user.WriteMsg(&data_struct.S2C_SetYananConfig{
		Error:        data_struct.S2C_SetYananConfig_Ok,
		WeChatNumber: wechatNumber,
	})
	log.Debug("userID %v 设置银滩延安客服微信号为: %v", user.data.userData.UserID, wechatNumber)
}

//设置用户角色
func (user *User) setRole(accountID int, role int) {
	otherUserData := new(UserData)
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		// load
		db.DB(DB).C("users").
			Find(bson.M{"accountid": accountID}).One(otherUserData)
	}, func() {
		if user.state == userLogout {
			return
		}
		if otherUserData.UserID == 0 {
			log.Debug("账户ID: %v 的用户不存在", accountID)
			user.WriteMsg(&data_struct.S2C_SetUserRole{
				Error: data_struct.S2C_SetUserRole_AccountIDInvalid,
			})
			return
		}
		if otherUser, ok := userIDUsers[otherUserData.UserID]; ok {
			if otherUser.data.userData.Role == role {
				log.Debug("账户ID: %v 已经是 %v", accountID, toRoleString(role))
				user.WriteMsg(&data_struct.S2C_SetUserRole{
					Error: data_struct.S2C_SetUserRole_SetRepeated,
					Role:  role,
				})
				return
			}
			if user.data.userData.Role > otherUser.data.userData.Role {
				otherUser.data.userData.JoinAgencyAt = time.Now().Unix()
				otherUser.data.userData.Role = role
				user.WriteMsg(&data_struct.S2C_SetUserRole{
					Error: data_struct.S2C_SetUserRole_OK,
					Role:  role,
				})
				log.Debug("userID %v 设置账号ID: %v为 %v", user.data.userData.UserID, accountID, toRoleString(role))
				if otherUser.data.userData.Role == -1 {
					otherUser.Close()
				}
				return
			}
		} else {
			if otherUserData.Role == role {
				log.Debug("账户ID: %v 已经是 %v", accountID, toRoleString(role))
				user.WriteMsg(&data_struct.S2C_SetUserRole{
					Error: data_struct.S2C_SetUserRole_SetRepeated,
					Role:  otherUserData.Role,
				})
				return
			}
			if user.data.userData.Role > otherUserData.Role {
				joinAgencyAt := time.Now().Unix()
				otherUserData.Role = role
				otherUserData.JoinAgencyAt = joinAgencyAt
				updateUserData(otherUserData.UserID, bson.M{"$set": bson.M{"role": role, "joinagencyat": joinAgencyAt}})
				user.WriteMsg(&data_struct.S2C_SetUserRole{
					Error: data_struct.S2C_SetUserRole_OK,
					Role:  role,
				})
				log.Debug("userID %v 设置账号ID: %v为 %v", user.data.userData.UserID, accountID, toRoleString(role))
				return
			}
		}
		log.Debug("userID: %v 权限不够", user.data.userData.UserID)
		user.WriteMsg(&data_struct.S2C_SetUserRole{
			Error: data_struct.S2C_SetUserRole_PermissionDenied,
		})
	})
}

//设置账密登录
func (user *User) setUsernamePassword(username string, password string) {
	userData := new(UserData)
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		//load
		db.DB(DB).C("users").
			Find(bson.M{"username": username}).One(userData)
	}, func() {
		if user.state == userLogout {
			return
		}
		switch userData.UserID {
		case 0:
			user.data.userData.Username = username
			user.data.userData.Password = password
			updateUserData(user.data.userData.UserID, bson.M{"$set": bson.M{"username": username, "password": password}})
			log.Debug("userID %v 设置用户名: %v, 密码: %v", user.data.userData.UserID, username, password)
			return
		case user.data.userData.UserID:
			if user.data.userData.Password == password {
				log.Debug("userID %v 新、旧密码不能相同", user.data.userData.UserID)
				return
			}
			user.data.userData.Password = password
			updateUserData(user.data.userData.UserID, bson.M{"$set": bson.M{"password": password}})
			log.Debug("userID %v 修改密码为: %v", user.data.userData.UserID, password)
			return
		}
		log.Debug("用户名: %v 已占用", username)

	})
}
