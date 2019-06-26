package backTrace

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

const (
	GenBeginFlag = "GenerateFlagBegin"
	GenEndFlag   = "GenerateFlagEnd"
)

type CodeCacheList struct {
	nodes []string
}

// 如果
func (list *CodeCacheList) getNeedLoadFlag(code string) bool {
	len := len(list.nodes)
	for index, codeN := range list.nodes {
		if codeN == code && index == len-1 {
			return true
		}
	}
	return false
}

//入参是新的需要缓存的Code，返回的是要移除的code
func (list *CodeCacheList) insert(code string) string {

	if len(list.nodes) == 0 {
		panic("CodeCacheList 's len must more than 0!")
	}

	delCode := list.nodes[0]

	for index := 0; index < len(list.nodes)-1; index++ {
		list.nodes[index] = list.nodes[index+1]
	}
	list.nodes[len(list.nodes)-1] = code

	return delCode
}

func TasksGenerate(buyReg *StrategyRegister, sellReg *StrategyRegister, codes []string, client *redis.Client, testFlag bool) uint32 {
	//var tasks []Task
	_, err = client.Set(GenBeginFlag, "true", 0).Result()
	if err != nil {
		panic(err)
	}

	count := uint32(0)
	//生成所有的Task
	for _, code := range codes {
		//获取买入策略
		for _, buyName := range buyReg.Names {
			//获取卖出策略
			for _, sellName := range sellReg.Names {
				//构建一个Task
				//tasks = append(tasks, Task{code: code, buyStragety: buyName, sellStragety: sellName})
				taskStr := code + "," + buyName + "," + sellName + "," + strconv.FormatInt(int64(count), 10)
				client.RPush(taskQueueName, taskStr)
				count += 1

				//TODO 临时加个测试标识
				if testFlag {
					return count
				}
			}
		}
	}

	_, err = client.Set(GenEndFlag, "true", 0).Result()
	if err != nil {
		panic(err)
	}

	return count
}

// 调度接口
type Scheduler interface {
	//入参是code数组，生成并启动Task
	schedulerTask(buyReg *StrategyRegister, sellReg *StrategyRegister, code []string, client *redis.Client)
}

// 单机调度器
type LocalScheduler struct {
	node            Node
	coresForPerTask int8
	cacheMap        StockMap
}

//根据股票ID进行任务调度
func (sc *LocalScheduler) schedulerTask(buyReg *StrategyRegister, sellReg *StrategyRegister, codes []string, client *redis.Client) {
	schedulerLogger := logrus.WithFields(logrus.Fields{
		"function": "schedulerTask()",
	})

	totalCores := sc.node.core
	canRunTaskNum := uint32(totalCores / sc.coresForPerTask) //计算总共有几个Task可以同时运行

	taskChan := make(chan Task, canRunTaskNum*2) //用于通知执行Task
	finalChan := make(chan int)

	//---------------------------------------预加载code

	//successedTaskForCode:=make(map[string]int32,defaultPreLoad)  //用于跟中code对应task的完成情况, 与默认加载code

	defaultPreLoad := uint32(3)
	preLoadStartIndex := uint32(0)
	preLoadEndIndex := defaultPreLoad //默认加载3个股票数据

	allCodesCount := uint32(len(codes))

	//fmt.Printf("allCodesCount: %d \n",allCodesCount)

	//一个code需要测试这么多个策略组合
	//oneCodeNeedTest := uint32(len(buyReg.Names) * len(sellReg.Names))

	//计算得到全部的Task数目，h个code 与 m个买策略 、 n个卖策略 笛卡尔积得到 h * m * n个 Task
	//allTaskCount := allCodesCount * oneCodeNeedTest

	//防止数组越界
	if allCodesCount < defaultPreLoad {
		preLoadEndIndex = allCodesCount
	}

	//请求缓存模块进行预加载
	_ = sc.cacheMap.Ready(codes[preLoadStartIndex:preLoadEndIndex])

	tmpList := make([]string, preLoadEndIndex, preLoadEndIndex)
	for index, c := range codes[preLoadStartIndex:preLoadEndIndex] {
		//fmt.Printf("index: %d \n",index)
		tmpList[index] = c
	}
	cacheCodes := CodeCacheList{tmpList}

	//等待分配Task执行
	for index := 0; index < int(canRunTaskNum); index++ {
		go func(workerId string) {

			for {
				task, ok := <-taskChan

				if ok != true {
					schedulerLogger.Errorf(" Worker %s exist. because the chan is closed. \n", workerId)
					break
				} else if task.id == -999 { //-999表示所有的Task都已经分发完了
					schedulerLogger.Infof(" Worker %s exist. because the tasks were done. \n", workerId)
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
					schedulerLogger.Debugf("%s start task (%s,%s,%s)", workerId, task.code, task.buyStragety, task.sellStragety)
					ana := Analyzer{BuyPolicies: []Strategy{buy},
						SellPolicies: []Strategy{sell}}
					agent := MoneyAgent{initMoney: 10000, Analyzer: ana}
					//经理需要做好准备后才能开始工作
					agent.Init()
					err = agent.WorkForSingle(*stock)

					if err != nil {
						schedulerLogger.Errorf("%s agent.WorkForSingle cased error by code %s, error: %s \n", workerId, task.code, err.Error())
						schedulerLogger.Errorf("%s aborted task (%s,%s,%s)", workerId, task.code, task.buyStragety, task.sellStragety)
						continue
					}

					//评估策略效果
					result := agent.GetProfileData()
					estimator, err := CreateEstimator(&result)
					if err != nil {
						schedulerLogger.Errorf("CreateEstimator caused Error: %s", err.Error())
						schedulerLogger.Errorf("%s aborted task (%s,%s,%s)", workerId, task.code, task.buyStragety, task.sellStragety)
						continue
					}

					record := RewardRecord{Code: task.code, SellStrategy: task.sellStragety, BuyStrategy: task.buyStragety,
						TotalReturnRate: estimator.TotalReturnRate, ReturnRatePerYear: estimator.ReturnRatePerYear,
						WinningProb: estimator.WinningProb, ProfitExpect: estimator.ProfitExpect, LossExpect: estimator.LossExpect,
						AlphaEarnings: estimator.AlphaEarnings, BetaEarnings: estimator.BetaEarnings}

					//将结果插入数据库
					/*_, err = stmt.Exec(task.code, task.sellStragety, task.buyStragety, estimator.TotalReturnRate,
						estimator.ReturnRatePerYear, estimator.WinningProb, estimator.ProfitExpect, estimator.LossExpect,
					estimator.AlphaEarnings, estimator.BetaEarnings)*/

					_, err = SaveRewardRecord(&record)

					if err != nil {
						schedulerLogger.Errorf("save RewardRecord caused error: %s", err.Error())
					}

					schedulerLogger.Debugf("%s Finish task (%s,%s,%s)", workerId, task.code, task.buyStragety, task.sellStragety)
				}
			}

			finalChan <- 1 //通知Scheduler,所有Task已经完成了

		}(fmt.Sprintf("worker_%d", index))
	}

	for {

		val, err := client.LPop(taskQueueName).Result()

		if err == redis.Nil {
			schedulerLogger.Info("all task already done!")
			break
		} else if err != nil {
			panic(errors.New("redis client can't work right."))
		}

		splits := strings.Split(val, ",")

		t := Task{
			code:         splits[0],
			buyStragety:  splits[1],
			sellStragety: splits[2],
		}

		//判断是否需要进行预加载
		needLoadFlag := cacheCodes.getNeedLoadFlag(splits[0])

		//如果preLoadEndIndex + 1 > = allCodesCount表示已经跑到最后一个code了，不需要再进行预加载了
		if needLoadFlag && preLoadEndIndex+1 < allCodesCount {
			preLoadStartIndex = preLoadEndIndex
			preLoadEndIndex += 1
			//更新正在缓存的code列表
			preCode := codes[preLoadEndIndex-1]
			delCode := cacheCodes.insert(preCode)

			schedulerLogger.Infof("gonna to pre load code %s,and del the code %s", preCode, delCode)

			//异步加载以及删除code
			go func(begin uint32, end uint32, delCode string) {
				_ = sc.cacheMap.Ready(codes[preLoadStartIndex:preLoadEndIndex])
				sc.cacheMap.Delete(delCode)
			}(preLoadStartIndex, preLoadEndIndex, delCode)
		}
		taskChan <- t

	}

	//通知各个goroutine已经完成任务了
	for x := uint32(0); x < canRunTaskNum; x++ {
		taskChan <- Task{id: -999}
	}

	//等待Task执行完毕
	for x := uint32(0); x < canRunTaskNum; x++ {
		_ = <-finalChan
	}
	close(taskChan)
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
