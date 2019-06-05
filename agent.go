package backTrace

import (
	"time"
)

type Agent interface {
	Run(Strategy)
}

type TransactionList struct {
	Date time.Time // 日期
	Opt  int8      // 操作类型，buy为1，sell为0
}

type HoldStock struct {
	Code string
	Vol  int //持有股票量
}

type StockAgent struct {
	StockData        []*Stock
	TransList        []*TransactionList //交易列表
	TotalCapital     float32            //总资金
	totalCash        float32            //总现金
	totalStockVolumn []*HoldStock       //各类股票总持有数
}

func (agent *StockAgent) Run(s Strategy) {

}
