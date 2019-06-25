package backTrace

func GenerateAllBuyStrage() (reg StrategyRegister) {
	for i := 5; i <= 70; i++ {
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
	for i := 5; i <= 70; i++ {
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
