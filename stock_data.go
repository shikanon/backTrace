package backTrace

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

type StockData struct {
	Date string `db:"date"`
	// stock code
	Code string `db:"code"`
	// stock open price
	Open float32 `db:"open"`
	// stock close price
	Close float32 `db:"close"`
	// stock high price
	High float32 `db:"high"`
	// stock low price
	Low float32 `db:"low"`
	// stock close price of last workday
	PreClose float32 `db:"pre_close"`
	// stock price change of absolute
	Change float32 `db:"change"`
	// stock change of percent
	ChangeRatio sql.NullFloat64 `db:"p_change"`
	// stock change of percent in forward answer authority
	ChangeForwardRatio float32 `db:"p_change_f"`
	// stock change of percent in backward answer authority
	ChangeBackwardRatio sql.NullFloat64 `db:"p_change_b"`
	// stock of volumes
	Volume float32 `db:"vol"`
	// volumes of all amount
	Amount float32 `db:"amount"`
	// stock of turnover rate
	Turnover float32 `db:"turnover"`
}

// date code open close high low pre_close change p_change p_change_f p_change_b vol amount turnover
func GetStock(code string) []*StockData {
	var stockDailyData []*StockData
	contextLogger := logrus.WithFields(logrus.Fields{
		"function": "GetStock()",
		"code":     code,
	})
	contextLogger.Info("star!")
	sqlstm := fmt.Sprintf("select date,code,open,close,high,low,pre_close,`change`,p_change,p_change_f,p_change_b,vol,amount,turnover from daily_data where code=%s limit 5", code)
	contextLogger.Info(sqlstm)
	err = DB.Select(&stockDailyData, sqlstm)
	if err != nil {
		logrus.Warn(err)
	}
	return stockDailyData
}
