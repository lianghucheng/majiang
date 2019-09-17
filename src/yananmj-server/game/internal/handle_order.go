package internal

import (
	"github.com/name5566/leaf/log"
	"gopkg.in/mgo.v2/bson"
	"time"
	"yananmj-server/msg"
)

// 验证用户是否存在，存在则存储订单信息
func startWXPayOrder(outTradeNo string, accountID, totalFee int, cb func()) {
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)

		userData := new(UserData)
		err := db.DB(DB).C("users").Find(bson.M{"accountid": accountID}).One(userData)
		if err != nil {
			log.Debug("find accountID %v error: %v", accountID, err)
			return
		}
		temp := &struct {
			UserID     int
			OutTradeNo string
			Success    bool
			TotalFee   int
			CreatedAT  int64
		}{}
		temp.UserID = userData.UserID
		temp.OutTradeNo = outTradeNo
		temp.TotalFee = totalFee
		temp.CreatedAT = time.Now().Unix()
		_, err = db.DB(DB).C("wxpayresult").Upsert(bson.M{"outtradeno": outTradeNo}, bson.M{"$set": temp})
		if err != nil {
			log.Debug("upsert userID: %v error: %v", userData.UserID, err)
		}
	}, func() {
		if cb != nil {
			cb()
		}
	})
}

func finishWXPayOrder(outTradeNo string, totalFee int, valid bool) {
	temp := &struct {
		UserID     int
		OutTradeNo string
		Success    bool
		TotalFee   int
		Valid      bool
		UpdatedAt  int64
	}{}
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)

		err := db.DB(DB).C("wxpayresult").Find(bson.M{"outtradeno": outTradeNo, "success": false}).One(&temp)
		if err != nil {
			temp = nil
			log.Debug("find out_trade_no: %v error: %v", temp.OutTradeNo, err)
			return
		}
		if temp.TotalFee == totalFee {
			temp.Success = true
			temp.Valid = valid
			temp.UpdatedAt = time.Now().Unix()
			err = db.DB(DB).C("wxpayresult").Update(bson.M{"outtradeno": temp.OutTradeNo, "success": false}, bson.M{"$set": temp})
			if err != nil {
				log.Debug("update out_trade_no: %v error: %v", temp.OutTradeNo, err)
				temp = nil
			}
		} else {
			temp = nil
		}
	}, func() {
		if temp == nil {
			return
		}
		addRoomCards := temp.TotalFee / 100
		switch temp.TotalFee {
		case 3000:
			addRoomCards = 33
		case 5000:
			addRoomCards = 60
		}
		if user, ok := userIDUsers[temp.UserID]; ok {
			user.WriteMsg(&data_struct.S2C_PayOK{
				RoomCards: addRoomCards,
			})
			user.data.userData.RoomCards += addRoomCards
			user.WriteMsg(&data_struct.S2C_UpdateRoomCards{
				RoomCards: user.data.userData.RoomCards,
			})
			if user.isRobot() {
				upsertRobotData(time.Now().Format("20060102"), bson.M{"$inc": bson.M{"recharge": addRoomCards}})
			}
		} else {
			updateUserData(temp.UserID, bson.M{"$inc": bson.M{"roomcards": addRoomCards}})
		}
	})
}
