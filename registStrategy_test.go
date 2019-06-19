package backTrace

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStrategyRegister(t *testing.T) {
	var reg StrategyRegister
	buy := BreakOutStrategyBuy{WindowsNum: 60}
	sell := BreakOutStrategySell{WindowsNum: 40}
	reg.Regist(buy)
	reg.Regist(sell)
	storeBuy, ok := reg.Value.Load("BreakOutStrategyBuy/WindowsNum:60")
	if !ok {
		log.Fatal("Error: store error!")
	}
	storeSell, ok := reg.Value.Load("BreakOutStrategySell/WindowsNum:40")
	if !ok {
		log.Fatal("Error: store error!")
	}
	assert.Equal(t, storeBuy, buy)
	assert.Equal(t, storeSell, sell)
}
