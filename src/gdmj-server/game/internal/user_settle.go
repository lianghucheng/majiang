package internal

import (
	"gopkg.in/mgo.v2/bson"
)

func (room *GDRoom) settlement() {
	for _, player := range room.userIDPlayerDatas {
		player.user.data.userData.Chips += int64(player.roundResult.TotalScore)
		updateUserData(player.user.data.userData.UserID, bson.M{"$inc": bson.M{"chips": player.roundResult.TotalScore}})
	}
}
