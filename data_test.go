package backTrace

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestGetAllSockCode(t *testing.T) {
	testLogger := logrus.WithFields(logrus.Fields{
		"function": "TestGetAllSockCode()",
	})
	code := "000001"
	stock, err := GetSockData(code)
	if err != nil {
		panic(err)
	}
	testLogger.Debug("finish GetSockData() test...")
	err = SaveLocalData(code, stock)
	if err != nil {
		panic(err)
	}
	testLogger.Debug("finish SaveLocalData() test...")
	stock2, err := LoadLocalData(code)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, stock, stock2)
	testLogger.Debug("finish LoadLocalData() test...")
	for _, c := range GetAllSockCode() {
		assert.Equal(t, 6, len(c))
	}
}

func TestSaveRewardRecord(t *testing.T) {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	var records []*RewardRecord
	for i := 0; i < 4; i++ {
		buy := fmt.Sprintf("Testbuy%d", random.Intn(1000))
		sell := fmt.Sprintf("Testbuy%d", random.Intn(1000))
		record := RewardRecord{
			Code:              "000001",
			SellStrategy:      sell,
			BuyStrategy:       buy,
			TotalReturnRate:   float32(2.2),
			ReturnRatePerYear: float32(2.2),
			WinningProb:       float32(2.2),
			ProfitExpect:      float32(2.2),
			LossExpect:        float32(2.2),
			AlphaEarnings:     float32(2.2),
			BetaEarnings:      float32(2.2),
		}
		records = append(records, &record)
	}
	res, err := SaveRewardRecord(records[0])
	if err != nil {
		panic(err)
	}
	logrus.Info(res)
	reses, err := SaveBatchRewardRecord(records[1:], 3)
	if err != nil {
		panic(err)
	}
	logrus.Info(reses)
}
