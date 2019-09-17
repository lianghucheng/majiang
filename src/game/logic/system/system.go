package logic

import (
	"game"
	. "msg/agent"
	. "msg/card"
	. "msg/config"
	. "msg/system"
	. "msg/user"
	"reflect"
)

func handler(m interface{}, h interface{}) {
	game.Skeleton.RegisterChanRPC(reflect.TypeOf(m), h)
}

func init() {
	handler(&C2S_SetSystemOn{}, handleSetSystemOn)
	handler(&C2S_SetUsernamePassword{}, handleSetUsernamePassword)
	handler(&C2S_SetGDConfig{}, handleSetGDConfig)
	handler(&C2S_SetUserRole{}, handleSetUserRole)
	handler(&C2S_GetUserInfo{}, handleGetUserInfo)
	handler(&C2S_TransferRoomCard{}, handleTransferRoomCard)
	handler(&C2S_GetTransferRoomCardRecord{}, handleGetTransferRoomCardRecord)
	handler(&C2S_GetAllTransferRoomCardRecord{}, handleGetAllTransferRoomCardRecord)
	handler(&C2S_GetAllAgentInfo{}, handleGetAllAgentInfo)
	handler(&C2S_GetAllUserInfo{}, handleGetAllUserInfo)
	handler(&C2S_GetBlackList{}, handleGetBlackList)
}
