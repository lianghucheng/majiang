package circle

import (
	"crypto/md5"
	"fmt"
	"github.com/name5566/leaf/log"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
	"yananmj-server/common"
)

var (
	partnerKey = "youxibi_game_yananmj"
	secretKey  = "B592C58AF6B9EF10C366240E8F19E491"
)

func DoRequest(gatewayUrl, method, params string) []byte {
	p := url.Values{}
	p.Add("device", "SERVER")
	p.Add("deviceId", "CZDDZ")
	p.Add("lang", "CN")
	p.Add("method", method)
	p.Add("params", params)
	p.Add("partnerKey", partnerKey)
	p.Add("secretKey", secretKey)
	p.Add("sendTime", strconv.Itoa(int(time.Now().Unix())))
	p.Add("signType", "NORMAL")
	p.Add("versionCode", "1")
	p.Add("versionName", "1.0")
	p.Add("sign", generateSign(p))

	r, err := http.PostForm(gatewayUrl, p)
	if err != nil {
		log.Debug("%v", err)
		return []byte{}
	}
	defer r.Body.Close()
	result, _ := ioutil.ReadAll(r.Body)
	return result
}

func generateSign(params url.Values) string {
	return sign(common.GetSignContent(params))
}

func sign(data string) string {
	m := md5.New()
	io.WriteString(m, data)
	// return hex.EncodeToString(m.Sum(nil))
	return fmt.Sprintf("%X", m.Sum(nil))
}
