package backTrace

imoprt(
	"errors"
)

//买入
const OPT_BUY = 0

//卖
const OPT_SELL = 1

//什么也不做
const OPT_HOLD = 2


func SliceOrOpter(fristArray []float32, secArray []float32) ([]float32,error){
	var result []float32
	if len(fristArray) != len(secArray){
		return nil, error.Error("When you use SliceOrOpter, the length of array must equal")
	}
	for i, v := range fristArray {
		result[i] = fristArray[i] || secArray
	}
	return result
}


type Analyzer struct {
	BuyPolicies  []Strategy
	SellPolicies []Strategy
}

func (ana *Analyzer) Analyse(data Stock) ([]int, error) {
	var result []int
	var err error
	var preStrategy bool //记录值，主要用于做多策略计算的
	var bs bool          // 是否买入
	var ss bool          // 是否卖出
	n := 0
	for _, strag := range ana.BuyPolicies {
		bs,err = strag.Do(data)
		if err != nil{
			return result, err
		}
		if n == 0 {
			preStrategy = bs
		} else {
			// 或策略
			bs = SliceOrOpter(preStrategy, bs)
			preStrategy = bs
		}
		n += 1
	}

	n = 0
	for _, strag := range ana.SellPolicies {
		ss = strag.Do(d)
		if n == 0 {
			preStrategy = ss
		} else {
			// 或策略
			ss = SliceOrOpter(preStrategy, ss)
			preStrategy = ss
		}
	}
	var r int // 决定最后是买入还是卖出
	if bs == ss {
		r = OPT_HOLD
	} else if bs == true {
		r = OPT_BUY
	} else {
		r = OPT_SELL
	}
	result = append(result, r)

	return result
}

type Strategy interface {
	Do(Stock)  ([]bool,error)
}

type BreakOutStrategyBuy struct{}

func Mean(value []float32) float32 {
	var sumValue float32
	for i, v := range value {
		sumValue += v
	}
	return sumValue / float32(len(value))
}

// 策略初加工所有股票数据
func (bos *BreakOutStrategyBuy) Process(slist []*Stock) []*Stock {
	return slist
}

// 根据特征字段判断是否买入
func (bos *BreakOutStrategyBuy) Do(s Stock) ([]bool,error) {
	length := len(s)
	N := 60
	if length < N {
		err := error.Error("stock data is too short and cann't use this strategy!")
		return nil,err
	}
	var result [length]bool
	var ma [length]float32
	for i, data := range s {
		if i >= N {
			ma[i] = Mean(data.Close[i-n : i])
			if data.Close > ma {
				result[i] = 1
			}
		}
	}
	return result, nil
}

type MACDStrategySell struct{}

// 策略初加工所有股票数据
func (bos *MACDStrategySell) Process(slist []*Stock) []*Stock {
	return slist
}

// 根据特征字段判断是否卖出
func (macd *MACDStrategySell) Do(s *StockDailyData) bool {
	return 0
}
