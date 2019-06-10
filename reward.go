package backTrace

import (
	"errors"
	"fmt"
)

//评估器
type Estimator struct {
	TotalReturnRate   float32 //收益率
	ReturnRatePerYear float32 // 年化收益率
	WinningProb       float32 //胜率
	ProfitExpect      float32 //盈利期望
	LossExpect        float32 //亏损期望
	AlphaEarnings     float32 // Alpha收益
	BetaEarnings      float32 // Beta收益
}

func (e *Estimator) String() string {
	return fmt.Sprintf("收益率TotalReturnRate:%f \n 年化收益率ReturnRatePerYear:%f \n 胜率WinningProb:%f\n 盈利期望ProfitExpect:%f\n 亏损期望LossExpect:%f\n",
		e.TotalReturnRate*100, e.ReturnRatePerYear*100, e.WinningProb*100, e.ProfitExpect*100, e.LossExpect*100)
}

func CreateEstimator(data *ProfileData) (*Estimator, error) {
	var subYearNum float32
	totalReturnRate := (data.FinalCapital - data.InitCapital) / data.InitCapital
	recordLength := len(data.HistoryMoney)
	if recordLength > 0 {
		subYearNum = float32(data.HistoryMoney[recordLength-1].Date.Sub(data.HistoryMoney[0].Date).Hours()) / (365 * 24)
	} else {
		return nil, errors.New("Transaction record cann't be null!")
	}
	returnRatePerYear := totalReturnRate / subYearNum
	win_n := 0
	var win_value float32
	var loss_value float32
	for _, r := range data.Record {
		if r.SellPrice-r.BuyPrice > 0 {
			win_n += 1
			win_value += (r.SellPrice - r.BuyPrice) / r.BuyPrice
		} else {
			loss_value += (r.SellPrice - r.BuyPrice) / r.BuyPrice
		}
	}
	winningProb := float32(win_n) / float32(recordLength)
	prefitExpect := win_value / float32(win_n)
	lossExpect := loss_value / float32(recordLength-win_n)
	return &Estimator{TotalReturnRate: totalReturnRate,
		ReturnRatePerYear: returnRatePerYear,
		WinningProb:       winningProb,
		ProfitExpect:      prefitExpect,
		LossExpect:        lossExpect}, nil
}
