package backTrace

func GenerateAllBuyStrage() (reg StrategyRegister) {
	for i := 5; i <= 90; i++ {
		buy := BreakOutStrategyBuy{WindowsNum: i}
		reg.Regist(buy)
	}
	for i := 1; i <= 3; i++ {
		kdjBuy := KDJtStrategyBuy{WindowsNum: 9 * i, KWindows: 3 * i, DWindows: 3 * i}
		reg.Regist(kdjBuy)
	}
	return
}

func GenerateAllSellStrage() (reg StrategyRegister) {
	for i := 5; i <= 90; i++ {
		sell := BreakOutStrategySell{WindowsNum: i}
		reg.Regist(sell)
	}
	return
}
