package sql

import (
	"github.com/aliyun/aliyun-tablestore-go-sdk/tablestore"
	"github.com/spf13/viper"
)

var DB *tablestore.TableStoreClient

func InitSql() {
	DB = tablestore.NewClient(viper.GetString("ali-table-store.endPoint"), viper.GetString("ali-table-store.instanceName"), viper.GetString("ali-table-store.accessKeyId"), viper.GetString("ali-table-store.accessKeySecret"))
}

//初始化``TableStoreClient``实例。
//endPoint是表格存储服务的地址（例如'https://instance.cn-hangzhou.ots.aliyun.com:80'），必须以'https://'开头。
//accessKeyId是访问表格存储服务的AccessKeyID，通过官方网站申请或通过管理员获取。
//accessKeySecret是访问表格存储服务的AccessKeySecret，通过官方网站申请或通过管理员获取。
//instanceName是要访问的实例名，通过官方网站控制台创建或通过管理员获取。
