package internal

import (
	"github.com/name5566/leaf/db/mongodb"
	"github.com/name5566/leaf/log"
	"hnzzmj-server/conf"
)

var mongoDB *mongodb.DialContext

const DB = "hnzzmj"

func init() {
	// mongodb
	if conf.Server.DBMaxConnNum <= 0 {
		conf.Server.DBMaxConnNum = 100
	}
	db, err := mongodb.Dial(conf.Server.DBUrl, conf.Server.DBMaxConnNum)
	if err != nil {
		log.Fatal("dial mongodb error: %v", err)
	}
	mongoDB = db
	initCollection()
	initConfigData()
}

func initCollection() {
	db := mongoDB
	err := db.EnsureCounter(DB, "counters", "users")
	if err != nil {
		log.Fatal("ensure counter error: %v", err)
	}
	err = db.EnsureCounter(DB, "counters", "configs")
	if err != nil {
		log.Fatal("ensure counter error: %v", err)
	}
	err = db.EnsureCounter(DB, "counters", "totalresult")
	if err != nil {
		log.Fatal("ensure counter error: %v", err)
	}
	err = db.EnsureCounter(DB, "counters", "roundresult")
	if err != nil {
		log.Fatal("ensure counter error: %v", err)
	}
	err = db.EnsureCounter(DB, "counters", "transferroomcard")
	if err != nil {
		log.Fatal("ensure counter error: %v", err)
	}
	err = db.EnsureCounter(DB, "counters", "shareroomcard")
	if err != nil {
		log.Fatal("ensure counter error: %v", err)
	}

	err = db.EnsureUniqueIndex(DB, "users", []string{"accountid"})
	if err != nil {
		log.Fatal("ensure index error: %v", err)
	}
	err = db.EnsureUniqueIndex(DB, "users", []string{"unionid"})
	if err != nil {
		log.Fatal("ensure index error: %v", err)
	}
}

func initConfigData() {
	hnzzConfigData = new(ConfigData)
	hnzzConfigData.initHNZZ()
}

func mongoDBDestroy() {
	mongoDB.Close()
	mongoDB = nil
}

func mongoDBNextSeq(id string) (int, error) {
	return mongoDB.NextSeq(DB, "counters", id)
}
