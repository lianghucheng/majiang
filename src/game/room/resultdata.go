package room

import (
	. "db"
	"fmt"
	. "game"
	"time"
	"util"

	"github.com/name5566/leaf/log"
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
	RedPacketType  int    // 红包种类(元): 1、5、10、50、100、200
	TotalRoomCards int    // 玩家房卡总和
}

func (data *TotalResultData) initValue(userID int) error {
	id, err := MongoDBNextSeq("totalresult")
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
	Skeleton.Go(func() {
		db := MongoDB.Ref()
		defer MongoDB.UnRef(db)
		id := data.(*TotalResultData).ID
		_, err := db.DB(DB).C("totalresult").UpsertId(id, data)
		if err != nil {
			log.Error("save totalresult %v data error: %v", id, err)
		}
	}, nil)
}

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
	id, err := MongoDBNextSeq("roundresult")
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
	Skeleton.Go(func() {
		db := MongoDB.Ref()
		defer MongoDB.UnRef(db)
		id := data.(*RoundResultData).ID
		_, err := db.DB(DB).C("roundresult").UpsertId(id, data)
		if err != nil {
			log.Error("save user %v data error: %v", id, err)
		}
	}, func() {

	})
}

func (r *GDRoom) SaveUserTotalResultData(results []PlayerResultData) {
	for pos := 0; pos < r.Rule.MaxPlayers; pos++ {
		userID := r.PositionUserIDs[pos]
		playerData := r.Useridplayerdatas[userID]
		if playerData.TotalResultData != nil {
			playerData.TotalResultData.RoomType = r.Rule.RoomType
			playerData.TotalResultData.EndTimestamp = r.EndTimestamp
			playerData.TotalResultData.Results = results
			playerData.TotalResultData.UpdatedAt = time.Now().Unix()

			saveTotalResultData(playerData.TotalResultData)
		}
	}
}

func (r *GDRoom) SaveUserRoundResultData(round int, results []PlayerResultData) {
	for pos := 0; pos < r.Rule.MaxPlayers; pos++ {
		userID := r.PositionUserIDs[pos]
		playerData := r.Useridplayerdatas[userID]
		if playerData.TotalResult != nil {
			playerData.RoundResultData = new(RoundResultData)
			err := playerData.RoundResultData.initValue(playerData.TotalResultData.ID)
			if err != nil {
				log.Error("init totalresult %v round result data error: %v", playerData.TotalResultData.ID, err)
				playerData.RoundResultData = nil
				return
			}
			playerData.RoundResultData.Round = round
			playerData.RoundResultData.StartTimestamp = r.EachRoundStartTimestamp
			playerData.RoundResultData.EndTimestamp = r.EndTimestamp
			playerData.RoundResultData.Position = playerData.Position
			playerData.RoundResultData.Results = results
			playerData.RoundResultData.UpdatedAt = time.Now().Unix()

			saveRoundResultData(playerData.RoundResultData)
		}
	}

}

func (r *GDRoom) initTotalResultData() {
	for _, userID := range r.PositionUserIDs {
		playerData := r.Useridplayerdatas[userID]
		playerData.TotalResultData = new(TotalResultData)
		err := playerData.TotalResultData.initValue(playerData.User.UserData.UserID)
		if err != nil {
			log.Error("init userID %v totalresult data error: %v", playerData.User.UserData.UserID, err)
			playerData.TotalResultData = nil
		}
		playerData.TotalResultData.UserID = playerData.User.UserData.UserID
		playerData.TotalResultData.RoomNumber = r.Number
		playerData.TotalResultData.RoomDesc = r.Desc
		playerData.TotalResultData.StartTimestamp = r.StartTimestamp
		playerData.TotalResultData.Position = playerData.Position
	}

}
