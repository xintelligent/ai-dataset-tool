package yolo

import (
	"ai-dataset-tool/log"
	"ai-dataset-tool/transform"
	"ai-dataset-tool/utils"
	"math"
	"math/rand"
	"os"
	"time"
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
		content += c.Name + "\n"
	}
	fErr := utils.WriteFile(content, classFile)
	if fErr != nil {
		log.Klog.Println("class文件写入错误:", fErr)
	}
}
func WriteYoloLabelsFile(labelFilePath string, imageOutPath string) {
	for _, label := range transform.PreLabelsData.LabSlice {
		var err error
		if labelFile, err = os.OpenFile(labelFilePath+"/"+label.Name+".txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777); err != nil {
			log.Klog.Println("文件操作err", err)
			os.Exit(1)
		}
		var content string
		for _, rect := range label.Rects {
			rw := utils.PxToString((rect.Xmax - rect.Xmin) / utils.ToFloat64(label.ImageWidth))
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
	splitDir(labelFilePath, imageOutPath)
}

func splitDir(annotationOutPath string, imageOutPath string) {
	os.MkdirAll(annotationOutPath+"/sub/", 0777)
	os.MkdirAll(imageOutPath+"/sub/", 0777)
	//os.MkdirAll("./split/sub/", 0777)
	imageCount := utils.ToFloat64(len(transform.PreLabelsData.LabSlice))
	testCount := math.Floor(imageCount*0.2 + 0.5)
	rand.Seed(time.Now().Unix())
	if imageCount > 2 {
		var testSlice []transform.Label
		for {
			index := rand.Intn(len(transform.PreLabelsData.LabSlice) - 1)
			testSlice = append(testSlice, remove(index))
			// 移动下annotation文件
			afile := transform.PreLabelsData.LabSlice[index].Name + ".txt"
			if fileIsExisted(annotationOutPath + "/" + afile) {
				os.Rename(annotationOutPath+"/"+afile, annotationOutPath+"/sub/"+afile)
			}
			// 移动下image文件
			ifile := transform.PreLabelsData.LabSlice[index].Name + "." + transform.PreLabelsData.LabSlice[index].Suffix
			log.Klog.Println(ifile)
			//if fileIsExisted("./split/" + ifile) {
			//	os.Rename("./split/"+ifile, "./split/sub/"+ifile)
			//
			//}
			if fileIsExisted(imageOutPath + "/" + ifile) {
				os.Rename(imageOutPath+"/"+ifile, imageOutPath+"/sub/"+ifile)
			} else {
				log.Klog.Println("没有:" + imageOutPath + "/" + ifile)
			}
			if len(testSlice) >= int(testCount) {
				break
			}
		}

	}
}
func remove(i int) (l transform.Label) {
	l = transform.PreLabelsData.LabSlice[i]
	transform.PreLabelsData.LabSlice[i] = transform.PreLabelsData.LabSlice[len(transform.PreLabelsData.LabSlice)-1]
	transform.PreLabelsData.LabSlice = transform.PreLabelsData.LabSlice[:len(transform.PreLabelsData.LabSlice)-1]
	return
}
func fileIsExisted(filename string) bool {
	existed := true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		existed = false
	}
	return existed
}
