package internal

import (
	"github.com/name5566/leaf/log"
	"gopkg.in/mgo.v2/bson"
)

type RedPacketMatchResultData struct {
	ID            bson.ObjectId `bson:"_id"`
	UserID        int
	RedPacketType int     // 红包种类(元): 1、5、10、50
	RedPacket     float64 // 红包奖励
	Taken         bool    // 奖励是否被领取
	Handling      bool    // 处理中
	CreatedAt     int64
	UpdatedAt     int64
}

func saveRedPacketMatchResultData(resultData *RedPacketMatchResultData) {
	temp := &struct {
		UserID        int
		RedPacketType int     // 红包种类(元): 1、5、10、50
		RedPacket     float64 // 红包奖励
		Taken         bool    // 是否领取
		CreatedAt     int64
	}{}
	temp.UserID = resultData.UserID
	temp.RedPacketType = resultData.RedPacketType
	temp.RedPacket = resultData.RedPacket
	temp.Taken = resultData.Taken
	temp.CreatedAt = resultData.CreatedAt
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		err := db.DB(DB).C("redpacketmatchresult").Insert(temp)
		if err != nil {
			log.Error("insert redpacketmatchresult data error: %v", err)
		}
	}, nil)
}

func updateRedPacketMatchResultData(id bson.ObjectId, update interface{}, cb func()) {
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		_, err := db.DB(DB).C("redpacketmatchresult").UpsertId(id, update)
		if err != nil {
			log.Error("update redpacketmatchresult %v data error: %v", id, err)
		}
	}, func() {
		if cb != nil {
			cb()
		}
	})
}
