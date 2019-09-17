package room

import (
	"game/player"
	"game/room"

	"gopkg.in/mgo.v2/bson"
)

func roomCardSettlement(r *room.GDRoom) {
	if len(r.Winneruserids) == 0 { // 流局
		for _, userID := range r.PositionUserIDs {
			playerData := r.Useridplayerdatas[userID]
			playerData.User.UserData.RoomCards += r.Rule.RoomCards
			player.UpdateUserData(userID, bson.M{"$set": bson.M{"roomcards": playerData.User.UserData.RoomCards}})
			if playerData.User.IsRobot() {
				//cards := gdRoom.rule.RoomCards
				//upsertRobotData(time.Now().Format("20060102"), bson.M{"$inc": bson.M{"roomcardmatchbalance": cards}})
			}
		}
	} else {
		winnerUserID := r.Winneruserids[0]
		playerData := r.Useridplayerdatas[winnerUserID]
		playerData.User.UserData.RoomCards += r.Rule.RoomCards * r.Rule.MaxPlayers
		player.UpdateUserData(winnerUserID, bson.M{"$set": bson.M{"roomcards": playerData.User.UserData.RoomCards}})
		if playerData.User.IsRobot() {
			//cards := gdRoom.rule.RoomCards * gdRoom.rule.MaxPlayers
			//upsertRobotData(time.Now().Format("20060102"), bson.M{"$inc": bson.M{"roomcardmatchbalance": cards}})
		}
	}
}
