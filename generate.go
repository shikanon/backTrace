package backTrace

func GenerateAllBuyStrage() (reg StrategyRegister) {
	var breakoutParam = []int{5, 6, 9, 10, 11, 12, 14, 19, 20, 21, 29, 30, 31, 39, 40, 41, 42, 49, 50, 51, 59, 60, 61, 62}
	for i := range breakoutParam {
		buy := BreakOutStrategyBuy{WindowsNum: i}
		reg.Regist(buy)
	}
	for i := 1; i <= 3; i++ {
		for j := 1; j <= 3; j++ {
			kdjBuy := KDJtStrategyBuy{
				WindowsNum: 9 * i,
				KWindows:   3 * i,
				DWindows:   3 * i,
				BuyRule:    j,
			}
			reg.Regist(kdjBuy)
		}
	}
	return
}

func GenerateAllSellStrage() (reg StrategyRegister) {
	var breakoutParam = []int{5, 6, 9, 10, 11, 12, 14, 19, 20, 21, 29, 30, 31, 39, 40, 41, 42, 49, 50, 51, 59, 60, 61, 62}
	for i := range breakoutParam {
		sell := BreakOutStrategySell{WindowsNum: i}
		reg.Regist(sell)
	}
	for i := 1; i <= 3; i++ {
		for j := 1; j <= 3; j++ {
			kdjSell := KDJtStrategySell{
				WindowsNum: 9 * i,
				KWindows:   3 * i,
				DWindows:   3 * i,
				SellRule:   j,
			}
			reg.Regist(kdjSell)
		}
	}
	return
}
