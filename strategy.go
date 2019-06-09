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
	var err error
	length := len(data)
	var result = make([]int, length)
	var bs = make([]bool, length) // 是否买入
	var ss = make([]bool, length) // 是否卖出
	n := 0
	var preStrategy = make([]bool, length) //记录值，主要用于做多策略计算的
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
	preStrategy = make([]bool, length)
	for _, strag := range ana.SellPolicies {
		ss, err = strag.Do(data)
		if err != nil {
			return nil, err
		}
		if n == 0 {
			copy(preStrategy, ss)
		} else {
			// 或策略
			ss, err = SliceOrOpter(preStrategy, ss)
			if err != nil {
				return nil, err
			}
			copy(preStrategy, ss)
		}
	}
	if len(ss) != len(bs) {
		return nil, errors.New("buy policy and sell policy length not equal!")
	}
	var r int // 决定最后是买入还是卖出
	for i, s := range bs {
		if bs[i] == ss[i] {
			r = OPT_HOLD
		} else if s == true {
			r = OPT_BUY
		} else {
			r = OPT_SELL
		}
		result[i] = r
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

	for i, data := range s {
		closeArray[i] = data.Close
	}
	for i, c := range closeArray {
		if i >= N {
			ma[i] = Mean(closeArray[i-N : i])
			if c > ma[i] {
				result[i] = true
			}
		}
	}
	return result, nil
}

type BreakOutStrategySell struct{}

// 策略初加工所有股票数据
func (bos *BreakOutStrategySell) Process(slist []*Stock) []*Stock {
	return slist
}

// 根据特征字段判断是否卖出
func (bos *BreakOutStrategySell) Do(s Stock) ([]bool, error) {
	length := len(s)
	N := 60
	if length < N {
		err := errors.New("stock data is too short and cann't use this strategy!")
		return nil, err
	}

	var closeArray = make([]float32, length)
	var result = make([]bool, length)
	var ma = make([]float32, length)

	for i, data := range s {
		closeArray[i] = data.Close
	}
	for i, c := range closeArray {
		if i >= N {
			ma[i] = Mean(closeArray[i-N : i])
			if c < ma[i] {
				result[i] = true
			}
		}
	}
	return result, nil
}
