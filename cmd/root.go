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
	"strconv"
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
	cobra.OnInitialize(prepare)
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
	viperClassMap := viper.Get("dataset.classMap")
	cm := []map[string]string{}
	// 类别映射关系 []map[string]int,转化设置的值中，索引值不能重复
	if cm != nil {
		switch t := viperClassMap.(type) {
		case []interface{}:
			for _, v := range t {
				switch f := v.(type) {
				case map[interface{}]interface{}:
					wip := make(map[string]string)
					for key, value := range f {
						wip[string(key.(string))] = string(value.(string))
					}
					cm = append(cm, wip)
				}
			}
		}
	}
	// 处理classMap
	for _, v := range sql.Classes {
		ca, err := rectConvertClassMap(strconv.Itoa(v.Index), &cm)
		if err != nil {
			transform.PreCategoriesData = append(transform.PreCategoriesData, transform.Category{
				v.Id,
				v.Index,
				v.Name,
			})
		} else {
			index, _ := strconv.Atoi(ca.newIndex)
			transform.PreCategoriesData = append(transform.PreCategoriesData, transform.Category{
				v.Id,
				index,
				ca.newName,
			})
		}

	}
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
			r, err := filterRectValue(val, bili, imageWidth, imageHeight, &cm)
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
			Image_path:   v.Image_path,
			ImageOutPath: utils.ImageOutPath,
			Name:         df.Name,
			Suffix:       df.Suffix,
			ImageWidth:   int(imageWidth),
			ImageHeight:  int(imageHeight),
			Rects:        rects,
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
		utils.ImageOutPath = "./split"
	}
}

func filterRectValue(s sql.Shape, bili float64, imageWidth float64, imageHeight float64, cm *[]map[string]string) (transform.Rect, error) {
	xmax := math.Floor(utils.ToFloat64(s.Xmax)*bili + 0.5)
	xmin := math.Floor(utils.ToFloat64(s.Xmin)*bili + 0.5)
	if xmin < 0 {
		xmin = 0
	}
	ymax := math.Floor(utils.ToFloat64(s.Ymax)*bili + 0.5)
	ymin := math.Floor(utils.ToFloat64(s.Ymin)*bili + 0.5)
	if zoomRect := viper.GetInt("dataset.zoomRect"); (zoomRect != 0) && (zoomRect > 0) {
		xmax = xmax * (utils.ToFloat64(zoomRect) * xmax / 100)
		xmin = xmin * (utils.ToFloat64(zoomRect) * xmin / 100)
		ymax = ymax * (utils.ToFloat64(zoomRect) * ymax / 100)
		ymin = xmin * (utils.ToFloat64(zoomRect) * xmin / 100)
	}
	if ymin < 0 {
		ymin = 0
	}
	t := transform.Rect{
		Xmax: xmax,
		Xmin: xmin,
		Ymax: ymax,
		Ymin: ymin,
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

	if ca, err := rectConvertClassMap(s.Category, cm); err != nil {
		t.Category = s.Category
	} else {
		t.Category = ca.newIndex

	}
	return t, nil
}

// 将标签中的Category转为目标数值
type category struct {
	oldIndex string
	newIndex string
	oldName  string
	newName  string
}

func rectConvertClassMap(i string, cm *[]map[string]string) (category, error) {
	for _, v := range *cm {
		if value, ok := v["oldIndex"]; ok && value == i {
			return category{oldIndex: v["oldIndex"], newIndex: v["newIndex"], oldName: v["oldName"], newName: v["newName"]}, nil
		}
	}
	return category{}, errors.New("no match")
}
