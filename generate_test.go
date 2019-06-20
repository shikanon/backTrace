package backTrace

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestGenerateAllStrage(t *testing.T) {
	testLogger := logrus.WithFields(logrus.Fields{
		"function": "TestGenerateAllStrage()",
	})
	regBuy := GenerateAllBuyStrage()
	assert.Equal(t, 86, len(regBuy.Names))
	regSell := GenerateAllSellStrage()
	assert.Equal(t, 86, len(regSell.Names))

	buyMethod, err := regBuy.Load(regBuy.Names[0])
	if err != nil {
		testLogger.Fatal(err)
		testLogger.Panic("StrategyRegister loaded failed!")
	}
	sellMethod, err := regSell.Load(regSell.Names[0])
	if err != nil {
		testLogger.Fatal(err)
		testLogger.Panic("StrategyRegister loaded failed!")
	}
	ana := Analyzer{BuyPolicies: []Strategy{buyMethod},
		SellPolicies: []Strategy{sellMethod}}

	agent := MoneyAgent{initMoney: 10000, Analyzer: ana}
	agent.Init()
}
