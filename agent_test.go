package backTrace

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestMoneyAgent_GetProfileData(t *testing.T) {

	/*testLogger := logrus.WithFields(logrus.Fields{
		"function": "TestMoneyAgent_GetProfileData()",
	})

	stockData, err := GetSockData("600018")
	if err != nil {
		testLogger.Fatal(err)
	}
	if stockData.Length > 0 {
		testLogger.Infof("find stock Code numbers is %d", stockData.Length)
		//初始化分析者
		buy := BreakOutStrategyBuy{WindowsNum: 60}
		sell := BreakOutStrategySell{WindowsNum: 40}
		ana := Analyzer{BuyPolicies: []Strategy{&buy},
			SellPolicies: []Strategy{&sell}}

		agent := MoneyAgent{initMoney: 10000, Analyzer: ana}

		//经理需要做好准备后才能开始工作
		agent.Init()

		//经理根据指定的策略对单只股票进行操作
		err = agent.WorkForSingle(stockData)

		assert.Equal(t, nil, err)

		result := agent.GetProfileData()
		estimator, err := CreateEstimator(&result)
		if err != nil {
			testLogger.Fatal(err)
		}
		testLogger.Println(estimator.String())

		if result.InitCapital < 0 {
			testLogger.Fatal("the InitCapital can't be less than 0!")
		}

		if result.FinalCapital < 0 {
			testLogger.Fatal("the InitCapital can't be less than 0!")
		}

		lenOfHistory := len(result.HistoryMoney)
		lenOfStocks := stockData.Length
		if lenOfHistory != lenOfStocks {
			testLogger.Fatalf("The len of HistoryMoneyRecord ( %d ) should be the same with the len of StockData ( %d ) !", lenOfHistory, lenOfStocks)
		}
	} else {
		testLogger.Fatal("can't find the stock in the database! the function GetSockData() is error!")
		testLogger.Println(err)
	}*/

}

func testForBenchmark(code string) error {
	stockData, err := GetSockData(code)
	if err != nil {
		return err
	}
	if stockData.Length > 0 {
		//初始化分析者
		buy := BreakOutStrategyBuy{WindowsNum: 60}
		sell := BreakOutStrategySell{WindowsNum: 40}
		ana := Analyzer{BuyPolicies: []Strategy{&buy},
			SellPolicies: []Strategy{&sell}}

		agent := MoneyAgent{initMoney: 10000, Analyzer: ana}

		//经理需要做好准备后才能开始工作
		agent.Init()

		//经理根据指定的策略对单只股票进行操作
		agent.WorkForSingle(stockData)

		result := agent.GetProfileData()
		_, err := CreateEstimator(&result)
		if err != nil {
			return err
		}

	} else {
		return err
	}
	return nil
}

func BenchmarkGetSockData(b *testing.B) {
	b.StartTimer() //重新开始时间
	for i := 0; i < b.N; i++ {
		GetSockData("600018")
	}
}

func BenchmarkAgent(b *testing.B) {
	testLogger := logrus.WithFields(logrus.Fields{
		"function": "BenchmarkAgent()",
	})
	code := "600018"
	stockData, err := GetSockData(code)
	if err != nil {
		testLogger.Fatal(err)
	}
	if stockData.Length > 0 {
		b.ResetTimer()
		b.StartTimer() //重新开始时间
		for i := 0; i < b.N; i++ {
			//初始化分析者
			buy := BreakOutStrategyBuy{WindowsNum: 60}
			sell := BreakOutStrategySell{WindowsNum: 40}
			ana := Analyzer{BuyPolicies: []Strategy{&buy},
				SellPolicies: []Strategy{&sell}}

			agent := MoneyAgent{initMoney: 10000, Analyzer: ana}

			//经理需要做好准备后才能开始工作
			agent.Init()

			//经理根据指定的策略对单只股票进行操作
			agent.WorkForSingle(stockData)

			result := agent.GetProfileData()
			_, err := CreateEstimator(&result)
			if err != nil {
				testLogger.Fatal(err)
			}
		}
	} else {
		testLogger.Fatal(err)
	}
}
