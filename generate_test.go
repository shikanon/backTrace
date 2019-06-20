package backTrace

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateAllStrage(t *testing.T) {
	regBuy := GenerateAllBuyStrage()
	assert.Equal(t, 86, len(regBuy.Names))
	regSell := GenerateAllSellStrage()
	assert.Equal(t, 86, len(regSell.Names))
	buyMethod, ok := regBuy.Value.Load(regBuy.Names[0])
	if !ok {
		panic("StrategyRegister loaded failed!")
	}
	sellMethod, ok := regSell.Value.Load(regSell.Names[0])
	if !ok {
		panic("StrategyRegister loaded failed!")
	}
	buy := buyMethod.(BreakOutStrategyBuy)
	sell := sellMethod.(BreakOutStrategySell)
	ana := Analyzer{BuyPolicies: []Strategy{&buy},
		SellPolicies: []Strategy{&sell}}

	agent := MoneyAgent{initMoney: 10000, Analyzer: ana}
	agent.Init()
}
