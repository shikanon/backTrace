package backTrace

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/astaxie/beego/config"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

var DB *sqlx.DB
var err error

type Stock []*StockDailyData

type StockDailyData struct {
	Date             time.Time       `db:"date"`
	Code             string          `db:"code"`
	Open             float32         `db:"open"`               // 开盘价/元                                                                            |
	Close            float32         `db:"close"`              // 收盘价/元                                                                          |
	High             float32         `db:"high"`               // 最高价/元                                                                            |
	Low              float32         `db:"low"`                // 最低价/元                                                                              |
	PreClose         float32         `db:"pre_close"`          // 昨日收价/元                                                                |
	Change           sql.NullFloat64 `db:"change"`             // 涨跌价格/元                                                                      |
	PChange          sql.NullFloat64 `db:"p_change"`           // 涨跌幅%(未复权)                                                              |
	PChangeF         sql.NullFloat64 `db:"p_change_f"`         // 前复权_涨跌幅%,当日收盘价 × 当日复权因子 / 最新复权因子                  |
	PChangeB         sql.NullFloat64 `db:"p_change_b"`         // 后复权_涨跌幅%,当日收盘价 × 当日复权因子                                 |
	OpenF            sql.NullFloat64 `db:"open_f"`             // 开盘价/元                                                                            |
	CloseF           sql.NullFloat64 `db:"close_f"`            // 收盘价/元                                                                          |
	HighF            sql.NullFloat64 `db:"high_f"`             // 最高价/元                                                                            |
	LowF             sql.NullFloat64 `db:"low_f"`              // 最低价/元                                                                              |
	PreCloseF        sql.NullFloat64 `db:"pre_close_f"`        // 昨日收价/元
	Volume           sql.NullFloat64 `db:"vol"`                // 成交量/手                                                                              |
	Amount           sql.NullFloat64 `db:"amount"`             // 成交额/千元                                                                      |
	Turnover         sql.NullFloat64 `db:"turnover"`           // 换手率%                                                                      |
	TurnoverF        sql.NullFloat64 `db:"turnover_f"`         // 自由流通股换手率(%), 成交量/发行总股数                                   |
	VolumeRatio      sql.NullFloat64 `db:"volume_ratio"`       // 量比,（现成交总手数 / 现累计开市时间(分) ）/ 过去5日平均每分钟成交量 |
	PE               sql.NullFloat64 `db:"pe"`                 // 市盈率, 总市值/过去一年净利润                                                            |
	PETTM            sql.NullFloat64 `db:"pe_ttm"`             // 滚动市盈率(TTM)                                                                  |
	PB               sql.NullFloat64 `db:"pb"`                 // 市净率, 总市值/净资产                                                                    |
	PS               sql.NullFloat64 `db:"ps"`                 // 市销率,  总市值/主营业务收入                                                             |
	PSTTM            sql.NullFloat64 `db:"ps_ttm"`             // 滚动市销率（TTM）                                                                |
	TotalShare       sql.NullFloat64 `db:"total_share"`        // 总股本（万股）                                                         |
	FloatShare       sql.NullFloat64 `db:"float_share"`        // 流通股本(万股)                                                         |
	FreeShare        sql.NullFloat64 `db:"free_share"`         // 自由流通股本(万股)                                                       |
	TotalMarketValue sql.NullFloat64 `db:"total_market_value"` // 总市值(万元)                                             |
	FloatMarketValue sql.NullFloat64 `db:"float_market_value"` // 流通市值(万元)
}

// select concat(UPPER(SUBSTRING(COLUMN_NAME,1,1)),
//     SUBSTRING(COLUMN_NAME,2,length(COLUMN_NAME)),
//     (case DATA_TYPE
//         when 'varchar' then ' string '
//         when 'int' then ' int '
//         when 'double' then ' sql.NullFloat64 '
//         when 'float' then ' sql.NullFloat64 '
//         when 'datetime' then ' string '
//         end),
//         '`db:"',
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

type DBinfo struct {
	USERNAME string
	PASSWORD string
	NETWORK  string
	SERVER   string
	PORT     int
	DATABASE string
}

func ReadConf(path string) *DBinfo {
	conf, err := config.NewConfig("ini", path)
	if err != nil {
		logrus.Fatal("Error:", err)
	}
	host := conf.String("db::host")
	username := conf.String("db::user")
	password := conf.String("db::password")
	database := conf.String("db::database")
	port, err := conf.Int("db::port")
	if err != nil {
		logrus.Fatal("Error:", err)
	}
	return &DBinfo{USERNAME: username,
		PASSWORD: password,
		SERVER:   host,
		PORT:     port,
		DATABASE: database,
		NETWORK:  "tcp",
	}
}

func init() {
	dbinfo := ReadConf("backTrace/config.conf")
	dataSource := fmt.Sprintf("%s:%s@%s(%s:%d)/%s?parseTime=true&loc=Local&sql_mode=''", dbinfo.USERNAME, dbinfo.PASSWORD, dbinfo.NETWORK, dbinfo.SERVER, dbinfo.PORT, dbinfo.DATABASE)
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
