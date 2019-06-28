package backTrace

import (
	"database/sql"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
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
		data.OpenF[i] = s.OpenF
		data.CloseF[i] = s.CloseF
		data.HighF[i] = s.HighF
		data.LowF[i] = s.LowF
		data.PreCloseF[i] = s.PreCloseF
		data.Volume[i] = s.Volume
		data.Amount[i] = s.Amount
		data.Turnover[i] = s.Turnover
		data.TurnoverF[i] = s.TurnoverF
		data.VolumeRatio[i] = s.VolumeRatio
		data.PE[i] = s.PE
		data.PETTM[i] = s.PETTM
		data.TotalShare[i] = s.TotalShare
		data.FloatShare[i] = s.FloatShare
		data.FreeShare[i] = s.FreeShare
		data.TotalMarketValue[i] = s.TotalMarketValue
		data.FloatMarketValue[i] = s.FloatMarketValue
	}
	return data
}

func ConvRowData(s *StockColumnData) Stock {
	var stock []*StockDailyData
	var length = s.Length
	for i := 0; i < length; i++ {
		data := &StockDailyData{
			Date:             s.Date[i],
			Code:             s.Code[i],
			Open:             s.Open[i],
			Close:            s.Close[i],
			High:             s.High[i],
			Low:              s.Low[i],
			PreClose:         s.PreClose[i],
			Change:           s.Change[i],
			PChange:          s.PChange[i],
			PChangeF:         s.PChangeF[i],
			PChangeB:         s.PChangeB[i],
			OpenF:            s.OpenF[i],
			CloseF:           s.CloseF[i],
			HighF:            s.HighF[i],
			LowF:             s.LowF[i],
			PreCloseF:        s.PreCloseF[i],
			Volume:           s.Volume[i],
			Amount:           s.Amount[i],
			Turnover:         s.Turnover[i],
			TurnoverF:        s.TurnoverF[i],
			VolumeRatio:      s.VolumeRatio[i],
			PE:               s.PE[i],
			PETTM:            s.PETTM[i],
			PB:               s.PB[i],
			PS:               s.PS[i],
			PSTTM:            s.PSTTM[i],
			TotalShare:       s.TotalShare[i],
			FloatShare:       s.FloatShare[i],
			FreeShare:        s.FreeShare[i],
			TotalMarketValue: s.TotalMarketValue[i],
			FloatMarketValue: s.FloatMarketValue[i],
		}
		stock = append(stock, data)
	}
	return stock
}

// get stock data
func GetSockData(code string) (StockColumnData, error) {
	var rowData Stock
	contextLogger := logrus.WithFields(logrus.Fields{
		"function": "GetStock()",
		"code":     code,
	})
	contextLogger.Infof("Begin to load stock code: %s", code)
	columnData, err := LoadLocalData(code)
	contextLogger.Debug(" the stock number is ", columnData.Length)
	if os.IsNotExist(err) {
		contextLogger.Info("star query database!", code)
		sqlstm := fmt.Sprintf("select * from stock_daily_data where code=%s", code)
		contextLogger.Info(sqlstm)
		err = DB.Select(&rowData, sqlstm)
		if err != nil {
			contextLogger.Warn(err)
			return columnData, err
		}
		columnData = *ConvColumnData(rowData)
		err = SaveLocalData(code, columnData)
		if err != nil {
			contextLogger.Warn(err)
			return columnData, err
		}
		return columnData, err
	}
	if err != nil {
		contextLogger.Warnf("other error: %v", err)
		return columnData, err
	}
	return columnData, err
}

// get stock data for row stock
func GetRowSockData(code string) (Stock, error) {
	var rowData Stock
	contextLogger := logrus.WithFields(logrus.Fields{
		"function": "GetStock()",
		"code":     code,
	})
	columnData, err := LoadLocalData(code)
	contextLogger.Info(" the stock number is ", columnData.Length)
	if os.IsNotExist(err) {
		contextLogger.Info("star query database!", code)
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
	var saveObj string
	contextLogger := logrus.WithFields(logrus.Fields{
		"function": "GetAllSockCode()",
	})
	filename := "./cache/all_stock_code_array.bin"
	loadfile, err := os.Open(filename)
	defer loadfile.Close()
	if os.IsNotExist(err) {
		contextLogger.Info("star to database select data!")
		// 查询
		err = DB.Select(&codes, "select code from stock_daily_data group by code")
		if err != nil {
			logrus.Error(err)
		}
		// 写缓存文件
		for _, c := range codes {
			if saveObj == "" {
				saveObj = c
			}
			saveObj = saveObj + ";" + c
		}
		savefile, err := os.Create(filename)
		defer savefile.Close()
		if err != nil {
			logrus.Warn(err)
		}
		_, err = savefile.WriteString(saveObj)
		if err != nil {
			logrus.Error(err)
		}
		return codes
	} else if err != nil {
		logrus.Errorf("can't find cache filename path:%v", err)
	}
	// 读缓存文件
	if contents, err := ioutil.ReadAll(loadfile); err == nil {
		codes = strings.Split(string(contents), ";")
	} else {
		logrus.Error(err)
	}
	return codes
}

func SaveRewardRecord(record *RewardRecord) (res sql.Result, err error) {
	InsertSQL := `INSERT INTO reward_record (code, SellStrategy, BuyStrategy, TotalReturnRate, ReturnRatePerYear,
		WinningProb, ProfitExpect, LossExpect, AlphaEarnings, BetaEarnings)
		values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	res, err = DB.Exec(InsertSQL, record.Code, record.SellStrategy, record.BuyStrategy, record.TotalReturnRate, record.ReturnRatePerYear,
		record.WinningProb, record.ProfitExpect, record.LossExpect, record.AlphaEarnings, record.BetaEarnings)
	return res, err
}

func SaveBatchRewardRecord(records []*RewardRecord, minbatch int) ([]*RewardRecord, error) {
	var err error
	if len(records) < minbatch {
		return records, err
	}
	InsertSQL := `INSERT INTO reward_record (code, SellStrategy, BuyStrategy, TotalReturnRate, ReturnRatePerYear,
		WinningProb, ProfitExpect, LossExpect, AlphaEarnings, BetaEarnings)
		values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	tx := DB.MustBegin()
	for _, record := range records {
		_, err = tx.Exec(InsertSQL, record.Code, record.SellStrategy, record.BuyStrategy, record.TotalReturnRate, record.ReturnRatePerYear,
			record.WinningProb, record.ProfitExpect, record.LossExpect, record.AlphaEarnings, record.BetaEarnings)
	}
	tx.Commit()
	return nil, err
}
