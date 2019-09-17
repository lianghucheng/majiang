package player

import (
	. "db"
	. "game"
	"math/rand"
	"time"
	"util"

	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/timer"
)

type User struct {
	gate.Agent
	State          int
	UserData       *UserData
	HeartbeatTimer *timer.Timer
	OwnerUserID    int // 所在房间的房主
	HeartbeatStop  bool
	Location       []float64 // 定位
}

func init() {
	rand.Seed(time.Now().UnixNano())

	result := new(UserData)

	m := make(map[int]bool)
	for _, v := range reservedAccountIDs {
		m[v] = true
	}

	Skeleton.Go(func() {
		db := MongoDB.Ref()
		defer MongoDB.UnRef(db)

		iter := db.DB(DB).C("users").Find(nil).Iter()
		for iter.Next(&result) {
			m[result.AccountID] = true
		}
		if err := iter.Close(); err != nil {
			log.Error("iter close error: %v", err)
		}
	}, func() {
		for i := 1000000; i < 10000000; i++ {
			if !m[i] {
				accountIDs = append(accountIDs, i)
			}
		}
		accountIDs = util.Shuffle(accountIDs)
		// log.Debug("%v %v", len(m), len(accountIDs))
	})

	cronExpr, _ := timer.NewCronExpr("10 0 0 * * *")
	Skeleton.CronFunc(cronExpr, func() {
		roomCardMatchOnlineNumber[0] = 0
		roomCardMatchOnlineNumber[1] = 0
		roomCardMatchOnlineNumber[2] = 0
		roomCardMatchOnlineNumber[3] = 0
	})
}

// 生成7位数的账号ID
func getAccountID() int {
	log.Debug("账号ID计数器: %v", accountIDCounter)
	accountID := accountIDs[accountIDCounter]
	accountIDCounter++
	return accountID
}

func InitUser(a gate.Agent) *User {
	user := new(User)
	user.Agent = a
	user.State = UserLogin
	user.UserData = new(UserData)
	user.Location = []float64{}
	return user
}

func (u *User) IsRobot() bool {
	if u.UserData.Role == RoleRobot {
		return true
	}
	return false
}
