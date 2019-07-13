package backTrace

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	LastestIndex      = "latestIndex"
	TaskStatuSucessed = int8(2)
	TaskStatuFailed   = int8(3)
	TaskStatuWait     = int8(0)
	TaskStatuRunning  = int8(1)
	TaskNone          = "-999"
	RedoDelim         = ";"
)

// 调度接口
type Scheduler interface {
	//入参是code数组，生成并启动Task
	schedulerTask(buyReg *StrategyRegister, sellReg *StrategyRegister, code []string, tm *TasksManager)
}

// 单机调度器
type LocalScheduler struct {
	node            Node
	coresForPerTask int32
	cacheMap        StockMap
}

//根据股票ID进行任务调度
func (sc *LocalScheduler) schedulerTask(buyReg *StrategyRegister, sellReg *StrategyRegister, codes []string,
	tm *TasksManager) {
	schedulerLogger := logrus.WithFields(logrus.Fields{
		"function": "schedulerTask()",
	})

	canRunTaskNum := uint32(sc.node.parallelism) //计算总共有几个Task可以同时运行

	taskChan := make(chan Task, canRunTaskNum)        //用于通知执行Task
	finalChan := make(chan TaskStatus, canRunTaskNum) //用于反馈Task运行结果的

	//等待分配Task执行
	//如果Task返回失败的结果，taskChan的另一端会停止发送task，整个程序会停掉
	//具体逻辑查看214行的for循环
	for index := 0; index < int(canRunTaskNum); index++ {
		go func(workerId string) {

			flag := true
			for {
				task, ok := <-taskChan

				msg := "ok"

				if ok != true {
					schedulerLogger.Errorf(" Worker %s exist. because the chan is closed. \n", workerId)
					break
				} else if task.Code == TaskNone { //-999表示所有的Task都已经分发完了
					schedulerLogger.Infof(" Worker %s exist. because the tasks were done. \n", workerId)
					break
				}

				buy, err := buyReg.Load(task.BuyStragety)
				if err != nil {
					flag = false
					msg = err.Error()
					schedulerLogger.Errorf(" There is no sell strategy named %s !! \n", task.BuyStragety)
				}

				sell, err := sellReg.Load(task.SellStragety)
				if err != nil {
					flag = false
					schedulerLogger.Errorf(" There is no buy strategy named %s !! \n", task.BuyStragety)
				}

				stock, err := sc.cacheMap.Load(task.Code)

				//   条件 flag == false 是为了不进入else分支
				if err != nil || flag == false {
					flag = false
					schedulerLogger.Errorf("%s cacheMap get data by Code %s caused error: %s \n", workerId, task.Code, err.Error())
				} else {
					schedulerLogger.Debugf("%s start Task (%s,%s,%s)", workerId, task.Code, task.BuyStragety, task.SellStragety)
					ana := Analyzer{BuyPolicies: []Strategy{buy},
						SellPolicies: []Strategy{sell}}
					agent := MoneyAgent{initMoney: 10000, Analyzer: ana}
					//经理需要做好准备后才能开始工作
					agent.Init()
					err = agent.WorkForSingle(*stock)

					if err != nil {
						//flag = false
						schedulerLogger.Errorf("%s agent.WorkForSingle cased error by Code %s, error: %s \n", workerId, task.Code, err.Error())
						schedulerLogger.Errorf("%s aborted Task (%s,%s,%s)", workerId, task.Code, task.BuyStragety, task.SellStragety)
					} else {

						//评估策略效果
						result := agent.GetProfileData()
						estimator, err := CreateEstimator(&result)
						if err != nil {
							schedulerLogger.Errorf("CreateEstimator caused Error: %s", err.Error())
							schedulerLogger.Errorf("%s aborted Task (%s,%s,%s)", workerId, task.Code, task.BuyStragety, task.SellStragety)
							continue
						}

						record := RewardRecord{Code: task.Code, SellStrategy: task.SellStragety, BuyStrategy: task.BuyStragety,
							TotalReturnRate: estimator.TotalReturnRate, ReturnRatePerYear: estimator.ReturnRatePerYear,
							WinningProb: estimator.WinningProb, ProfitExpect: estimator.ProfitExpect, LossExpect: estimator.LossExpect,
							AlphaEarnings: estimator.AlphaEarnings, BetaEarnings: estimator.BetaEarnings}

						//将结果插入数据库
						_, err = SaveRewardRecord(&record)

						if err != nil {
							flag = false
							msg = err.Error()
							schedulerLogger.Errorf("save RewardRecord caused error: %s", err.Error())
						} else {
							schedulerLogger.Debugf("%s Finish Task (%s,%s,%s)", workerId, task.Code, task.BuyStragety, task.SellStragety)
						}
					}
				}

				statu := TaskStatuSucessed
				if flag == false {
					statu = TaskStatuFailed
				}

				taskStatu := TaskStatus{
					Task:  &task,
					Statu: statu,
					Msg:   msg,
				}
				finalChan <- taskStatu //返回task执行结果
			}

			//通知taskManager,该goroutine已经完成所有任务，退出了
			taskStatu := TaskStatus{
				Task: &Task{Code: TaskNone},
			}
			finalChan <- taskStatu

		}(fmt.Sprintf("worker_%d", index))
	}

	go func(tm *TasksManager) {
		for {
			taskStatu, ok := <-finalChan
			//管道关闭了,退出
			if ok == false {
				break
			}
			tm.UpdateStatu(&taskStatu)
		}
	}(tm)

	allCodesCount := int32(len(codes))
	schedulerLogger.Infof("code's len is %d", allCodesCount)
	//预加载1
	preLoadStratIndex := tm.LastCodeIndex
	preLoadEndIndex := tm.LastCodeIndex + 5

	if preLoadEndIndex > allCodesCount {
		preLoadEndIndex = allCodesCount
	}

	err := sc.cacheMap.Ready(codes[preLoadStratIndex:preLoadEndIndex])
	if err != nil {
		panic(errors.Errorf("cacheMap reday load data caused error, %s", err.Error()))
	}

	//预加载2
	preLoadStratIndex2 := preLoadEndIndex
	preLoadEndIndex2 := preLoadEndIndex + 5

	if preLoadEndIndex2 > allCodesCount {
		preLoadEndIndex2 = allCodesCount
	}

	err = sc.cacheMap.Ready(codes[preLoadStratIndex2:preLoadEndIndex2])
	if err != nil {
		panic(errors.Errorf("cacheMap reday load data caused error, %s", err.Error()))
	}

	lastCodeIndex := tm.LastCodeIndex
	lastBuyIndex := tm.LastBuyIndex
	lastSellIndex := tm.LastSellIndex

	successedFlag := true

	schedulerLogger.Infof("Task will start from index: %d,%d,%d", lastCodeIndex, lastBuyIndex, lastSellIndex)

	for codeIndex, code := range codes[lastCodeIndex:] {
		//获取买入策略
		for buyIndex, buyName := range buyReg.Names[lastBuyIndex:] {
			//获取卖出策略
			for sellIndex, sellName := range sellReg.Names[lastSellIndex:] {
				//构建一个Task
				//tasks = append(tasks, Task{Code: Code, buyStragety: buyName, sellStragety: sellName})
				//taskStr := Code + "," + buyName + "," + sellName
				//id := (CodeIndex + 1) * (buyIndex + 1) * (sellIndex + 1)

				actualCodeIndex := int32(codeIndex) + lastCodeIndex
				actualBuyIndex := int32(buyIndex) + lastBuyIndex
				actualSellIndex := int32(sellIndex) + lastSellIndex

				if atomic.LoadUint32(&tm.stop) > 0 { //tm检查到有task失败,程序中断
					schedulerLogger.Errorf("Generate Task is breakout ,because tm.Stop > 0.")
					successedFlag = false
					break
				}
				t := Task{CodeIndex: actualCodeIndex, Code: code, BuyIndex: actualBuyIndex, BuyStragety: buyName,
					SellIndex: actualSellIndex, SellStragety: sellName}
				schedulerLogger.Debugf("add Task %s", fmt.Sprintf("%d,%d,%d", actualCodeIndex, actualBuyIndex, actualSellIndex))

				isFinished := tm.AddTask(t) //记录task,并检查是否已经完成了的
				if isFinished == false {
					taskChan <- t
					schedulerLogger.Debug(t.String(","))
				}

				newIndex := atomic.LoadInt32(&tm.LastCodeIndex)

				if newIndex > preLoadEndIndex && preLoadEndIndex != allCodesCount &&
					preLoadStratIndex2 == allCodesCount {
					waitForDeleteCodes := codes[preLoadStratIndex:preLoadEndIndex]
					go func(delCodes []string) {
						for _, code := range delCodes {
							sc.cacheMap.Delete(code)
						}
					}(waitForDeleteCodes)

					preLoadStratIndex = preLoadStratIndex2
					preLoadEndIndex = preLoadEndIndex2

					preLoadStratIndex2 = preLoadEndIndex2
					preLoadEndIndex2 = preLoadEndIndex2 + 5

					if preLoadEndIndex2 > int32(len(codes)) {
						preLoadStratIndex2 = int32(len(codes))
					}
					err = sc.cacheMap.Ready(codes[preLoadStratIndex2:preLoadEndIndex2])
					if err != nil {
						panic(errors.Errorf("cacheMap reday load data caused error, %s", err.Error()))
					}
				}
			}
		}
	}

	if successedFlag { //如果是中断程序退出的，直接跳过发送完成所有task的信号
		//通知各个goroutine所有的task已经分配完了,已经没有task了
		for x := uint32(0); x < canRunTaskNum; x++ {
			taskChan <- Task{Code: TaskNone}
		}
	}

	for {
		if atomic.LoadUint32(&tm.stop) > 0 { //强行停止
			break
		} else if tm.finishCount == canRunTaskNum { //等待所有goroutine完成task退出
			schedulerLogger.Info("all Task were done.")
			tm.clean()
			break
		} else {
			time.Sleep(time.Millisecond * 500)
		}
	}
	close(taskChan)
}

//节点信息，表示一个节点
type Node struct {
	nodeId      string
	hostName    string
	parallelism int32
	ip          string
	port        int32
}

func (n *Node) string() string {
	return n.nodeId
}

//创建节点的方法
func NewNode(hostName string, ip string, port int32, parallelism int32) Node {
	return Node{
		nodeId:      fmt.Sprintf("%s_%s_%d", hostName, ip, port),
		hostName:    hostName,
		ip:          ip,
		port:        port,
		parallelism: parallelism,
	}
}

// ---------------------------Task
type Task struct {
	//id           int64
	Code         string
	CodeIndex    int32
	BuyStragety  string
	BuyIndex     int32
	SellStragety string
	SellIndex    int32
}

func (t *Task) String(delim string) string {
	return t.Code + delim +
		strconv.FormatInt(int64(t.CodeIndex), 10) + delim + t.BuyStragety + delim +
		strconv.FormatInt(int64(t.BuyIndex), 10) + delim + t.SellStragety + delim +
		strconv.FormatInt(int64(t.SellIndex), 10)
}

//Task状态
type TaskStatus struct {
	Task  *Task
	Statu int8 //
	Msg   string
}

type IndexNode struct {
	T    *Task
	Key  string
	Prev *IndexNode
	Next *IndexNode
}

// -1 表示当前node比other小
// 0 表示相等
// 1 表示当前node大于other
func (node *IndexNode) compare(other *IndexNode) int {

	if node.T.CodeIndex == other.T.CodeIndex && node.T.BuyIndex == other.T.BuyIndex &&
		node.T.SellIndex == other.T.SellIndex {
		return 0
	}

	//codeIndex小的Task在前面
	if node.T.CodeIndex < other.T.CodeIndex {
		return -1
	} else if node.T.CodeIndex > other.T.CodeIndex {
		return 1
	}

	//buyIndex小的Task在前面
	if node.T.BuyIndex < other.T.BuyIndex {
		return -1
	} else if node.T.BuyIndex > other.T.BuyIndex {
		return 1
	}

	//sellIndex小的Task在前面
	if node.T.SellIndex < other.T.SellIndex {
		return -1
	}
	return 1
}

type IndexQueue struct {
	Head *IndexNode
	Tail *IndexNode
}

func (q *IndexQueue) Insert(newNode *IndexNode) {
	if q.Head == nil {
		q.Head = newNode
		q.Tail = newNode
	} else {
		for currentNode := q.Head; ; currentNode = currentNode.Next {
			if currentNode.compare(newNode) > 0 { //当前节点ID比新节点ID大，因此将新节点插入在当前节点前
				if currentNode == q.Head { //如果当前节点前置指针为空，则是头节点
					q.Head = newNode
					newNode.Prev = nil
				} else {
					currentNode.Prev.Next = newNode
					newNode.Prev = currentNode.Prev
				}
				newNode.Next = currentNode
				currentNode.Prev = newNode
				break
			}
			//已经到达尾部，因此只需要将新节点添加尾部即可
			if currentNode == q.Tail {
				newNode.Prev = currentNode
				newNode.Next = nil
				currentNode.Next = newNode
				q.Tail = newNode //设置尾节点为最新添加的节点
				break
			}
		}
	}
}

//移除head
func (q *IndexQueue) Pop() {
	if q.Head != nil {
		//fmt.Printf("delete key %s \n",q.Head.Key)
		q.Head.Prev = nil
		q.Head = q.Head.Next

		if q.Head == nil {
			q.Tail = nil
		}

	}
}

//用于跟踪Task的运行状况
type TasksManager struct {
	logger            *logrus.Entry
	client            *redis.Client
	runningTasks      sync.Map
	waitForCheckPoint IndexQueue
	LastCodeIndex     int32
	LastBuyIndex      int32
	LastSellIndex     int32
	finishCount       uint32
	//clean             uint32
	lock     sync.RWMutex
	initFlag int64  //初始化标志
	stop     uint32 //停止标志
	redisKey string
}

// 分配task之前先记录task,如果task已经存,表示该task已经存在
// task存在的情况可能是断点重跑的原因，应该返回false表示应该跳过该task的分配
func (tm *TasksManager) AddTask(t Task) bool {

	key := fmt.Sprintf("%d,%d,%d", t.CodeIndex, t.BuyIndex, t.SellIndex)
	status, ok := tm.runningTasks.Load(key)
	if ok {
		if status == TaskStatuSucessed {
			return true
		}
	}
	tm.runningTasks.Store(key, TaskStatuRunning)
	tm.waitForCheckPoint.Insert(&IndexNode{T: &t, Key: key})
	return false
}

func (tm *TasksManager) UpdateStatu(s *TaskStatus) {

	if s.Task.Code == TaskNone {
		atomic.AddUint32(&tm.finishCount, 1)
	} else {
		switch s.Statu {
		case TaskStatuFailed:
			tm.forceStop()
			tm.logger.Infof("Program is going to shutdown. %s", s.Msg)
		case TaskStatuSucessed:
			tm.SaveStatus(s)
		}
	}

}

func (tm *TasksManager) forceStop() {
	atomic.AddUint32(&tm.stop, 1)
}

func (tm *TasksManager) SaveStatus(s *TaskStatus) {
	//写REDO日志
	key := fmt.Sprintf("%d,%d,%d", s.Task.CodeIndex, s.Task.BuyIndex, s.Task.SellIndex)
	tm.runningTasks.Store(key, TaskStatuSucessed)

	//循环更新下标
	currentNode := tm.waitForCheckPoint.Head
	nextDoneKey := key

	//needToUpdateIndex := false

	//mt.Printf("-------------------------- update key %s \n",key)

	updateIndexFlag := false
	for {
		val, ok := tm.runningTasks.Load(nextDoneKey)
		if ok != true {
			panic(errors.Errorf("when tasksManager update Task status by key %s ,it can't find the key in map", nextDoneKey))
		}

		//fmt.Printf("key %s:val %d \n",nextDoneKey,val)
		//fmt.Printf("currentNode.Key %s \n",currentNode.Key)
		if val == TaskStatuSucessed && currentNode.Key == nextDoneKey {

			updateIndexFlag = true

			atomic.StoreInt32(&tm.LastCodeIndex, currentNode.T.CodeIndex)
			atomic.StoreInt32(&tm.LastBuyIndex, currentNode.T.BuyIndex)
			atomic.StoreInt32(&tm.LastSellIndex, currentNode.T.SellIndex)
			currentNode = currentNode.Next
			tm.waitForCheckPoint.Pop()          //移除head
			tm.runningTasks.Delete(nextDoneKey) //移除runningtask中完成了的task
			if tm.waitForCheckPoint.Head != nil {
				nextDoneKey = tm.waitForCheckPoint.Head.Key
				//fmt.Printf("tm.waitForCheckPoint.Head.Key %s \n",nextDoneKey)
			} else {
				break
			}
			//needToUpdateIndex = true
		} else {
			break
		}
	}

	if updateIndexFlag {
		newIndex := fmt.Sprintf("%d%s%d%s%d", atomic.LoadInt32(&tm.LastCodeIndex), RedoDelim,
			atomic.LoadInt32(&tm.LastBuyIndex), RedoDelim, atomic.LoadInt32(&tm.LastSellIndex))
		_, err := tm.client.Set(tm.redisKey, newIndex, 0).Result()
		if err != nil {
			tm.logger.Infof("write LatestIndex to redis caused error ,%s", err.Error())
			tm.forceStop() //强制终止程序
		}
	}

}

//TasksManager必须要要调用了该方法后才能正常运作
func (tm *TasksManager) Recover() {
	tm.lock.Lock()
	if tm.initFlag != 0 {
		fmt.Println("tasksManager already init!")
		return
	}

	tm.logger = logrus.WithFields(logrus.Fields{
		"function": "TaskManager",
	})

	//获取redis client
	tm.client, err = CreateRedisClient()
	if err != nil {
		tm.logger.Errorf("TaskManager can't get redis client. error: %s\n", err.Error())
		panic(err)
	}

	indexStr, err := tm.client.Get(tm.redisKey).Result()
	beginFromZero := false
	if err == redis.Nil {
		tm.logger.Info("TaskManger Recover can't find latestIndex in redis.")
		tm.logger.Info("TaskManger gonna scheduler Task form the 0.")
		beginFromZero = true
	} else if err != nil {
		tm.logger.Errorf("TaskManger Recover get the latestIndex caused error, %s", err.Error())
		panic(err)
	}

	if beginFromZero {
		tm.LastCodeIndex = 0
		tm.LastBuyIndex = 0
		tm.LastSellIndex = 0
	} else {
		array := strings.Split(indexStr, RedoDelim)
		if len(array) != 3 {
			tm.logger.Errorf("TaskManger Recover get the latestIndex is wrong , %s", indexStr)
			panic(err)
		}

		tmpCodeIndex, err := strconv.ParseInt(array[0], 10, 64)
		if err != nil {
			tm.logger.Errorf("TaskManger Recover get the latestCodeIndex is wrong , %s", array[0])
			panic(err)
		}
		tm.LastCodeIndex = int32(tmpCodeIndex)

		tmpBuyIndex, err := strconv.ParseInt(array[1], 10, 64)
		if err != nil {
			tm.logger.Errorf("TaskManger Recover get the latestBuyIndex is wrong , %s", array[1])
			panic(err)
		}
		tm.LastBuyIndex = int32(tmpBuyIndex)

		tmpSellIndex, err := strconv.ParseInt(array[2], 10, 64)
		if err != nil {
			tm.logger.Errorf("TaskManger Recover get the latestSellIndex is wrong , %s", array[2])
			panic(err)
		}
		tm.LastSellIndex = int32(tmpSellIndex)
		tm.logger.Infof("Task will start from index: %s", indexStr)
	}

	if beginFromZero == false {
		//增加这个task状态，lastIndex对应的Task其实是上次断点前最后完成的Task，这里添加状态是避免重复跑这个Task
		key := fmt.Sprintf("%d,%d,%d", tm.LastCodeIndex, tm.LastBuyIndex, tm.LastSellIndex)
		tm.runningTasks.Store(key, TaskStatuSucessed)
	}
	tm.initFlag = 1
	tm.lock.Unlock()

}

func (tm *TasksManager) clean() {
	err := tm.client.Close()
	if err != nil {
		tm.logger.Errorf("TaskManager close client caused error : %s ", err.Error())
	} else {
		tm.logger.Info("TaskManager close client successed. ")
	}
}
