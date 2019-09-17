package room

import (
	"sync"
)

//////////////////////////////////////////////////////////////
//! 玩家管理者
type RoomMgr struct {
	MapPerson map[int]interface{}
	lock      *sync.RWMutex
}

var personmgrSingleton *RoomMgr = nil
var once sync.Once

//! 得到服务器指针
func GetRoomMgr() *RoomMgr {
	once.Do(func() {
		personmgrSingleton = new(RoomMgr)
		personmgrSingleton.MapPerson = make(map[int]interface{})
		personmgrSingleton.lock = new(sync.RWMutex)
	})
	return personmgrSingleton
}

//! 加入玩家
func (self *RoomMgr) AddPerson(uid int, r interface{}) {
	self.lock.Lock()
	defer self.lock.Unlock()

	self.MapPerson[uid] = r
}

//! 删玩家
func (self *RoomMgr) DelPerson(uid int) {
	self.lock.Lock()
	defer self.lock.Unlock()

	delete(self.MapPerson, uid)
}

//! 该玩家是否存在
func (self *RoomMgr) GetRoom(uid int) interface{} {
	self.lock.RLock()
	defer self.lock.RUnlock()

	r, ok := self.MapPerson[uid]
	if ok {
		return r
	}

	return nil
}
