package log

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
)

var Klog *logrus.Logger

type Field = logrus.Fields

func InitLog() {
	Klog = logrus.New()
	if viper.GetString("console") == "file" {
		file, err := os.OpenFile("tool.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
		if err != nil {
			panic("不能写日志文件")
		}
		Klog.Out = file
	} else {
		Klog.Info("Using default stderr")
	}
}
