package yolo

import (
	"ai-dataset-tool/log"
	"ai-dataset-tool/transform"
	"ai-dataset-tool/utils"
	"os"
)

var labelFile *os.File
var classFile *os.File

func WriteYoloClassFile(classFilePath string) {
	var err error
	if classFile, err = os.OpenFile(classFilePath+"/class.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777); err != nil {
		log.Klog.Println("文件操作err", err)
		os.Exit(1)
	}
	var content string
	for _, c := range transform.PreCategoriesData {
		content += c.Name
	}
	fErr := utils.WriteFile(content, classFile)
	if fErr != nil {
		log.Klog.Println("class文件写入错误:", fErr)
	}
}
func WriteYoloLabelsFile(labelFilePath string) {
	for _, label := range transform.PreLabelsData.LabSlice {
		var err error
		if labelFile, err = os.OpenFile(labelFilePath+"/"+label.Name+".txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777); err != nil {
			log.Klog.Println("文件操作err", err)
			os.Exit(1)
		}
		var content string
		for _, rect := range label.Rects {
			rw := utils.PxToString((rect.Xmax - rect.Xmin) / utils.ToFloat64(label.ImageHeight))
			rh := utils.PxToString((rect.Ymax - rect.Ymin) / utils.ToFloat64(label.ImageHeight))
			rx := utils.PxToString((rect.Xmin + (rect.Xmax-rect.Xmin)/2) / utils.ToFloat64(label.ImageWidth))
			ry := utils.PxToString((rect.Ymin + (rect.Ymax-rect.Ymin)/2) / utils.ToFloat64(label.ImageHeight))
			content += rect.Category + " " + rx + " " + ry + " " + rw + " " + rh + "\n"
		}
		fErr := utils.WriteFile(content, labelFile)
		if fErr != nil {
			log.Klog.Println("标签文件写入错误:", fErr)
		}
		labelFile.Close()
	}
}
