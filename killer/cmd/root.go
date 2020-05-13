package cmd

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var Db *sqlx.DB

var Log = logrus.New()
var RootCmd = &cobra.Command{
	Use:   "killer",
	Short: "killer root",
	Run: func(c *cobra.Command, args []string) {

	},
}

func Exec() {
	if viper.GetString("console") == "file" {
		file, err := os.OpenFile("killer.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
		if err != nil {
			panic("不能写日志文件")
		}
		Log.Out = file
	} else {
		Log.Info("Using default stderr")
	}
	Log.Infoln("log init")
	RootCmd.AddCommand(
		versionCmd,
		translate,
	)
	dsn := fmt.Sprintf("%s:%s@%s(%s:%d)/%s", viper.GetString("mysql.username"), viper.GetString("mysql.password"), viper.GetString("mysql.network"), viper.GetString("mysql.server"), viper.GetInt("mysql.port"), viper.GetString("mysql.database"))
	DB, err := sqlx.Open("mysql", dsn)
	if err != nil {
		Log.Printf("Open mysql failed,err:%v\n", err)
		return
	}
	Db = DB
	RootCmd.Execute()
}
func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	ex, osErr := os.Executable()
	if osErr != nil {
		panic(fmt.Errorf("os err: %s \n", osErr))
	}
	viper.AddConfigPath(filepath.Dir(ex))
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Read config file error: %s \n", err))
	}
}
