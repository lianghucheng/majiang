package internal

import (
	"crypto/md5"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/name5566/leaf/log"
	"gopkg.in/mgo.v2/bson"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"time"
	"yananmj-server/msg"
)

type WeChatPayResult struct {
	AppID         string `xml:"appid"`
	BankType      string `xml:"bank_type"` // 付款银行
	CashFee       int    `xml:"cash_fee"`  // 现金支付金额
	FeeType       string `xml:"fee_type"`  // 货币种类
	IsSubscribe   string `xml:"is_subscribe"`
	MchID         string `xml:"mch_id"`       // 商户号
	NonceStr      string `xml:"nonce_str"`    // 随机字符串
	OpenID        string `xml:"open_id"`      // 用户标识
	OutTradeNo    string `xml:"out_trade_no"` // 商户订单号
	ResultCode    string `xml:"result_code"`
	ReturnCode    string `xml:"return_code"`
	Sign          string `xml:"sign"`           // 签名
	TimeEnd       string `xml:"time_end"`       // 支付完成时间
	TotalFee      int    `xml:"total_fee"`      // 总金额
	TradeType     string `xml:"trade_type"`     // 交易类型
	TransactionID string `xml:"transaction_id"` // 微信支付订单号
}

func handleWeChatPay(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		result, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		handleWeChatPayResult(result)
		return
	}
	m := map[string]interface{}{}
	m["appid"] = "wx8345bec34f4c1192"
	m["body"] = "车主斗地主在线支付"
	m["key"] = "987601DFA5CC4D61A0245AC9C4C66037"
	m["mch_id"] = "1445975802"
	m["notify_url"] = "http://139.199.180.94:8082/czddz/wechatpay"
	m["spbill_create_ip"] = "113.92.153.13"

	data, err := json.Marshal(m)
	if err != nil {
		log.Error("marshal message %v error: %v", reflect.TypeOf(m), err)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprintf(w, "%s", data)
}

func handleWeChatPayResult(result []byte) {
	wechatPayResult := new(WeChatPayResult)
	err := xml.Unmarshal(result, &wechatPayResult)
	if err != nil {
		log.Error("unmarshal wechatpay result error: %v", err)
		return
	}
	log.Debug("wechatPayResult: %v", wechatPayResult)
	// 签名验证
	signTemp := "appid=" + wechatPayResult.AppID + "&bank_type=" + wechatPayResult.BankType +
		"&cash_fee=" + strconv.Itoa(wechatPayResult.CashFee) + "&fee_type=" + wechatPayResult.FeeType +
		"&is_subscribe=" + wechatPayResult.IsSubscribe + "&mch_id=" + wechatPayResult.MchID +
		"&nonce_str=" + wechatPayResult.NonceStr + "&openid=" + wechatPayResult.OpenID +
		"&out_trade_no=" + wechatPayResult.OutTradeNo + "&result_code=" + wechatPayResult.ResultCode +
		"&return_code=" + wechatPayResult.ReturnCode + "&time_end=" + wechatPayResult.TimeEnd +
		"&total_fee=" + strconv.Itoa(wechatPayResult.TotalFee) + "&trade_type=" + wechatPayResult.TradeType +
		"&transaction_id=" + wechatPayResult.TransactionID
	h := md5.New()
	io.WriteString(h, signTemp+"&key=987601DFA5CC4D61A0245AC9C4C66037")
	sign := fmt.Sprintf("%x", h.Sum(nil))
	if sign != wechatPayResult.Sign {
		return
	}
	firstPay := false
	userData := new(UserData)
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)

		count, _ := db.DB(DB).C("wechatrechargeresult").
			Find(bson.M{"outtradeno": wechatPayResult.OutTradeNo}).Count()
		if count == 0 {
			firstPay = true
		} else {
			return
		}
		err := db.DB(DB).C("users").Find(bson.M{"openid": wechatPayResult.OpenID}).One(userData)
		if err != nil {
			userData = nil
			log.Error("load openid %v user data error: %v", wechatPayResult.OpenID, err)
		}
	}, func() {
		if !firstPay || userData == nil {
			return
		}
		if user, ok := userIDUsers[userData.UserID]; ok {
			user.data.userData.RoomCards += wechatPayResult.TotalFee * 5
			user.saveUserRechargeResultData(wechatPayResult.TotalFee*5, &RechargeResultData{
				AccountID:  user.data.userData.AccountID,
				Nickname:   user.data.userData.Nickname,
				Headimgurl: user.data.userData.Headimgurl,
				OutTradeNo: wechatPayResult.OutTradeNo,
				TotalFee:   wechatPayResult.TotalFee,
				RoomCards:  wechatPayResult.TotalFee * 5,
			})
			user.WriteMsg(&data_struct.S2C_UpdateRoomCards{
				RoomCards: user.data.userData.RoomCards,
			})
		} else {
			userData.RoomCards += wechatPayResult.TotalFee * 5
			user.saveUserRechargeResultData(wechatPayResult.TotalFee*5, &RechargeResultData{
				AccountID:  userData.AccountID,
				Nickname:   userData.Nickname,
				Headimgurl: userData.Headimgurl,
				OutTradeNo: wechatPayResult.OutTradeNo,
				TotalFee:   wechatPayResult.TotalFee,
				RoomCards:  wechatPayResult.TotalFee * 5,
			})
		}
	})
}

// 用户充值(微信充值)
func (user *User) saveUserRechargeResultData(roomCards int, info *RechargeResultData) {
	rechargeResultData := new(RechargeResultData)
	skeleton.Go(func() {
		err := rechargeResultData.initValue(roomCards)
		if err != nil {
			log.Error("init rechargeResultData data error: %v", err)
			rechargeResultData = nil
		}
	}, func() {
		if rechargeResultData != nil {
			rechargeResultData.AccountID = info.AccountID
			rechargeResultData.Nickname = info.Nickname
			rechargeResultData.Headimgurl = info.Headimgurl
			rechargeResultData.TotalFee = info.TotalFee
			rechargeResultData.RoomCards = info.RoomCards
			rechargeResultData.UpdateAt = time.Now().Unix()

			saveRechargeResultData(rechargeResultData)
		}
	})
}
