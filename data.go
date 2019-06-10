package backTrace

import (
	"database/sql"
	"encoding/gob"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

type StockColumnData struct {
	Length           int
	Date             []time.Time
	Code             []string
	Open             []float32
	Close            []float32
	High             []float32
	Low              []float32
	PreClose         []float32
	Change           []sql.NullFloat64
	PChange          []sql.NullFloat64
	PChangeF         []sql.NullFloat64
	PChangeB         []sql.NullFloat64
	OpenF            []sql.NullFloat64
	CloseF           []sql.NullFloat64
	HighF            []sql.NullFloat64
	LowF             []sql.NullFloat64
	PreCloseF        []sql.NullFloat64
	Volume           []sql.NullFloat64
	Amount           []sql.NullFloat64
	Turnover         []sql.NullFloat64
	TurnoverF        []sql.NullFloat64
	VolumeRatio      []sql.NullFloat64
	PE               []sql.NullFloat64
	PETTM            []sql.NullFloat64
	PB               []sql.NullFloat64
	PS               []sql.NullFloat64
	PSTTM            []sql.NullFloat64
	TotalShare       []sql.NullFloat64
	FloatShare       []sql.NullFloat64
	FreeShare        []sql.NullFloat64
	TotalMarketValue []sql.NullFloat64
	FloatMarketValue []sql.NullFloat64
}

func ConvColumnData(stock Stock) *StockColumnData {
	var length = len(stock)
	var data = &StockColumnData{
		Length:           length,
		Date:             make([]time.Time, length),
		Code:             make([]string, length),
		Open:             make([]float32, length),
		Close:            make([]float32, length),
		High:             make([]float32, length),
		Low:              make([]float32, length),
		PreClose:         make([]float32, length),
		Change:           make([]sql.NullFloat64, length),
		PChange:          make([]sql.NullFloat64, length),
		PChangeF:         make([]sql.NullFloat64, length),
		PChangeB:         make([]sql.NullFloat64, length),
		OpenF:            make([]sql.NullFloat64, length),
		CloseF:           make([]sql.NullFloat64, length),
		HighF:            make([]sql.NullFloat64, length),
		LowF:             make([]sql.NullFloat64, length),
		PreCloseF:        make([]sql.NullFloat64, length),
		Volume:           make([]sql.NullFloat64, length),
		Amount:           make([]sql.NullFloat64, length),
		Turnover:         make([]sql.NullFloat64, length),
		TurnoverF:        make([]sql.NullFloat64, length),
		VolumeRatio:      make([]sql.NullFloat64, length),
		PE:               make([]sql.NullFloat64, length),
		PETTM:            make([]sql.NullFloat64, length),
		PB:               make([]sql.NullFloat64, length),
		PS:               make([]sql.NullFloat64, length),
		PSTTM:            make([]sql.NullFloat64, length),
		TotalShare:       make([]sql.NullFloat64, length),
		FloatShare:       make([]sql.NullFloat64, length),
		FreeShare:        make([]sql.NullFloat64, length),
		TotalMarketValue: make([]sql.NullFloat64, length),
		FloatMarketValue: make([]sql.NullFloat64, length),
	}
	for i, s := range stock {
		data.Date[i] = s.Date
		data.Code[i] = s.Code
		data.Open[i] = s.Open
		data.Close[i] = s.Close
		data.High[i] = s.High
		data.Low[i] = s.Low
		data.PreClose[i] = s.PreClose
		data.Change[i] = s.Change
		data.PChange[i] = s.PChange
		data.PChangeF[i] = s.PChangeF
		data.PChangeB[i] = s.PChangeB
	}
	return data
}

func ConvRowData(s *StockColumnData) Stock {
	var stock []*StockDailyData
	var length = s.Length
	for i := 0; i < length; i++ {
		data := &StockDailyData{
			Date:     s.Date[i],
			Code:     s.Code[i],
			Open:     s.Open[i],
			Close:    s.Close[i],
			High:     s.High[i],
			Low:      s.Low[i],
			PreClose: s.PreClose[i],
			Change:   s.Change[i],
			PChange:  s.PChange[i],
			PChangeF: s.PChangeF[i],
			PChangeB: s.PChangeB[i],
		}
		stock = append(stock, data)
	}
	return stock
}

// get stock data
func GetSockData(code string) (Stock, error) {
	var rowData Stock
	contextLogger := logrus.WithFields(logrus.Fields{
		"function": "GetStock()",
		"code":     code,
	})
	columnData, err := LoadLocalData(code)
	contextLogger.Info(" the stock number is ", columnData.Length)
	if os.IsNotExist(err) {
		contextLogger.Info("star query database!")
		sqlstm := fmt.Sprintf("select * from stock_daily_data where code=%s", code)
		contextLogger.Info(sqlstm)
		err = DB.Select(&rowData, sqlstm)
		if err != nil {
			contextLogger.Warn(err)
			return rowData, nil
		}
		columnData = *ConvColumnData(rowData)
		err = SaveLocalData(code, columnData)
		if err != nil {
			contextLogger.Warn(err)
			return rowData, nil
		}
		return rowData, nil
	}
	if err != nil {
		contextLogger.Fatal(err)
		return rowData, err
	}
	rowData = ConvRowData(&columnData)

	return rowData, nil
}

func SaveLocalData(code string, stock StockColumnData) error {
	filename := fmt.Sprintf("./cache/%s.bin", code)
	file, err := os.Create(filename)
	defer file.Close()
	if err != nil {
		return err
	}
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(stock)
	if err != nil {
		return err
	}
	return nil
}

func LoadLocalData(code string) (StockColumnData, error) {
	var stock StockColumnData
	filename := fmt.Sprintf("./cache/%s.bin", code)
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		return stock, err
	}
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&stock)
	if err != nil {
		return stock, err
	}
	return stock, err
}

func GetAllSockCode() []string {
	var codes []string
	contextLogger := logrus.WithFields(logrus.Fields{
		"function": "GetAllSockCode()",
	})
	contextLogger.Info("star!")
	err = DB.Select(&codes, "select code from stock_daily_data group by code")
	if err != nil {
		logrus.Warn(err)
	}
	return codes
}
