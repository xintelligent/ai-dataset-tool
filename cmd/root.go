package cmd

import (
	"ai-dataset-tool/log"
	"ai-dataset-tool/sql"
	"ai-dataset-tool/transform"
	"ai-dataset-tool/utils"
	"errors"
	"fmt"
	"image"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
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
var rectIndex int = 1

func Exec() {
	// 准备日志服务
	log.InitLog()
	// 准备数据库连接
	sql.InitSql()
	// 准本数据源
	var classErr error
	classErr = sql.NewClassesFromMysql()
	if classErr != nil {
		log.Klog.Println("标签类别查询失败", classErr)
		os.Exit(1)
	}
	var labelErr error
	labelErr = sql.NewLabelsFromMysql()
	if labelErr != nil {
		log.Klog.Println("查询标签数据失败", labelErr)
		os.Exit(1)
	}
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
				v.Index,
				v.Name,
			})
		} else {
			index, _ := strconv.Atoi(ca.newIndex)
			transform.PreCategoriesData = append(transform.PreCategoriesData, transform.Category{
				index,
				ca.newName,
			})
		}

	}
	// 计数,当作索引用
	var polygonCount int
	var imageCount int
	// 验证每一条数据
	for _, v := range sql.Labels {
		// 拆分文件名
		df := utils.TransformFile(v.ImagePath)
		df.Suffix = "jpg"
		// 验证图片名称重复（不带后缀）
		if _, ok := images[df.Name]; !ok {
			images[df.Name] = true
		} else {
			df.Name = utils.RandStringBytesMaskImprSrcUnsafe(16)
		}

		var bili float64
		configImageWidth := viper.GetInt("alibucket.imageWidth")
		configImageHeight := viper.GetInt("alibucket.imageHeight")
		info := utils.GetImageInfoFromOSS(v.ImagePath)
		fmt.Println("阿里云响应图片尺寸:::::", info)
		v.ImageWidth, _ = strconv.Atoi(info.ImageWidth.Value)
		v.ImageHeight, _ = strconv.Atoi(info.ImageHeight.Value)
		if (configImageWidth > 0 || configImageHeight > 0) && ((v.ImageWidth > configImageWidth) || (v.ImageHeight > configImageHeight)) {
			bili = float64(configImageHeight) / float64(v.ImageHeight)
			if v.ImageWidth > v.ImageHeight {
				bili = float64(configImageWidth) / float64(v.ImageWidth)
			}
		} else {
			bili = 1
		}
		imageWidth := math.Floor(utils.ToFloat64(v.ImageWidth)*bili + 0.5)
		imageHeight := math.Floor(utils.ToFloat64(v.ImageHeight)*bili + 0.5)
		var polygons []transform.Polygon
		// TODO::::: 标签类型 不去考虑具体形状(将Rect当作两个点的多边形)
		for _, val := range v.LabelPolygon {
			err := filterRectValue(val, bili, imageWidth, imageHeight, &cm)
			if err != nil {
				log.Klog.Println(v.ImagePath, err)
				continue
			}
			polygonCount++
			polygons = append(polygons, transform.Polygon{
				Id:        val.Id,
				Index:     polygonCount,
				Point:     val.Point,
				Category:  val.Category,
				LabelType: val.LabelType,
			})
		}
		if len(polygons) == 0 {
			log.Klog.Println("图片没有任何正确的标签:" + v.ImagePath)
			continue
		}
		imageCount++
		transform.PreLabelsData.Push(transform.Label{
			Index:        imageCount,
			Image_path:   v.ImagePath,
			ImageOutPath: utils.ImageOutPath,
			Name:         df.Name,
			Suffix:       df.Suffix,
			ImageWidth:   int(imageWidth),
			ImageHeight:  int(imageHeight),
			Polygons:     polygons,
		})
		if needDownloadImageFile {
			utils.DownloadIns.Goroutine_cnt <- 1
			wg.Add(1)
			go utils.DownloadIns.DGoroutine(&wg, df)
		}
	}
	wg.Wait()
	if needDownloadImageFile {
		checkImageFile()
	}
	// 图片全部下载完毕 TODO:::
	if needSplitImage {
		fmt.Println("切分功能关闭了TODO")
		//transform.SlitImage(utils.ImageOutPath)
		//utils.ImageOutPath = "./split"
	}
}
func checkImageFile() {
	var needRemove []int
	for k, label := range transform.PreLabelsData.LabSlice {
		f, err := os.Open(label.ImageOutPath + "/" + label.Name + "." + label.Suffix)
		if err != nil {
			removeImageFromPre(k)
		}
		imgFile, _, err := image.DecodeConfig(f)
		if err != nil {
			log.Klog.Println(err.Error())
		}
		log.Klog.Println("标签数据宽高: ", label.ImageWidth, label.ImageHeight, label.Name)
		log.Klog.Println("图片实际宽高: ", imgFile.Width, imgFile.Height)
		if imgFile.Width != label.ImageWidth || imgFile.Height != label.ImageHeight {
			needRemove = append(needRemove, k)
			log.Klog.Println("不符合实际情况的图片: ", label)
		}
		f.Close()
	}
	if len(needRemove) > 0 {
		for _, v := range needRemove {
			removeImageFromPre(v)
		}
	}
	log.Klog.Println("校验原文件结束")
}
func removeImageFromPre(i int) (l transform.Label) {
	l = transform.PreLabelsData.LabSlice[i]
	transform.PreLabelsData.LabSlice[i] = transform.PreLabelsData.LabSlice[len(transform.PreLabelsData.LabSlice)-1]
	transform.PreLabelsData.LabSlice = transform.PreLabelsData.LabSlice[:len(transform.PreLabelsData.LabSlice)-1]
	return
}

// 发现有错误数据，log记录后直接放弃 TODO::: 缩放
func filterRectValue(s sql.Polygon, bili float64, imageWidth float64, imageHeight float64, cm *[]map[string]string) error {
	for _, value := range s.Point {
		x := value[0]
		y := value[1]

		if x <= 0 || y <= 0 {
			return errors.New("max 类型坐标有0值")
		}
		if x > imageWidth || y > imageHeight {
			fmt.Println("图片高度", imageWidth, imageHeight, x, y)
			return errors.New("坐标值大于图片宽高")
		}

		if ca, err := rectConvertClassMap(s.Category, cm); err == nil {
			s.Category = ca.newIndex
		}
	}
	rectIndex++
	return nil
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
