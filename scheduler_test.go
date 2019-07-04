package backTrace

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSchedulerTask(t *testing.T) {

	node := NewNode("localhost", "127.0.0.1", 5555, 1)
	var stocks StockMap
	sc := LocalScheduler{node: node, coresForPerTask: 1, cacheMap: stocks}
	// allStocks := GetAllSockCode()
	testStock := []string{"000001", "000002"}

	//获取买策略
	buyReg := GenerateAllBuyStrage()

	//获取卖策略
	sellReg := GenerateAllSellStrage()

	tm := TasksManager{
		redoLogFile:       "test_checkpoint.txt",
		runningTasks:      make(map[string]int8),
		waitForCheckPoint: IndexQueue{},
		lastCodeIndex:     0,
		lastBuyIndex:      int32(len(buyReg.Names) - 1),
		lastSellIndex:     int32(len(sellReg.Names) - 1),
	}

	tm.recover()
	sc.schedulerTask(&buyReg, &sellReg, testStock, tm)
}

func TestIndexQueue(t *testing.T) {

	testLogger := logrus.WithFields(logrus.Fields{
		"function": "TestIndexQueue()",
	})

	t1 := Task{Code: "001", CodeIndex: 1, BuyIndex: 1, SellIndex: 1}
	t2 := Task{Code: "002", CodeIndex: 1, BuyIndex: 1, SellIndex: 2}
	t3 := Task{Code: "003", CodeIndex: 1, BuyIndex: 2, SellIndex: 1}
	i1 := IndexNode{Key: "001", T: &t1}
	i2 := IndexNode{Key: "002", T: &t2}
	i3 := IndexNode{Key: "003", T: &t3}

	queue := IndexQueue{}

	queue.Insert(&i3)
	queue.Insert(&i1)
	queue.Insert(&i2)

	head := queue.Head

	for {
		if head == nil {
			break
		}
		testLogger.Info(head.Key)
		head = head.Next

	}

	cur := queue.Head
	assert.Equal(t, "001", cur.T.Code)

	cur = cur.Next
	assert.Equal(t, "002", cur.T.Code)

	cur = cur.Next
	assert.Equal(t, "003", cur.T.Code)
}

func TestTaskManager(t *testing.T) {

	/*testLogger := logrus.WithFields(logrus.Fields{
		"function": "TestTaskManager()",
	})
	*/
	testFile := "test_checkpoint.txt"
	tm := TasksManager{
		redoLogFile: testFile,
	}
	tm.recover()

	t1 := Task{Code: "001", CodeIndex: 0, BuyIndex: 0, BuyStragety: "buy0", SellIndex: 0, SellStragety: "sell0"}
	t2 := Task{Code: "002", CodeIndex: 0, BuyIndex: 0, BuyStragety: "buy0", SellIndex: 2, SellStragety: "sell2"}
	t3 := Task{Code: "003", CodeIndex: 1, BuyIndex: 0, BuyStragety: "buy0", SellIndex: 1, SellStragety: "sell1"}
	t4 := Task{Code: "006", CodeIndex: 2, BuyIndex: 1, BuyStragety: "buy1", SellIndex: 0, SellStragety: "sell0"}
	//分配task
	tm.AddTask(t1)
	tm.AddTask(t2)
	tm.AddTask(t3)
	tm.AddTask(t4)

	//模拟task完成了,更新task状态
	ts1 := TaskStatus{task: &t1, statu: TaskStatuSucessed, msg: "ok"}
	ts2 := TaskStatus{task: &t2, statu: TaskStatuSucessed, msg: "ok"}
	ts3 := TaskStatus{task: &t3, statu: TaskStatuSucessed, msg: "ok"}
	ts4 := TaskStatus{task: &t4, statu: TaskStatuSucessed, msg: "ok"}

	tm.SaveStatus(&ts3)

	assert.Equal(t, int32(0), tm.lastCodeIndex)
	assert.Equal(t, int32(0), tm.lastBuyIndex)
	assert.Equal(t, int32(0), tm.lastSellIndex)

	tm.SaveStatus(&ts1)
	tm.SaveStatus(&ts2)

	assert.Equal(t, int32(1), tm.lastCodeIndex)
	assert.Equal(t, int32(0), tm.lastBuyIndex)
	assert.Equal(t, int32(1), tm.lastSellIndex)

	tm.SaveStatus(&ts4)
	tm.clean()

}
