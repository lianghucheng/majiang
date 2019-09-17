package internal

//获取玩家
func (user *User) getAllPlayers(r interface{}) {
	yananRoom := r.(*YananRoom)
	yananRoom.GetAllPlayers(user)
}

//玩家准备
func (user *User) doPrepare(r interface{}) {
	yananRoom := r.(*YananRoom)
	if yananRoom.state == roomGame {
		yananRoom.reconnect(user)
	} else {
		yananRoom.doPrepare(user.data.userData.UserID)
	}
}

//玩家下炮子
func (user *User) doSetGun(r interface{}, gun int) {
	yananRoom := r.(*YananRoom)
	yananRoom.doSetGun(user.data.userData.UserID, gun)
}

//玩家出牌
func (user *User) doDiscard(r interface{}, tile int) {
	yananRoom := r.(*YananRoom)
	if yananRoom.state == roomGame {
		playerData := yananRoom.userIDPlayerDatas[user.data.userData.UserID]
		// 托管计数清0
		playerData.discardsCount = 0
		playerData.managed = false
		yananRoom.doDiscard(user.data.userData.UserID, tile)
	}
}

//玩家胡牌
func (user *User) doWin(r interface{}) {
	yananRoom := r.(*YananRoom)
	if yananRoom.state == roomGame && yananRoom.prepareWin(user.data.userData.UserID) {
		yananRoom.doWin()
	}
}

//玩家扛牌
func (user *User) doKong(r interface{}, meld []int) {
	yananRoom := r.(*YananRoom)
	if yananRoom.state == roomGame && yananRoom.prepareKong(user.data.userData.UserID, meld) {
		yananRoom.doKong()
	}
}

//玩家碰牌
func (user *User) doPong(r interface{}) {
	yananRoom := r.(*YananRoom)
	if yananRoom.state == roomGame && yananRoom.preparePong(user.data.userData.UserID) {
		yananRoom.doPong()
	}
}

//玩家过牌
func (user *User) doPass(r interface{}) {
	yananRoom := r.(*YananRoom)
	if yananRoom.state == roomGame {
		yananRoom.doPass(user.data.userData.UserID)
	}
}

//玩家取消托管
func (user *User) doCancelTrusteeship(r interface{}, managed bool) {
	yananRoom := r.(*YananRoom)
	if yananRoom.state == roomGame {
		yananRoom.doCancelTrusteeship(user.data.userData.UserID, managed)
	}
}
