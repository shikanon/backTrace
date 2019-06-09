package backTrace

import (
	"fmt"
	"time"
)

const OPT_BUY = 0 //买入

const OPT_SELL = 1 //卖

const OPT_HOLD = 2 //什么也不做

const RateBuy = 0.003  //买入手续费
const RateSell = 0.003 //卖出手续费

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

type IncomeRecord struct {
	buyDate    time.Time //买入日期
	buyVol     int       //买入股数
	buyPrice   float32   //买入价格
	sellDate   time.Time //卖出日期
	sellVol    int       //卖出股数
	sellPrice  float32   //卖出价格
	initMoney  float32   //起步资金
	finalMoney float32   //退场资金
}

type MoneyAgent struct {
	initMoney    float32
	currentMoney MoneyRecord    //该经理当前资金状况
	historyMoney []MoneyRecord  //该投资经理的资金变化记录
	historyTrans []TransRecord  //该投资经理的交易记录
	incomeRecord []IncomeRecord //每次完整交易的记录,一次买入一次卖出
	Analyzer
}

func (agent *MoneyAgent) init() {
	//经理第一天上班，资金状态需要初始化
	agent.currentMoney = MoneyRecord{totalMoney: agent.initMoney, freeMoney: agent.initMoney, myStocks: make(map[string]*MyStock)}
}

//资金经理开始干活了,他需要对这个股票的所有数据进行分析
func (agent *MoneyAgent) WorkForSingle(stocks Stock) {
	//分析数据，获得买入卖出操作指令
	points, err := agent.Analyzer.Analyse(stocks)
	if err != nil {
		fmt.Printf("出现异常，退出程序：%s", err)
		return
	}

	//TODO 取出昨天的股票数据
	/*for index,value:= range points{
		fmt.Printf("date: %s,code :%s, 策略返回结果是： %s\n",stocks[index].Date,stocks[index].Code,value)
	}*/

	lastDay := len(stocks) - 1

	for index, dayData := range stocks {

		var opStready int
		var yestoday *StockDailyData

		//最后一天,必需要全部卖出
		if index == lastDay {
			opStready = OPT_SELL
			yestoday = stocks[index-1]
		} else if index == 0 { //第一天默认不做任何操作
			yestoday = nil
			opStready = OPT_HOLD
		} else { //第一天与最后一天以外，策略按照分析者的建议执行
			yestoday = stocks[index-1]
			opStready = points[index]
		}

		switch opStready {
		case OPT_BUY:
			//修改交易记录，资金状态
			agent.buy(yestoday, dayData, opStready)
			//fmt.Printf("buy %s\n", dayData.Code)
		case OPT_SELL:
			//修改交易记录，资金状态
			agent.sell(yestoday, dayData, opStready)
			//fmt.Printf("sell %s\n", dayData.Code)
		case OPT_HOLD:
			//记录操作
			agent.hold(yestoday, dayData, opStready)
			//fmt.Printf("hold %s\n", dayData.Code)
		}
	}

}

//买入
func (agent *MoneyAgent) buy(yestoday *StockDailyData, today *StockDailyData, opStready int) {
	//判断是否涨停，如果涨停则无法交易，需要跳转到hold
	//如果yestoday = nil 则说明股票第一天上市，直接转到hold,然后结束
	if yestoday == nil {
		agent.hold(yestoday, today, opStready)
		return
	} else {
		//涨停，买入不了,跳转到hold，然后结束
		if (today.Close-yestoday.Close)/yestoday.Close > 0.0998 {
			agent.hold(yestoday, today, opStready)
			fmt.Printf("涨停!!!,无法买入, 时间：%s, 昨天股价: %.2f, 今天股价: %.2f\n", today.Date, yestoday.Close, today.Close)
			return
		}
	}

	maxPiece := agent.currentMoney.freeMoney / today.Close //粗略计算可买股数
	totalCost := maxPiece * today.Close                    //粗略估计总的费用
	fee := totalCost * RateBuy                             //计算交易费

	for true {
		//判断扣除手续费之后，算上手续费后，不足购买maxPiece 股，并且maxPiece>1,可以减少买入股数，以完成交易
		if totalCost > agent.currentMoney.freeMoney-fee {
			maxPiece -= 1
			totalCost = maxPiece * today.Close
			fee = totalCost * RateBuy
		} else {
			break //有足够的钱支付  购买超过1股以及 手续费
		}

		//不够钱交易1股，跳转到持有不变
		if maxPiece <= 0 {
			agent.hold(yestoday, today, opStready)
			return
		}

	}

	//先检查是否有持有当前股票
	myStock := agent.currentMoney.myStocks[today.Code]

	// 为空则表示之前没有持有该股票
	if myStock == nil {
		agent.currentMoney.myStocks[today.Code] = &MyStock{0, 0}
		myStock = agent.currentMoney.myStocks[today.Code]
	} else { //有值说明之前持有股票，股价的变化需要更新总资金数
		oldTotal := myStock.price * myStock.vol
		newTotal := myStock.vol * today.Close
		agent.currentMoney.totalMoney -= oldTotal
		agent.currentMoney.totalMoney += newTotal
	}

	agent.currentMoney.freeMoney -= totalCost //扣除买股的钱
	agent.currentMoney.freeMoney -= fee       //扣除手续费
	agent.currentMoney.date = today.Date

	agent.currentMoney.totalMoney -= fee //购买了股票，总资金的变化只有减去

	//记录股数的变化
	myStock.vol += maxPiece
	myStock.price = today.Close

	//保存状态
	agent.saveStatus(today.Date, opStready, OPT_BUY, yestoday, today)

}

//卖出股票
func (agent *MoneyAgent) sell(yestoday *StockDailyData, today *StockDailyData, opStready int) {

	//先检查是否有持有当前股票
	myStock := agent.currentMoney.myStocks[today.Code]

	//没有持有当前股票,什么也做不了  或者  第一天的交易命令，默认什么都不做
	if myStock == nil || yestoday == nil {
		agent.hold(yestoday, today, opStready)
		return
	} else if (today.Close-yestoday.Close)/yestoday.Close < -0.0998 { //跌停，无法交易
		agent.hold(yestoday, today, opStready)
		fmt.Printf("跌停!!!,无法卖出,%s,昨天股价: %.2f, 今天股价: %.2f\n", today.Date, yestoday.Close, today.Close)
		return
	}

	//修改当前状态
	oldPrice := myStock.price
	vol := myStock.vol

	//不再持有股票
	delete(agent.currentMoney.myStocks, today.Code)

	totalSell := vol * today.Close //计算卖出后可得多少钱
	fee := totalSell * RateSell
	totalSell -= fee                          //扣除卖出手续费
	agent.currentMoney.freeMoney += totalSell //更新空闲资金
	agent.currentMoney.date = today.Date      //记录资金变化的时间

	agent.currentMoney.totalMoney -= oldPrice * vol //减去股票的钱(昨天的价位计算得到的)
	agent.currentMoney.totalMoney += totalSell      //加上卖出后得到的钱

	//保存状态
	agent.saveStatus(today.Date, opStready, OPT_SELL, yestoday, today)
}

//继续持有股票
func (agent *MoneyAgent) hold(yestoday *StockDailyData, today *StockDailyData, opStready int) {
	myStock := agent.currentMoney.myStocks[today.Code]
	//只有已经持有股票才需要更新资金
	if myStock != nil {
		vol := myStock.vol
		//判断是否持有股票
		if vol > 0 { //持有股票
			//记录股价的变化导致资金变化
			oldPrice := myStock.price
			myStock.price = today.Close                                //更新股价
			agent.currentMoney.totalMoney -= oldPrice * myStock.vol    //减去旧的股票资金
			agent.currentMoney.totalMoney += today.Close * myStock.vol //更新新的股票资金
		}
	}
	agent.currentMoney.date = today.Date //更新时间

	//保存状态
	agent.saveStatus(today.Date, opStready, OPT_HOLD, yestoday, today)

}

func (agent *MoneyAgent) saveStatus(date time.Time, opStready int, optFinal int, yestoday *StockDailyData, today *StockDailyData) {

	//将当前资金变化状态记录到历史状态中
	agent.historyMoney = append(agent.historyMoney, agent.currentMoney)

	//将当天的交易记录追加到历史状态中
	newTran := TransRecord{Date: date, OptStrategy: opStready, OptFinal: optFinal}
	agent.historyTrans = append(agent.historyTrans, newTran)

	opStrategy := codeToStr(opStready)
	opFinal := codeToStr(optFinal)

	var todayPrice float32
	if yestoday == nil {
		todayPrice = 0 //上市第一天，昨天为0
	} else {
		todayPrice = yestoday.Close
	}

	yesterdayPrice := today.Close

	var vol float32
	myStock := agent.currentMoney.myStocks[today.Code]
	if myStock != nil {
		vol = myStock.vol
	}

	fmt.Printf("date: %s, 策略: %s, 实际: %s ,"+
		"空闲资金为: %.2f, 总资产为: %.2f ,昨天股价为： %.2f ,今天股价：%.2f\n ,持有股数： %.1f \n", newTran.Date, opStrategy,
		opFinal, agent.currentMoney.freeMoney, agent.currentMoney.totalMoney, yesterdayPrice, todayPrice, vol)

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
}

type ProfileData struct {
	InitCapital  float32
	FinalCapital float32
	HistoryMoney []*MoneyRecord
	Record       []*IncomeRecord
}

func (agent *MoneyAgent) GetProfileData() {

}
