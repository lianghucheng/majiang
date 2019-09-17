package internal

import (
	"fmt"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/util"
	"time"
)

type RechargeResultData struct {
	ID         int "_id"
	AccountID  int
	Nickname   string // 昵称
	Headimgurl string // 头像
	OutTradeNo string // 商户订单号
	TotalFee   int    // 充值总金额
	RoomCards  int    // 充房卡数
	CreateAt   int64
	UpdateAt   int64
}

func (data *RechargeResultData) initValue(roomcards int) error {
	id, err := mongoDBNextSeq("wechatrechargeresult")
	if err != nil {
		return fmt.Errorf("get next rechargeresult id error: %v", err)
	}
	data.ID = id
	data.RoomCards = roomcards
	data.CreateAt = time.Now().Unix()
	return nil
}

func saveRechargeResultData(rechargeResultData *RechargeResultData) {
	data := util.DeepClone(rechargeResultData)
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		id := data.(*RechargeResultData).ID
		_, err := db.DB(DB).C("wechatrechargeresult").UpsertId(id, data)
		if err != nil {
			log.Error("save rechargeresult %v data error: %v", id, err)
		}
	}, nil)
}
