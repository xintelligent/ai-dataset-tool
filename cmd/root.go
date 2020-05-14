package cmd

import (
	"ai-dataset-tool/log"
	"ai-dataset-tool/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var RootCmd = &cobra.Command{
	Use:   "killer",
	Short: "killer root",
	Run: func(c *cobra.Command, args []string) {

	},
}

func Exec() {
	// 准备日志服务
	log.InitLog()
	// 准备数据库连接
	sql.InitSql()
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
