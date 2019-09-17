package system

import (
	"msg"
)

func init() {
	msg.MsgRegister(&C2S_SetSystemOn{})
	msg.MsgRegister(&S2C_SetSystemOn{})
	msg.MsgRegister(&C2S_SetUserRole{})
	msg.MsgRegister(&S2C_SetUserRole{})
	msg.MsgRegister(&S2C_UpdateNotice{})
	msg.MsgRegister(&S2C_UpdateRadio{})
	msg.MsgRegister(&C2S_SetRobotData{})
}

// 系统开关
type C2S_SetSystemOn struct {
	On bool
}

const (
	S2C_SetSystemOn_OK               = 0
	S2C_SetSystemOn_PermissionDenied = 1 // 没有权限
)

type S2C_SetSystemOn struct {
	Error int
	On    bool
}

type C2S_SetUserRole struct {
	AccountID int
	Role      int //-1 黑名单 0 机器人 1 玩家 2 代理 3 管理员 4 超管
}

const (
	S2C_SetUserRole_OK               = 0
	S2C_SetUserRole_AccountIDInvalid = 1 // 账户ID无效
	S2C_SetUserRole_NotYourself      = 2 // 不能设置自己
	S2C_SetUserRole_RoleInvalid      = 3 // 角色 + S2C_SetUserRole.Role + 无效
	S2C_SetUserRole_PermissionDenied = 4 // 没有权限
	S2C_SetUserRole_SetRepeated      = 5 // 用户已经是 S2C_SetUserRole.Role(1 玩家 2 二级代理 3 一级代理)
)

type S2C_SetUserRole struct {
	Error int
	Role  int
}

// 更新公告
type S2C_UpdateNotice struct {
	Notice string
}

// 更新广播
type S2C_UpdateRadio struct {
	Radio string
}

// robot
type C2S_SetRobotData struct {
	LoginIP   string
	RoomCards int
}
