package internal

import (
	"encoding/json"
	"fmt"
	"github.com/name5566/leaf/log"
	"net/http"
	"reflect"
	"yananmj-server/conf"
)

func init() {
	go startHTTPServer()
}

func startHTTPServer() {
	mux := http.NewServeMux()
	mux.Handle("/yanan/android", http.HandlerFunc(handleYananAndroid))
	mux.Handle("/yanan/ios", http.HandlerFunc(handleYananIOS))
	mux.Handle("/czddz/wechatpay", http.HandlerFunc(handleWeChatPay))
	mux.Handle("/czddz/alipay", http.HandlerFunc(handleAliPay))
	mux.Handle("/wxpay", http.HandlerFunc(handleWXPay))
	err := http.ListenAndServe(conf.Server.HTTPAddr, mux)
	if err != nil {
		log.Fatal("%v", err)
	}
}

func handleYananAndroid(w http.ResponseWriter, req *http.Request) {
	m := map[string]interface{}{}
	m["version"] = yananConfigData.AndriodVersion
	m["downloadurl"] = yananConfigData.AndriodDownloadUrl
	m["guestlogin"] = yananConfigData.AndriodGuestLogin

	data, err := json.Marshal(m)
	if err != nil {
		log.Error("marshal message %v error: %v", reflect.TypeOf(m), err)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*") //解决跨域问题
	fmt.Fprintf(w, "%s", data)
}

func handleYananIOS(w http.ResponseWriter, req *http.Request) {
	m := map[string]interface{}{}
	m["version"] = yananConfigData.IOSVersion
	m["downloadurl"] = yananConfigData.IOSDownloadUrl
	m["guestlogin"] = yananConfigData.IOSGuestLogin

	data, err := json.Marshal(m)
	if err != nil {
		log.Error("marshal message %v error: %x", reflect.TypeOf(m), err)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprintf(w, "%s", data)
}
