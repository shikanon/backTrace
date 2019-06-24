package backTrace

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSchedulerTask(t *testing.T) {

	node := NewNode("localhost", "127.0.0.1", 5555, 1)
	var stocks StockMap
	sc := LocalScheduler{node: node, coresForPerTask: 1, cacheMap: stocks}
	allStocks := GetAllSockCode()

	c, _ := CreateRedisClient()
	_, err = c.Del(GenBeginFlag, GenEndFlag, taskQueueName).Result()

	//获取买策略
	buyReg := GenerateAllBuyStrage()

	//获取卖策略
	sellReg := GenerateAllSellStrage()

	TasksGenerate(&buyReg, &sellReg, allStocks[:1], c)

	sc.schedulerTask(&buyReg, &sellReg, allStocks[:1], c)
}

func TestCodeCacheList(t *testing.T) {
	x := CodeCacheList{
		nodes: []string{"123", "456", "789"},
	}
	assert.Equal(t, false, x.getNeedLoadFlag("456"))
	assert.Equal(t, true, x.getNeedLoadFlag("789"))
	assert.Equal(t, "123", x.insert("zxc"))
}
