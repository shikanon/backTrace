package backTrace

import (
	"bufio"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	TaskStatuSucessed = int8(2)
	TaskStatuFailed   = int8(3)
	TaskStatuWait     = int8(0)
	TaskStatuRunning  = int8(1)
	TaskNone          = "-999"
	RedoTaskEvent     = "EVENT"
	RedoCheckPoint    = "POINT"
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

	//---------------------------------------预加载code

	//successedTaskForCode:=make(map[string]int32,defaultPreLoad)  //用于跟中code对应task的完成情况, 与默认加载code

	//defaultPreLoad := uint32(10)
	//preLoadStartIndex := uint32(0)
	//preLoadEndIndex := defaultPreLoad //默认加载3个股票数据

	//fmt.Printf("allCodesCount: %d \n",allCodesCount)

	//一个code需要测试这么多个策略组合
	//oneCodeNeedTest := uint32(len(buyReg.Names) * len(sellReg.Names))

	//计算得到全部的Task数目，h个code 与 m个买策略 、 n个卖策略 笛卡尔积得到 h * m * n个 Task
	//allTaskCount := allCodesCount * oneCodeNeedTest

	//等待分配Task执行
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
					schedulerLogger.Debugf("%s start task (%s,%s,%s)", workerId, task.Code, task.BuyStragety, task.SellStragety)
					ana := Analyzer{BuyPolicies: []Strategy{buy},
						SellPolicies: []Strategy{sell}}
					agent := MoneyAgent{initMoney: 10000, Analyzer: ana}
					//经理需要做好准备后才能开始工作
					agent.Init()
					err = agent.WorkForSingle(*stock)

					if err != nil {
						//flag = false
						schedulerLogger.Errorf("%s agent.WorkForSingle cased error by Code %s, error: %s \n", workerId, task.Code, err.Error())
						schedulerLogger.Errorf("%s aborted task (%s,%s,%s)", workerId, task.Code, task.BuyStragety, task.SellStragety)
					} else {

						//评估策略效果
						result := agent.GetProfileData()
						estimator, err := CreateEstimator(&result)
						if err != nil {
							schedulerLogger.Errorf("CreateEstimator caused Error: %s", err.Error())
							schedulerLogger.Errorf("%s aborted task (%s,%s,%s)", workerId, task.Code, task.BuyStragety, task.SellStragety)
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
							schedulerLogger.Errorf("save RewardRecord caused error: %s", err.Error())
						} else {
							schedulerLogger.Debugf("%s Finish task (%s,%s,%s)", workerId, task.Code, task.BuyStragety, task.SellStragety)
						}
					}
				}

				statu := TaskStatuSucessed
				if flag == false {
					statu = TaskStatuFailed
				}

				taskStatu := TaskStatus{
					task:  &task,
					statu: statu,
					msg:   msg,
				}
				finalChan <- taskStatu //返回task执行结果
			}

			if flag == false {
				//通知taskManager,该goroutine已经完成所有任务，退出了
				taskStatu := TaskStatus{
					task: &Task{Code: TaskNone},
				}
				finalChan <- taskStatu
			}

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
	preLoadStratIndex := tm.lastCodeIndex
	preLoadEndIndex := tm.lastCodeIndex + 5

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

	lastCodeIndex := tm.lastCodeIndex
	lastBuyIndex := tm.lastBuyIndex
	lastSellIndex := tm.lastSellIndex

	successedFlag := true
	for codeIndex, code := range codes[lastCodeIndex:] {
		//获取买入策略
		for buyIndex, buyName := range buyReg.Names[lastBuyIndex:] {
			//获取卖出策略
			for sellIndex, sellName := range sellReg.Names[lastSellIndex:] {
				//构建一个Task
				//tasks = append(tasks, Task{Code: Code, buyStragety: buyName, sellStragety: sellName})
				//taskStr := Code + "," + buyName + "," + sellName
				//id := (CodeIndex + 1) * (buyIndex + 1) * (sellIndex + 1)

				if atomic.LoadUint32(&tm.stop) > 0 { //tm检查到有task失败,程序中断
					schedulerLogger.Errorf("Generate Task is breakout ,because tm.Stop > 0.")
					successedFlag = false
					break
				}
				t := Task{CodeIndex: int32(codeIndex), Code: code, BuyIndex: int32(buyIndex), BuyStragety: buyName,
					SellIndex: int32(sellIndex), SellStragety: sellName}
				schedulerLogger.Infof("add task %s", fmt.Sprintf("%s,%s,%s", code, buyName, sellName))
				isFinished := tm.AddTask(t) //记录task,并检查是否已经完成了的
				if isFinished == false {
					taskChan <- t
					schedulerLogger.Debug(t.String(","))
				}

				newIndex := atomic.LoadInt32(&tm.lastCodeIndex)

				if newIndex > preLoadEndIndex && preLoadEndIndex != allCodesCount {
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
		if tm.finishCount == canRunTaskNum { //等待所有goroutine完成task退出
			schedulerLogger.Infof("all task were done ,now Taskmanager clean it's tmp file %s", tm.redoLogFile)
			tm.clean()
			break
		} else if atomic.LoadUint32(&tm.stop) > 0 { //强行停止
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
	task  *Task
	statu int8 //
	msg   string
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
		q.Head.Prev = nil
		q.Head = q.Head.Next
	}
}

//用于跟踪Task的运行状况
type TasksManager struct {
	logger            *logrus.Entry
	writer            *os.File
	redoLogFile       string
	runningTasks      sync.Map
	waitForCheckPoint IndexQueue
	lastCodeIndex     int32
	lastBuyIndex      int32
	lastSellIndex     int32
	finishCount       uint32
	//clean             uint32
	recoverModel bool
	lock         sync.RWMutex
	initFlag     int64
	stop         uint32
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

	if s.task.Code == TaskNone {
		atomic.AddUint32(&tm.finishCount, 1)
	} else {
		switch s.statu {
		case TaskStatuFailed:
			atomic.AddUint32(&tm.stop, 1)
			tm.logger.Infof("Program is going to shutdown. %s", s.msg)
		case TaskStatuSucessed:
			tm.SaveStatus(s)
		}
	}

}

func (tm *TasksManager) SaveStatus(s *TaskStatus) {
	//写REDO日志
	key := fmt.Sprintf("%d,%d,%d", s.task.CodeIndex, s.task.BuyIndex, s.task.SellIndex)
	tm.runningTasks.Store(key, TaskStatuSucessed)

	//在恢复的过程中不产生redo日志
	if !tm.recoverModel {

		//IO写入task完成事件
		log := RedoTaskEvent + RedoDelim + s.task.String(RedoDelim) + "\n"
		_, err := tm.writer.WriteString(log)

		if err != nil {
			panic(errors.Errorf("TaskManager can't write redo log by err: %s", err.Error()))
		}
	}

	//checkpoint 并 移除map已经写入checkpoint的task
	var done []string
	//循环更新下标
	currentNode := tm.waitForCheckPoint.Head
	nextDoneKey := key

	//needToUpdateIndex := false

	for {

		val, ok := tm.runningTasks.Load(nextDoneKey)
		if ok != true {
			panic(errors.Errorf("when tasksManager update task status by key %s ,it can't find the key in map", nextDoneKey))
		}

		if val == TaskStatuSucessed && (currentNode.Key == nextDoneKey) {
			done = append(done, currentNode.Key)
			atomic.CompareAndSwapInt32(&tm.lastCodeIndex, tm.lastCodeIndex, currentNode.T.CodeIndex)
			tm.lastBuyIndex = currentNode.T.BuyIndex
			tm.lastSellIndex = currentNode.T.SellIndex
			currentNode = currentNode.Next
			tm.waitForCheckPoint.Pop() //移除head

			if tm.waitForCheckPoint.Head != nil {
				nextDoneKey = tm.waitForCheckPoint.Head.Key
			} else {
				break
			}
			//needToUpdateIndex = true
		} else {
			break
		}
	}

	//IO 写入更新后的index
	/*log = RedoCheckPoint + RedoDelim + strconv.FormatInt(int64(tm.lastCodeIndex), 10) + RedoDelim +
		strconv.FormatInt(int64(tm.lastBuyIndex), 10) + RedoDelim + strconv.FormatInt(int64(tm.lastSellIndex), 10) + "\n"
	_, err := tm.writer.WriteString(log)

	if err != nil {
		panic(errors.Errorf("TaskManager can't write checkpoint log by error : %s", err.Error()))
	}
	*/
	//移除map中的已经完成的task
	for _, val := range done {
		tm.runningTasks.Delete(val)
	}

}

//TasksManager必须要要调用了该方法后才能正常运作
func (tm *TasksManager) recover() {
	tm.lock.Lock()
	if tm.initFlag != 0 {
		fmt.Println("tasksManager already init!")
		return
	}

	tm.logger = logrus.WithFields(logrus.Fields{
		"function": "TaskManager",
	})

	//进入恢复模式
	tm.recoverModel = true

	file, err := os.OpenFile(tm.redoLogFile, os.O_RDWR, 0666)
	if os.IsNotExist(err) {
		file, err = os.Create(tm.redoLogFile)
	}

	if err != nil {
		panic(errors.Errorf("TaskManager can't open or create file by error: %s", err.Error()))
	}

	tm.writer = file
	reader := bufio.NewReader(file)

	for {
		line, err := reader.ReadString('\n')
		if err != nil || err == io.EOF {
			break
		}
		line = strings.TrimSuffix(line, "\n")
		strArry := strings.Split(line, RedoDelim)

		if len(strArry) < 1 {
			continue
		} else {
			switch strArry[0] {
			case RedoTaskEvent:
				//复现事件
				tm.redoTask(strArry)
				/*case RedoCheckPoint:
				tm.redoCheckPoint(strArry)*/
			}
		}

	}

	tm.recoverModel = false //恢复模式关闭,后续saveStatus操作会产生redo日志
	tm.initFlag = 1
	tm.lock.Unlock()

}

func (tm *TasksManager) redoTask(array []string) {

	code := array[1]
	codeIndex, err := strconv.ParseInt(array[2], 10, 32)
	if err != nil {
		panic(errors.Errorf("redoTask caused error by ParseInt CodeIndex : %s", err.Error()))
	}
	buy := array[3]
	buyIndex, err := strconv.ParseInt(array[4], 10, 32)
	if err != nil {
		panic(errors.Errorf("redoTask caused error by ParseInt buyIndex : %s", err.Error()))
	}
	sell := array[5]
	sellIndex, err := strconv.ParseInt(array[6], 10, 32)
	if err != nil {
		panic(errors.Errorf("redoTask caused error by ParseInt sellIndex : %s", err.Error()))
	}
	t := Task{Code: code, CodeIndex: int32(codeIndex),
		BuyStragety: buy, BuyIndex: int32(buyIndex), SellStragety: sell, SellIndex: int32(sellIndex)}

	tm.AddTask(t)

	ts := TaskStatus{
		task:  &t,
		statu: TaskStatuSucessed,
		msg:   "ok",
	}
	tm.UpdateStatu(&ts)

}

func (tm *TasksManager) clean() {

	err = tm.writer.Close()

	err := os.Remove(tm.redoLogFile)
	if err != nil {
		//如果删除失败则输出 file remove Error!
		fmt.Println("file remove Error!")
		//输出错误详细信息
		fmt.Printf("%s", err)
	} else {
		//如果删除成功则输出 file remove OK!
		fmt.Print("file remove OK!")
	}
}

/*func (tm *TasksManager) redoCheckPoint(array []string) {

}*/
