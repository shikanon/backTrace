package backTrace

func Mean(value []float32) float32 {
	var sumValue float32
	for _, v := range value {
		sumValue += v
	}
	return sumValue / float32(len(value))
}

func Max(value []float32) float32 {
	var maxnum float32
	for i, v := range value {
		if i == 0 {
			maxnum = v
		} else {
			if maxnum < v {
				maxnum = v
			}
		}
	}
	return maxnum
}

func Min(value []float32) float32 {
	var minnum float32
	for i, v := range value {
		if i == 0 {
			minnum = v
		} else {
			if minnum > v {
				minnum = v
			}
		}
	}
	return minnum
}
