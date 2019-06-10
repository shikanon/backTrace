package backTrace

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestGetStock(t *testing.T) {
	testLogger := logrus.WithFields(logrus.Fields{
		"function": "TestGetStock()",
	})
	result, err := GetSockData("600018")
	if err != nil {
		testLogger.Fatal(err)
	}
	if result.Length > 0 {
		testLogger.Infof("find stock code numbers is %d", result.Length)
		assert.Equal(t, result.Code[0], "600018")
	} else {
		testLogger.Fatal("can't find the stock in the database!")
	}

}
