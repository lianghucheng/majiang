package internal

import (
	"fmt"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/util"
	"time"
)

type RoundResultData struct {
	ID             int "_id"
	TotalResultID  int
	Round          int
	StartTimestamp int64 // 开始时间
	EndTimestamp   int64 // 结束时间
	Position       int
	Results        []PlayerResultData
	CreatedAt      int64
	UpdatedAt      int64
}

func (data *RoundResultData) initValue(totalResultID int) error {
	id, err := mongoDBNextSeq("roundresult")
	if err != nil {
		return fmt.Errorf("get next roundresult id error: %v", err)
	}
	data.ID = id
	data.TotalResultID = totalResultID
	data.CreatedAt = time.Now().Unix()
	return nil
}

func saveRoundResultData(roundResultData *RoundResultData) {
	data := util.DeepClone(roundResultData)
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		id := data.(*RoundResultData).ID
		_, err := db.DB(DB).C("roundresult").UpsertId(id, data)
		if err != nil {
			log.Error("save user %v data error: %v", id, err)
		}
	}, func() {

	})
}
