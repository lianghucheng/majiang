package songs

import (
	"github.com/PuerkitoBio/goquery"
	"go-wyy/models"
	"go-wyy/service/comment"
	"log"
	"net/http"
	"sync"
	"time"
)

/**
userId 用户Id

*/
func Songs(userId string) {
	req, err := http.NewRequest("GET", "http://music.163.com/playlist?id="+userId, nil)

	if err != nil {
		panic(err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.84 Safari/537.36")
	req.Header.Set("Referer", "http://music.163.com/")
	req.Header.Set("Host", "music.163.com")

	c := &http.Client{}
	res, err := c.Do(req)
	if err != nil {
		panic(err)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)

	if err != nil {
		panic(err)
	}

	g := 0
	wg := &sync.WaitGroup{}

	doc.Find("ul[class=f-hide] a").Each(func(i int, selection *goquery.Selection) {
		/*开启协程插入数据库，并且开启协程请求每首歌的评论*/
		songIdUrl, _ := selection.Attr("href")
		title := selection.Text()
		var song models.Song
		//歌曲id
		songId := songIdUrl[9:len(songIdUrl)]
		song.SongId = songId

		///song?id=歌曲id
		song.SongUrlId = songIdUrl

		//歌曲标题
		song.Title = title

		//获取歌曲下载链接
		//download, err := GetDownloadUrl(songId, "320000")
		//if err != nil {
		//	panic(err)
		//}
		//if len(download.Data) != 0 {
		//	song.DownloadUrl = download.Data[0].Url
		//}
		//fmt.Println(download.Data[0].Url)

		song.SongId = songId
		//song.DownloadUrl = download.Data[0].Url
		song.Title = title
		song.SongUrlId = songIdUrl

		song.UserId = userId

		//if err := models.DB.Create(&song).Error; err != nil {
		//	beego.Debug(err)
		//}
		log.Printf("正在获取第%d首歌曲", i+1)
		log.Printf("正在开启%d个协程", g+1)
		if i%200 == 0 && i >= 200 {
			time.Sleep(450 * time.Second)
		} else {
			go comment.GetAllComment(songId, wg)
		}

		g++
		wg.Add(1)
	})
	wg.Wait()
}
