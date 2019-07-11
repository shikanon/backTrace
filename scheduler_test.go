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
		waitForCheckPoint: IndexQueue{},
		LastCodeIndex:     0,
		//LastBuyIndex:      int32(len(buyReg.Names) - 1),
		LastBuyIndex: 0,
		//LastSellIndex:     int32(len(sellReg.Names) - 1),
		LastSellIndex: 0,
	}

	tm.Recover()
	sc.schedulerTask(&buyReg, &sellReg, testStock, &tm)
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

	testLogger := logrus.WithFields(logrus.Fields{
		"function": "TestTaskManager()",
	})

	testKey := "test_latestIndex"

	//获取redis client
	client, err := CreateRedisClient()
	if err != nil {
		testLogger.Fatalf("TaskManager can't get redis client. error: %s\n", err.Error())
		panic(err)
	}
	client.Del(testKey)

	tm := TasksManager{
		redisKey: testKey,
	}
	tm.Recover()

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
	ts1 := TaskStatus{Task: &t1, Statu: TaskStatuSucessed, Msg: "ok"}
	ts2 := TaskStatus{Task: &t2, Statu: TaskStatuSucessed, Msg: "ok"}
	ts3 := TaskStatus{Task: &t3, Statu: TaskStatuSucessed, Msg: "ok"}
	ts4 := TaskStatus{Task: &t4, Statu: TaskStatuSucessed, Msg: "ok"}

	/*key:= fmt.Sprintf("%d,%d,%d",tm.LastCodeIndex,tm.LastBuyIndex,tm.LastSellIndex)
	fmt.Println(key)*/

	tm.SaveStatus(&ts3)

	assert.Equal(t, int32(0), tm.LastCodeIndex)
	assert.Equal(t, int32(0), tm.LastBuyIndex)
	assert.Equal(t, int32(0), tm.LastSellIndex)

	/*	key = fmt.Sprintf("%d,%d,%d",tm.LastCodeIndex,tm.LastBuyIndex,tm.LastSellIndex)
		fmt.Println(key)
	*/
	tm.SaveStatus(&ts1)
	tm.SaveStatus(&ts2)

	/*key = fmt.Sprintf("%d,%d,%d",tm.LastCodeIndex,tm.LastBuyIndex,tm.LastSellIndex)
	fmt.Println(key)
	*/
	assert.Equal(t, int32(1), tm.LastCodeIndex)
	assert.Equal(t, int32(0), tm.LastBuyIndex)
	assert.Equal(t, int32(1), tm.LastSellIndex)

	tm.SaveStatus(&ts4)
	tm.clean()

}
