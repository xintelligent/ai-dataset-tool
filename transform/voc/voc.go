package voc

import (
	"ai-dataset-tool/log"
	"ai-dataset-tool/sql"
	"ai-dataset-tool/utils"
	"encoding/json"
	"encoding/xml"
	"github.com/spf13/viper"
	"math"
	"math/rand"
	"os"
	"strings"
	"sync"
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
type source struct {
	Database   string `xml:"database"`
	Annotation string `xml:"annotation"`
	Image      string `xml:"image"`
	Flickrid   int    `xml:"flickrid"`
}
type owner struct {
	Flickrid string `xml:"flickrid"`
	Name     string `xml:"name"`
}
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

func WriteVocLabelsFile() {
	log.Klog.Printf("本次总图片数量：%d", len(sql.Labels))
	var wg sync.WaitGroup
	for _, value := range sql.Labels {
		df := utils.TransformFile(value.Image_path)
		labelData := sql.Data{}
		err := json.Unmarshal([]byte(value.Data), &labelData)
		if err != nil {
			log.Klog.Println("Label data field Unmarshal err", err)
		}
		log.Klog.Println(df.Name)

		utils.DownloadIns.Goroutine_cnt <- 1
		wg.Add(1)
		go utils.DownloadIns.DVGoroutine(&wg, df)

		var vocLabelsTrainFile *os.File // 训练集
		// 构建xml数据
		vocAnnotation := Annotation{
			Folder:   utils.DownloadIns.ImageOutPath,
			Filename: df.Name,
			Size: Size{
				labelData.ImageWidth,
				labelData.ImageHeight,
				3,
			},
			Segmented: 0,
		}
		var objs []Object
		var bili float64
		configImageWidth := viper.GetInt("alibucket.imageWidth")
		configImageHeight := viper.GetInt("alibucket.imageHeight")
		if (labelData.ImageWidth > configImageWidth) || (labelData.ImageHeight > configImageHeight) {
			bili = float64(configImageWidth) / float64(labelData.ImageWidth)
		} else {
			bili = 1
		}
		for _, val := range labelData.Label {
			obj := Object{
				utils.GetCategoryName(val.Category, &sql.Classes),
				"Unspecified",
				0,
				0,
				Bndbox{
					utils.PxToString(math.Floor(utils.ToFloat64(val.Xmin)*bili + 0.5)),
					utils.PxToString(math.Floor(utils.ToFloat64(val.Ymin)*bili + 0.5)),
					utils.PxToString(math.Floor(utils.ToFloat64(val.Xmax)*bili + 0.5)),
					utils.PxToString(math.Floor(utils.ToFloat64(val.Ymax)*bili + 0.5)),
				},
			}
			objs = append(objs, obj)
			if err != nil {
				log.Klog.Errorln("标签文件写入失败")
				os.Exit(1)
			}
		}
		vocAnnotation.Object = objs
		vocXml, xmlErr := xml.MarshalIndent(&vocAnnotation, "", "  ")
		if xmlErr != nil {
			log.Klog.Println("xml编码失败")
		}
		defer vocLabelsTrainFile.Close()
		// 一个图片文件
		if vocLabelsTrainFile, err = os.OpenFile(utils.DownloadIns.AnnotationOutPath+"/"+df.Name+".xml", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777); err != nil {
			log.Klog.Println("文件操作err", err)
			os.Exit(1)
		}
		fileErr := utils.WriteFile(string(vocXml), vocLabelsTrainFile)
		if fileErr != nil {
			log.Klog.Errorln("标签文件写入失败")
			os.Exit(1)
		}
	}
	wg.Wait()
	writeVocTraniAndTestFile()
}

func writeVocTraniAndTestFile() {
	imageCount := utils.ToFloat64(len(sql.Labels))
	testCount := math.Floor(imageCount*0.2 + 0.5)
	rand.Seed(time.Now().Unix())
	var testSlice []sql.Lab
	for {
		index := rand.Intn(len(sql.Labels) - 1)
		testSlice = append(testSlice, remove(index))
		if len(testSlice) >= int(testCount) {
			break
		}
	}
	writeTxt(&testSlice, "test.txt")
	writeTxt(&sql.Labels, "trainval.txt")

}

func remove(i int) (l sql.Lab) {
	l = sql.Labels[i]
	sql.Labels[i] = sql.Labels[len(sql.Labels)-1]
	sql.Labels = sql.Labels[:len(sql.Labels)-1]
	return
}
func writeTxt(l *[]sql.Lab, fileName string) {
	var fileContext string
	for _, v := range *l {
		firstIndex := strings.LastIndex(v.Image_path, "/")
		lastIndex := strings.LastIndex(v.Image_path, ".")
		fileContext += v.Image_path[firstIndex+1:lastIndex] + "\n"
	}
	if file, err := os.OpenFile(utils.DownloadIns.AnnotationOutPath+"/"+fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777); err == nil {
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
