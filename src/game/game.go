package game

import (
	"config"
	. "db"

	"github.com/name5566/leaf/chanrpc"
	"github.com/name5566/leaf/module"
)

func NewSkeleton() *module.Skeleton {
	skeleton := &module.Skeleton{
		GoLen:              conf.GoLen,
		TimerDispatcherLen: conf.TimerDispatcherLen,
		AsynCallLen:        conf.AsynCallLen,
		ChanRPCServer:      chanrpc.NewServer(conf.ChanRPCLen),
	}
	skeleton.Init()
	return skeleton
}

var (
	Skeleton   = NewSkeleton()
	ChanRPC    = Skeleton.ChanRPCServer
	ModuleGame = new(Module)
)

type Module struct {
	*module.Skeleton
}

func (m *Module) OnInit() {
	m.Skeleton = Skeleton
}

func (m *Module) OnDestroy() {
	MongoDBDestroy()
}

func HandleRegister(id interface{}, f interface{}) {
	Skeleton.RegisterChanRPC(id, f)
}
