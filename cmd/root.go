package cmd

import (
	"ai-dataset-tool/log"
	"ai-dataset-tool/sql"
	"ai-dataset-tool/transform"
	"ai-dataset-tool/utils"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

/**
就准备只适配我们自己的数据源
*/
var RootCmd = &cobra.Command{
	Use:   "killer",
	Short: "killer root",
	Run: func(c *cobra.Command, args []string) {

	},
}
var needDownloadImageFile bool
var needSplitImage bool

func Exec() {
	// 准备日志服务
	log.InitLog()
	// 准备数据库连接
	sql.InitSql()
	// 准本数据源
	var classErr error
	sql.Classes, classErr = sql.NewClassesFromMysql()
	if classErr != nil {
		fmt.Println("标签类别查询失败")
		os.Exit(1)
	}
	var labelErr error
	sql.Labels, labelErr = sql.NewLabelsFromMysql()
	if labelErr != nil {
		fmt.Println("查询标签数据失败")
		os.Exit(1)
	}
	// 下载oss 图片文件到本地
	preDownload()
	RootCmd.Execute()
}
func init() {
	RootCmd.PersistentFlags().BoolVarP(&needDownloadImageFile, "needDownloadImageFile", "d", false, "need download imageTool file")
	RootCmd.PersistentFlags().BoolVarP(&needSplitImage, "needDownloadImageFile", "s", false, "need download imageTool file")
	RootCmd.PersistentFlags().StringVarP(&utils.AnnotationOutPath, "AnnotationOutPath", "a", "./data", "label file out path")
	RootCmd.PersistentFlags().StringVarP(&utils.ImageOutPath, "imageOutPath", "i", "./images", "images file out path")
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

func preDownload() {
	log.Klog.Printf("本次总图片数量：%d", len(sql.Labels))
	var wg sync.WaitGroup
	images := make(map[string]bool)
	// 验证每一条数据
	for _, v := range sql.Labels {
		df := utils.TransformFile(v.Image_path)
		data, err := v.JsonToStruct()
		if err != nil || len(data.Label) == 0 {
			continue
		}
		if _, ok := images[df.Name]; ok {
			images[df.Name] = true
		} else {
			df.Name = utils.RandStringBytesMaskImprSrcUnsafe(16)
		}
		var labels []transform.Shape
		for _, v := range data.Label {
			labels = append(labels, transform.Shape{
				utils.ToFloat64(v.Xmax),
				utils.ToFloat64(v.Xmin),
				utils.ToFloat64(v.Ymax),
				utils.ToFloat64(v.Ymin),
				v.Category,
			})
		}
		transform.LabelsData.Push(transform.Lab{
			v.Image_path,
			utils.ImageOutPath,
			df.Name,
			df.Suffix,
			transform.Data{
				labels,
				data.ImageWidth,
				data.ImageHeight,
			},
		})
		if needDownloadImageFile || needSplitImage {
			utils.DownloadIns.Goroutine_cnt <- 1
			wg.Add(1)
			go utils.DownloadIns.DGoroutine(&wg, df)
		}
	}
	wg.Wait()
	if needSplitImage {

	}
}
