package conf

import (
	. "db"
	"game"

	"github.com/name5566/leaf/log"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func init() {
	GdConfigData = new(ConfigData)
	GdConfigData.initGD()
}

var GdConfigData *ConfigData

type ConfigData struct {
	ID                 int "_id"
	Game               string
	AndroidVersion     int    // Android 版本号
	AndroidDownloadUrl string // Android 下载链接
	IOSVersion         int    // iOS 版本号
	IOSDownloadUrl     string // iOS 下载链接
	AndroidGuestLogin  bool   // Android 游客登录
	IOSGuestLogin      bool   // iOS 游客登录
	Notice             string // 公告
	Radio              string // 广播
	WeChatNumber       string // 客服微信号
}

const (
	defaultGDAndroidDownloadUrl = "https://www.shenzhouxing.com/gdmj/dl/"
	defaultGDIOSDownloadUrl     = "https://www.shenzhouxing.com/gdmj/dl/"
)

func (data *ConfigData) initGD() {
	data.Game = "广东麻将"

	db := MongoDB.Ref()
	defer MongoDB.UnRef(db)
	err := db.DB(DB).C("configs").
		Find(bson.M{"game": data.Game}).One(data)
	if err == nil {
		return
	}
	if err != mgo.ErrNotFound {
		log.Error("init %v config data error: %v", data.Game, err)
		return
	}
	id, err := MongoDBNextSeq("configs")
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
	game.Skeleton.Go(func() {
		db := MongoDB.Ref()
		defer MongoDB.UnRef(db)
		_, err := db.DB(DB).C("configs").UpsertId(configData.ID, configData)
		if err != nil {
			log.Error("save %v config data error: %v", configData.ID, err)
		}
	}, nil)
}
