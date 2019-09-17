package config

import (
	"msg"
)

func init() {
	msg.MsgRegister(&C2S_SetGDConfig{})
	msg.MsgRegister(&S2C_SetGDConfig{})
}

type C2S_SetGDConfig struct {
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
	S2C_SetGDConfig_OK                  = 0
	S2C_SetGDConfig_PermissionDenied    = 1 // 没有权限
	S2C_SetGDConfig_VersionInvalid      = 2 // 版本 + S2C_SetGDConfig.AndroidVersion + 无效
	S2C_SetGDConfig_DownloadUrlInvalid  = 3 // 下载地址 + S2C_SetGDConfig.AndroidDownloadUrl + 无效
	S2C_SetGDConfig_WeChatNumberInvalid = 4 // 客服微信号 + S2C_SetGDConfig.WeChatNumberOfCustomerService + 无效
)

type S2C_SetGDConfig struct {
	Error              int
	AndroidVersion     int    // Android 版本号
	AndroidDownloadUrl string // Android 下载链接
	IOSVersion         int    // iOS 版本号
	IOSDownloadUrl     string // iOS 下载链接
	Notice             string // 公告
	Radio              string // 广播
	WeChatNumber       string // 客服微信号
}
