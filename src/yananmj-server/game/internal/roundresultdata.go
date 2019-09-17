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

func (data *RoundResultData) initValue(userID int) error {
	id, err := mongoDBNextSeq("roundresult")
	if err != nil {
		return fmt.Errorf("get next roundresult error: %v", err)
	}

	data.ID = id
	data.TotalResultID = userID
	data.CreatedAt = time.Now().Unix()
	return nil
}

func saveRoundResultData(resultdata *RoundResultData) {
	data := util.DeepClone(resultdata)
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)

		id := data.(*RoundResultData).ID
		_, err := db.DB(DB).C("roundresult").UpsertId(id, data)
		if err != nil {
			log.Fatal("save: %v roundresult error: %v", id, err)
		}
	}, func() {

	})
}
