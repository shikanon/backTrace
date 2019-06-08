package backTrace

//评估器
type Estimator struct {
	TotalReturnRate   float32 //收益率
	ReturnRatePerYear float32 // 年化收益率
	WinningProb       float32 //胜率
	AlphaEarnings     float32 // Alpha收益
	BetaEarnings      float32 // Beta收益
}

//交易
type Transaction struct {
	initCapital  float32
	finalCapital float32
	record       []*TransRecord
}

func (e *Estimator) Estimate(trans *Transaction) {
	var yearNum int
	e.TotalReturnRate = (trans.finalCapital - trans.initCapital) / trans.initCapital
	recordLength := len(trans.record)
	if recordLength > 0 {
		yearNum = trans.record[recordLength].Date.Year() - trans.record[0].Date.Year()
	} else {
		panic("Transaction record cann't be null!")
	}
	e.ReturnRatePerYear = e.TotalReturnRate / float32(yearNum)
}
