package internal

func (user *User) doPrepare(r interface{}) {
	switch r.(type) {
	case *HNZZRoom:
		hnzzRoom := r.(*HNZZRoom)
		if hnzzRoom.state == roomGame {
			hnzzRoom.reconnect(user)
		} else {
			hnzzRoom.doPrepare(user.data.userData.UserID)
		}
	}
}

func (user *User) doDiscard(r interface{}, tile int) {
	switch r.(type) {
	case *HNZZRoom:
		hnzzRoom := r.(*HNZZRoom)
		if hnzzRoom.state == roomGame {
			hnzzRoom.doDiscard(user.data.userData.UserID, tile)
		}
	}
}

func (user *User) doWin(r interface{}) {
	switch r.(type) {
	case *HNZZRoom:
		hnzzRoom := r.(*HNZZRoom)
		if hnzzRoom.state == roomGame && hnzzRoom.prepareWin(user.data.userData.UserID) {
			hnzzRoom.doWin()
		}
	}
}

func (user *User) doKong(r interface{}, meld []int) {
	switch r.(type) {
	case *HNZZRoom:
		hnzzRoom := r.(*HNZZRoom)
		if hnzzRoom.state == roomGame && hnzzRoom.prepareKong(user.data.userData.UserID, meld) {
			hnzzRoom.doKong()
		}
	}
}

func (user *User) doPong(r interface{}) {
	switch r.(type) {
	case *HNZZRoom:
		hnzzRoom := r.(*HNZZRoom)
		if hnzzRoom.state == roomGame && hnzzRoom.preparePong(user.data.userData.UserID) {
			hnzzRoom.doPong()
		}
	}
}

func (user *User) doChow(r interface{}, meld []int) {
	switch r.(type) {
	case *HNZZRoom:
		hnzzRoom := r.(*HNZZRoom)
		if hnzzRoom.state == roomGame && hnzzRoom.prepareChow(user.data.userData.UserID, meld) {
			hnzzRoom.doChow()
		}
	}
}

func (user *User) doPass(r interface{}) {
	switch r.(type) {
	case *HNZZRoom:
		hnzzRoom := r.(*HNZZRoom)
		if hnzzRoom.state == roomGame {
			hnzzRoom.doPass(user.data.userData.UserID)
		}
	}
}
