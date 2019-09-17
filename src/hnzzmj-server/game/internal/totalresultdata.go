package internal

import (
	"fmt"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/util"
	"time"
)

type TotalResultData struct {
	ID             int "_id"
	UserID         int
	RoomType       int    // 房间类型 0 练习、1 房卡匹配、2 私人
	RoomNumber     string // 房号
	RoomDesc       string // 房间描述
	StartTimestamp int64  // 开始时间
	EndTimestamp   int64  // 结束时间
	Position       int
	Results        []PlayerResultData
	CreatedAt      int64
	UpdatedAt      int64
}

type PlayerResultData struct {
	UserID         int
	Nickname       string // 昵称
	Score          int    // 分数
	RoomCards      int    // (房卡匹配场有效)
	TotalRoomCards int    // 玩家房卡总和
}

func (data *TotalResultData) initValue(userID int) error {
	id, err := mongoDBNextSeq("totalresult")
	if err != nil {
		return fmt.Errorf("get next totalresult id error: %v", err)
	}
	data.ID = id
	data.UserID = userID
	data.CreatedAt = time.Now().Unix()
	return nil
}

func saveTotalResultData(totalResultData *TotalResultData) {
	data := util.DeepClone(totalResultData)
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		id := data.(*TotalResultData).ID
		_, err := db.DB(DB).C("totalresult").UpsertId(id, data)
		if err != nil {
			log.Error("save totalresult %v data error: %v", id, err)
		}
	}, nil)
}
