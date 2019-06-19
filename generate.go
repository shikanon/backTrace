package backTrace

func GenerateAllStrage() (reg StrategyRegister) {
	for i := 5; i <= 90; i++ {
		buy := BreakOutStrategyBuy{WindowsNum: i}
		reg.Regist(buy)
	}
	for i := 5; i <= 90; i++ {
		sell := BreakOutStrategySell{WindowsNum: i}
		reg.Regist(sell)
	}
	return
}
