package internal

import (
	"fmt"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/util"
	"time"
)

type AliRechargeResultData struct {
	ID         int "_id"
	AccountID  int
	OutTradeNo string // 商户订单号
	TotalAmout string // 订单金额
	RoomCards  int    // 充房卡数
	Success    bool   // 是否充值成功
	CreateAt   int64
	UpdateAt   int64
}

func (data *AliRechargeResultData) initValue() error {
	id, err := mongoDBNextSeq("alirechargeresult")
	if err != nil {
		return fmt.Errorf("get next alirechargeresult id error: %v", err)
	}
	data.ID = id
	data.CreateAt = time.Now().Unix()
	return nil
}

func saveAliRechargeResultData(aliRechargeResultData *AliRechargeResultData) {
	data := util.DeepClone(aliRechargeResultData)
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		id := data.(*RechargeResultData).ID
		_, err := db.DB(DB).C("alirechargeresult").UpsertId(id, data)
		if err != nil {
			log.Error("save rechargeresult %v data error: %v", id, err)
		}
	}, nil)
}

func updateAliRechargeData(accountID int, update interface{}) {
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		_, err := db.DB(DB).C("alirechargeresult").Upsert(accountID, update)
		if err != nil {
			log.Error("update user %v data error: %v", accountID, err)
		}
	}, nil)
}
