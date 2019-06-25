package backTrace

import (
	"errors"
	"fmt"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
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

	//构建任务调度器
	nodeName := gConf.String("node::hostname")
	nodeIp := gConf.String("node::host")
	tmpPort, err := gConf.Int64("node::port")
	if err != nil {
		contextLogger.Errorf("config is Error: %v, can't not find node::port of int type", err)
	}

	tmpCore, err := gConf.Int64("node::core")
	if err != nil {
		contextLogger.Errorf("config is Error: %v, can't not find node::core of int type", err)
	}
	node := NewNode(nodeName, nodeIp, int32(tmpPort), int8(tmpCore))
	var stocks StockMap
	sc := LocalScheduler{node: node, coresForPerTask: 1, cacheMap: stocks}

	//获取买策略
	buyReg := GenerateAllBuyStrage()

	//获取卖策略
	sellReg := GenerateAllSellStrage()

	//获取redis client
	client, err := CreateRedisClient()
	if err != nil {
		contextLogger.Errorf(" can't get redis client. error: %s\n", err.Error())
		panic(err)
	}

	// 根据开始与结束标识确定是否需要重新生成Task
	// needReGenerateFlag = true ， = false 表示之前已经生成过了，不需要重新生成
	needReGenerateFlag := false

	_, err = client.Get(GenBeginFlag).Result()
	if err == redis.Nil { // 开始标识为空表示,还没有生成Task
		needReGenerateFlag = true
	} else if err != nil { //异常报错,redis client 不能正常工作,中断程序
		panic(errors.New(" redis client can't get value by key 'GenerateFlagBegin' ."))
	}

	if needReGenerateFlag == false {
		_, err = client.Get(GenEndFlag).Result()
		if err == redis.Nil { // 结束标识为空表示,Task生成了一半程序故障退出了,没有完全生成,需要重新生成
			needReGenerateFlag = true
		} else if err != nil { //异常报错,redis client 不能正常工作,中断程序
			panic(errors.New(" redis client can't get value by key 'GenerateFlagEnd' ."))
		}
	}

	//清空task队列以及生成标志
	if needReGenerateFlag {
		delCount, err := client.Del(GenBeginFlag, GenEndFlag, taskQueueName).Result()
		if err != nil && delCount != 3 { //删除key数目应该等于3
			panic(err)
		}
	}

	//获取所有股票
	allCodes := GetAllSockCode()

	allCodes = allCodes

	//一个code需要测试这么多个策略组合
	oneCodeNeedTest := uint32(len(buyReg.Names) * len(sellReg.Names))

	//计算得到全部的Task数目，h个code 与 m个买策略 、 n个卖策略 笛卡尔积得到 h * m * n个 Task
	allTaskCount := uint32(len(allCodes)) * oneCodeNeedTest

	if needReGenerateFlag {
		genCount := TasksGenerate(&buyReg, &sellReg, allCodes, client)

		//正常生成的Task与计算得到的结果应该是一致的
		if allTaskCount != genCount {
			contextLogger.Errorf("Generate task 's count %d should be the same with allTaskCount : %d\n", genCount, allTaskCount)
			panic(errors.New(fmt.Sprintf("Generate task 's count %d should be the same with allTaskCount : %d\n", genCount, allTaskCount)))
		}
	}

	contextLogger.Infof("Generate tasks total : %d .", allTaskCount)

	//分批生成Task并调度
	contextLogger.Info("start to scheduler tasks")

	sc.schedulerTask(&buyReg, &sellReg, allCodes, client)

	contextLogger.Info("scheduler finished!")

}
