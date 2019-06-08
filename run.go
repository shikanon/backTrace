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



	initMoney:= MoneyRecord{totalMoney:10000,freeMoney:10000}

	//初始化分析者
	ana:= Analyzer{BreakOutStrategyBuy{},}

	agent := MoneyAgent{currentMoney:initMoney,Analyzer: ana}

	//经理根据指定的策略对单只股票进行操作
	for _,stock:= range stocks{
		agent.WorkForSingle(*stock)
	}

	//输出交易信息
	agent.PrintHistoryInfo()






}
