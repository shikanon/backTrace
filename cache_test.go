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
