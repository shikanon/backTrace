package backTrace

import (
	"fmt"
	"time"
)

//买入
const BUY  = 0

//卖
const SELL  = 1

//什么也不做
const HOLD = 2



type MyStock struct {
	Code string
	Vol  int //持有股票量
}

//交易记录
type TransRecord struct {
	Date time.Time // 日期
	Opt  int8      // 操作类型，buy为1，sell为0, 什么也不做为2
	Stock Stock   // 股票
}


//资金状态记录
type MoneyRecord struct {
	Date time.Time //日期
	freezeMoney int8 // 当前
	freeMoney int8  //空闲资金
	totalMoney int8   //总资金, 等于股票 + 空闲资金
}


type MoneyAgent struct {
	myStocks []*MyStock  //持有的股票数
	currentMoney MoneyRecord    //该经理当前资金状况
	historyMoney []MoneyRecord  //该投资经理的资金变化记录
	historyTrans []TransRecord  //该投资经理的交易记录
	Analyzer
}

//资金经理开始干活了,他需要对这个股票的所有数据进行分析
func (agent *MoneyAgent) WorkForSingle(stocks Stock){
	//分析数据，获得买入卖出操作指令
	points:= agent.Analyzer.Analyse(stocks)

	for index,dayData := range stocks{
		switch points[index] {
		case BUY:
			//修改交易记录，资金状态
			agent.buy(dayData)
			fmt.Printf("buy %s\n",dayData.Code)
		case SELL:
			//修改交易记录，资金状态
			fmt.Printf("sell %s\n",dayData.Code)
		case HOLD:
			//记录操作
			fmt.Printf("hold %s\n",dayData.Code)
		}
	}
}


func (agent *MoneyAgent) buy(dayData *StockDailyData){

	//保存当前状态

	//修改当前状态

}