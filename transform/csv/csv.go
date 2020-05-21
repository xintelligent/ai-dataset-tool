package csv

import (
	"ai-dataset-tool/log"
	"ai-dataset-tool/sql"
	"ai-dataset-tool/transform"
	"ai-dataset-tool/utils"
	"fmt"
	"os"
	"strings"
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
func WriteCsvLabelsFile(labelFilePath string) {
	defer labelsmapFile.Close()
	var err error
	if labelsmapFile, err = os.OpenFile(labelFilePath+"/labelsmap.csv", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777); err != nil {
		log.Klog.Println("文件操作err", err)
		os.Exit(1)
	}
	log.Klog.Printf("本次总图片数量：%d", len(transform.PreLabelsData.LabSlice))
	var labelCount int64
	for _, value := range transform.PreLabelsData.LabSlice {
		// 一个图片文件
		subIndex := strings.LastIndex(value.Image_path, "/")
		for _, val := range value.Rects {
			atomic.AddInt64(&labelCount, 1)
			content := value.Image_path[subIndex+1:] + "," + utils.PxToString(val.Xmin) + "," + utils.PxToString(val.Ymin) + "," + utils.PxToString(val.Xmax) + "," + utils.PxToString(val.Ymax) + "," + utils.GetCategoryName(val.Category, &sql.Classes) + "\n"
			err := utils.WriteFile(content, labelsmapFile)
			if err != nil {
				log.Klog.Errorln("标签文件写入失败")
				os.Exit(1)
			}
		}
	}
	log.Klog.Printf("标签总数：%d", labelCount)
}
