package songs

import (
	"encoding/json"
	"fmt"
	"go-wyy/models"
	"go-wyy/service/encrypt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

/**
ids: [歌曲id]
rate :320000 普通品质
	  640000 高级品质
      160000 低级品质
*/
func GetDownloadUrl(id string, rate string) (data *models.DownloadData, err error) {
	var DownloadData *models.DownloadData
	initStr := `{"ids": "` + "[" + id + "]" + `", "br": "` + rate + `", "csrf_token": ""}`
	params, key, err := encrypt.EncParams(initStr)
	if err != nil {
		panic(err)
	}
	DownloadData, err = Download(params, key)
	return DownloadData, err
}

// Download 根据传入id返回生成的mp3地址
func Download(params string, encSecKey string) (data *models.DownloadData, err error) {
	var DownloadData *models.DownloadData
	client := &http.Client{}
	form := url.Values{}
	form.Set("params", params)
	form.Set("encSecKey", encSecKey)
	body := strings.NewReader(form.Encode())
	time.Sleep(500 * time.Microsecond)
	request, _ := http.NewRequest("POST", "http://music.163.com/weapi/song/enhance/player/url?csrf_token=", body)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Content-Length", (string)(body.Len()))
	request.Header.Set("Referer", "http://music.163.com")
	// 发起请求
	response, reqErr := client.Do(request)
	// 错误处理
	if reqErr != nil {
		fmt.Println("Fatal error ", reqErr.Error())
		return DownloadData, err
	}
	defer response.Body.Close()
	resBody, _ := ioutil.ReadAll(response.Body)
	err = json.Unmarshal(resBody, &DownloadData)
	if err != nil {
		panic(err)
	}
	return DownloadData, err
}
