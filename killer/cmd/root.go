package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var Log = logrus.New()
var RootCmd = &cobra.Command{
	Use:   "killer",
	Short: "killer root",
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
