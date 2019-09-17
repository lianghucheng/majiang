package internal

import (
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/util"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"yananmj-server/msg"
)

var yananConfigData *ConfigData

type ConfigData struct {
	ID   int "_id"
	Game string
	data_struct.C2S_SetYananConfig
}

const (
	defaultYananAndroidDownloadUrl = "https://www.shenzhouxing.com/yanan/dl/"
	defaultYananIOSDownloadUrl     = "https://www.shenzhouxing.com/yanan/dl/"
)

func (data *ConfigData) initYanan() {
	data.Game = "延安麻将"

	db := mongoDB.Ref()
	defer mongoDB.UnRef(db)

	err := db.DB(DB).C("configs").Find(bson.M{"game": data.Game}).One(data)
	if err == nil {
		return
	}
	if err != mgo.ErrNotFound {
		log.Fatal("init: %v config data error: %v", data.Game, err)
		return
	}
	id, err := mongoDBNextSeq("configs")
	if err != nil {
		log.Fatal("get next config id error:%v", err)
		return
	}

	data.ID = id
	data.AndriodVersion = 2
	data.AndriodDownloadUrl = defaultYananAndroidDownloadUrl
	data.IOSVersion = 2
	data.IOSDownloadUrl = defaultYananIOSDownloadUrl
	data.AndriodGuestLogin = false
	data.IOSGuestLogin = false
	data.Notice = "房卡兑奖活动时间为11.17-12.17，快快进入房卡比赛场，赢取更多房卡，兑换大奖吧！兑奖请联系唯一官方微信：yintan19。同时关注“银滩雀神大赛”，及时了解获取“现金大奖赛”的最新消息！"
	data.Radio = "诚招代理，咨询详情请加微信: yintan19"
	data.WeChatNumber = "yintan19"
	saveConfigData(data)
}

func saveConfigData(configdata *ConfigData) {
	data := util.DeepClone(configdata)
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		id := data.(*ConfigData).ID
		_, err := db.DB(DB).C("configs").UpsertId(id, data)
		if err != nil {
			log.Fatal("save %v config data error:%v", id, err)
		}
	}, func() {})
}
