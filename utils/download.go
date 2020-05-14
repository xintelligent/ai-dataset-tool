package utils

import (
	"ai-dataset-tool/log"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/spf13/viper"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"
)

var r *rand.Rand
var DownloadIns *Download

type Download struct {
	Goroutine_cnt         chan int
	AnnotationOutPath     string
	NeedDownloadImageFile bool
	ImageOutPath          string
}

type DownloadFile struct {
	ObjectName string // Oss 存储的文件名
	Name       string // 不包含后缀
	Suffix     string // 后缀
}

// 定义进度变更事件处理函数。
func (listener *OssProgressListener) ProgressChanged(event *oss.ProgressEvent) {
	switch event.EventType {
	case oss.TransferStartedEvent:
		fmt.Printf("Transfer Started, ConsumedBytes: %d, TotalBytes %d.\n",
			event.ConsumedBytes, event.TotalBytes)
	case oss.TransferDataEvent:
		fmt.Printf("\rTransfer Data, ConsumedBytes: %d, TotalBytes %d, %d%%.",
			event.ConsumedBytes, event.TotalBytes, event.ConsumedBytes*100/event.TotalBytes)
	case oss.TransferCompletedEvent:
		fmt.Printf("\nTransfer Completed, ConsumedBytes: %d, TotalBytes %d.\n",
			event.ConsumedBytes, event.TotalBytes)
	case oss.TransferFailedEvent:
		fmt.Printf("\nTransfer Failed, ConsumedBytes: %d, TotalBytes %d.\n",
			event.ConsumedBytes, event.TotalBytes)
	default:
	}
}

// 定义进度条监听器。
type OssProgressListener struct {
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
func download(df DownloadFile) {
	derr := Bucket.GetObjectToFile(df.ObjectName, DownloadIns.ImageOutPath+"/"+df.Name+"."+df.Suffix, oss.Process(viper.GetString("alibucket.style")), oss.Progress(&OssProgressListener{}))
	if derr != nil {
		handleError(derr, df.ObjectName)
	}
	return
}

func handleError(err error, objectName string) {
	log.Klog.Println("Error:", err)
	log.Klog.Info("下载失败：" + objectName)
}

func InitDownload() {
	DownloadIns.Goroutine_cnt = make(chan int, 10)
}

func (this *Download) DGoroutine(file DownloadFile) {
	download(file)
	<-this.Goroutine_cnt
}
func (this *Download) DVGoroutine(wg *sync.WaitGroup, file DownloadFile) {
	download(file)
	wg.Done()
	<-this.Goroutine_cnt
}

// 判断xml文件是否存在，存在的话就返回重新命名，因为图片文件有可能只是后缀不同
func rename(fileName string) (name string) {
	if _, err := os.Stat(DownloadIns.AnnotationOutPath + "/" + fileName + ".xml"); os.IsNotExist(err) {
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
