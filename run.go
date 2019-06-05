package backTrace

import (
	"os"

	"github.com/sirupsen/logrus"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	logrus.SetFormatter(&logrus.TextFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	logrus.SetOutput(os.Stdout)

	// Only log level.
	logrus.SetLevel(logrus.InfoLevel)
}

func RunBacktrace() {
	var stocks []*Stock
	for _, code := range GetAllSockCode() {
		stock := GetSockData(code)
		stocks = append(stocks, &stock)
	}
	// 用策略对股票数据做预处理
	buy := BreakOutStrategyBuy{}
	buy.Process(stocks)
	sell := MACDStrategySell{}
	sell.Process(stocks)
	// 执行策略
	agent := StockAgent{}
	agent.Run(&buy, &sell)
}
