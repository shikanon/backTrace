package backTrace

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
	var stocks []*Stock
	for _, code := range GetAllSockCode() {
		stock := GetSockData(code)
		stocks = append(stocks, &stock)
		break
	}

	//初始化分析者
	buy := BreakOutStrategyBuy{}
	sell := BreakOutStrategySell{}
	ana := Analyzer{BuyPolicies: []Strategy{&buy},
		SellPolicies: []Strategy{&sell}}

	agent := MoneyAgent{initMoney: 10000, Analyzer: ana}

	//经理需要做好准备后才能开始工作
	agent.init()

	//经理根据指定的策略对单只股票进行操作
	for _, stock := range stocks {
		agent.WorkForSingle(*stock)
	}

	//输出交易信息
	agent.PrintHistoryInfo()

}
