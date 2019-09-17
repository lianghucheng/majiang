package internal

import (
	"github.com/name5566/leaf/log"
	"gopkg.in/mgo.v2/bson"
	"time"
	"yananmj-server/game/mahjong"
	"yananmj-server/msg"
)

//转房卡
func (user *User) transferRoomCard(accountID int, roomCards int) {
	otherUserData := new(UserData)

	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		//load
		db.DB(DB).C("users").
			Find(bson.M{"accountid": accountID}).One(otherUserData)

	}, func() {
		if user.state == userLogout {
			return
		}

		if otherUserData.UserID == 0 {
			log.Debug("账户ID: %v的userID不存在", accountID)
			user.WriteMsg(&data_struct.S2C_TransferRoomCard{
				Error: data_struct.S2C_TransferRoomCard_AccountIDInvalid,
			})
			return
		}
		if otherUser, ok := userIDUsers[otherUserData.UserID]; ok {
			otherUser.data.userData.RoomCards += roomCards
			otherUser.data.userData.PurchasedRoomCards += roomCards
			user.saveTransferRoomCardData(roomCards, &data_struct.TransferRoomCardUserInfo{
				FromAccountID:  user.data.userData.AccountID,
				FromNickName:   user.data.userData.Nickname,
				FromHeadimgurl: user.data.userData.Headimgurl,
				FromRole:       user.data.userData.Role,
				ToAccountID:    otherUser.data.userData.AccountID,
				ToNickName:     otherUser.data.userData.Nickname,
				ToHeadimgurl:   otherUser.data.userData.Headimgurl,
				ToRole:         otherUser.data.userData.Role,
			})
			//updateUserData(otherUser.data.UserID, bson.M{"$set": bson.M{"roomcards": otherUser.data.RoomCards, "purchasedroomcards": otherUser.data.PurchasedRoomCards}})
			otherUser.WriteMsg(&data_struct.S2C_UpdateRoomCards{
				RoomCards: otherUser.data.userData.RoomCards,
			})
		} else {
			otherUserData.RoomCards += roomCards
			otherUserData.PurchasedRoomCards += roomCards
			user.saveTransferRoomCardData(roomCards, &data_struct.TransferRoomCardUserInfo{
				FromAccountID:  user.data.userData.AccountID,
				FromNickName:   user.data.userData.Nickname,
				FromHeadimgurl: user.data.userData.Headimgurl,
				FromRole:       user.data.userData.Role,
				ToAccountID:    otherUserData.AccountID,
				ToNickName:     otherUserData.Nickname,
				ToHeadimgurl:   otherUserData.Headimgurl,
				ToRole:         otherUserData.Role,
			})
			updateUserData(otherUserData.UserID, bson.M{"$set": bson.M{"roomcards": otherUserData.RoomCards, "purchasedroomcards": otherUserData.PurchasedRoomCards}})
		}

		user.data.userData.RoomCards -= roomCards
		user.data.userData.SaleRoomCardNumber += roomCards
		//updateUserData(user.data.userData.UserID, bson.M{"$set": bson.M{"roomcards": user.data.userData.RoomCards, "saleroomcards": user.data.userData.SaleRoomCardNumber}})
		user.WriteMsg(&data_struct.S2C_UpdateRoomCards{
			RoomCards: user.data.userData.RoomCards,
		})
		user.WriteMsg(&data_struct.S2C_TransferRoomCard{
			Error:     data_struct.S2C_TransferRoomCard_OK,
			RoomCards: roomCards,
		})
		log.Debug("userID %v 给账号ID: %v 转了 %v张房卡", user.data.userData.UserID, accountID, roomCards)
	})
}

func (user *User) saveTransferRoomCardData(roomCards int, info *data_struct.TransferRoomCardUserInfo) {
	transferRoomCardData := new(TransferRoomCardData)
	skeleton.Go(func() {
		err := transferRoomCardData.initValue(roomCards)
		if err != nil {
			log.Error("init transferroomcard data error: %v", err)
			transferRoomCardData = nil
		}
	}, func() {
		if transferRoomCardData != nil {
			transferRoomCardData.FromAccountID = info.FromAccountID
			transferRoomCardData.FromNickName = info.FromNickName
			transferRoomCardData.FromHeadimgurl = info.FromHeadimgurl
			transferRoomCardData.FromRole = info.FromRole
			transferRoomCardData.ToAccountID = info.ToAccountID
			transferRoomCardData.ToNickName = info.ToNickName
			transferRoomCardData.ToHeadimgurl = info.ToHeadimgurl
			transferRoomCardData.ToRole = info.ToRole
			transferRoomCardData.UpdatedAt = time.Now().Unix()

			saveTransferRoomCardData(transferRoomCardData)
		}
	})
}

//获取历史总成绩
func (user *User) getTotalResult() {
	totalResultDatas := []TotalResultData{}
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		//load
		db.DB(DB).C("totalresult").
			Find(bson.M{"userid": user.data.userData.UserID}).Sort("-_id").Limit(15).All(&totalResultDatas)
	}, func() {
		totalResults := []data_struct.TotalResult{}
		for _, totalResultData := range totalResultDatas {
			playerResults := []data_struct.PlayerResult{}

			for _, playerResultData := range totalResultData.Results {
				playerResults = append(playerResults, data_struct.PlayerResult{
					Nickname:  playerResultData.Nickname,
					Score:     playerResultData.Score,
					RoomCards: playerResultData.RoomCards,
				})
			}
			result := mahjong.ResultLose
			score := totalResultData.Results[totalResultData.Position].Score
			if score == 0 {
				result = mahjong.ResultDraw
			} else if score > 0 {
				result = mahjong.ResultWin
			}
			startTime := time.Unix(totalResultData.StartTimestamp, 0)
			endTime := time.Unix(totalResultData.EndTimestamp, 0)
			totalResult := data_struct.TotalResult{
				TotalResultID: totalResultData.ID,
				RoomType:      totalResultData.RoomType,
				RoomNumber:    totalResultData.RoomNumber,
				RoomDesc:      totalResultData.RoomDesc,
				Result:        result,
				Duration:      startTime.Format("2006/01/02 15:04:05") + "-" + endTime.Format("15:04:05"),
				Position:      totalResultData.Position,
				PlayerResults: playerResults,
			}
			totalResults = append(totalResults, totalResult)
		}
		user.WriteMsg(&data_struct.S2C_TotalResults{
			Results: totalResults,
		})
	})
}

//获取历史单局成绩
func (user *User) getRoundResult(totalResultID int) {
	roundResultDatas := []RoundResultData{}
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		db.DB(DB).C("roundresult").
			Find(bson.M{"totalresultid": totalResultID}).All(&roundResultDatas)
	}, func() {
		roundResults := []data_struct.RoundResult{}
		for _, roundResultData := range roundResultDatas {
			playerResults := []data_struct.PlayerResult{}

			for _, playerResultData := range roundResultData.Results {
				playerResults = append(playerResults, data_struct.PlayerResult{
					Nickname: playerResultData.Nickname,
					Score:    playerResultData.Score,
				})
			}

			startTime := time.Unix(roundResultData.StartTimestamp, 0)
			endTime := time.Unix(roundResultData.EndTimestamp, 0)

			roundResult := data_struct.RoundResult{
				Position:      roundResultData.Position,
				Round:         roundResultData.Round,
				Duration:      startTime.Format("2006/01/02 15:04:05") + "-" + endTime.Format("15:04:05"),
				PlayerResults: playerResults,
			}
			roundResults = append(roundResults, roundResult)
		}
		user.WriteMsg(&data_struct.S2C_RoundResults{
			Results: roundResults,
		})
	})
}

func (user *User) getUserInfo(accountID int) {
	otherUserData := new(UserData)
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		// load
		db.DB(DB).C("users").
			Find(bson.M{"accountid": accountID}).One(otherUserData)
	}, func() {
		if otherUserData.UserID < 1 {
			log.Debug("账户ID: %v 的用户不存在", accountID)
			user.WriteMsg(&data_struct.S2C_UserInfo{
				Error: data_struct.S2C_UserInfo_AccountIDInvalid,
			})
			return
		}
		if otherUser, ok := userIDUsers[otherUserData.UserID]; ok {
			joinAgencyTime := ""
			if otherUser.data.userData.JoinAgencyAt > 1509465600 {
				joinAgencyTime = time.Unix(otherUser.data.userData.JoinAgencyAt, 0).Format("2006/01/02 15:04:05")
			}
			user.WriteMsg(&data_struct.S2C_UserInfo{
				Error:              data_struct.S2C_UserInfo_OK,
				AccountID:          otherUser.data.userData.AccountID,
				Nickname:           otherUser.data.userData.Nickname,
				Headimgurl:         otherUser.data.userData.Headimgurl,
				Sex:                otherUser.data.userData.Sex,
				RoomCards:          otherUser.data.userData.RoomCards,
				JoinAgencyAt:       joinAgencyTime,
				Role:               otherUser.data.userData.Role,
				GameScore:          otherUser.data.userData.GameScore,
				ConsumedRoomCards:  otherUser.data.userData.ConsumedRoomCards,
				PurchasedRoomCards: otherUser.data.userData.PurchasedRoomCards,
				LastLogin:          time.Unix(otherUser.data.userData.LastLoginAt, 0).Format("2006/01/02 15:04:05"),
			})
		} else {
			joinAgencyTime := ""
			if otherUserData.JoinAgencyAt > 1509465600 {
				joinAgencyTime = time.Unix(otherUserData.JoinAgencyAt, 0).Format("2006/01/02 15:04:05")
			}
			user.WriteMsg(&data_struct.S2C_UserInfo{
				Error:              data_struct.S2C_UserInfo_OK,
				AccountID:          otherUserData.AccountID,
				Nickname:           otherUserData.Nickname,
				Headimgurl:         otherUserData.Headimgurl,
				Sex:                otherUserData.Sex,
				RoomCards:          otherUserData.RoomCards,
				JoinAgencyAt:       joinAgencyTime,
				Role:               otherUserData.Role,
				GameScore:          otherUserData.GameScore,
				ConsumedRoomCards:  otherUserData.ConsumedRoomCards,
				PurchasedRoomCards: otherUserData.PurchasedRoomCards,
				LastLogin:          time.Unix(otherUserData.LastLoginAt, 0).Format("2006/01/02 15:04:05"),
			})
		}
	})
}

func (user *User) getTransferRoomCardRecordByPage(info *data_struct.C2S_GetTransferRoomCardRecord) {
	transferRoomCardData := new(TransferRoomCardData)
	transferRoomCardUserInfos := []data_struct.TransferRoomCardUserInfo{}
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)

		count, _ := db.DB(DB).C("transferroomcard").
			Find(bson.M{"$or": []bson.M{bson.M{"fromaccountid": info.AccountID}, bson.M{"toaccountid": info.AccountID}}}).
			Count()
		iter := db.DB(DB).C("transferroomcard").
			Find(bson.M{"$or": []bson.M{bson.M{"fromaccountid": info.AccountID}, bson.M{"toaccountid": info.AccountID}}}).
			Sort("-createdat").Skip((info.PageNumber - 1) * info.PageSize).Limit(info.PageSize).Iter()
		if err := iter.Close(); err != nil {
			log.Error("iter close error: %v", err)
		}
		for iter.Next(&transferRoomCardData) {
			transferRoomCardUserInfos = append(transferRoomCardUserInfos, data_struct.TransferRoomCardUserInfo{
				FromAccountID: transferRoomCardData.FromAccountID,
				FromNickName:  transferRoomCardData.FromNickName,
				FromRole:      transferRoomCardData.FromRole,
				ToAccountID:   transferRoomCardData.ToAccountID,
				ToNickName:    transferRoomCardData.ToNickName,
				ToRole:        transferRoomCardData.ToRole,
				Total:         count,
				PageNumber:    info.PageNumber,
			})
		}
	}, func() {
		user.WriteMsg(&data_struct.S2C_TransferRoomCardRecord{
			Infos: transferRoomCardUserInfos,
		})
	})
}

func (user *User) getTransferRoomCardRecord(accountID int) {
	theUserData := new(UserData)
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		// load
		db.DB(DB).C("users").
			Find(bson.M{"accountid": accountID}).One(theUserData)
	}, func() {
		if theUserData.UserID < 1 {
			log.Debug("账户ID: %v 的用户不存在", accountID)
			user.WriteMsg(&data_struct.S2C_TransferRoomCardRecord{
				Error: data_struct.S2C_TransferRoomCardRecord_AccountIDInvalid,
			})
			return
		}
		user.sendTransferRoomCardRecord(accountID)
	})
}

func (user *User) sendTransferRoomCardRecord(accountID int) {
	transferRoomCardDatas := []TransferRoomCardData{}
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		// load
		db.DB(DB).C("transferroomcard").
			Find(bson.M{"$or": []bson.M{bson.M{"fromaccountid": accountID}, bson.M{"toaccountid": accountID}}}).
			Limit(20).All(&transferRoomCardDatas)
	}, func() {
		transferRoomCardUserInfos := []data_struct.TransferRoomCardUserInfo{}
		for _, transferRoomCardData := range transferRoomCardDatas {
			if transferRoomCardData.FromAccountID == accountID {
				transferRoomCardUserInfos = append(transferRoomCardUserInfos, data_struct.TransferRoomCardUserInfo{
					ToAccountID:  transferRoomCardData.ToAccountID,
					ToNickName:   transferRoomCardData.ToNickName,
					ToHeadimgurl: transferRoomCardData.ToHeadimgurl,
					RoomCards:    transferRoomCardData.RoomCards,
					Date:         time.Unix(transferRoomCardData.CreatedAt, 0).Format("2006/01/02 15:04:05"),
				})
			} else if transferRoomCardData.ToAccountID == accountID {
				transferRoomCardUserInfos = append(transferRoomCardUserInfos, data_struct.TransferRoomCardUserInfo{
					FromAccountID:  transferRoomCardData.FromAccountID,
					FromNickName:   transferRoomCardData.FromNickName,
					FromHeadimgurl: transferRoomCardData.FromHeadimgurl,
					RoomCards:      transferRoomCardData.RoomCards,
					Date:           time.Unix(transferRoomCardData.CreatedAt, 0).Format("2006/01/02 15:04:05"),
				})
			}
		}
		user.WriteMsg(&data_struct.S2C_TransferRoomCardRecord{
			Error: data_struct.S2C_TransferRoomCardRecord_OK,
			Infos: transferRoomCardUserInfos,
		})
	})
}

func (user *User) getAllTransferRoomCardRecord(info *data_struct.C2S_GetAllTransferRoomCardRecord) {
	transferRoomCardData := new(TransferRoomCardData)
	transferRoomCardUserInfos := []data_struct.TransferRoomCardUserInfo{}
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)

		//load
		count, _ := db.DB(DB).C("transferroomcard").Find(nil).Count()

		iter := db.DB(DB).C("transferroomcard").Find(nil).
			Sort("-createdat").Skip((info.PageNumber - 1) * info.PageSize).Limit(info.PageSize).Iter()
		if err := iter.Close(); err != nil {
			log.Error("iter close error: %v", err)
		}
		for iter.Next(&transferRoomCardData) {
			transferRoomCardUserInfos = append(transferRoomCardUserInfos, data_struct.TransferRoomCardUserInfo{
				FromAccountID: transferRoomCardData.FromAccountID,
				FromNickName:  transferRoomCardData.FromNickName,
				FromRole:      transferRoomCardData.FromRole,
				ToAccountID:   transferRoomCardData.ToAccountID,
				ToNickName:    transferRoomCardData.ToNickName,
				ToRole:        transferRoomCardData.ToRole,
				RoomCards:     transferRoomCardData.RoomCards,
				Date:          time.Unix(transferRoomCardData.CreatedAt, 0).Format("2006/01/02 15:04:05"),
				Total:         count,
				PageNumber:    info.PageNumber,
			})
		}
	}, func() {
		user.WriteMsg(&data_struct.S2C_AllTransferRoomCardRecord{
			Infos: transferRoomCardUserInfos,
		})
	})
}

func (user *User) getAllTransferRoomCardRecordByTime(info *data_struct.C2S_GetAllTransferRoomCardRecord) {
	transferRoomCardData := new(TransferRoomCardData)
	transferRoomCardUserInfos := []data_struct.TransferRoomCardUserInfo{}
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)

		// load
		count, _ := db.DB(DB).C("transferroomcard").Find(bson.M{"createdat": bson.M{"$gte": info.StartTime, "$lte": info.EndTime}}).Count()

		iter := db.DB(DB).C("transferroomcard").Find(bson.M{"createdat": bson.M{"$gte": info.StartTime, "$lte": info.EndTime}}).
			Sort("-createdat").Skip((info.PageNumber - 1) * info.PageSize).Limit(info.PageSize).Iter()
		if err := iter.Close(); err != nil {
			log.Error("iter close error: %v", err)
		}
		for iter.Next(&transferRoomCardData) {
			transferRoomCardUserInfos = append(transferRoomCardUserInfos, data_struct.TransferRoomCardUserInfo{
				FromAccountID: transferRoomCardData.FromAccountID,
				FromNickName:  transferRoomCardData.FromNickName,
				FromRole:      transferRoomCardData.FromRole,
				ToAccountID:   transferRoomCardData.ToAccountID,
				ToNickName:    transferRoomCardData.ToNickName,
				ToRole:        transferRoomCardData.ToRole,
				RoomCards:     transferRoomCardData.RoomCards,
				Date:          time.Unix(transferRoomCardData.CreatedAt, 0).Format("2006/01/02 15:04:05"),
				Total:         count,
				PageNumber:    info.PageNumber,
			})
		}

	}, func() {
		user.WriteMsg(&data_struct.S2C_AllTransferRoomCardRecord{
			Infos: transferRoomCardUserInfos,
		})
	})
}

func (user *User) getAllAgent(info *data_struct.C2S_GetAllAgentInfo) {
	userData := new(UserData)
	agentInfo := []data_struct.AgentInfo{}
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)

		//load
		count, _ := db.DB(DB).C("users").Find(bson.M{"role": bson.M{"$gt": 1, "$lt": 4}}).Count()

		iter := db.DB(DB).C("users").Find(bson.M{"role": bson.M{"$gt": 1, "$lt": 4}}).
			Sort("-createdat").Skip((info.PageNumber - 1) * info.PageSize).Limit(info.PageSize).Iter()
		if err := iter.Close(); err != nil {
			log.Error("iter close error: %v", err)
		}
		for iter.Next(&userData) {
			joinAgencyTime := ""
			if userData.JoinAgencyAt > 1509465600 {
				joinAgencyTime = time.Unix(userData.JoinAgencyAt, 0).Format("2006/01/02 15:04:05")
			}
			agentInfo = append(agentInfo, data_struct.AgentInfo{
				JoinAgencyAt: joinAgencyTime,
				Role:         userData.Role,
				AccountID:    userData.AccountID,
				Nickname:     userData.Nickname,
				RoomCards:    userData.RoomCards,
				Total:        count,
				PageNumber:   info.PageNumber,
			})
		}
	}, func() {
		user.WriteMsg(&data_struct.S2C_AllAgentInfo{
			Infos: agentInfo,
		})
	})
}

func (user *User) getAllAgentInfoByTime(info *data_struct.C2S_GetAllAgentInfo) {
	userData := new(UserData)
	agentInfo := []data_struct.AgentInfo{}
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)

		//load
		count, _ := db.DB(DB).C("users").
			Find(bson.M{"role": bson.M{"$gt": 1, "$lt": 4}, "joinagencyat": bson.M{"$gte": info.StartTime, "$lte": info.EndTime}}).Count()

		iter := db.DB(DB).C("users").Find(bson.M{"role": bson.M{"$gt": 1, "$lt": 4}, "joinagencyat": bson.M{"$gte": info.StartTime, "$lte": info.EndTime}}).
			Sort("-createdat").Skip((info.PageNumber - 1) * info.PageSize).Limit(info.PageSize).Iter()
		if err := iter.Close(); err != nil {
			log.Error("iter close error: %v", err)
		}
		for iter.Next(&userData) {
			joinAgencyTime := ""
			if userData.JoinAgencyAt > 1509465600 {
				joinAgencyTime = time.Unix(userData.JoinAgencyAt, 0).Format("2006/01/02 15:04:05")
			}
			agentInfo = append(agentInfo, data_struct.AgentInfo{
				JoinAgencyAt: joinAgencyTime,
				Role:         userData.Role,
				AccountID:    userData.AccountID,
				Nickname:     userData.Nickname,
				RoomCards:    userData.RoomCards,
				Total:        count,
				PageNumber:   info.PageNumber,
			})
		}
	}, func() {
		user.WriteMsg(&data_struct.S2C_AllAgentInfo{
			Infos: agentInfo,
		})
	})
}

func (user *User) getAllUserInfoByNickname(info *data_struct.C2S_GetAllUserInfo) {
	userData := new(UserData)
	userInfo := []data_struct.UserInfo{}
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)

		count, _ := db.DB(DB).C("users").
			Find(bson.M{"nickname": bson.M{"$regex": info.Nickname}}).Count()
		iter := db.DB(DB).C("users").
			Find(bson.M{"nickname": bson.M{"$regex": info.Nickname}}).
			Skip((info.PageNumber - 1) * info.PageSize).Limit(info.PageSize).Iter()
		if err := iter.Close(); err != nil {
			log.Error("iter close error: %v", err)
		}
		for iter.Next(&userData) {
			userInfo = append(userInfo, data_struct.UserInfo{
				AccountID:          userData.AccountID,
				Headimgurl:         userData.Headimgurl,
				Nickname:           userData.Nickname,
				Sex:                userData.Sex,
				RoomCards:          userData.RoomCards,
				GameScore:          userData.GameScore,
				ConsumedRoomCards:  userData.ConsumedRoomCards,
				PurchasedRoomCards: userData.PurchasedRoomCards,
				Role:               userData.Role,
				LastLogin:          time.Unix(userData.LastLoginAt, 0).Format("2006/01/02 15:04:05"),
				Total:              count,
				PageNumber:         info.PageNumber,
			})
		}
	}, func() {
		user.WriteMsg(&data_struct.S2C_AllUserInfo{
			Infos: userInfo,
		})
	})
}

func (user *User) getAllUserInfo(info *data_struct.C2S_GetAllUserInfo) {
	userData := new(UserData)
	userInfo := []data_struct.UserInfo{}
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)

		//load
		t := time.Now()
		zeroHour := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
		zeroHourYesterday := zeroHour.AddDate(0, 0, -1).Unix()

		twentyFourHour := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 0, 0, time.Local)
		twentyFourHourYesterday := twentyFourHour.AddDate(0, 0, -1).Unix()

		newUserCount, _ := db.DB(DB).C("users").
			Find(bson.M{"createdat": bson.M{"$gte": zeroHourYesterday, "$lte": twentyFourHourYesterday}}).Count()

		count, _ := db.DB(DB).C("users").Find(bson.M{"role": bson.M{"$lt": 4}}).Count()

		iter := db.DB(DB).C("users").Find(bson.M{"role": bson.M{"$lt": 4}}).
			Skip((info.PageNumber - 1) * info.PageSize).Limit(info.PageSize).Iter()
		if err := iter.Close(); err != nil {
			log.Error("iter close error: %v", err)
		}
		for iter.Next(&userData) {
			userInfo = append(userInfo, data_struct.UserInfo{
				AccountID:          userData.AccountID,
				Headimgurl:         userData.Headimgurl,
				Nickname:           userData.Nickname,
				Sex:                userData.Sex,
				RoomCards:          userData.RoomCards,
				GameScore:          userData.GameScore,
				ConsumedRoomCards:  userData.ConsumedRoomCards,
				PurchasedRoomCards: userData.PurchasedRoomCards,
				OnlineUser:         len(userIDUsers),
				NewUserYesterday:   newUserCount,
				Role:               userData.Role,
				LastLogin:          time.Unix(userData.LastLoginAt, 0).Format("2006/01/02 15:04:05"),
				Total:              count,
				PageNumber:         info.PageNumber,
			})
		}
	}, func() {
		user.WriteMsg(&data_struct.S2C_AllUserInfo{
			Infos: userInfo,
		})
	})
}

func (user *User) getBlackList(info *data_struct.C2S_GetBlackList) {
	userData := new(UserData)
	userInfo := []data_struct.UserInfo{}
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		//load
		count, _ := db.DB(DB).C("users").Find(bson.M{"role": -1}).Count()

		iter := db.DB(DB).C("users").Find(bson.M{"role": -1}).
			Sort("-createdat").Skip((info.PageNumber - 1) * info.PageSize).Limit(info.PageSize).Iter()
		if err := iter.Close(); err != nil {
			log.Error("iter close error: %v", err)
		}
		for iter.Next(&userData) {
			userInfo = append(userInfo, data_struct.UserInfo{
				AccountID:          userData.AccountID,
				Headimgurl:         userData.Headimgurl,
				Nickname:           userData.Nickname,
				Sex:                userData.Sex,
				RoomCards:          userData.RoomCards,
				GameScore:          userData.GameScore,
				ConsumedRoomCards:  userData.ConsumedRoomCards,
				PurchasedRoomCards: userData.PurchasedRoomCards,
				Role:               userData.Role,
				LastLogin:          time.Unix(userData.LastLoginAt, 0).Format("2006/01/02 15:04:05"),
				Total:              count,
				PageNumber:         info.PageNumber,
			})
		}
	}, func() {
		user.WriteMsg(&data_struct.S2C_BlackList{
			Infos: userInfo,
		})
	})
}

func (user *User) saveShareRoomcardData(createAt int64, giftRoomCards int) {
	shareRoomCardData := new(ShareRoomCardData)
	skeleton.Go(func() {
		err := shareRoomCardData.initValue()
		if err != nil {
			log.Error("init shareroomcard data error: %v", err)
			shareRoomCardData = nil
		}
	}, func() {
		if shareRoomCardData != nil {
			shareRoomCardData.AccountID = user.data.userData.AccountID
			shareRoomCardData.Nickname = user.data.userData.Nickname
			shareRoomCardData.GiftRoomCards = giftRoomCards
			shareRoomCardData.TotalRoomCards = user.data.userData.RoomCards
			shareRoomCardData.CreateAt = createAt
			shareRoomCardData.UpdateAt = time.Now().Unix()
			saveShareRoomCardData(shareRoomCardData)
		}
	})
}
