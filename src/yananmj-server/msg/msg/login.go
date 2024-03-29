package msg

//微信登录
type C2S_WeChatLogin struct {
	Nickname   string
	Headimgurl string
	Sex        int    // 1 男性、2 女性
	Serial     string // 安卓设备序列号
	Model      string // 安卓手机型号
	Unionid    string // 微信unionid
	Openid     string // 用户唯一标识
}

//Token登录
type C2S_TokenLogin struct {
	Token string
}

//账密登录
type C2S_UsernamePasswordLogin struct {
	Username string
	Password string
}

type S2C_Login struct {
	AccountID          int
	Nickname           string
	Username           string
	JoinAgencyAT       string
	SaleRoomCardNumber int // 售卡数量
	Headimgurl         string
	Sex                int // 1 男性、2 女性
	RoomCards          int // 房卡数量
	Role               int // 1 玩家、2 代理、3 管理员、4 超管
	Token              string
	AnotherLogin       bool   // 其他设备登录
	AnotherRoom        bool   // 在其他房间
	Notice             string // 公告
	Radio              string // 广播
	WeChatNumber       string // 客服微信号
}

//##############广东麻将。湖南转转麻将#################
type S2C_Close struct {
	Error        int
	WeChatNumber string // 客服微信号
}

// Close
const (
	S2C_Close_LoginRepeated   = 1 // 您的账号在其他设备上线，非本人操作请注意修改密码
	S2C_Close_InnerError      = 2 // 登录出错，请重新登录
	S2C_Close_TokenInvalid    = 3 // 登录状态失效，请重新登录
	S2C_Close_UnionIDInvalid  = 4 // 登录出错，微信ID无效
	S2C_Close_UsernameInvalid = 5 // 登录出错，用户名无效
	S2C_Close_SystemOff       = 6 // 系统升级维护中，请稍后重试
	S2C_Close_RoleBlack       = 7 // 账号已冻结，请联系客服微信 S2C_Close.WeChatNumber
	S2C_Close_IPChanged       = 8 // 登录IP发生变化，非本人操作请注意修改密码
)
