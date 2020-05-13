package cmd

import (
	"math"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"
)

type annotation struct {
	Folder   string `xml:"folder"`
	Filename string `xml:"filename"`
	//Source source `xml:"source"`
	//Owner owner `xml:"owner"`
	Size      size     `xml:"size"`
	Segmented int      `xml:"segmented"`
	Object    []object `xml:"object"`
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
type size struct {
	Width  int `xml:"width"`
	Height int `xml:"height"`
	Depth  int `xml:"depth"`
}
type object struct {
	Name      string `xml:"name"`
	Pose      string `xml:"pose"`
	Truncated int    `xml:"truncated"`
	Difficult int    `xml:"difficult"`
	Bndbox    bndbox `xml:"bndbox"`
}
type bndbox struct {
	Xmin string `xml:"xmin"`
	Ymin string `xml:"ymin"`
	Xmax string `xml:"xmax"`
	Ymax string `xml:"ymax"`
}

func writeVocLabelsFile() {
	Log.Printf("本次总图片数量：%d", len(labels))
	var wg sync.WaitGroup
	for _, value := range labels {
		DownloadPoolIns.goroutine_cnt <- 1
		wg.Add(1)
		go DownloadPoolIns.DVGoroutine(&wg, transformFile(value.Image_path), value)
	}
	wg.Wait()
	writeVocTraniAndTestFile()
}

func writeVocTraniAndTestFile() {
	imageCount := toFloat64(len(labels))
	testCount := math.Floor(imageCount*0.2 + 0.5)
	rand.Seed(time.Now().Unix())
	var testSlice []lab
	for {
		index := rand.Intn(len(labels) - 1)
		testSlice = append(testSlice, remove(index))
		if len(testSlice) >= int(testCount) {
			break
		}
	}
	writeTxt(&testSlice, "test.txt")
	writeTxt(&labels, "trainval.txt")

}

func remove(i int) (l lab) {
	l = labels[i]
	labels[i] = labels[len(labels)-1]
	labels = labels[:len(labels)-1]
	return
}
func writeTxt(l *[]lab, fileName string) {
	var fileContext string
	for _, v := range *l {
		firstIndex := strings.LastIndex(v.Image_path, "/")
		lastIndex := strings.LastIndex(v.Image_path, ".")
		fileContext += v.Image_path[firstIndex+1:lastIndex] + ".jpg" + "\n"
	}
	if file, err := os.OpenFile(AnnotationOutPath+"/"+fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777); err == nil {
		err := WriteFile(fileContext, file)
		if err != nil {
			Log.Errorln("标签文件写入失败")
			os.Exit(1)
		}
	} else {
		Log.Println("文件操作err", err)
		os.Exit(1)
	}
}
