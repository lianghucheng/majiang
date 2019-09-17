package internal

import (
	"fmt"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/util"
	"time"
)

//售卡记录数据
type TransferRoomCardData struct {
	ID             int    "_id"
	FromAccountID  int    // 玩家ID
	FromNickName   string // 玩家昵称
	FromHeadimgurl string // 用户头像
	FromRole       int    // 代理级别
	ToAccountID    int    // 被转玩家ID
	ToNickName     string // 被转玩家昵称
	ToHeadimgurl   string // 被转玩家头像
	ToRole         int    // 被转玩家角色
	RoomCards      int    // 转给玩家的卡数
	CreatedAt      int64
	UpdatedAt      int64
}

func (data *TransferRoomCardData) initValue(roomCards int) error {
	id, err := mongoDBNextSeq("transferroomcard")
	if err != nil {
		return fmt.Errorf("get next transferroomcard id error: %v", err)
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
