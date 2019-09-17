package msg

type C2S_SetHNZZConfig struct {
	AndroidVersion     int    // Android 版本号
	AndroidDownloadUrl string // Android 下载链接
	IOSVersion         int    // iOS 版本号
	IOSDownloadUrl     string // iOS 下载链接
	AndroidGuestLogin  bool   // Android 游客登录
	IOSGuestLogin      bool   // iOS 游客登录
	Notice             string // 公告
	Radio              string // 广播
	WeChatNumber       string // 客服微信号
}

const (
	S2C_SetHNZZConfig_OK                  = 0
	S2C_SetHNZZConfig_PermissionDenied    = 1 // 没有权限
	S2C_SetHNZZConfig_VersionInvalid      = 2 // 版本 + S2C_SetHNZZConfig.AndroidVersion + 无效
	S2C_SetHNZZConfig_DownloadUrlInvalid  = 3 // 下载地址 + S2C_SetHNZZConfig.AndroidDownloadUrl + 无效
	S2C_SetHNZZConfig_WeChatNumberInvalid = 4 // 客服微信号 + S2C_SetHNZZConfig.WeChatNumberOfCustomerService + 无效
)

type S2C_SetHNZZConfig struct {
	Error              int
	AndroidVersion     int    // Android 版本号
	AndroidDownloadUrl string // Android 下载链接
	IOSVersion         int    // iOS 版本号
	IOSDownloadUrl     string // iOS 下载链接
	Notice             string // 公告
	Radio              string // 广播
	WeChatNumber       string // 客服微信号
}
