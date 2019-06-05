package backTrace

type Strategy interface {
	Buy(StockDailyData) int
	Sell(StockDailyData) int
}

type BuyStrategy interface {
	Buy(StockDailyData) int
}

type SellStrategy interface {
	Sell(StockDailyData) int
}

type BreakOutStrategyBuy struct{}

// 策略初加工所有股票数据
func (bos *BreakOutStrategyBuy) Process(slist []*Stock) []*Stock {
	return slist
}

// 根据特征字段判断是否买入
func (bos *BreakOutStrategyBuy) Buy(s StockDailyData) int {
	return 0
}

type MACDStrategySell struct{}

// 策略初加工所有股票数据
func (bos *MACDStrategySell) Process(slist []*Stock) []*Stock {
	return slist
}

// 根据特征字段判断是否卖出
func (macd *MACDStrategySell) Sell(s StockDailyData) int {
	return 0
}
