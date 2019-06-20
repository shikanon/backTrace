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

	//Tasks切片
	var tasks []Task

	//生成所有的Task  h个code 与 m个买策略 、 n个卖策略 笛卡尔积得到 h * m * n个 Task
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
	canRunTaskNum := totalCores / sc.coresForPerTask //计算总共有几个Task可以同时运行

	taskChan := make(chan Task, canRunTaskNum*2) //用于

	//TODO 需要增加预加载逻辑
	preLoadIndex := 10 //默认加载10个股票数据
	//预请求数据加载
	if canRunTaskNum > 10 {
		preLoadIndex = int(canRunTaskNum)
	}
	//预加载
	_ = sc.cacheMap.Ready(allCodes[0:preLoadIndex])

	//等待分配Task执行
	for index := 0; index < int(canRunTaskNum); index++ {
		go func(taskChan chan Task, workerId string) {

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
					schedulerLogger.Errorf("cacheMap get data by code %s got error: %s \n", task.code, err.Error())
				} else {
					ana := Analyzer{BuyPolicies: []Strategy{buy},
						SellPolicies: []Strategy{sell}}
					agent := MoneyAgent{initMoney: 10000, Analyzer: ana}
					//经理需要做好准备后才能开始工作
					agent.Init()
					agent.WorkForSingle(*stock)
				}
			}
		}(taskChan, fmt.Sprintf("worker_%d", index))
	}

	for _, task := range tasks {
		taskChan <- task
		//TODO 增加跟踪任务执行的状态
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
	code         string
	buyStragety  string
	sellStragety string
}

//TODO task方法完善
func (task *Task) runTask() {

}
