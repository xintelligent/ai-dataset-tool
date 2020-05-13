package cmd

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"sync/atomic"
)

var labelsmapFile *os.File
var classesFile *os.File

func writeCsvClassFile() {
	defer classesFile.Close()
	var err error
	if classesFile, err = os.OpenFile(AnnotationOutPath+"/classes.csv", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777); err != nil {
		fmt.Println("文件操作err", err)
		os.Exit(1)
	}
	Log.Printf("本次总类别数量：%d", len(classes))
	for _, value := range classes {
		content := value.Name + "," + pxToString(value.Index) + "\n"
		err := WriteFile(content, classesFile)
		if err != nil {
			os.Exit(1)
		}
	}
}
func writeCsvLabelsFile() {
	labelData := data{}
	defer labelsmapFile.Close()
	var err error
	if labelsmapFile, err = os.OpenFile(AnnotationOutPath+"/labelsmap.csv", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777); err != nil {
		Log.Println("文件操作err", err)
		os.Exit(1)
	}
	Log.Printf("本次总图片数量：%d", len(labels))
	var labelCount int64
	for _, value := range labels {
		err := json.Unmarshal([]byte(value.Data), &labelData)
		if err != nil {
			Log.Println("Label data field Unmarshal err", err)
		}
		// 一个图片文件
		subIndex := strings.LastIndex(value.Image_path, "/")
		// fullFileIndex := strings.LastIndex(value.Image_path, ".")
		if NeedDownloadImageFile {
			DownloadPoolIns.goroutine_cnt <- 1
			go DownloadPoolIns.DGoroutine(transformFile(value.Image_path))
		}
		var bili float64
		if labelData.ImageWidth > 1024 {
			bili = float64(1024) / float64(labelData.ImageWidth)
		} else {
			bili = 1
		}
		for _, val := range labelData.Label {
			atomic.AddInt64(&labelCount, 1)
			content := value.Image_path[subIndex+1:] + "," + pxToString(math.Floor(toFloat64(val.Xmin)*bili+0.5)) + "," + pxToString(math.Floor(toFloat64(val.Ymin)*bili+0.5)) + "," + pxToString(math.Floor(toFloat64(val.Xmax)*bili+0.5)) + "," + pxToString(math.Floor(toFloat64(val.Ymax)*bili+0.5)) + "," + getCategoryName(val.Category) + "\n"
			err := WriteFile(content, labelsmapFile)
			if err != nil {
				Log.Errorln("标签文件写入失败")
				os.Exit(1)
			}
		}
	}
	Log.Printf("标签总数：%d", labelCount)
}
