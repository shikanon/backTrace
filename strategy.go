package backTrace


type Analyzer struct {
	BuyPolicies 	[]*Strategy
	SellPolicies	[]*Strategy
}

type (ana *Analyzer) Analyse(data Stock) []int{
	var result []int
	var buys []int
	for d := range data{
		var preBuy Strategy
		n = 0
		for strag:= range ana.BuyPolicies{
			bs := strag.Do(d)
			if n == 0 {
				preBuy = bs
			}else{
				bs = preBuy | bs
				preBuy = bs
			}
			// 或策略
			n += 1
		}
		
		ss := ana.SellOpter.Do(d)
		result = append(result, d)
	}
	return result
}

type Strategy interface{
	Do(StockDailyData) int
}

type BreakOutStrategyBuy struct{}

// 策略初加工所有股票数据
func (bos *BreakOutStrategyBuy) Process(slist []*Stock) []*Stock {
	return slist
}

// 根据特征字段判断是否买入
func (bos *BreakOutStrategyBuy) Do(s StockDailyData) int {
	return 0
}

type MACDStrategySell struct{}

// 策略初加工所有股票数据
func (bos *MACDStrategySell) Process(slist []*Stock) []*Stock {
	return slist
}

// 根据特征字段判断是否卖出
func (macd *MACDStrategySell) Do(s StockDailyData) int {
	return 0
}
