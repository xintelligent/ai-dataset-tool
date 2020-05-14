package sql

import (
	"ai-dataset-tool/log"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
)

var Db *sqlx.DB

func InitSql() {
	dsn := fmt.Sprintf("%s:%s@%s(%s:%d)/%s", viper.GetString("mysql.username"), viper.GetString("mysql.password"), viper.GetString("mysql.network"), viper.GetString("mysql.server"), viper.GetInt("mysql.port"), viper.GetString("mysql.database"))
	DB, err := sqlx.Open("mysql", dsn)
	if err != nil {
		log.Klog.Printf("Open mysql failed,err:%v\n", err)
		return
	}
	Db = DB
}
