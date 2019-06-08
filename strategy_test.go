package backTrace

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestAnalyzer(t *testing.T) {
	testLogger := logrus.WithFields(logrus.Fields{
		"function": "TestGetStock()",
	})
	stockData := GetSockData("600018")
	if len(stockData) > 0 {
		testLogger.Infof("find stock code numbers is %d", len(stockData))
		buy := BreakOutStrategyBuy{}
		ana := Analyzer{BuyPolicies: []Strategy{&buy},
			SellPolicies: []Strategy{&buy}}
		res, err := ana.Analyse(stockData)
		if err != nil {
			testLogger.Fatal("analyse function is err!", err)
		}
		assert.Equal(t, len(res), len(stockData))
	} else {
		testLogger.Fatal("can't find the stock in the database!")
	}

}
