package db

import (
	"config"

	"github.com/name5566/leaf/db/mongodb"
	"github.com/name5566/leaf/log"
)

var MongoDB *mongodb.DialContext

const DB = "gdmj"

func init() {
	if conf.Server.DBMaxConnNum <= 0 {
		conf.Server.DBMaxConnNum = 100
	}
	db, err := mongodb.Dial(conf.Server.DBUrl, conf.Server.DBMaxConnNum)
	if err != nil {
		log.Fatal("dial mongodb error: %v", err)
	}
	MongoDB = db
	initCollection()
}

func initCollection() {
	db := MongoDB
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
	err = db.EnsureCounter(DB, "counters", "redpacketmatchresult")
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

func MongoDBDestroy() {
	MongoDB.Close()
	MongoDB = nil
}

func MongoDBNextSeq(id string) (int, error) {
	return MongoDB.NextSeq(DB, "counters", id)
}
