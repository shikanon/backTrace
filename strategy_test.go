package backTrace

import (
	"testing"

	"github.com/sirupsen/logrus"
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
		ana.Analyse(stockData)
	} else {
		testLogger.Fatal("can't find the stock in the database!")
	}

}
