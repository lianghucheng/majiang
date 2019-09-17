package room

import (
	"sync"
	"util"
)

type RoomNumber struct {
	RoomId  int
	Lock    *sync.RWMutex
	RoomMap map[string]*GDRoom
}

var roomSingleton *RoomNumber = nil

//var roomNumber map[string]
//! 得到服务器指针
var once1 sync.Once

func GetRoom() *RoomNumber {

	once1.Do(func() {
		roomSingleton = new(RoomNumber)
		roomSingleton.RoomMap = make(map[string]*GDRoom)
		roomSingleton.Lock = new(sync.RWMutex)
	})
	return roomSingleton
}

//! 得到一个id
func (self *RoomNumber) GetID() int {
	self.Lock.Lock()
	defer self.Lock.Unlock()

	self.RoomId++
	if self.RoomId > 9999 {
		self.RoomId = 0
	}
	roomNumber := (util.Random(9)+1)*100000 + self.RoomId*10 + util.Random(10)
	return roomNumber
}

func (self *RoomNumber) SetRoom(roomNumber string, r *GDRoom) {
	self.Lock.Lock()
	defer self.Lock.Unlock()
	self.RoomMap[roomNumber] = r
}

func (self *RoomNumber) DelRoom(roomNumber string) {
	self.Lock.Lock()
	defer self.Lock.Unlock()
	r := self.RoomMap[roomNumber]
	if r != nil {
		for i := 0; i < r.Rule.MaxPlayers; i++ {
			GetRoomMgr().DelPerson(r.PositionUserIDs[i])
		}
	}
	delete(self.RoomMap, roomNumber)
}
