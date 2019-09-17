package room

import (
	"game/player"
	"game/room"

	"gopkg.in/mgo.v2/bson"
)

func CostCard(r *room.GDRoom) {
	//私人场，房主扣卡
	if r.Rule.RoomType == room.RoomPrivate {
		playerData := r.Useridplayerdatas[r.OwnerUserID]
		playerData.User.UserData.RoomCards -= r.Rule.RoomCards
		playerData.User.UserData.ConsumedRoomCards += r.Rule.RoomCards
		player.UpdateUserData(r.OwnerUserID, bson.M{"$set": bson.M{"roomcards": playerData.User.UserData.RoomCards, "consumedroomcards": playerData.User.UserData.ConsumedRoomCards}})
		return
	}
	for _, userID := range r.PositionUserIDs {
		playerData := r.Useridplayerdatas[userID]

		playerData.User.UserData.RoomCards -= r.Rule.RoomCards
		playerData.User.UserData.ConsumedRoomCards += r.Rule.RoomCards
		player.UpdateUserData(userID, bson.M{"$set": bson.M{"roomcards": playerData.User.UserData.RoomCards, "consumedroomcards": playerData.User.UserData.ConsumedRoomCards}})
		if playerData.User.IsRobot() {
			//cards := -r.Rule.RoomCards
			switch r.Rule.RoomType {
			case room.RoomRoomCardMatch:
				//upsertRobotData(time.Now().Format("20060102"), bson.M{"$inc": bson.M{"roomcardmatchbalance": cards}})
			case room.RoomRedPacketMatching:
				//upsertRobotData(time.Now().Format("20060102"), bson.M{"$inc": bson.M{"redpacketmatchbalance": cards}})
			}
		}
	}
}
