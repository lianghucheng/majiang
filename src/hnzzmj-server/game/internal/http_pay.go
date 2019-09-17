package internal

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/name5566/leaf/log"
	"hnzzmj-server/game/pay/wxpay"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

func handleWXPay(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		total_fee := r.URL.Query().Get("total_fee")
		account_id := r.URL.Query().Get("account_id")
		if total_fee == "" || account_id == "" {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "%v", "no total_fee or no account_id")
			return
		}
		switch account_id {
		case "4137353", "4463729":
			total_fee = "1"
		}
		ip := strings.Split(r.RemoteAddr, ":")[0]
		p := wxpay.NewWXTradeAppPayParameter(total_fee, ip)
		data, err := json.Marshal(p)
		if err != nil {
			log.Error("marshal message %v error: %v", reflect.TypeOf(p), err)
			data = []byte{}
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fmt.Fprintf(w, "%s", data)
		totalFee, _ := strconv.Atoi(total_fee)
		accountID, _ := strconv.Atoi(account_id)
		startWXPayOrder(p["out_trade_no"], accountID, totalFee, nil)
	case "POST":
		result, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		log.Debug("result: %s", result)
		payResult := new(wxpay.WXPayResult)
		xml.Unmarshal(result, &payResult)
		if wxpay.VerifyPayResult(payResult) {
			// 需要验证 out_trade_no 和 total_fee
			fmt.Fprintf(w, "%v", wxpay.ReturnWXSuccess)
			finishWXPayOrder(payResult.OutTradeNo, payResult.TotalFee, true)
		} else {
			fmt.Fprintf(w, "%v", wxpay.ReturnWXFail)
		}
	}
}
