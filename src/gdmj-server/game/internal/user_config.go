package internal

import (
	"gdmj-server/msg"
	"github.com/name5566/leaf/log"
	"gopkg.in/mgo.v2/bson"
	"net/url"
	"time"
)

var (
	gdIOSProductInfos = []msg.ProductInfo{
		{ID: "com.youxibi.hnzzmj1", Desc: "房卡48张", Price: 6},
		{ID: "com.youxibi.hnzzmj2", Desc: "房卡240张", Price: 30},
		{ID: "com.youxibi.hnzzmj3", Desc: "房卡500张", Price: 60},
	}

	gdAndroidProductInfos = []msg.ProductInfo{
		{Desc: "房卡48张", Price: 6},
		{Desc: "房卡240张", Price: 30},
		{Desc: "房卡500张", Price: 60},
	}
)

func (user *User) setGDAndroidVersion(version int) {
	if version < gdConfigData.AndroidVersion && user.data.userData.Role < roleRoot {
		log.Debug("设置广东麻将安卓版本: %v 无效", version)
		user.WriteMsg(&msg.S2C_SetGDConfig{
			Error:          msg.S2C_SetGDConfig_VersionInvalid,
			AndroidVersion: version,
		})
		return
	}
	gdConfigData.AndroidVersion = version
	saveConfigData(gdConfigData)
	user.WriteMsg(&msg.S2C_SetGDConfig{
		Error:          msg.S2C_SetGDConfig_OK,
		AndroidVersion: version,
	})
	log.Debug("userID %v 设置广东麻将安卓版本: %v 成功", user.data.userData.UserID, version)
}

func (user *User) setGDIOSVersion(version int) {
	if version < gdConfigData.IOSVersion && user.data.userData.Role < roleRoot {
		log.Debug("设置广东麻将iOS版本: %v 无效", version)
		user.WriteMsg(&msg.S2C_SetGDConfig{
			Error:      msg.S2C_SetGDConfig_VersionInvalid,
			IOSVersion: version,
		})
		return
	}
	gdConfigData.IOSVersion = version
	saveConfigData(gdConfigData)
	user.WriteMsg(&msg.S2C_SetGDConfig{
		Error:      msg.S2C_SetGDConfig_OK,
		IOSVersion: version,
	})
	log.Debug("userID %v 设置广东麻将iOS版本: %v 成功", user.data.userData.UserID, version)
}

func (user *User) setGDAndroidDownloadUrl(downloadUrl string) {
	surl, err := url.Parse(downloadUrl)
	if err == nil && surl.Scheme == "" || err != nil || downloadUrl == gdConfigData.AndroidDownloadUrl {
		log.Debug("设置广东麻将安卓下载地址: %v 无效", downloadUrl)
		user.WriteMsg(&msg.S2C_SetGDConfig{
			Error:              msg.S2C_SetGDConfig_DownloadUrlInvalid,
			AndroidDownloadUrl: downloadUrl,
		})
		return
	}
	gdConfigData.AndroidDownloadUrl = downloadUrl
	saveConfigData(gdConfigData)
	user.WriteMsg(&msg.S2C_SetGDConfig{
		Error:              msg.S2C_SetGDConfig_OK,
		AndroidDownloadUrl: downloadUrl,
	})
	log.Debug("userID %v 设置广东麻将安卓下载地址: %v 成功", user.data.userData.UserID, downloadUrl)
}

func (user *User) setGDIOSDownloadUrl(downloadUrl string) {
	surl, err := url.Parse(downloadUrl)
	if err == nil && surl.Scheme == "" || err != nil || downloadUrl == gdConfigData.IOSDownloadUrl {
		log.Debug("设置广东麻将iOS下载地址: %v无效", downloadUrl)
		user.WriteMsg(&msg.S2C_SetGDConfig{
			Error:          msg.S2C_SetGDConfig_DownloadUrlInvalid,
			IOSDownloadUrl: downloadUrl,
		})
		return
	}
	gdConfigData.IOSDownloadUrl = downloadUrl
	saveConfigData(gdConfigData)
	user.WriteMsg(&msg.S2C_SetGDConfig{
		Error:          msg.S2C_SetGDConfig_OK,
		IOSDownloadUrl: downloadUrl,
	})
	log.Debug("userID %v 设置广东麻将iOS下载地址: %v 成功", user.data.userData.UserID, downloadUrl)
}

func (user *User) setGDNotice(notice string) {
	gdConfigData.Notice = notice
	saveConfigData(gdConfigData)
	user.WriteMsg(&msg.S2C_SetGDConfig{
		Error:  msg.S2C_SetGDConfig_OK,
		Notice: notice,
	})
	broadcastAll(&msg.S2C_UpdateNotice{
		Notice: notice,
	})
	log.Debug("userID %v 设置广东麻将公告: %v 成功", user.data.userData.UserID, notice)
}

func (user *User) setGDRadio(radio string) {
	gdConfigData.Radio = radio
	saveConfigData(gdConfigData)
	user.WriteMsg(&msg.S2C_SetGDConfig{
		Error: msg.S2C_SetGDConfig_OK,
		Radio: radio,
	})
	broadcastAll(&msg.S2C_UpdateRadio{})
	log.Debug("userID %v 设置广东麻将广播: %v 成功", user.data.userData.UserID, radio)
}

func (user *User) setGDWeChatNumber(wechatNumber string) {
	if wechatNumber == gdConfigData.WeChatNumber {
		log.Debug("设置广东麻将客服微信号: %v 无效", wechatNumber)
		user.WriteMsg(&msg.S2C_SetGDConfig{
			Error:        msg.S2C_SetGDConfig_WeChatNumberInvalid,
			WeChatNumber: wechatNumber,
		})
		return
	}
	gdConfigData.WeChatNumber = wechatNumber
	saveConfigData(gdConfigData)
	user.WriteMsg(&msg.S2C_SetGDConfig{
		Error:        msg.S2C_SetGDConfig_OK,
		WeChatNumber: wechatNumber,
	})
	log.Debug("userID %v 设置广东麻将客服微信号: %v 成功", user.data.userData.UserID, wechatNumber)
}

func (user *User) setUsernamePassword(username string, password string) {
	userData := new(UserData)
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)

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
			log.Debug("userID: %v 设置用户名: %v 密码: %v", user.data.userData.UserID, username, password)
			return
		case user.data.userData.UserID:
			if user.data.userData.Password == password {
				log.Debug("新旧密码不能相同")
				return
			}
			user.data.userData.Password = password
			updateUserData(user.data.userData.UserID, bson.M{"$set": bson.M{"password": password}})
			log.Debug("userID: %v 修改密码为: %v", user.data.userData.UserID, password)
			return
		}
		log.Debug("用户名: %v 已占用", username)
	})
}

func (user *User) setRole(accountID int, role int) {
	otherUserData := new(UserData)
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)

		db.DB(DB).C("users").
			Find(bson.M{"accountid": accountID}).One(otherUserData)
	}, func() {
		if user.state == userLogout {
			return
		}
		if otherUserData.UserID == 0 {
			log.Debug("账户ID: %v 的用户不存在", accountID)
			user.WriteMsg(&msg.S2C_SetUserRole{
				Error: msg.S2C_SetUserRole_AccountIDInvalid,
			})
			return
		}
		if otherUser, ok := userIDUsers[otherUserData.UserID]; ok {
			if otherUser.data.userData.Role == role {
				log.Debug("账户ID: %v 已经是: %v", accountID, toRoleString(role))
				user.WriteMsg(&msg.S2C_SetUserRole{
					Error: msg.S2C_SetUserRole_SetRepeated,
					Role:  role,
				})
				return
			}
			if user.data.userData.Role > otherUser.data.userData.Role {
				otherUser.data.userData.JoinAgencyAt = time.Now().Unix()
				otherUser.data.userData.Role = role
				user.WriteMsg(&msg.S2C_SetUserRole{
					Error: msg.S2C_SetUserRole_OK,
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
				log.Debug("账户ID: %v 已经是: %v", accountID, toRoleString(role))
				user.WriteMsg(&msg.S2C_SetUserRole{
					Error: msg.S2C_SetUserRole_SetRepeated,
					Role:  role,
				})
				return
			}
			if user.data.userData.Role > otherUserData.Role {
				otherUserData.JoinAgencyAt = time.Now().Unix()
				otherUserData.Role = role
				user.WriteMsg(&msg.S2C_SetUserRole{
					Error: msg.S2C_SetUserRole_OK,
					Role:  role,
				})
				log.Debug("userID %v 设置账号ID: %v为 %v", user.data.userData.UserID, accountID, toRoleString(role))
				return
			}
		}
		log.Debug("userID: %v 权限不够", user.data.userData.UserID)
		user.WriteMsg(&msg.S2C_SetUserRole{
			Error: msg.S2C_SetUserRole_PermissionDenied,
		})
	})
}
