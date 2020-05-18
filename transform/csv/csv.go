package csv

import (
	"ai-dataset-tool/log"
	"ai-dataset-tool/sql"
	"ai-dataset-tool/utils"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"math"
	"os"
	"strings"
	"sync"
	"sync/atomic"
)

var labelsmapFile *os.File
var classesFile *os.File

func WriteCsvClassFile(classesFilePath string) {
	defer classesFile.Close()
	var err error
	if classesFile, err = os.OpenFile(classesFilePath+"/classes.csv", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777); err != nil {
		fmt.Println("文件操作err", err)
		os.Exit(1)
	}
	log.Klog.Printf("本次总类别数量：%d", len(sql.Classes))
	for _, value := range sql.Classes {
		content := value.Name + "," + utils.PxToString(value.Index) + "\n"
		err := utils.WriteFile(content, classesFile)
		if err != nil {
			os.Exit(1)
		}
	}
}
func WriteCsvLabelsFile(labelFilePath string, needDownloadImageFile bool) {
	labelData := sql.Data{}
	defer labelsmapFile.Close()
	var err error
	var wg sync.WaitGroup
	if labelsmapFile, err = os.OpenFile(labelFilePath+"/labelsmap.csv", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777); err != nil {
		log.Klog.Println("文件操作err", err)
		os.Exit(1)
	}
	log.Klog.Printf("本次总图片数量：%d", len(sql.Labels))
	var labelCount int64
	for _, value := range sql.Labels {
		err := json.Unmarshal([]byte(value.Data), &labelData)
		if err != nil {
			log.Klog.Println("Label data field Unmarshal err", err)
		}
		// 一个图片文件
		subIndex := strings.LastIndex(value.Image_path, "/")
		// fullFileIndex := strings.LastIndex(value.Image_path, ".")
		utils.DownloadIns.Goroutine_cnt <- 1
		go utils.DownloadIns.DGoroutine(&wg, utils.TransformFile(value.Image_path))
		var bili float64
		configImageWidth := viper.GetInt("alibucket.imageWidth")
		configImageHeight := viper.GetInt("alibucket.imageHeight")
		if (labelData.ImageWidth > configImageWidth) || (labelData.ImageHeight > configImageHeight) {
			bili = float64(configImageWidth) / float64(labelData.ImageWidth)
		} else {
			bili = 1
		}
		for _, val := range labelData.Label {
			atomic.AddInt64(&labelCount, 1)
			content := value.Image_path[subIndex+1:] + "," + utils.PxToString(math.Floor(utils.ToFloat64(val.Xmin)*bili+0.5)) + "," + utils.PxToString(math.Floor(utils.ToFloat64(val.Ymin)*bili+0.5)) + "," + utils.PxToString(math.Floor(utils.ToFloat64(val.Xmax)*bili+0.5)) + "," + utils.PxToString(math.Floor(utils.ToFloat64(val.Ymax)*bili+0.5)) + "," + utils.GetCategoryName(val.Category, &sql.Classes) + "\n"
			err := utils.WriteFile(content, labelsmapFile)
			if err != nil {
				log.Klog.Errorln("标签文件写入失败")
				os.Exit(1)
			}
		}
	}
	wg.Wait()
	log.Klog.Printf("标签总数：%d", labelCount)
}
