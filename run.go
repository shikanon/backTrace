package backTrace

import (
	"github.com/sirupsen/logrus"
	"runtime"
)

func init() {
	/*	// Log as JSON instead of the default ASCII formatter.
		logrus.SetFormatter(&logrus.TextFormatter{})

		// Output to stdout instead of the default stderr
		// Can be any io.Writer, see below for File example
		logrus.SetOutput(os.Stdout)

		// Only log level.
		logrus.SetLevel(logrus.InfoLevel)*/
}

func RunBacktrace() {
	contextLogger := logrus.WithFields(logrus.Fields{
		"function": "RunBacktrace()",
	})

	//核数
	cpus := gConf.DefaultInt("go::cpu", 1)
	runtime.GOMAXPROCS(cpus)

	//构建任务调度器
	nodeName := gConf.String("node::hostname")
	nodeIp := gConf.String("node::host")
	// 设置core默认设置19999
	tmpPort := gConf.DefaultInt("node::port", 19999)
	// 设置core默认设置为CPU核心数
	tmpCore := gConf.DefaultInt("node::core", 2)
	node := NewNode(nodeName, nodeIp, int32(tmpPort), int32(tmpCore))
	var stocks StockMap
	sc := LocalScheduler{node: node, coresForPerTask: 1, cacheMap: stocks}

	//获取买策略
	buyReg := GenerateAllBuyStrage()

	//获取卖策略
	sellReg := GenerateAllSellStrage()

	/*	//获取redis client
		client, err := CreateRedisClient()
		if err != nil {
			contextLogger.Errorf(" can't get redis client. error: %s\n", err.Error())
			panic(err)
			}*/

	//获取所有股票
	allCodes := GetAllSockCode()

	//一个code需要测试这么多个策略组合
	oneCodeNeedTest := uint32(len(buyReg.Names) * len(sellReg.Names))

	//计算得到全部的Task数目，h个code 与 m个买策略 、 n个卖策略 笛卡尔积得到 h * m * n个 Task
	allTaskCount := uint32(len(allCodes)) * oneCodeNeedTest

	contextLogger.Infof("tasks total : %d .", allTaskCount)

	//断点恢复
	tm := TasksManager{
		redoLogFile: "checkpoint.txt",
	}
	tm.recover()

	//分批生成Task并调度
	contextLogger.Info("start to scheduler tasks")

	sc.schedulerTask(&buyReg, &sellReg, allCodes, tm)

	contextLogger.Info("scheduler finished!")

}
