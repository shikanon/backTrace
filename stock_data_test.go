package backTrace

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestGetStock(t *testing.T) {
	var stock *StockDailyData
	testLogger := logrus.WithFields(logrus.Fields{
		"function": "TestGetStock()",
	})
	result, err := GetSockData("600018")
	if err != nil {
		testLogger.Fatal(err)
	}
	if len(result) > 0 {
		testLogger.Infof("find stock code numbers is %d", len(result))
		stock = result[0]
		assert.Equal(t, stock.Code, "600018")
	} else {
		testLogger.Fatal("can't find the stock in the database!")
	}

}
