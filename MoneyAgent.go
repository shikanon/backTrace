package backTrace

import (
	"fmt"
	"time"
)

//买入
const OPT_BUY  = 0

//卖
const OPT_SELL  = 1

//什么也不做
const OPT_HOLD = 2


type MyStock struct {
	vol  float64 //持有股票量
	price float64 //价位
}

//交易记录
type TransRecord struct {
	Date time.Time // 日期
	OptStrategy  int   // 分析者给出的指令，操作类型，buy为1，sell为0, 什么也不做为2
	OptFinal  int      // 最终执行的操作类型，buy为1，sell为0, 什么也不做为2
	//Stock Stock   // 股票
}

func (tran *TransRecord) string() string {
	return fmt.Sprintf("date: %s, total: %s, free: %s",tran.OptStrategy,tran.OptFinal)
}


//资金状态记录
type MoneyRecord struct {
	date time.Time //日期
	myStocks map[string]*MyStock //当前持有的股票
	freeMoney float64  //空闲资金
	totalMoney float64   //总资金, 等于股票 + 空闲资金
}


func (mr *MoneyRecord) string() string {
	return fmt.Sprintf("date: %s, total: %s, free: %s",mr.date,mr.totalMoney,mr.freeMoney)
}

type MoneyAgent struct {
	currentMoney  MoneyRecord //该经理当前资金状况
	historyMoney []MoneyRecord  //该投资经理的资金变化记录
	historyTrans []TransRecord  //该投资经理的交易记录
	Analyzer
}

//资金经理开始干活了,他需要对这个股票的所有数据进行分析
func (agent *MoneyAgent) WorkForSingle(stocks Stock){
	//分析数据，获得买入卖出操作指令
	points:= agent.Analyzer.Analyse(stocks)

	for index,dayData := range stocks{

		opStready, opFinal := points[index], points[index]

		//判断资金是否足够买入
		if float64(agent.currentMoney.freeMoney) < dayData.Close {
			opFinal = OPT_HOLD
		}else if agent.currentMoney.myStocks[dayData.Code].vol ==0 && opFinal == OPT_SELL {
			opFinal = OPT_HOLD
		}

		switch opStready {
		case OPT_BUY:
			//修改交易记录，资金状态
			agent.buy(dayData, opFinal, opStready)
			fmt.Printf("buy %s\n",dayData.Code)
		case OPT_SELL:
			//修改交易记录，资金状态
			agent.sell(dayData, opFinal, opStready)
			fmt.Printf("sell %s\n",dayData.Code)
		case OPT_HOLD:
			//记录操作
			agent.hold(dayData, opFinal, opStready)
			fmt.Printf("hold %s\n",dayData.Code)
		}
	}
}



//买入
func (agent *MoneyAgent) buy(dayData *StockDailyData, opFinal int, opStready int){

	//修改当前状态
	totalPiece := agent.currentMoney.freeMoney/dayData.Close  //计算可买股数
	agent.currentMoney.freeMoney -= totalPiece*dayData.Close  //更新空闲资金
	agent.currentMoney.date = dayData.Date   //记录资金变化的时间
	agent.currentMoney.myStocks[dayData.Code].vol += totalPiece //更新持有股数
	agent.currentMoney.myStocks[dayData.Code].price = dayData.Close  //记录持有股票的价格

	//将当前资金变化状态记录到历史状态中
	agent.historyMoney = append(agent.historyMoney,agent.currentMoney)


	//将当天的交易记录追加到历史状态中
	newTran := TransRecord{Date: dayData.Date,OptStrategy: opStready}
	agent.historyTrans = append(agent.historyTrans,newTran)


}


//卖出股票
func (agent *MoneyAgent) sell(dayData *StockDailyData, opFinal int, opStready int){
	//修改当前状态
	price:= agent.currentMoney.myStocks[dayData.Code].price
	vol := agent.currentMoney.myStocks[dayData.Code].vol

	//不再持有股票
	delete(agent.currentMoney.myStocks,dayData.Code)

	totalSell := vol*dayData.Close  //计算卖出后可得多少钱
	agent.currentMoney.freeMoney += totalSell  //更新空闲资金
	agent.currentMoney.date = dayData.Date   //记录资金变化的时间

	agent.currentMoney.totalMoney -= price*vol  //减去股票的钱
	agent.currentMoney.totalMoney += totalSell  //加上卖出后得到的钱

	//将当前资金变化状态记录到历史状态中
	agent.historyMoney = append(agent.historyMoney,agent.currentMoney)


	//将当天的交易记录追加到历史状态中
	newTran := TransRecord{Date: dayData.Date,OptStrategy: opStready}
	agent.historyTrans = append(agent.historyTrans,newTran)


}

//继续持有股票
func  (agent *MoneyAgent) hold(dayData *StockDailyData, opFinal int, opStready int){
	vol := agent.currentMoney.myStocks[dayData.Code].vol

	//判断是否持有股票
	if vol > 0 {//持有股票
		//记录股价的变化导致资金变化
		oldPrice := agent.currentMoney.myStocks[dayData.Code].price
		agent.currentMoney.myStocks[dayData.Code].price = dayData.Close  //更新股价
		agent.currentMoney.totalMoney -= oldPrice*agent.currentMoney.myStocks[dayData.Code].vol //减去旧的股票资金
		agent.currentMoney.totalMoney += dayData.Close*agent.currentMoney.myStocks[dayData.Code].vol //更新新的股票资金
	}
	agent.currentMoney.date = dayData.Date  //更新时间

	//将当前资金变化状态记录到历史状态中
	agent.historyMoney = append(agent.historyMoney,agent.currentMoney)

	//将当天的交易记录追加到历史状态中
	newTran := TransRecord{Date: dayData.Date,OptStrategy: opStready}
	agent.historyTrans = append(agent.historyTrans,newTran)


}

func codeToStr(op int) string{
	switch op {
	case OPT_BUY:
		return "买入"
	case OPT_SELL:
		return "卖出"
	case OPT_HOLD:
		return "不做操作"
	}
	return "操作类型错误,默认为不做操作"
}


func (agent *MoneyAgent) PrintHistoryInfo(){
	for _,tran:= range agent.historyTrans{
		opStrategy:= codeToStr(tran.OptStrategy)
		opFinal:= codeToStr(tran.OptFinal)
		fmt.Printf("date: %s, 策略: %s, 实际: %s",tran.Date,opStrategy,opFinal)
	}
}
