package voc

import (
	"ai-dataset-tool/log"
	"ai-dataset-tool/sql"
	"ai-dataset-tool/transform"
	"ai-dataset-tool/utils"
	"encoding/xml"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"
)

type Annotation struct {
	Folder   string `xml:"folder"`
	Filename string `xml:"filename"`
	//Source source `xml:"source"`
	//Owner owner `xml:"owner"`
	Size      Size     `xml:"size"`
	Segmented int      `xml:"segmented"`
	Object    []Object `xml:"object"`
}

//type source struct {
//	Database   string `xml:"database"`
//	Annotation string `xml:"annotation"`
//	Image      string `xml:"imageTool"`
//	Flickrid   int    `xml:"flickrid"`
//}
//type owner struct {
//	Flickrid string `xml:"flickrid"`
//	Name     string `xml:"name"`
//}
type Size struct {
	Width  int `xml:"width"`
	Height int `xml:"height"`
	Depth  int `xml:"depth"`
}
type Object struct {
	Name      string `xml:"name"`
	Pose      string `xml:"pose"`
	Truncated int    `xml:"truncated"`
	Difficult int    `xml:"difficult"`
	Bndbox    Bndbox `xml:"bndbox"`
}
type Bndbox struct {
	Xmin string `xml:"xmin"`
	Ymin string `xml:"ymin"`
	Xmax string `xml:"xmax"`
	Ymax string `xml:"ymax"`
}

func WriteVocLabelsFile(annotationOutPath string, imageOutPath string) {
	for _, value := range transform.PreLabelsData.LabSlice {
		df := utils.TransformFile(value.Image_path)
		var vocLabelsTrainFile *os.File // 训练集
		// 构建xml数据
		vocAnnotation := Annotation{
			Folder:   imageOutPath,
			Filename: df.Name,
			Size: Size{
				value.ImageWidth,
				value.ImageHeight,
				3,
			},
			Segmented: 0,
		}
		var objs []Object
		for _, val := range value.Rects {
			obj := Object{
				utils.GetCategoryName(val.Category, &sql.Classes),
				"Unspecified",
				0,
				0,
				Bndbox{
					utils.PxToString(val.Xmin),
					utils.PxToString(val.Ymin),
					utils.PxToString(val.Xmax),
					utils.PxToString(val.Ymax),
				},
			}
			objs = append(objs, obj)
		}
		vocAnnotation.Object = objs
		vocXml, xmlErr := xml.MarshalIndent(&vocAnnotation, "", "  ")
		if xmlErr != nil {
			log.Klog.Println("xml编码失败")
		}
		defer vocLabelsTrainFile.Close()
		// 一个图片文件
		var vocErr error
		if vocLabelsTrainFile, vocErr = os.OpenFile(annotationOutPath+"/"+df.Name+".xml", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777); vocErr != nil {
			log.Klog.Println("文件操作err", vocErr)
			os.Exit(1)
		}
		fileErr := utils.WriteFile(string(vocXml), vocLabelsTrainFile)
		if fileErr != nil {
			log.Klog.Errorln("标签文件写入失败")
			os.Exit(1)
		}
	}
	writeVocTrainAndTestFile(annotationOutPath)
}

func writeVocTrainAndTestFile(annotationOutPath string) {
	imageCount := utils.ToFloat64(len(transform.PreLabelsData.LabSlice))
	testCount := math.Floor(imageCount*0.2 + 0.5)
	rand.Seed(time.Now().Unix())
	if imageCount > 2 {
		var testSlice []transform.Label
		for {
			index := rand.Intn(len(transform.PreLabelsData.LabSlice) - 1)
			testSlice = append(testSlice, remove(index))
			if len(testSlice) >= int(testCount) {
				break
			}
		}
		writeTxt(&testSlice, annotationOutPath+"/test.txt")
	}
	writeTxt(&transform.PreLabelsData.LabSlice, annotationOutPath+"/trainval.txt")

}

func remove(i int) (l transform.Label) {
	l = transform.PreLabelsData.LabSlice[i]
	transform.PreLabelsData.LabSlice[i] = transform.PreLabelsData.LabSlice[len(transform.PreLabelsData.LabSlice)-1]
	transform.PreLabelsData.LabSlice = transform.PreLabelsData.LabSlice[:len(transform.PreLabelsData.LabSlice)-1]
	return
}
func writeTxt(l *[]transform.Label, fileName string) {
	var fileContext string
	for _, v := range *l {
		firstIndex := strings.LastIndex(v.Image_path, "/")
		lastIndex := strings.LastIndex(v.Image_path, ".")
		fileContext += v.Image_path[firstIndex+1:lastIndex] + "\n"
	}
	if file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777); err == nil {
		err := utils.WriteFile(fileContext, file)
		if err != nil {
			log.Klog.Errorln("标签文件写入失败")
			os.Exit(1)
		}
	} else {
		log.Klog.Println("文件操作err", err)
		os.Exit(1)
	}
}
