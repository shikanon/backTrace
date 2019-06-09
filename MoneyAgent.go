package backTrace

import (
	"fmt"
	"time"
)

//买入
const OPT_BUY = 0

//卖
const OPT_SELL = 1

//什么也不做
const OPT_HOLD = 2

const RateBuy = 0.003
const RateSell = 0.003

type MyStock struct {
	vol   float32 //持有股票量
	price float32 //价位
}

//交易记录
type TransRecord struct {
	Date        time.Time // 日期
	OptStrategy int       // 分析者给出的指令，操作类型，buy为1，sell为0, 什么也不做为2
	OptFinal    int       // 最终执行的操作类型，buy为1，sell为0, 什么也不做为2
	//Stock Stock   // 股票
}

func (tran *TransRecord) string() string {
	return fmt.Sprintf("date: %s, total: %d, free: %d", tran.Date, tran.OptStrategy, tran.OptFinal)
}

//资金状态记录
type MoneyRecord struct {
	date       time.Time           //日期
	myStocks   map[string]*MyStock //当前持有的股票
	freeMoney  float32             //空闲资金
	totalMoney float32             //总资金, 等于股票 + 空闲资金
}

func (mr *MoneyRecord) string() string {
	return fmt.Sprintf("date: %s, total: %f, free: %f", mr.date, mr.totalMoney, mr.freeMoney)
}

type MoneyAgent struct {
	currentMoney MoneyRecord   //该经理当前资金状况
	historyMoney []MoneyRecord //该投资经理的资金变化记录
	historyTrans []TransRecord //该投资经理的交易记录
	Analyzer
}

//资金经理开始干活了,他需要对这个股票的所有数据进行分析
func (agent *MoneyAgent) WorkForSingle(stocks Stock) {
	//分析数据，获得买入卖出操作指令
	points, err := agent.Analyzer.Analyse(stocks)
	if err != nil {
		fmt.Printf("出现异常，退出程序：%s", err)
		return
	}

	/*for index,value:= range points{
		fmt.Printf("date: %s,code :%s, 策略返回结果是： %s\n",stocks[index].Date,stocks[index].Code,value)
	}*/

	for index, dayData := range stocks {

		switch points[index] {
		case OPT_BUY:
			//修改交易记录，资金状态
			agent.buy(dayData, points[index])
			//fmt.Printf("buy %s\n", dayData.Code)
		case OPT_SELL:
			//修改交易记录，资金状态
			agent.sell(dayData, points[index])
			//fmt.Printf("sell %s\n", dayData.Code)
		case OPT_HOLD:
			//记录操作
			agent.hold(dayData, points[index])
			//fmt.Printf("hold %s\n", dayData.Code)
		}
	}
}

//买入
func (agent *MoneyAgent) buy(dayData *StockDailyData, opStready int) {

	//修改当前状态
	maxPiece := agent.currentMoney.freeMoney / dayData.Close //粗略计算可买股数
	totalCost := maxPiece * dayData.Close                    //粗略估计总的费用
	fee := totalCost * RateBuy                               //计算交易费

	for true {
		//判断扣除手续费之后，算上手续费后，不足购买maxPiece 股，并且maxPiece>1,可以减少买入股数，以完成交易
		if totalCost > agent.currentMoney.freeMoney-fee {
			maxPiece -= 1
			totalCost = maxPiece * dayData.Close
			fee = totalCost * RateBuy
		} else {
			break //有足够的钱支付  购买超过1股以及 手续费
		}

		//不够钱交易1股，跳转到持有不变
		if maxPiece <= 0 {
			agent.hold(dayData, opStready)
			return
		}

	}

	//先检查是否有持有当前股票
	myStock := agent.currentMoney.myStocks[dayData.Code]

	// 为空则表示之前没有持有该股票
	if myStock == nil {
		agent.currentMoney.myStocks[dayData.Code] = &MyStock{0, 0}
		myStock = agent.currentMoney.myStocks[dayData.Code]
	} else { //有值说明之前持有股票，股价的变化需要更新总资金数
		oldTotal := myStock.price * myStock.vol
		newTotal := myStock.vol * dayData.Close
		agent.currentMoney.totalMoney -= oldTotal
		agent.currentMoney.totalMoney += newTotal
	}

	agent.currentMoney.freeMoney -= totalCost //扣除买股的钱
	agent.currentMoney.freeMoney -= fee       //扣除手续费
	agent.currentMoney.date = dayData.Date

	agent.currentMoney.totalMoney -= fee //购买了股票，总资金的变化只有减去

	//记录股数的变化
	myStock.vol += maxPiece
	myStock.price = dayData.Close

	//将当前资金变化状态记录到历史状态中
	agent.historyMoney = append(agent.historyMoney, agent.currentMoney)

	//将当天的交易记录追加到历史状态中
	newTran := TransRecord{Date: dayData.Date, OptStrategy: opStready, OptFinal: OPT_BUY}
	agent.historyTrans = append(agent.historyTrans, newTran)

}

//卖出股票
func (agent *MoneyAgent) sell(dayData *StockDailyData, opStready int) {

	//先检查是否有持有当前股票
	myStock := agent.currentMoney.myStocks[dayData.Code]

	//没有持有当前股票,什么也做不了
	if myStock == nil {
		agent.hold(dayData, opStready)
		return
	}

	//修改当前状态
	oldPrice := myStock.price
	vol := myStock.vol

	//不再持有股票
	delete(agent.currentMoney.myStocks, dayData.Code)

	totalSell := vol * dayData.Close //计算卖出后可得多少钱
	fee := totalSell * RateSell
	totalSell -= fee                          //扣除卖出手续费
	agent.currentMoney.freeMoney += totalSell //更新空闲资金
	agent.currentMoney.date = dayData.Date    //记录资金变化的时间

	agent.currentMoney.totalMoney -= oldPrice * vol //减去股票的钱(昨天的价位计算得到的)
	agent.currentMoney.totalMoney += totalSell      //加上卖出后得到的钱

	//保存状态
	agent.saveStatus(dayData.Date, opStready, OPT_SELL)
}

//继续持有股票
func (agent *MoneyAgent) hold(dayData *StockDailyData, opStready int) {
	myStock := agent.currentMoney.myStocks[dayData.Code]

	if myStock == nil {
		agent.currentMoney.myStocks[dayData.Code] = &MyStock{0, 0}
		myStock = agent.currentMoney.myStocks[dayData.Code]
	}
	vol := myStock.vol

	//判断是否持有股票
	if vol > 0 { //持有股票
		//记录股价的变化导致资金变化
		oldPrice := myStock.price
		myStock.price = dayData.Close                                //更新股价
		agent.currentMoney.totalMoney -= oldPrice * myStock.vol      //减去旧的股票资金
		agent.currentMoney.totalMoney += dayData.Close * myStock.vol //更新新的股票资金
	}
	agent.currentMoney.date = dayData.Date //更新时间

	//保存状态
	agent.saveStatus(dayData.Date, opStready, OPT_HOLD)

}

func (agent *MoneyAgent) saveStatus(date time.Time, opStready int, optFinal int) {
	//将当前资金变化状态记录到历史状态中
	agent.historyMoney = append(agent.historyMoney, agent.currentMoney)

	//将当天的交易记录追加到历史状态中
	newTran := TransRecord{Date: date, OptStrategy: opStready, OptFinal: optFinal}
	agent.historyTrans = append(agent.historyTrans, newTran)
}

func codeToStr(op int) string {
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

func (agent *MoneyAgent) PrintHistoryInfo() {
	for index, tran := range agent.historyTrans {
		opStrategy := codeToStr(tran.OptStrategy)
		opFinal := codeToStr(tran.OptFinal)

		mr := agent.historyMoney[index]

		fmt.Printf("date: %s, 策略: %s, 实际: %s ,"+
			"空闲资金为: %.2f, 总资产为: %.2f \n ", tran.Date, opStrategy, opFinal, mr.freeMoney, mr.totalMoney)
	}
}
