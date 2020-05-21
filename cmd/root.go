package cmd

import (
	"ai-dataset-tool/log"
	"ai-dataset-tool/sql"
	"ai-dataset-tool/transform"
	"ai-dataset-tool/utils"
	"fmt"
	"math"
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
	Use: "ai-dataset-tool",
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
	sql.Db.Close()
	// 下载oss 图片文件到本地
	utils.InitDownload()
	defer close(utils.DownloadIns.Goroutine_cnt)
	utils.InitOss()
	os.MkdirAll(utils.AnnotationOutPath, 0777)
	os.MkdirAll(utils.ImageOutPath, 0777)
	os.MkdirAll("split", 0777)
	prepare()
	RootCmd.Execute()
}
func init() {
	RootCmd.PersistentFlags().BoolVarP(&needDownloadImageFile, "needDownloadImageFile", "d", false, "need download image file")
	RootCmd.PersistentFlags().BoolVarP(&needSplitImage, "needSplitImage", "s", false, "need split image file")
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

func prepare() {
	log.Klog.Printf("本次总图片数量：%d", len(sql.Labels))
	var wg sync.WaitGroup
	images := make(map[string]bool)
	// 验证每一条数据
	for _, v := range sql.Labels {
		// 拆分文件名
		df := utils.TransformFile(v.Image_path)
		// json 转 struct
		data, err := v.JsonToStruct()
		if err != nil || len(data.Label) == 0 {
			continue
		}
		// 验证图片名称重复（不带后缀）
		if _, ok := images[df.Name]; !ok {
			images[df.Name] = true
		} else {
			df.Name = utils.RandStringBytesMaskImprSrcUnsafe(16)
		}
		var bili float64
		configImageWidth := viper.GetInt("alibucket.imageWidth")
		configImageHeight := viper.GetInt("alibucket.imageHeight")
		if (data.ImageWidth > configImageWidth) || (data.ImageHeight > configImageHeight) {
			bili = float64(configImageWidth) / float64(data.ImageWidth)
		} else {
			bili = 1
		}
		var rects []transform.Rect
		for _, v := range data.Label {
			rects = append(rects, transform.Rect{
				math.Floor(utils.ToFloat64(v.Xmax)*bili + 0.5),
				math.Floor(utils.ToFloat64(v.Xmin)*bili + 0.5),
				math.Floor(utils.ToFloat64(v.Ymax)*bili + 0.5),
				math.Floor(utils.ToFloat64(v.Ymin)*bili + 0.5),
				v.Category,
			})
		}
		transform.PreLabelsData.Push(transform.Label{
			v.Image_path,
			utils.ImageOutPath,
			df.Name,
			df.Suffix,
			int(math.Floor(utils.ToFloat64(data.ImageWidth)*bili + 0.5)),
			int(math.Floor(utils.ToFloat64(data.ImageHeight)*bili + 0.5)),
			rects,
		})
		if needDownloadImageFile || needSplitImage {
			utils.DownloadIns.Goroutine_cnt <- 1
			wg.Add(1)
			go utils.DownloadIns.DGoroutine(&wg, df)
		}
	}
	wg.Wait()
	// 图片全部下载完毕
	if needSplitImage {
		transform.SlitImage(utils.ImageOutPath)
	}

	for _, v := range sql.Classes {
		transform.PreCategoriesData = append(transform.PreCategoriesData, transform.Category{
			v.Id,
			v.Index,
			v.Name,
		})
	}
}
