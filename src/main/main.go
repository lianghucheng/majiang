package main

import (
	"config"
	"game"
	_ "game/config"
	_ "game/logic/room"
	"gate"
	"msg"

	"github.com/name5566/leaf"
	lconf "github.com/name5566/leaf/conf"
)

func main() {
	msg.MsgRegisterInit()
	gate.GateInit()
	lconf.LogLevel = conf.Server.LogLevel
	lconf.LogPath = conf.Server.LogPath
	lconf.LogFlag = conf.LogFlag
	lconf.ConsolePort = conf.Server.ConsolePort
	lconf.ProfilePath = conf.Server.ProfilePath
	leaf.Run(
		game.ModuleGame,
		gate.ModuleGate,
	)
}
