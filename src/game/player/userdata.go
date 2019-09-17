package player

import (
	. "db"
	"fmt"
	. "game"
	"msg/login"
	"net/url"
	"strings"
	"time"

	"github.com/name5566/leaf/log"
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

const defaultAvatar = "https://www.shenzhouxing.com/gd/dl/img/avatar.jpg"

func (data *UserData) InitValue() error {
	userID, err := MongoDBNextSeq("users")
	if err != nil {
		return fmt.Errorf("get next users id error: %v", err)
	}
	data.UserID = userID
	data.Role = RolePlayer
	data.AccountID = getAccountID()
	data.CreatedAt = time.Now().Unix()
	return nil
}

func SaveUserData(userData *UserData) {
	Skeleton.Go(func() {
		log.Debug("%v", userData)
		db := MongoDB.Ref()
		defer MongoDB.UnRef(db)
		_, err := db.DB(DB).C("users").UpsertId(userData.UserID, userData)
		if err != nil {
			log.Error("save user %v data error: %v", userData.UserID, err)
		}
	}, nil)
}

func UpdateUserData(id int, update interface{}) {
	Skeleton.Go(func() {
		db := MongoDB.Ref()
		defer MongoDB.UnRef(db)
		_, err := db.DB(DB).C("users").UpsertId(id, update)
		if err != nil {
			log.Error("update user %v data error: %v", id, err)
		}
	}, nil)
}

func (data *UserData) UpdateWeChatInfo(info *login.C2S_WeChatLogin) {
	if data.Unionid == "" {
		data.Unionid = info.Unionid
		switch data.Unionid {
		case "o8c-nt6tO8aIBNPoxvXOQTVJUxY0":
			data.Role = RoleRoot
			data.RoomCards = 99999
		default:
			data.Role = RolePlayer
			data.RoomCards = 5
		}
	}
	if info.Nickname != "" {
		data.Nickname = info.Nickname
	}

	surl, err := url.Parse(info.Headimgurl)
	if err == nil {
		if surl.Scheme == "" {
			if data.Headimgurl == "" {
				data.Headimgurl = defaultAvatar
			}
		} else {
			if strings.HasSuffix(info.Headimgurl, "/0") {
				data.Headimgurl = info.Headimgurl[:len(info.Headimgurl)-1] + "132"
			} else {
				data.Headimgurl = info.Headimgurl
			}
		}
	} else {
		if data.Headimgurl == "" {
			data.Headimgurl = defaultAvatar
		}
	}
	if info.Sex == 1 {
		data.Sex = info.Sex
	} else {
		data.Sex = 2
	}
	data.Serial = info.Serial
	data.Model = info.Model

	data.UpdatedAt = time.Now().Unix()
}
