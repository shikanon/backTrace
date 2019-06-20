package backTrace

func Mean(value []float32) float32 {
	var sumValue float32
	for _, v := range value {
		sumValue += v
	}
	return sumValue / float32(len(value))
}
