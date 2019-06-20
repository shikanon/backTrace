package backTrace

import "testing"

func TestSchedulerTask(t *testing.T) {

	node := NewNode("localhost", "127.0.0.1", 5555, 2)
	var stocks StockMap
	sc := LocalScheduler{node: node, coresForPerTask: 1, cacheMap: stocks}
	allStocks := GetAllSockCode()
	sc.schedulerTask(allStocks)
}
