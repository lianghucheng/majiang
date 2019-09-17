package internal

import (
	"fmt"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/util"
)

// 房卡分享数据
type ShareRoomCardData struct {
	ID             int "_id"
	AccountID      int
	Nickname       string
	GiftRoomCards  int // 分享获得的房卡
	TotalRoomCards int // 分享之后的房卡数
	CreateAt       int64
	UpdateAt       int64
}

func (data *ShareRoomCardData) initValue() error {
	id, err := mongoDBNextSeq("shareroomcard")
	if err != nil {
		return fmt.Errorf("get next shareroomcard id error: %v", err)
	}
	data.ID = id
	return nil
}

func saveShareRoomCardData(shareRoomCardData *ShareRoomCardData) {
	data := util.DeepClone(shareRoomCardData)
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)

		id := data.(*ShareRoomCardData).ID
		_, err := db.DB(DB).C("shareroomcard").UpsertId(id, data)
		if err != nil {
			log.Error("save shareroomcard %v data error: %v", id, err)
		}
	}, nil)
}
