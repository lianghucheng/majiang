package internal

import (
	"fmt"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/util"
	"time"
)

type TransferRoomCardData struct {
	ID             int "_id"
	FromAccountID  int
	FromNickName   string
	FromHeadimgurl string
	FromRole       int
	ToAccountID    int
	ToNickName     string
	ToHeadimgurl   string
	ToRole         int
	RoomCards      int // 房卡数量
	CreatedAt      int64
	UpdatedAt      int64
}

func (data *TransferRoomCardData) initValue(roomCards int) error {
	id, err := mongoDBNextSeq("transferroomcard")
	if err != nil {
		return fmt.Errorf("get next totalresult id error: %v", err)
	}
	data.ID = id
	data.RoomCards = roomCards
	data.CreatedAt = time.Now().Unix()
	return nil
}

func saveTransferRoomCardData(transferRoomCardData *TransferRoomCardData) {
	data := util.DeepClone(transferRoomCardData)
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		id := data.(*TransferRoomCardData).ID
		_, err := db.DB(DB).C("transferroomcard").UpsertId(id, data)
		if err != nil {
			log.Error("save transferroomcard %v data error: %v", id, err)
		}
	}, nil)
}
