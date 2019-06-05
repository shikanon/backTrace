package backTrace

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestGetStock(t *testing.T) {
	var stock *StockData
	testLogger := logrus.WithFields(logrus.Fields{
		"function": "TestGetStock()",
	})
	result := GetStock("600018")
	if len(result) > 0 {
		testLogger.Infof("find stock code numbers is %d", len(result))
		stock = result[0]
		assert.Equal(t, stock.Code, "600018")
	} else {
		testLogger.Fatal("can't find the stock in the database!")
	}

}
