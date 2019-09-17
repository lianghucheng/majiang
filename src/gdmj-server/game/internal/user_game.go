package internal

func (user *User) doPrepare(r interface{}) {
	switch r.(type) {
	case *GDRoom:
		gdRoom := r.(*GDRoom)
		if gdRoom.state == roomGame {
			gdRoom.reconnect(user)
		} else {
			gdRoom.doPrepare(user.data.userData.UserID)
		}
	}
}

func (user *User) doChow(r interface{}, meld []int) {
	switch r.(type) {
	case *GDRoom:
		gdRoom := r.(*GDRoom)
		if gdRoom.state == roomGame && gdRoom.prepareChow(user.data.userData.UserID, meld) {
			gdRoom.doChow()
		}
	}
}

func (user *User) doPass(r interface{}) {
	switch r.(type) {
	case *GDRoom:
		gdRoom := r.(*GDRoom)
		if gdRoom.state == roomGame {
			gdRoom.doPass(user.data.userData.UserID)
		}
	}
}

//玩家取消托管
func (user *User) doCancelTrusteeship(r interface{}, managed bool) {
	gdRoom := r.(*GDRoom)
	if gdRoom.state == roomGame {
		gdRoom.doCancelTrusteeship(user.data.userData.UserID)
	}
}

//玩家下炮子
func (user *User) doSetGun(r interface{}, gun int) {
	gdRoom := r.(*GDRoom)
	gdRoom.doSetGun(user.data.userData.UserID, gun)
}
