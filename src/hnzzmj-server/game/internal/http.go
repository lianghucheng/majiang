package internal

import (
	"encoding/json"
	"fmt"
	"github.com/name5566/leaf/log"
	"hnzzmj-server/conf"
	"net/http"
	"reflect"
)

func init() {
	go startHTTPServer()
}

func startHTTPServer() {
	mux := http.NewServeMux()
	mux.Handle("/hnzz/android", http.HandlerFunc(handleHNZZAndroid))
	mux.Handle("/hnzz/ios", http.HandlerFunc(handleHNZZIOS))
	// mux.Handle("/price", http.HandlerFunc(handlePrice))
	mux.Handle("/wxpay", http.HandlerFunc(handleWXPay))
	err := http.ListenAndServe(conf.Server.HTTPAddr, mux)
	if err != nil {
		log.Fatal("%v", err)
	}
}

func handleHNZZAndroid(w http.ResponseWriter, req *http.Request) {
	m := map[string]interface{}{}
	m["version"] = hnzzConfigData.AndroidVersion
	m["downloadurl"] = hnzzConfigData.AndroidDownloadUrl
	m["guestlogin"] = hnzzConfigData.AndroidGuestLogin
	m["online"] = len(userIDUsers)
	data, err := json.Marshal(m)
	if err != nil {
		log.Error("marshal message %v error: %v", reflect.TypeOf(m), err)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*") //解决跨域问题
	fmt.Fprintf(w, "%s", data)
}

func handleHNZZIOS(w http.ResponseWriter, req *http.Request) {
	m := map[string]interface{}{}
	m["version"] = hnzzConfigData.IOSVersion
	m["downloadurl"] = hnzzConfigData.IOSDownloadUrl
	m["guestlogin"] = hnzzConfigData.IOSGuestLogin
	m["online"] = len(userIDUsers)
	data, err := json.Marshal(m)
	if err != nil {
		log.Error("marshal message %v error: %v", reflect.TypeOf(m), err)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprintf(w, "%s", data)
}
