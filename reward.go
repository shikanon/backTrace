package backTrace

import (
	"errors"
	"time"
)

//评估器
type Estimator struct {
	TotalReturnRate   float32 //收益率
	ReturnRatePerYear float32 // 年化收益率
	WinningProb       float32 //胜率
	AlphaEarnings     float32 // Alpha收益
	BetaEarnings      float32 // Beta收益
}

type IncommeRecord struct {
	buyDate    time.Time
	buyVol     int
	initMoney  float32
	finalMoney float32
}

//交易
type Transaction struct {
	InitCapital  float32
	FinalCapital float32
	HistoryMoney []*MoneyRecord
	Record       []*IncommeRecord
}

func (e *Estimator) Estimate(trans *Transaction) error {
	var yearNum int
	e.TotalReturnRate = (trans.FinalCapital - trans.InitCapital) / trans.InitCapital
	recordLength := len(trans.Record)
	if recordLength > 0 {
		yearNum = trans.HistoryMoney[recordLength].date.Year() - trans.HistoryMoney[0].date.Year()
	} else {
		return errors.New("Transaction record cann't be null!")
	}
	e.ReturnRatePerYear = e.TotalReturnRate / float32(yearNum)
	return nil
}
