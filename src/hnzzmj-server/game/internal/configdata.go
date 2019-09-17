package internal

import (
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/util"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"hnzzmj-server/msg"
)

var hnzzConfigData *ConfigData

type ConfigData struct {
	ID   int "_id"
	Game string
	msg.C2S_SetHNZZConfig
}

const (
	defaultHNZZAndroidDownloadUrl = "https://www.shenzhouxing.com/hnzz/dl/"
	defaultHNZZIOSDownloadUrl     = "https://www.shenzhouxing.com/hnzz/dl/"
)

func (data *ConfigData) initHNZZ() {
	data.Game = "湖南转转麻将"

	db := mongoDB.Ref()
	defer mongoDB.UnRef(db)
	err := db.DB(DB).C("configs").
		Find(bson.M{"game": data.Game}).One(data)
	if err == nil {
		return
	}
	if err != mgo.ErrNotFound {
		log.Error("init %v config data error: %v", data.Game, err)
		return
	}
	id, err := mongoDBNextSeq("configs")
	if err != nil {
		log.Error("get next configs id error: %v", err)
		return
	}
	data.ID = id
	data.AndroidVersion = 4
	data.AndroidDownloadUrl = defaultHNZZAndroidDownloadUrl
	data.IOSVersion = 1
	data.IOSDownloadUrl = defaultHNZZIOSDownloadUrl
	data.AndroidGuestLogin = false
	data.IOSGuestLogin = false
	data.Notice = `房卡兑奖活动将在2017年11月30号日结束。
多多参与房卡比赛场，赢取更多房卡吧。
关注微信公众号“银滩雀神大赛”及时获取最新活动资讯！
兑奖请联系客服微信：yintan66`
	data.Radio = "请各位玩家文明游戏，未成年人勿过度沉迷"
	data.WeChatNumber = "yintan66"
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
			log.Error("save %v config data error: %v", id, err)
		}
	}, nil)
}
