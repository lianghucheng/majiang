package internal

import (
	"gdmj-server/msg"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/util"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var gdConfigData *ConfigData

type ConfigData struct {
	ID   int "_id"
	Game string
	msg.C2S_SetGDConfig
}

const (
	defaultGDAndroidDownloadUrl = "https://www.shenzhouxing.com/gdmj/dl/"
	defaultGDIOSDownloadUrl     = "https://www.shenzhouxing.com/gdmj/dl/"
)

func (data *ConfigData) initGD() {
	data.Game = "广东麻将"

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
	data.AndroidVersion = 1
	data.AndroidDownloadUrl = defaultGDAndroidDownloadUrl
	data.IOSVersion = 1
	data.IOSDownloadUrl = defaultGDIOSDownloadUrl
	data.AndroidGuestLogin = false
	data.IOSGuestLogin = false
	data.Notice = "广东麻将"
	data.Radio = "请各位玩家文明游戏，禁止赌博，未成年人勿过度沉迷"
	data.WeChatNumber = "yintan19"
	saveConfigData(data)
}

func saveConfigData(configData *ConfigData) {
	data := util.DeepClone(configData)
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
