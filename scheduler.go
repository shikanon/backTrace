package backTrace

import (
	"fmt"
	"github.com/sirupsen/logrus"
)

// 调度接口
type Scheduler interface {
	//入参是code数组，生成并启动Task
	schedulerTask([]string)
}

// 单机调度器
type LocalScheduler struct {
	node            Node
	coresForPerTask int8
	cacheMap        StockMap
	//runningTasks    map[string]bool
}

//根据股票ID进行任务调度
func (sc *LocalScheduler) schedulerTask(allCodes []string) {
	schedulerLogger := logrus.WithFields(logrus.Fields{
		"function": "schedulerTask()",
	})

	buyReg := GenerateAllBuyStrage()
	sellReg := GenerateAllSellStrage()

	//计算总共的Task有多少
	oneCodeNeedTest := uint32(len(buyReg.Names) * len(sellReg.Names)) //一个code需要测试这么多个策略组合
	//allTaskCount := len(allCodes) * oneCodeNeedTest  //计算得到全部的Task数目，h个code 与 m个买策略 、 n个卖策略 笛卡尔积得到 h * m * n个 Task

	//TODO Tasks切片在策略多了以后会很大,后续考虑指定长度,循环利用
	//tasks := make([]Task,0,1000)
	var tasks []Task

	//生成所有的Task
	for _, code := range allCodes {
		//获取买入策略
		for _, buyName := range buyReg.Names {
			//获取卖出策略
			for _, sellName := range sellReg.Names {
				//构建一个Task
				tasks = append(tasks, Task{code: code, buyStragety: buyName, sellStragety: sellName})
			}
		}
	}

	schedulerLogger.Infof("total tasks %d", len(tasks))

	totalCores := sc.node.core
	canRunTaskNum := uint32(totalCores / sc.coresForPerTask) //计算总共有几个Task可以同时运行

	taskChan := make(chan Task, canRunTaskNum*2) //用于通知执行Task
	finalChan := make(chan int)

	//code数组的下标
	preLoadStartIndex := uint32(0)
	preLoadEndIndex := uint32(10) //默认加载10个股票数据

	allCodesCount := uint32(len(allCodes))

	//防止数组越界
	if allCodesCount < 10 {
		preLoadEndIndex = uint32(len(allCodes))
	}

	//请求缓存模块进行预加载
	_ = sc.cacheMap.Ready(allCodes[preLoadStartIndex:preLoadEndIndex])

	//等待分配Task执行
	for index := 0; index < int(canRunTaskNum); index++ {
		go func(workerId string) {

			stmt, err := DB.Prepare("insert into RewardRecord(code,SellStrategy,BuyStrategy,TotalReturnRate,ReturnRatePerYear," +
				"WinningProb,ProfitExpect,LossExpect,AlphaEarnings,BetaEarnings) values (?,?,?,?,?,?,?,?,?,?);")

			if err != nil {
				//TODO 待优化代码，需要增加逻辑：如果异常发生了，反馈给Scheduler终止任务调度
				schedulerLogger.Errorf(" Worker %s exist. because the DB.Prepare caused error: %s \n", workerId, err.Error())
				close(taskChan)
				return
			}

			for {
				task, ok := <-taskChan

				if ok != true {
					schedulerLogger.Errorf(" Worker %s exist. because the chan is closed. \n", workerId)
					break
				}

				buy, err := buyReg.Load(task.buyStragety)
				if err != nil {
					schedulerLogger.Errorf(" There is no sell strategy named %s !! \n", task.buyStragety)
				}

				sell, err := sellReg.Load(task.sellStragety)
				if err != nil {
					schedulerLogger.Errorf(" There is no buy strategy named %s !! \n", task.buyStragety)
				}

				stock, err := sc.cacheMap.Load(task.code)

				if err != nil {
					schedulerLogger.Errorf("%s cacheMap get data by code %s got error: %s \n", workerId, task.code, err.Error())
				} else {
					schedulerLogger.Infof("%s start task (%s,%s,%s)", workerId, task.code, task.buyStragety, task.sellStragety)
					ana := Analyzer{BuyPolicies: []Strategy{buy},
						SellPolicies: []Strategy{sell}}
					agent := MoneyAgent{initMoney: 10000, Analyzer: ana}
					//经理需要做好准备后才能开始工作
					agent.Init()
					agent.WorkForSingle(*stock)

					//评估策略效果
					result := agent.GetProfileData()
					estimator, err := CreateEstimator(&result)
					if err != nil {
						schedulerLogger.Errorf("CreateEstimator caused Error: %s", err.Error())
						schedulerLogger.Infof("%s aborted task (%s,%s,%s)", workerId, task.code, task.buyStragety, task.sellStragety)
						continue
					}

					/*record := RewardRecord{Code: task.code, SellStrategy: task.sellStragety, BuyStrategy: task.buyStragety,
					TotalReturnRate: estimator.TotalReturnRate, ReturnRatePerYear: estimator.ReturnRatePerYear,
					WinningProb: estimator.WinningProb, ProfitExpect: estimator.ProfitExpect, LossExpect: estimator.LossExpect,
					AlphaEarnings: estimator.AlphaEarnings, BetaEarnings: estimator.BetaEarnings}
					*/
					//将结果插入数据库
					_, err = stmt.Exec(task.code, task.sellStragety, task.buyStragety, estimator.TotalReturnRate,
						estimator.ReturnRatePerYear, estimator.WinningProb, estimator.ProfitExpect, estimator.LossExpect,
						estimator.AlphaEarnings, estimator.BetaEarnings)

					schedulerLogger.Infof("%s Finish task (%s,%s,%s)", workerId, task.code, task.buyStragety, task.sellStragety)
				}
			}

			finalChan <- 1 //通知Scheduler,所有Task已经完成了

		}(fmt.Sprintf("worker_%d", index))
	}

	var alreadyDone uint32 = 0 //已经调度的任务数
	for _, task := range tasks {
		taskChan <- task
		//schedulerLogger.Infof("assign task ")
		alreadyDone += 1 //已经调度的任务数 +1

		// 判断是否需要进行预加载数据
		if alreadyDone+canRunTaskNum >= preLoadEndIndex*oneCodeNeedTest && preLoadEndIndex < allCodesCount {
			preLoadStartIndex = preLoadEndIndex
			if preLoadEndIndex+10 > allCodesCount {
				preLoadEndIndex = allCodesCount
			} else {
				preLoadEndIndex += 10
			}
			_ = sc.cacheMap.Ready(allCodes[preLoadStartIndex:preLoadEndIndex])
		}

	}
	close(taskChan)

	//等待Task执行完毕
	for x := uint32(0); x < canRunTaskNum; x++ {
		_ = <-finalChan
	}

}

//资源管理接口,负责节点的管理（节点注册、通信等等）
type backend interface {
	Register(hostName string, ip string, port int32, core int8) (bool, error)
	Leave(nodeId string) (bool, error)
	GetAliveNode() []*Node
}

/*
//单机版的计算资源管理
type LocalBackend struct {
	node Node
}

//注册节点
func (lb *LocalBackend) register(hostName string, ip string, port int32, core int8) (bool, error) {
	if lb.node == (Node{}) {
		lb.node = NewNode(hostName, ip, port, core)
		return true, nil
	} else {
		return false, fmt.Errorf(" LocalNode %s is already registered!", lb.node)
	}
}

func (lb *LocalBackend) leave(nodeId string) (bool, error) {
	if lb.node.nodeId != nodeId {
		return false, fmt.Errorf(" There is no node named %s .", nodeId)
	} else {
		lb.node = Node{}
		return true, nil
	}
}

func (lb *LocalBackend) getAliveNode() []*Node {
	return []*Node{&lb.node}
}*/

//节点信息，表示一个节点
type Node struct {
	nodeId   string
	hostName string
	core     int8
	ip       string
	port     int32
}

func (n *Node) string() string {
	return n.nodeId
}

//创建节点的方法
func NewNode(hostName string, ip string, port int32, core int8) Node {
	return Node{
		nodeId:   fmt.Sprintf("%s_%s_%d", hostName, ip, port),
		hostName: hostName,
		ip:       ip,
		port:     port,
		core:     core,
	}
}

// ---------------------------Task

type Task struct {
	id           int32
	code         string
	buyStragety  string
	sellStragety string
}

//TODO task方法完善
func (task *Task) runTask() {

}
