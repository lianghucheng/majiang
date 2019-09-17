package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"gdmj-server/common"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"time"
)

var (
	packetCreate = struct {
		Urll   string
		Method string
	}{
		Urll:   "http://192.168.1.150:7006/server.do",
		Method: "shenzhouxing.circle.server.packet.create.normal",
	}

	login = struct {
		Urll   string
		Method string
	}{
		Urll:   "http://192.168.1.150:8002/server.do",
		Method: "youxibi.user.server.third.create.wechat",
	}

	sign = struct {
		SignType    string
		VersionCode string
		VersionName string
		Lang        string
		DeviceId    string
		Device      string
		PartnerKey  string
		SecretKey   string
	}{
		SignType:    "NORMAL",
		VersionCode: "1",
		VersionName: "1.0",
		Lang:        "CN",
		Device:      "SERVER",
		DeviceId:    "test",
		PartnerKey:  "youxibi_game_chezhu_ddz",
		SecretKey:   "F957BC19502E301F8FDB8BF192116AFD",
	}
)

type CircleRegister struct {
	UnionId    string
	Nickname   string
	Sex        int
	Language   string
	City       string
	Province   string
	Country    string
	Headimgurl string
}

type CircleCreatePacket struct {
	UserId int64
	Sum    float32
	Desc   string
}

func serverRequest(v interface{}) ([]byte, error) {
	now := strconv.FormatInt(time.Now().Unix(), 10)
	urll, method := "", ""
	switch v.(type) {
	case CircleRegister:
		method = login.Method
		urll = login.Urll
	case CircleCreatePacket:
		method = packetCreate.Method
		urll = packetCreate.Urll
	default:
		return nil, errors.New("invalid type assertion :" + fmt.Sprint(reflect.TypeOf(v)))
	}
	params, _ := json.Marshal(v)
	resp, err := http.PostForm(urll, url.Values{
		"params":     {string(params)},
		"partnerKey": {sign.PartnerKey},
		"sendTime":   {now},
		"sign": {common.Md5Encrypt(
			"device=" + sign.Device +
				"&deviceId=" + sign.DeviceId +
				"&lang=" + sign.Lang +
				"&method=" + method +
				"&params=" + string(params) +
				"&partnerKey=" + sign.PartnerKey +
				"&secretKey=" + sign.SecretKey +
				"&sendTime=" + now +
				"&signType=" + sign.SignType +
				"&versionCode=" + sign.VersionCode +
				"&versionName=" + sign.VersionName,
		)},
		"signType":    {sign.SignType},
		"versionCode": {sign.VersionCode},
		"versionName": {sign.VersionName},
		"lang":        {sign.Lang},
		"deviceId":    {sign.DeviceId},
		"device":      {sign.Device},
		"method":      {method},
	})
	if err != nil {
		return nil, errors.New(fmt.Sprint(err))
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New(fmt.Sprint(err))
	}
	return body, nil
}
