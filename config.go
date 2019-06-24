package backTrace

import (
	"github.com/astaxie/beego/config"
	"github.com/sirupsen/logrus"
)

var gConf config.Configer

//var confPath = "backTrace/config.conf"
var confPath = "config.conf"

func init() {
	conf, err := config.NewConfig("ini", confPath)
	if err != nil {
		logrus.Fatal("Error:", err)
		panic(err)
	}
	gConf = conf
}
