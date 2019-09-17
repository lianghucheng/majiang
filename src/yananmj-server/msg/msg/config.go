package msg

//获取麻将数据
type C2S_SetConfig struct {
	AndroidVersion     int    //安卓版本
	AndroidDownloadUrl string //安卓下载地址
	IOSVersion         int    //ios版本
	IOSDownloadUrl     string //ios下载地址
	AndroidGuestLogin  bool   //安卓游客登录
	IOSGuestLogin      bool   //ios游客登录
	Notice             string //公告
	Radio              string //广播
	WeChatNumber       string //客服微信号
}

//麻将数据返回常量
const (
	S2C_SetConfig_Ok                  = 0
	S2C_SetConfig_PermissionDenied    = 1 // 没有权限
	S2C_SetConfig_VersionInvalid      = 2 // 版本+S2C_SetYananConfig.AndriodVersion+无效
	S2C_SetConfig_DownloadUrlInvalid  = 3 // 下载地址+S2C_SetYananConfig.AdriodDownloadUrl+无效
	S2C_SetConfig_WeChatNumberInvalid = 4 // 客服微信告+S2C_SetYananConfig.WeChatNumber+无效
)

type S2C_SetConfig struct {
	Error              int
	AndriodVersion     int    // 安卓版本
	AndriodDownloadUrl string // 安卓下载地址
	IOSVersion         int    // iOS 版本
	IOSDownloadUrl     string // iOS 下载地址
	Notice             string // 公告
	Radio              string // 广播
	WeChatNumber       string // 客服微信号
}

//#############################广东麻将无
//广东麻将无#############################

//#############################广东麻将
//广东麻将#############################
