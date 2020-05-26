package cmd

import (
	"ai-dataset-tool/log"
	"ai-dataset-tool/sql"
	"ai-dataset-tool/transform"
	"ai-dataset-tool/utils"
	"errors"
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
	Run: func(cmd *cobra.Command, args []string) {

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
	sql.Db.Close()
	// 下载oss 图片文件到本地
	utils.InitDownload()
	defer close(utils.DownloadIns.Goroutine_cnt)
	utils.InitOss()
	os.MkdirAll(utils.AnnotationOutPath, 0777)
	os.MkdirAll(utils.ImageOutPath, 0777)
	os.MkdirAll("split", 0777)
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
		df.Suffix = "jpg"
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
			bili = float64(configImageHeight) / float64(data.ImageHeight)
			if data.ImageWidth > data.ImageHeight {
				bili = float64(configImageWidth) / float64(data.ImageWidth)
			}
		} else {
			bili = 1
		}
		imageWidth := math.Floor(utils.ToFloat64(data.ImageWidth)*bili + 0.5)
		imageHeight := math.Floor(utils.ToFloat64(data.ImageHeight)*bili + 0.5)
		var rects []transform.Rect
		for _, val := range data.Label {
			r, err := verifyRectValue(val, bili, imageWidth, imageHeight)
			if err != nil {
				log.Klog.Println(v.Image_path, err)
				continue
			}
			rects = append(rects, r)
		}
		if len(rects) == 0 {
			log.Klog.Println("图片没有任何正确的标签:" + v.Image_path)
			continue
		}
		transform.PreLabelsData.Push(transform.Label{
			v.Image_path,
			utils.ImageOutPath,
			df.Name,
			df.Suffix,
			int(imageWidth),
			int(imageHeight),
			rects,
		})
		if needDownloadImageFile {
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

func verifyRectValue(s sql.Shape, bili float64, imageWidth float64, imageHeight float64) (transform.Rect, error) {
	xmax := math.Floor(utils.ToFloat64(s.Xmax)*bili + 0.5)
	xmin := math.Floor(utils.ToFloat64(s.Xmin)*bili + 0.5)
	if xmin < 0 {
		xmin = 0
	}
	ymax := math.Floor(utils.ToFloat64(s.Ymax)*bili + 0.5)
	ymin := math.Floor(utils.ToFloat64(s.Ymin)*bili + 0.5)
	if ymin < 0 {
		ymin = 0
	}

	t := transform.Rect{
		xmax,
		xmin,
		ymax,
		ymin,
		s.Category,
	}
	if xmax <= xmin || ymax <= ymin {
		return t, errors.New("最大值 小于 最小值")
	}
	if xmax <= 0 || ymax <= 0 {
		return t, errors.New("max 类型坐标有0值")
	}
	if xmax > imageWidth || ymax > imageHeight || xmin >= imageWidth || ymin >= imageHeight {
		return t, errors.New("坐标值大于图片宽高")
	}
	return t, nil
}
