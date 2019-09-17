package main

import (
	"fmt"
	"github.com/name5566/leaf/db/mongodb"
	"gopkg.in/mgo.v2/bson"
	"strconv"
)

type UserData struct {
	UserID               int "_id"
	AccountID            int
	Nickname             string
	Headimgurl           string
	Sex                  int // 1 男性，2 女性
	Unionid              string
	CircleID             int // 圈圈ID
	Serial               string
	Model                string
	LoginIP              string
	Token                string
	ExpireAt             int64 // token 过期时间
	Role                 int   // 1 玩家 2 代理 3 管理员 4 超管
	Username             string
	Password             string
	RoomCards            int   // 房卡数量
	CompleteDailyShareAt int64 // 每日分享完成的时间
	GiftRoomCards        int   // 获赠的所有房卡
	JoinAgencyAt         int64 // 加入代理的时间
	SaleRoomCards        int   // 售卡数量
	GameScore            int   // 游戏积分
	TotalRounds          int   // 总共多少局
	WinRounds            int   // 赢了多少局
	PurchasedRoomCards   int   // 购买的房卡
	ConsumedRoomCards    int   // 玩游戏消耗的房卡
	LastLoginAt          int64 // 上一次登录的时间
	CreatedAt            int64
	UpdatedAt            int64
	Chips                int64
}

func main() {
	db, err := mongodb.Dial("mongodb://localhost", 1)
	fmt.Println(err)
	s := db.Ref()
	defer db.UnRef(s)
	userdata := UserData{}
	for i := 0; i < 10; i++ {
		userdata.Unionid = strconv.Itoa(i)
		userdata.AccountID = i
		userdata.UserID = i + 1
		userdata.Role = -2
		_, err := s.DB("gdmj").C("users").Upsert(bson.M{`unionid`: userdata.Unionid}, bson.M{`$set`: &userdata})
		fmt.Println(err)
	}
}
