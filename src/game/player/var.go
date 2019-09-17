package player

// 用户状态
const (
	UserLogin  = iota
	UserLogout // 1
)

const (
	RoleRobot  = -2
	RoleBlack  = -1
	RolePlayer = 1
	RoleAgent  = 2
	RoleAdmin  = 3
	RoleRoot   = 4
)

var (
	userIDUsers = make(map[int]*User)

	userIDRooms     = make(map[int]interface{})
	roomNumberRooms = make(map[string]interface{})

	gdPracticeRooms = make(map[int]interface{}) // key: userID
	gdMatchRooms    = make(map[int]interface{}) // 匹配场 (房卡匹配、红包匹配)

	SystemOn = true // 系统开关

	accountIDs         = []int{}
	accountIDCounter   = 0
	reservedAccountIDs = []int{6666666, 8888888, 9999999}

	roomCardMatchOnlineNumber  = []int{0, 0, 0, 0} // 房卡比赛场在线人数
	redPacketMatchOnlineNumber = []int{0, 0, 0, 0} // 红包比赛在线人数
)
