package conf

import (
	"encoding/json"
	"github.com/name5566/leaf/log"
	"io/ioutil"
)

var Server struct {
	LogLevel     string
	LogPath      string
	WSAddr       string
	CertFile     string
	KeyFile      string
	TCPAddr      string
	MaxConnNum   int
	DBUrl        string
	DBMaxConnNum int
	ConsolePort  int
	ProfilePath  string
	HTTPAddr     string
}

func init() {
	data, err := ioutil.ReadFile("conf/hnzzmj-server.json")
	if err != nil {
		log.Fatal("%v", err)
	}
	err = json.Unmarshal(data, &Server)
	if err != nil {
		log.Fatal("%v", err)
	}
}
