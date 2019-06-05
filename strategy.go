package backTrace

type Strategy interface {
	BuyOrSell(StockDailyData) int
}

type BreakOutStrategy struct{}

// 策略初加工所有股票数据
func (bos *BreakOutStrategy) Process(slist []*Stock) []*Stock {
	return slist
}

// 根据特征字段判断是否买入
func (bos *BreakOutStrategy) BuyOrSell(s StockDailyData) int {
	return 0
}
