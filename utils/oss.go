package utils

import (
	"ai-dataset-tool/log"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/spf13/viper"
)

var Bucket *oss.Bucket

func InitOss() {
	client, err := oss.New(viper.GetString("alibucket.endpoint"), viper.GetString("alibucket.accessKeyId"), viper.GetString("alibucket.accessKeySecret"))
	if err != nil {
		log.Klog.Println(err)
	}
	var berr error
	Bucket, berr = client.Bucket(viper.GetString("alibucket.bucketName"))
	if berr != nil {
		log.Klog.Println(berr)
	}
}
