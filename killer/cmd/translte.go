package cmd

import (
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var Db *sqlx.DB

var Format string
var AnnotationOutPath string
var NeedDownloadImageFile bool
var ImageOutPath string
var DownloadPoolIns *downloadPool
var Bucket *oss.Bucket
var translate = &cobra.Command{
	Use: "translate",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			Log.Println("缺少项目id")
			os.Exit(1)
		}
		os.MkdirAll(AnnotationOutPath, 0777)
		os.MkdirAll(ImageOutPath, 0777)
		if NeedDownloadImageFile {
			DownloadPoolIns = NewDownloadPool()
			defer close(DownloadPoolIns.goroutine_cnt)
		}
		client, err := oss.New(viper.GetString("alibucket.endpoint"), viper.GetString("alibucket.accessKeyId"), viper.GetString("alibucket.accessKeySecret"))
		if err != nil {
			Log.Println(err)
		}
		var berr error
		Bucket, berr = client.Bucket(viper.GetString("alibucket.bucketName"))
		if berr != nil {
			Log.Println(berr)
		}
		baseOnFormat(args[0])
	},
}

func init() {
	translate.Flags().StringVarP(&Format, "Format", "f", "csv", "Format(csv or coco)")
	translate.Flags().StringVarP(&AnnotationOutPath, "AnnotationOutPath", "a", "./data", "label file out path")
	translate.Flags().StringVarP(&ImageOutPath, "imageOutPath", "i", "./images", "images file out path")
	translate.Flags().BoolVarP(&NeedDownloadImageFile, "needDownloadImageFile", "n", false, "need download image file")
	dsn := fmt.Sprintf("%s:%s@%s(%s:%d)/%s", viper.GetString("mysql.username"), viper.GetString("mysql.password"), viper.GetString("mysql.network"), viper.GetString("mysql.server"), viper.GetInt("mysql.port"), viper.GetString("mysql.database"))
	DB, err := sqlx.Open("mysql", dsn)
	if err != nil {
		Log.Printf("Open mysql failed,err:%v\n", err)
		return
	}
	Db = DB
}
func baseOnFormat(pid string) {
	getClassesFromMysql(pid)
	getLabelsFromMysql(pid)
	switch Format {
	case "csv":
		writeCsvClassFile()
		writeCsvLabelsFile()
	case "coco":
		WriteCocoFile()
	case "voc":
		writeVocLabelsFile()
		writeVocLabelsFile()
	}
}
