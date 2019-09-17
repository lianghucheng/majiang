package internal

import (
	"encoding/json"
	"fmt"
	"gdmj-server/conf"
	"net/http"
	"reflect"

	"github.com/name5566/leaf/log"
)

func init() {
	go startHTTPServer()
}

func startHTTPServer() {
	mux := http.NewServeMux()
	mux.Handle("/gd/android", http.HandlerFunc(handleGDAndroid))
	mux.Handle("/gd/ios", http.HandlerFunc(handleGDIOS))
	mux.Handle("/wxpay", http.HandlerFunc(handleWXPay))
	err := http.ListenAndServe(conf.Server.HTTPAddr, mux)
	if err != nil {
		log.Fatal("%v", err)
	}
}

func handleGDAndroid(w http.ResponseWriter, req *http.Request) {
	m := map[string]interface{}{}
	m["version"] = gdConfigData.AndroidVersion
	m["downloadurl"] = gdConfigData.AndroidDownloadUrl
	m["guestlogin"] = gdConfigData.AndroidGuestLogin
	m["online"] = len(userIDUsers)
	data, err := json.Marshal(m)
	if err != nil {
		log.Error("marshal message %v error: %v", reflect.TypeOf(m), err)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*") //解决跨域问题
	fmt.Fprintf(w, "%s", data)
}

func handleGDIOS(w http.ResponseWriter, req *http.Request) {
	m := map[string]interface{}{}
	m["version"] = gdConfigData.IOSVersion
	m["downloadurl"] = gdConfigData.IOSDownloadUrl
	m["guestlogin"] = gdConfigData.IOSGuestLogin
	m["online"] = len(userIDUsers)
	data, err := json.Marshal(m)
	if err != nil {
		log.Error("marshal message %v error: %v", reflect.TypeOf(m), err)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprintf(w, "%s", data)
}
