package internal

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/name5566/leaf/log"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
	"yananmj-server/common"
	"yananmj-server/msg"
)

type AliPayResult struct {
	GmtCreate      string `json:"gmt_create"` // 交易创建时间
	Charset        string `json:"charset"`
	SellerEmail    string `json:"seller_email"` // 卖家支付宝账号
	Subject        string `json:"subject"`      // 订单标题
	Sign           string `json:"sign"`
	BuyerID        string `json:"buyer_id"`       // 卖家支付宝用户号
	InvoiceAmount  string `json:"invoice_amount"` // 开票金额
	NotifyID       string `json:"notify_id"`      // 通知校验ID
	FundBillList   string `json:"fund_bill_list"` // 支付金额信息
	NotifyType     string `json:"notify_type"`    // 通知类型
	TradeStatus    string `json:"trade_status"`   // 交易状态
	ReceipAmout    string `json:"receip_amout"`   // 实收金额
	AppID          string `json:"app_id"`
	BuyerPayAmount string `json:"buyer_pay_amount"` // 付款金额
	SignType       string `json:"sign_type"`
	SellerID       string `json:"seller_id"`   // 卖家支付宝用户号
	GmtPayment     string `json:"gmt_payment"` // 交易付款时间
	NotifyTime     string `json:"notify_time"` // 通知时间
	Version        string `json:"version"`
	OutTradeNo     string `json:"out_trade_no"` // 商户订单号
	TotalAmount    string `json:"total_amount"` // 订单金额
	TradeNo        string `json:"trade_no"`     // 支付宝交易号
	AuthAppID      string `json:"auth_app_id"`
	BuyLogonID     string `json:"buy_logon_id"` // 买家支付宝账号
	PointAmonut    string `json:"point_amonut"` // 集分宝金额
}

var (
	rsaPrivateKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA0WSGFirHp7WRCDoNb9gdVzMBA+2E6wXtDdHoCvUvlZvBMUmzbns26TOJb4CYuKcb/rZRU9bl5CvrBS9qEW4CFkDJdWpI4sfCWWrTCCgd+DZH0NTi5xYoW2r5xOC2WPqjF678qu1tb2KFcal90aPy8/gChkfxYVCaZJKxcMblK/R68xdx0w8EO7vwwGkKgNSI3cXNgLRjNGl8iOFjHMXxpP6JLZJehlFOTEKWzCfvS4XPc6OhanoaQCqdGdqFqUjyGxsk3sIy2w8IuloqWlES9mEoyDKq1F1gK5zhsLuZvVI3LK+eKCRxgeFVFXw7uUqTfcB3D6wRhqR8CXAa8RA/pwIDAQABAoIBAQDMAyI1hNbkSx4UouMmnqzvoc0Sc5/2kN6HgYWQ75S+MnQHvqQpN7mnesQkNGoYNxEqmZ4hjpaMOlIQykKQ2tsDrXnbgYOkGTb9gfw8zUFt7g0IpfKxbkBB2bejH8HqbcDruV2KeCwQwy/7L0VcNV3oYDKtfHjs9OiIpvhlRhRRPmyeMgZWr7TLp9G/x4iF0eg320pxWwurkW7IKOeFOsyO77EvT/8MnzHzgbSMoc5PmbUpKmbThLh/F/dKcUY0lxUeiUCCt8w/aLhKLO85qngypRegVJkkb30Ad34gpi6akauZUazNg6HViwJ5HzF7qudp71Eoi63NmQLlUKCXSa15AoGBAPxBCf6XYm90RJrblfQyeMAMIjej2/mqYVtQs+o1WaGEiS3DrH/iselesnx595A5lc8PtrEM2vhcDI1U/98DcDAJ9VXEvQZwnw78iQaY/GCfi3WnY3quiNiwvA/Da3FcOW6CfFqe57pFyisiSkdPhF/kxD+d9BMlOH0H5a1mWLXtAoGBANSAi1RBTpE6xL5jMI5+I3iQQkCtKnf4SeEfMYjuqZYqKvALA+jDkSjFYD+pVfAY+T1c5i72ONA1TVMaG4+8Wm7eke3cQ9+nlgMaeAeGIdD6JAS97u5cS/lML1qh5xVYNRFxrWVWQylCiru2VjEJd6xAWch6VtRALJrDMQHDPNljAoGBANSY9BVv/PQ2J4PkQXN3/jDNiSEfpsu6fyb400k3AX2ROBQr7/wwUQWAXClwmechwVKryataTEo5OhL7alLIkQrLucs5bp442LVGvS2kTkAY9u6Hzt2cr5UBDt6yMqFturGao7e0aVSicQr9cWC8cbJoGcYMF6LzIbKurzH/KhDZAoGAG34yCImWf6Wp1LQCkTzym+OWHsYIq5LdBBpED2JJYJs+COZz8AZ1XmAC7tmau8CPZogBY+wJN67dvTWwgS0uSg/Ts4F+6o3FE8u14ctRzra+ODrWkdIxJiTcL46o1hMeco5Rj73UXJ82UcjqZ9fAuvFsbEqft0BCRReh3IeE9N8CgYArV5RndC3yz7GLPeyd8scUf6qzeZ6zbqkPytG/QJS+v2F9PFOu2hNjLCs+l+ZoQCcLx21dQhkARNp4TJuWu5PqHoG4Sg/aeOxdAQoXK61fEyRUzF+5DoYR3674fo3qb2HVdAKznV+YXShEEvEH8GFrkRurI9mArZpQTI8jHBEBww==
-----END RSA PRIVATE KEY-----`)

	alipayrsaPublicKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAhZ5tUwVnou3dfmginCDmRX8Lfu3HwOOitBEY0buaAb65C6dL9xXtnKJp5QzOuylfRNy4sdXXWXXd6SlkK0Xb8tuBxLcTW67TBz1SmzUstxvWbH18o8e6MED1VhvEt8puNdcBInfgDdeKx+x8X/u8zBldzL4K9Yieb67fMsfvDszWV2rI+BvxaKnNmYYBpC8oyJpq261d43WZosbuFghoT7hYks8NuLfPc+T6+xGRNfl0BrGMsVE4xAvE7E79AvXLCkZh+AV2FPGvy7TB0Dxbnn0mpNt2NrcwvMM7sbIlDL3hPdtCXl2/vY5KIA87qIyBQHpR+w9BTNIW5mkXm36ZQQIDAQAB
-----END RSA PRIVATE KEY-----`)

	PID        = "2088811602965802"
	appID      = "2017110809804067"
	gatewayUrl = "https://openapi.alipay.com/gateway.do"
	notifyUrl  = "http://139.199.180.94:8082/czddz/alipay"
)

func handleAliPay(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		total_amount := r.URL.Query().Get("total_amount")
		accountID, _ := strconv.Atoi(r.URL.Query().Get("accountID"))
		if total_amount == "" || accountID < 1 {
			return
		}
		data := requestAlipayGateway(total_amount, accountID)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fmt.Fprintf(w, "%s", data)
		return
	}
	if r.Method == "POST" {
		result, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		log.Debug("result: %s", result)
		m, _ := url.ParseQuery(string(result))

		p := url.Values{}
		p.Add("app_id", m.Get("app_id"))
		p.Add("auth_app_id", m.Get("auth_app_id"))
		p.Add("buyer_id", m.Get("buyer_id"))
		p.Add("buyer_logon_id", m.Get("buyer_logon_id"))
		p.Add("buyer_pay_amount", m.Get("buyer_pay_amount"))
		p.Add("charset", m.Get("charset"))
		p.Add("fund_bill_list", m.Get("fund_bill_list"))
		p.Add("gmt_create", m.Get("gmt_create"))
		p.Add("gmt_payment", m.Get("gmt_payment"))
		p.Add("notify_time", m.Get("notify_time"))
		p.Add("notify_type", m.Get("notify_type"))
		p.Add("out_trade_no", m.Get("out_trade_no"))
		p.Add("point_amount", m.Get("point_amount"))
		p.Add("receipt_amount", m.Get("receipt_amount"))
		p.Add("seller_email", m.Get("seller_email"))
		p.Add("seller_id", m.Get("seller_id"))
		p.Add("subject", m.Get("subject"))
		p.Add("total_amount", m.Get("total_amount"))
		p.Add("trade_no", m.Get("trade_no"))
		p.Add("trade_status", m.Get("trade_status"))
		p.Add("version", m.Get("version"))
		// 签名验证
		signTemp := buildParam(p)
		log.Debug("signTemp: %v", signTemp)

		if ok := verifyResponseData(signTemp, m.Get("sign"), m.Get("sign_type")); !ok {
			fmt.Fprint(w, "failure")
			return
		}
		userData := new(UserData)
		rechargeResultData := new(AliRechargeResultData)
		skeleton.Go(func() {
			db := mongoDB.Ref()
			defer mongoDB.UnRef(db)

			outTradeNo := m.Get("out_trade_no")
			err := db.DB(DB).C("alirechargeresult").
				Find(bson.M{"outtradeno": outTradeNo}).One(rechargeResultData)
			if err != nil {
				rechargeResultData = nil
				userData = nil
				return
			}
			err = db.DB(DB).C("users").Find(bson.M{"accountid": rechargeResultData.AccountID}).One(userData)
			if err != nil {
				userData = nil
				log.Error("load accountid %v user data error: %v", "", err)
			}
		}, func() {
			if rechargeResultData == nil || userData == nil {
				fmt.Fprint(w, "failure")
				return
			}
			if rechargeResultData.TotalAmout == m.Get("total_amount") &&
				appID == m.Get("app_id") && PID == m.Get("seller_id") {
				if rechargeResultData.Success {
					fmt.Fprint(w, "success")
					return
				}
				t, _ := strconv.ParseFloat(rechargeResultData.TotalAmout, 64)
				if user, ok := userIDUsers[userData.UserID]; ok {
					user.data.userData.RoomCards += int(t * 500)
					updateAliRechargeData(user.data.userData.AccountID, bson.M{"$set": bson.M{"roomcards": int(t * 500), "success": true}})
					user.WriteMsg(&data_struct.S2C_UpdateRoomCards{
						RoomCards: user.data.userData.RoomCards,
					})
				} else {
					userData.RoomCards += int(t * 500)
					updateAliRechargeData(user.data.userData.AccountID, bson.M{"$set": bson.M{"roomcards": int(t * 500), "success": true}})
				}
				fmt.Fprint(w, "success")
				return
			}
			fmt.Fprint(w, "failure")
			return
		})
	}
}

func requestAlipayGateway(totalAmount string, accountID int) []byte {

	p := url.Values{}
	outTradeNo := getOutTradeNo()
	biz_content := `{"product_code":"QUICK_MSECURITY_PAY","total_amount":"` + totalAmount + `","subject":"车主斗地主游戏充值","out_trade_no":"` + outTradeNo + `"}`
	//biz_content := "{\"product_code\":\"QUICK_MSECURITY_PAY\",\"total_amount\":\"" + totalAmount + "\",\"subject\":\"车主斗地主游戏充值\",\"out_trade_no\":\"" + outTradeNo + "\"}"
	p.Add("app_id", appID)
	p.Add("biz_content", biz_content)
	p.Add("charset", "utf-8")
	p.Add("method", "alipay.trade.app.pay")
	p.Add("notify_url", notifyUrl)
	p.Add("sign_type", "RSA2")
	p.Add("timestamp", time.Now().Format("2006-01-02 15:04:05"))
	p.Add("version", "1.0")
	block, _ := pem.Decode(rsaPrivateKey)
	priv, _ := x509.ParsePKCS1PrivateKey(block.Bytes)

	h := sha256.New()
	h.Write([]byte(buildParam(p)))

	sign, _ := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, h.Sum(nil))
	p.Add("sign", base64.StdEncoding.EncodeToString(sign))

	r, err := http.NewRequest("POST", gatewayUrl, strings.NewReader(p.Encode()))
	if err != nil {
		log.Error("post gatewayUrl error: %v", err)
		return []byte{}
	}
	defer r.Body.Close()
	result, _ := ioutil.ReadAll(r.Body)

	aliRechargeResultData := new(AliRechargeResultData)
	skeleton.Go(func() {
		err := aliRechargeResultData.initValue()
		if err != nil {
			log.Error("init rechargeResultData data error: %v", err)
			aliRechargeResultData = nil
		}
	}, func() {
		if aliRechargeResultData != nil {
			aliRechargeResultData.AccountID = accountID
			aliRechargeResultData.OutTradeNo = outTradeNo
			aliRechargeResultData.TotalAmout = totalAmount
			aliRechargeResultData.Success = false
			aliRechargeResultData.RoomCards = 0
			aliRechargeResultData.UpdateAt = time.Now().Unix()
			saveAliRechargeResultData(aliRechargeResultData)
		}
	})
	return result
}

func getOutTradeNo() string {
	return time.Now().Format("0102150405") + common.GetID(5)
}

func buildParam(v url.Values) string {
	if v == nil {
		return ""
	}
	var buf bytes.Buffer
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vs := v[k]
		prefix := k + "="
		for _, v := range vs {
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(prefix)
			buf.WriteString(v)
		}
	}
	return buf.String()
}

func verifyResponseData(data string, sign string, signType string) bool {
	s, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		log.Debug("%v", err)
		return false
	}
	log.Debug("sign: %v", s)
	if signType == "RSA2" {
		err = VerifyPKCS1v15([]byte(data), s, alipayrsaPublicKey, crypto.SHA256)
		if err == nil {
			return true
		}
		log.Debug("%v", err)
		return false
	}
	return false
}

func VerifyPKCS1v15(src, sig, key []byte, hash crypto.Hash) error {
	h := hash.New()
	h.Write(src)
	block, _ := pem.Decode(key)
	if block == nil {
		return errors.New("public key error")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}
	return rsa.VerifyPKCS1v15(pub.(*rsa.PublicKey), hash, h.Sum(nil), sig)
}
