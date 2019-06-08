package backTrace

import "errors"

func SliceOrOpter(fristArray []bool, secArray []bool) ([]bool, error) {
	if len(fristArray) != len(secArray) {
		return nil, errors.New("When you use SliceOrOpter, the length of array must equal")
	}
	var result = make([]bool, len(fristArray))
	for i, _ := range fristArray {
		result[i] = fristArray[i] || secArray[i]
	}
	return result, nil
}

type Analyzer struct {
	BuyPolicies  []Strategy
	SellPolicies []Strategy
}

func (ana *Analyzer) Analyse(data Stock) ([]int, error) {
	var result []int
	var err error
	var preStrategy []bool //记录值，主要用于做多策略计算的
	var bs []bool          // 是否买入
	var ss []bool          // 是否卖出
	n := 0
	for _, strag := range ana.BuyPolicies {
		bs, err = strag.Do(data)
		if err != nil {
			return nil, err
		}
		if n == 0 {
			copy(preStrategy, bs)
		} else {
			// 或策略
			bs, err = SliceOrOpter(preStrategy, bs)
			if err != nil {
				return nil, err
			}
			copy(preStrategy, bs)
		}
		n += 1
	}

	n = 0
	for _, strag := range ana.SellPolicies {
		ss, err = strag.Do(data)
		if err != nil {
			return nil, err
		}
		if n == 0 {
			preStrategy = ss
		} else {
			// 或策略
			ss, err = SliceOrOpter(preStrategy, ss)
			if err != nil {
				return nil, err
			}
			preStrategy = ss
		}
	}
	var r int // 决定最后是买入还是卖出
	for s := range bs {
		if bs[s] == ss[s] {
			r = OPT_HOLD
		} else if bs[s] == true {
			r = OPT_BUY
		} else {
			r = OPT_SELL
		}
		result = append(result, r)
	}

	return result, nil
}

type Strategy interface {
	Do(Stock) ([]bool, error)
}

type BreakOutStrategyBuy struct{}

func Mean(value []float32) float32 {
	var sumValue float32
	for _, v := range value {
		sumValue += v
	}
	return sumValue / float32(len(value))
}

// 策略初加工所有股票数据
func (bos *BreakOutStrategyBuy) Process(slist []*Stock) []*Stock {
	return slist
}

// 根据特征字段判断是否买入
func (bos *BreakOutStrategyBuy) Do(s Stock) ([]bool, error) {
	length := len(s)
	N := 60
	if length < N {
		err := errors.New("stock data is too short and cann't use this strategy!")
		return nil, err
	}

	var closeArray = make([]float32, length)
	var result = make([]bool, length)
	var ma = make([]float32, length)

	for _, data := range s {
		closeArray = append(closeArray, data.Close)
	}
	for i, c := range closeArray {
		if i >= N {
			ma = append(ma, Mean(closeArray[i-N:i]))
			if c > ma[i] {
				result = append(result, true)
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
func (macd *MACDStrategySell) Do(s Stock) ([]bool, error) {
	var result = make([]bool, len(s))
	return result, nil
}
