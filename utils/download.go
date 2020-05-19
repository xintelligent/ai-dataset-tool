package utils

import (
	"ai-dataset-tool/log"
	"github.com/spf13/viper"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"
)

var r *rand.Rand
var DownloadIns *Download
var (
	AnnotationOutPath string
	ImageOutPath      string
)

type Download struct {
	Goroutine_cnt chan int
}

type DownloadFile struct {
	ObjectName string // Oss 存储的文件名
	Name       string // 不包含后缀
	Suffix     string // 后缀
}

func TransformFile(path string) DownloadFile {
	slashIndex := strings.LastIndex(path, "/")
	dotIndex := strings.LastIndex(path, ".")
	return DownloadFile{
		path,
		rename(path[slashIndex+1 : dotIndex]),
		path[dotIndex+1:],
	}
}

func InitDownload() {
	DownloadIns = &Download{
		make(chan int, viper.GetInt("concurrent")),
	}
}

func (this *Download) DGoroutine(wg *sync.WaitGroup, file DownloadFile) {
	download(file)
	wg.Done()
	<-this.Goroutine_cnt
}
func (this *Download) GDownload(file DownloadFile) {
	download(file)
}

// ----------------------------------------------------------------
// 判断xml文件是否存在，存在的话就返回重新命名，因为图片文件有可能只是后缀不同
func rename(fileName string) (name string) {
	if _, err := os.Stat(AnnotationOutPath + "/" + fileName + ".xml"); os.IsNotExist(err) {
		log.Klog.Println("原文件名:" + fileName)
		return fileName
	}
	name = getRandString(16)
	log.Klog.Println("已存在，需要重命名,原文件名:" + fileName + "新文件名:" + name)
	return
}
func getRandString(len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		b := r.Intn(26) + 65
		bytes[i] = byte(b)
	}
	return string(bytes)
}

func init() {
	r = rand.New(rand.NewSource(time.Now().Unix()))
}
