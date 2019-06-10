package backTrace

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

func init() {
	/*	// Log as JSON instead of the default ASCII formatter.
		logrus.SetFormatter(&logrus.TextFormatter{})

		// Output to stdout instead of the default stderr
		// Can be any io.Writer, see below for File example
		logrus.SetOutput(os.Stdout)

		// Only log level.
		logrus.SetLevel(logrus.InfoLevel)*/
}

func RunBacktrace() {
	contextLogger := logrus.WithFields(logrus.Fields{
		"function": "RunBacktrace()",
	})
	var stocks []*StockColumnData
	for _, code := range GetAllSockCode() {
		stock, err := GetSockData(code)
		if err != nil {
			contextLogger.Fatal(err)
		}
		stocks = append(stocks, &stock)
		break
	}

	//初始化分析者
	buy := BreakOutStrategyBuy{WindowsNum: 60}
	sell := BreakOutStrategySell{WindowsNum: 60}
	ana := Analyzer{BuyPolicies: []Strategy{&buy},
		SellPolicies: []Strategy{&sell}}

	agent := MoneyAgent{initMoney: 10000, Analyzer: ana}

	//经理需要做好准备后才能开始工作
	agent.Init()

	//经理根据指定的策略对单只股票进行操作
	for _, stock := range stocks {
		agent.WorkForSingle(*stock)
	}

	//输出交易信息
	r := agent.GetProfileData()
	fmt.Printf("Init:%.2f, final: %.2f \n", r.InitCapital, r.FinalCapital)
	for _, record := range r.Record {
		contextLogger.Printf("beforBuy: %.2f , afterSell: %.2f, buyDate: %s, buyVol: %d, buyPrice: %.2f, sellDate:%s,"+
			" sellVol: %d, sellPrice:%.2f \n", record.InitMoney, record.FinalMoney, record.BuyDate, record.BuyVol, record.BuyPrice,
			record.SellDate, record.SellVol, record.SellPrice)
	}
	estimator, err := CreateEstimator(&r)
	if err != nil {
		contextLogger.Fatal(err)
	}
	contextLogger.Println(estimator.String())
}
