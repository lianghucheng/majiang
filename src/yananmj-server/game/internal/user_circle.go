package internal

import (
	"encoding/json"
	"github.com/name5566/leaf/log"
	"yananmj-server/game/circle"
)

func (user *User) requestCircleID() {
	if user.isRobot() || user.data.userData.CircleID > 0 {
		return
	}
	temp := &struct {
		Code     string
		CircleID int `json:"data"`
	}{}
	skeleton.Go(func() {
		req := circle.NewCircleLoginRequest(&circle.CircleUserInfo{
			UnionID:    user.data.userData.Unionid,
			Nickname:   user.data.userData.Nickname,
			Headimgurl: user.data.userData.Headimgurl,
			Sex:        user.data.userData.Sex,
		})
		data := circle.DoRequest(req.GatewayUrl, req.Method, req.Params)
		// {"id":null,"code":"0","model":null,"message":"ok","data":174}
		// log.Debug("%s", data)
		err := json.Unmarshal(data, temp)
		if err != nil || temp.Code != "0" {
			temp = nil
			return
		}
	}, func() {
		if temp != nil {
			user.data.userData.CircleID = temp.CircleID
		}
	})
}

func (user *User) requestCircleLoginCode(successCb func(loginCode string), failCb func()) {
	if user.data.userData.CircleID < 1 {
		if failCb != nil {
			failCb()
		}
		return
	}
	temp := &struct {
		Code      string
		LoginCode string `json:"data"`
	}{}
	skeleton.Go(func() {
		req := circle.NewCircleAuthorize(user.data.userData.CircleID)
		data := circle.DoRequest(req.GatewayUrl, req.Method, req.Params)
		// {"id":null,"code":"0","model":null,"message":"ok","data":"1730c016-01c7-4e15-85b8-986f5d812dd9"}
		// log.Debug("%s", data)
		err := json.Unmarshal(data, temp)
		if err != nil || temp.Code != "0" {
			temp = nil
			return
		}
	}, func() {
		if temp == nil {
			if failCb != nil {
				failCb()
			}
		} else {
			if successCb != nil {
				successCb(temp.LoginCode)
			}
		}
	})
}

// 请求生成一个圈圈红包
func (user *User) requestCircleRedPacket(redPacket float64, desc string, successCb func(), failCb func()) {
	temp := &struct {
		Code string
		Data string
	}{}
	skeleton.Go(func() {
		req := circle.NewCircleCreateRedPacketRequest(&circle.RedPacketInfo{
			UserID: user.data.userData.CircleID,
			Sum:    redPacket,
			Desc:   desc,
		})
		data := circle.DoRequest(req.GatewayUrl, req.Method, req.Params)
		// {"id":null,"code":"0","model":null,"message":"ok","data":"SUCCESS"}
		log.Debug("%s", data)
		err := json.Unmarshal(data, temp)
		if err != nil || temp.Code != "0" || temp.Data != "SUCCESS" {
			temp = nil
			return
		}
	}, func() {
		if temp == nil {
			if failCb != nil {
				failCb()
			}
		} else {
			if successCb != nil {
				successCb()
			}
		}
	})
}
