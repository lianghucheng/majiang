package player

import (
	"sync"
)

//////////////////////////////////////////////////////////////
//! 玩家管理者
type PersonMgr struct {
	MapPerson map[int]*User
	lock      *sync.RWMutex
}

var personmgrSingleton *PersonMgr = nil
var once sync.Once

//! 得到服务器指针
func GetPersonMgr() *PersonMgr {
	once.Do(func() {
		personmgrSingleton = new(PersonMgr)
		personmgrSingleton.MapPerson = make(map[int]*User)
		personmgrSingleton.lock = new(sync.RWMutex)
	})
	return personmgrSingleton
}

//! 加入玩家
func (self *PersonMgr) AddPerson(uid int, person *User) {
	self.lock.Lock()
	defer self.lock.Unlock()

	self.MapPerson[uid] = person
}

//! 删玩家
func (self *PersonMgr) DelPerson(uid int) {
	self.lock.Lock()
	defer self.lock.Unlock()

	delete(self.MapPerson, uid)
}

//! 该玩家是否存在
func (self *PersonMgr) GetPerson(uid int) *User {
	self.lock.RLock()
	defer self.lock.RUnlock()

	person, ok := self.MapPerson[uid]
	if ok {
		return person
	}

	return nil
}
