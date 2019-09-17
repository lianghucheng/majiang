package internal

import (
	"fmt"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/util"
	url2 "net/url"
	"strings"
	"time"
	"yananmj-server/msg"
)

const defaultAvatar = "https://www.shenzhouxing.com/ruijin/dl/img/avatar.jpg"

//用户数据
type UserData struct {
	UserID               int "_id"
	AccountID            int
	Nickname             string
	Headimgurl           string
	Sex                  int //1为男性，2为女性
	Unionid              string
	CircleID             int // 圈圈ID
	Serial               string
	Model                string
	LoginIP              string
	Token                string
	Role                 int // 1 玩家 2 代理 3 管理员 4 超管
	Username             string
	Password             string
	RoomCards            int   // 房卡数量
	CompleteDailyShareAt int64 // 每日分享完成的时间
	GiftRoomCards        int   // 获赠的所有房卡
	JoinAgencyAt         int64 // 加入代理的时间
	SaleRoomCardNumber   int   // 售卡数量
	GameScore            int   // 游戏积分
	PurchasedRoomCards   int   // 购买的房卡
	ConsumedRoomCards    int   // 玩游戏消耗的房卡
	LastLoginAt          int64 // 上一次登录的时间
	CreatedAt            int64
	UpdatedAt            int64
}

func (data *UserData) initValue() error {
	userId, err := mongoDBNextSeq("users")
	if err != nil {
		return fmt.Errorf("get next users id error: %v", err)
	}
	data.UserID = userId
	data.Role = rolePlayer
	//data.AccountID = common.GetID(4) + strconv.Itoa(data.UserID)
	data.AccountID = getAccountID()
	data.CreatedAt = time.Now().Unix()
	return nil
}

func saveUserData(userdata *UserData) {
	data := util.DeepClone(userdata)
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		id := data.(*UserData).UserID
		_, err := db.DB(DB).C("users").UpsertId(id, data)
		if err != nil {
			log.Error("save user:%v data error:%v", id, err)
		}
	}, func() {})
}

func updateUserData(id int, update interface{}) {
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		_, err := db.DB(DB).C("users").UpsertId(id, update)
		if err != nil {
			log.Error("update user %v data error: %v", id, err)
		}
	}, nil)
}

func (data *UserData) updateWeChatInfo(info *data_struct.C2S_WeChatLogin) {
	if data.Unionid == "" {
		data.Unionid = info.Unionid
		switch data.Unionid {
		case "o8c-nt6tO8aIBNPoxvXOQTVJUxY0":
			data.Role = roleRoot
			data.RoomCards = 99999
			data.Username = "银滩麻将"
			data.Password = "123456"
		case "o8c-nt2jC5loIHg1BQGgYW6aqe60":
			data.Role = roleRoot
			data.RoomCards = 5
			data.Username = "超级管理员"
			data.Password = "123456"
		default:
			data.Role = rolePlayer
			data.RoomCards = 5
		}
	}
	data.Nickname = info.Nickname

	url, err := url2.Parse(info.Headimgurl)
	if err == nil {
		if url.Scheme == "" {
			data.Headimgurl = defaultAvatar
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
