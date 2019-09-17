package internal

import (
	"github.com/name5566/leaf/log"
	"gopkg.in/mgo.v2/bson"
	"hnzzmj-server/msg"
	"net/url"
	"time"
)

var (
	hnzzIOSProductInfos = []msg.ProductInfo{
		{ID: "com.youxibi.hnzzmj1", Desc: "房卡48张", Price: 6},
		{ID: "com.youxibi.hnzzmj2", Desc: "房卡240张", Price: 30},
		{ID: "com.youxibi.hnzzmj3", Desc: "房卡500张", Price: 60},
	}

	hnzzAndroidProductInfos = []msg.ProductInfo{
		{Desc: "房卡48张", Price: 6},
		{Desc: "房卡240张", Price: 30},
		{Desc: "房卡500张", Price: 60},
	}
)

func (user *User) setHNZZAndroidVersion(version int) {
	if version <= hnzzConfigData.AndroidVersion && user.data.userData.Role < roleRoot {
		log.Debug("设置的银滩转转安卓版本: %v 无效", version)
		user.WriteMsg(&msg.S2C_SetHNZZConfig{
			Error:          msg.S2C_SetHNZZConfig_VersionInvalid,
			AndroidVersion: version,
		})
		return
	}
	hnzzConfigData.AndroidVersion = version
	saveConfigData(hnzzConfigData)
	user.WriteMsg(&msg.S2C_SetHNZZConfig{
		Error:          msg.S2C_SetHNZZConfig_OK,
		AndroidVersion: version,
	})
	log.Debug("userID %v 设置银滩转转安卓新版本为: %v", user.data.userData.UserID, version)
}

func (user *User) setHNZZIOSVersion(version int) {
	if version <= hnzzConfigData.IOSVersion && user.data.userData.Role < roleRoot {
		log.Debug("设置的银滩转转iOS版本: %v 无效", version)
		user.WriteMsg(&msg.S2C_SetHNZZConfig{
			Error:      msg.S2C_SetHNZZConfig_VersionInvalid,
			IOSVersion: version,
		})
		return
	}
	hnzzConfigData.IOSVersion = version
	saveConfigData(hnzzConfigData)
	user.WriteMsg(&msg.S2C_SetHNZZConfig{
		Error:      msg.S2C_SetHNZZConfig_OK,
		IOSVersion: version,
	})
	log.Debug("userID %v 设置银滩转转iOS新版本为: %v", user.data.userData.UserID, version)
}

func (user *User) setHNZZAndroidDownloadUrl(downloadUrl string) {
	surl, err := url.Parse(downloadUrl)
	if err == nil && surl.Scheme == "" || err != nil || downloadUrl == hnzzConfigData.AndroidDownloadUrl {
		log.Debug("设置的银滩转转安卓下载地址: %v 无效", downloadUrl)
		user.WriteMsg(&msg.S2C_SetHNZZConfig{
			Error:              msg.S2C_SetHNZZConfig_DownloadUrlInvalid,
			AndroidDownloadUrl: downloadUrl,
		})
		return
	}
	hnzzConfigData.AndroidDownloadUrl = downloadUrl
	saveConfigData(hnzzConfigData)
	user.WriteMsg(&msg.S2C_SetHNZZConfig{
		Error:              msg.S2C_SetHNZZConfig_OK,
		AndroidDownloadUrl: downloadUrl,
	})
	log.Debug("userID %v 设置银滩转转安卓下载地址为: %v", user.data.userData.UserID, downloadUrl)
}

func (user *User) setHNZZIOSDownloadUrl(downloadUrl string) {
	surl, err := url.Parse(downloadUrl)
	if err == nil && surl.Scheme == "" || err != nil || downloadUrl == hnzzConfigData.IOSDownloadUrl {
		log.Debug("设置的银滩转转iOS下载地址: %v 无效", downloadUrl)
		user.WriteMsg(&msg.S2C_SetHNZZConfig{
			Error:          msg.S2C_SetHNZZConfig_DownloadUrlInvalid,
			IOSDownloadUrl: downloadUrl,
		})
		return
	}
	hnzzConfigData.IOSDownloadUrl = downloadUrl
	saveConfigData(hnzzConfigData)
	user.WriteMsg(&msg.S2C_SetHNZZConfig{
		Error:          msg.S2C_SetHNZZConfig_OK,
		IOSDownloadUrl: downloadUrl,
	})
	log.Debug("userID %v 设置银滩转转iOS下载地址为: %v", user.data.userData.UserID, downloadUrl)
}

func (user *User) setHNZZNotice(notice string) {
	hnzzConfigData.Notice = notice
	saveConfigData(hnzzConfigData)
	user.WriteMsg(&msg.S2C_SetHNZZConfig{
		Error:  msg.S2C_SetHNZZConfig_OK,
		Notice: notice,
	})
	broadcastAll(&msg.S2C_UpdateNotice{Notice: notice})
	log.Debug("userID %v 设置银滩转转公告成功", user.data.userData.UserID)
}

func (user *User) setHNZZRadio(radio string) {
	hnzzConfigData.Radio = radio
	saveConfigData(hnzzConfigData)
	user.WriteMsg(&msg.S2C_SetHNZZConfig{
		Error: msg.S2C_SetHNZZConfig_OK,
		Radio: radio,
	})
	broadcastAll(&msg.S2C_UpdateRadio{Radio: radio})
	log.Debug("userID %v 设置银滩转转广播成功", user.data.userData.UserID)
}

func (user *User) setHNZZWeChatNumber(wechatNumber string) {
	if wechatNumber == hnzzConfigData.WeChatNumber {
		log.Debug("设置的银滩转转客服微信号: %v 无效", wechatNumber)
		user.WriteMsg(&msg.S2C_SetHNZZConfig{
			Error:        msg.S2C_SetHNZZConfig_WeChatNumberInvalid,
			WeChatNumber: wechatNumber,
		})
		return
	}
	hnzzConfigData.WeChatNumber = wechatNumber
	saveConfigData(hnzzConfigData)
	user.WriteMsg(&msg.S2C_SetHNZZConfig{
		Error:        msg.S2C_SetHNZZConfig_OK,
		WeChatNumber: wechatNumber,
	})
	log.Debug("userID %v 设置银滩转转客服微信号为: %v", user.data.userData.UserID, wechatNumber)
}

func (user *User) setUsernamePassword(username string, password string) {
	userData := new(UserData)
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		// load
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
			user.WriteMsg(&msg.S2C_SetUserRole{
				Error: msg.S2C_SetUserRole_AccountIDInvalid,
			})
			return
		}
		if otherUser, ok := userIDUsers[otherUserData.UserID]; ok {
			if otherUser.data.userData.Role == role {
				log.Debug("账户ID: %v 已经是 %v", accountID, toRoleString(role))
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
				log.Debug("账户ID: %v 已经是 %v", accountID, toRoleString(role))
				user.WriteMsg(&msg.S2C_SetUserRole{
					Error: msg.S2C_SetUserRole_SetRepeated,
					Role:  otherUserData.Role,
				})
				return
			}
			if user.data.userData.Role > otherUserData.Role {
				joinAgencyAt := time.Now().Unix()
				otherUserData.Role = role
				otherUserData.JoinAgencyAt = joinAgencyAt
				updateUserData(otherUserData.UserID, bson.M{"$set": bson.M{"role": role, "joinagencyat": joinAgencyAt}})
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
