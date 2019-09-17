package circle

import (
	"encoding/json"
)

type CircleRequest struct {
	GatewayUrl string
	Method     string
	Params     string
}

type CircleUserInfo struct {
	UnionID    string `json:"unionid"`
	Nickname   string `json:"nickname"`
	Headimgurl string `json:"headimgurl"`
	Sex        int    `json:"sex"`
	Language   string `json:"language"`
	City       string `json:"city"`
	Province   string `json:"province"`
	Country    string `json:"country"`
}

type RedPacketInfo struct {
	UserID int     `json:"userid"`
	Sum    float64 `json:"sum"`
	Desc   string  `json:"desc"`
}

func NewCircleLoginRequest(info *CircleUserInfo) *CircleRequest {
	data, _ := json.Marshal(info)
	req := new(CircleRequest)
	req.GatewayUrl = "http://api.user.youxibi.com/server.do"
	req.Method = "youxibi.user.server.third.create.wechat"
	req.Params = string(data)
	return req
}

func NewCircleCreateRedPacketRequest(info *RedPacketInfo) *CircleRequest {
	data, _ := json.Marshal(info)
	req := new(CircleRequest)
	req.GatewayUrl = "http://api.circle.shenzhouxing.com/server.do"
	req.Method = "shenzhouxing.circle.server.packet.create.normal"
	req.Params = string(data)
	return req
}

func NewCircleAuthorize(userID int) *CircleRequest {
	temp := &struct {
		UserID int `json:"userid"`
	}{
		UserID: userID,
	}
	data, _ := json.Marshal(temp)
	req := new(CircleRequest)
	req.GatewayUrl = "http://api.user.youxibi.com/server.do"
	req.Method = "youxibi.user.server.create.login.code"
	req.Params = string(data)
	return req
}
