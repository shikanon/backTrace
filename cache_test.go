package backTrace

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func BenchmarkStockMapr(b *testing.B) {
	testLogger := logrus.WithFields(logrus.Fields{
		"function": "BenchmarkStockMapr()",
	})
	var stocks StockMap
	codes := GetAllSockCode()
	var ReadCodes = make([]string, 5)
	copy(ReadCodes, codes[0:5])
	stocks.Ready(ReadCodes)
	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		pos := i % 5
		c := ReadCodes[pos]
		_, err := stocks.Load(c)
		if err != nil {
			testLogger.Fatal(err)
		}
	}
}

func TestLoadAndSave(t *testing.T) {
	testLogger := logrus.WithFields(logrus.Fields{
		"function": "TestLoadAndSave()",
	})
	var stocks StockMap
	stock, err := stocks.Load("1000001")
	testLogger.Info(stock)
	if stock != nil {
		testLogger.Fatal("错误,应返回空类型")
	}
	if err != nil {
		testLogger.Infof("这仅仅是一条错误测试信息：%v", err)
	}
}
