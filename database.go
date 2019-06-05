package backTrace

import (
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

var DB *sqlx.DB
var err error

const (
	USERNAME = "root"
	PASSWORD = "asdfQWER!@#$"
	NETWORK  = "tcp"
	SERVER   = "gz-cdb-73f4xwbv.sql.tencentcdb.com"
	PORT     = 61347
	DATABASE = "stock"
)

type Stock []*StockDailyData

type StockDailyData struct {
	Date             time.Time `db:"date"`
	Code             string    `db:"code"`
	Open             float64   `json:"open"`               // 开盘价/元                                                                            |
	Close            float64   `json:"close"`              // 收盘价/元                                                                          |
	High             float64   `json:"high"`               // 最高价/元                                                                            |
	Low              float64   `json:"low"`                // 最低价/元                                                                              |
	PreClose         float64   `json:"pre_close"`          // 昨日收价/元                                                                |
	Change           float64   `json:"change"`             // 涨跌价格/元                                                                      |
	PChange          float64   `json:"p_change"`           // 涨跌幅%(未复权)                                                              |
	PChangeF         float64   `json:"p_change_f"`         // 前复权_涨跌幅%,当日收盘价 × 当日复权因子 / 最新复权因子                  |
	PChangeB         float64   `json:"p_change_b"`         // 后复权_涨跌幅%,当日收盘价 × 当日复权因子                                 |
	Volume           float64   `json:"vol"`                // 成交量/手                                                                              |
	Amount           float64   `json:"amount"`             // 成交额/千元                                                                      |
	Turnover         float64   `json:"turnover"`           // 换手率%                                                                      |
	TurnoverF        float64   `json:"turnover_f"`         // 自由流通股换手率(%), 成交量/发行总股数                                   |
	VolumeRatio      float64   `json:"volume_ratio"`       // 量比,（现成交总手数 / 现累计开市时间(分) ）/ 过去5日平均每分钟成交量 |
	PE               float64   `json:"pe"`                 // 市盈率, 总市值/过去一年净利润                                                            |
	PETTM            float64   `json:"pe_ttm"`             // 滚动市盈率(TTM)                                                                  |
	PB               float64   `json:"pb"`                 // 市净率, 总市值/净资产                                                                    |
	PS               float64   `json:"ps"`                 // 市销率,  总市值/主营业务收入                                                             |
	PSTTM            float64   `json:"ps_ttm"`             // 滚动市销率（TTM）                                                                |
	TotalShare       float64   `json:"total_share"`        // 总股本（万股）                                                         |
	FloatShare       float64   `json:"float_share"`        // 流通股本(万股)                                                         |
	FreeShare        float64   `json:"free_share"`         // 自由流通股本(万股)                                                       |
	TotalMarketValue float64   `json:"total_market_value"` // 总市值(万元)                                             |
	FloatMarketValue float64   `json:"float_market_value"` // 流通市值(万元)
}

// select concat(UPPER(SUBSTRING(COLUMN_NAME,1,1)),
//     SUBSTRING(COLUMN_NAME,2,length(COLUMN_NAME)),
//     (case DATA_TYPE
//         when 'varchar' then ' string '
//         when 'int' then ' int '
//         when 'double' then ' float32 '
//         when 'float' then ' float64 '
//         when 'datetime' then ' string '
//         end),
//         '`json:"',
//         COLUMN_NAME,
//         '"`',
//         '// ',
//         COLUMN_COMMENT
// ) as golang_variable
// from (
//     select
//     COLUMN_NAME,
//     DATA_TYPE,
//     COLUMN_COMMENT from information_schema.COLUMNS
//     where table_name = 'stock_daily_data' -- 填写所需要的表名
// ) a;

func init() {
	dataSource := fmt.Sprintf("%s:%s@%s(%s:%d)/%s?parseTime=true&loc=Local&sql_mode=''", USERNAME, PASSWORD, NETWORK, SERVER, PORT, DATABASE)
	fmt.Println(dataSource)
	DB, err = sqlx.Open("mysql", dataSource)
	if err != nil {
		logrus.Error(err)
	}
	err = DB.Ping()
	if err != nil {
		logrus.Error(err)
	}
}
