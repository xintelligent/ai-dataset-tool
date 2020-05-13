package cmd

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/spf13/viper"
	"math"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"
)

var r *rand.Rand

type downloadPool struct {
	goroutine_cnt chan int
}

type downloadFile struct {
	objectName string // Oss 存储的文件名
	name       string // 不包含后缀
	suffix     string // 后缀
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

func transformFile(path string) downloadFile {
	slashIndex := strings.LastIndex(path, "/")
	dotIndex := strings.LastIndex(path, ".")
	return downloadFile{
		path,
		rename(path[slashIndex+1 : dotIndex]),
		path[dotIndex+1:],
	}
}
func download(df downloadFile) {
	derr := Bucket.GetObjectToFile(df.objectName, ImageOutPath+"/"+df.name+"."+df.suffix, oss.Process(viper.GetString("alibucket.style")), oss.Progress(&OssProgressListener{}))
	if derr != nil {
		handleError(derr, df.objectName)
	}
	return
}

func handleError(err error, objectName string) {
	Log.Println("Error:", err)
	Log.Info("下载失败：" + objectName)
}

func NewDownloadPool() (d *downloadPool) {
	d = new(downloadPool)
	d.goroutine_cnt = make(chan int, 10)
	return
}

func (this *downloadPool) DGoroutine(file downloadFile) {
	download(file)
	<-this.goroutine_cnt
}
func (this *downloadPool) DVGoroutine(wg *sync.WaitGroup, file downloadFile, label lab) {
	type vocLabels struct {
		annotation
	}
	labelData := data{}
	err := json.Unmarshal([]byte(label.Data), &labelData)
	if err != nil {
		Log.Println("Label data field Unmarshal err", err)
	}
	Log.Println(file.objectName)
	//download(file)
	var vocLabelsTrainFile *os.File // 训练集
	var reader *os.File
	defer reader.Close()
	// 构建xml数据
	vocAnnotation := annotation{
		Folder:   ImageOutPath,
		Filename: file.name,
		Size: size{
			labelData.ImageWidth,
			labelData.ImageHeight,
			3,
		},
		Segmented: 0,
	}

	var objs []object
	var bili float64
	configImageWidth := viper.GetInt("alibucket.imageWidth")
	configImageHeight := viper.GetInt("alibucket.imageHeight")
	if (labelData.ImageWidth > configImageWidth) || (labelData.ImageHeight > configImageHeight) {
		bili = float64(configImageWidth) / float64(labelData.ImageWidth)
	} else {
		bili = 1
	}
	for _, val := range labelData.Label {
		obj := object{
			getCategoryName(val.Category),
			"Unspecified",
			0,
			0,
			bndbox{
				pxToString(math.Floor(toFloat64(val.Xmin)*bili + 0.5)),
				pxToString(math.Floor(toFloat64(val.Ymin)*bili + 0.5)),
				pxToString(math.Floor(toFloat64(val.Xmax)*bili + 0.5)),
				pxToString(math.Floor(toFloat64(val.Ymax)*bili + 0.5)),
			},
		}
		objs = append(objs, obj)
		if err != nil {
			Log.Errorln("标签文件写入失败")
			os.Exit(1)
		}
	}
	vocAnnotation.Object = objs
	vocXml, xmlErr := xml.MarshalIndent(&vocAnnotation, "", "  ")
	if xmlErr != nil {
		Log.Println("xml编码失败")
	}
	defer vocLabelsTrainFile.Close()
	// 一个图片文件
	if vocLabelsTrainFile, err = os.OpenFile(AnnotationOutPath+"/"+file.name+".xml", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777); err != nil {
		Log.Println("文件操作err", err)
		os.Exit(1)
	}
	fileErr := WriteFile(string(vocXml), vocLabelsTrainFile)
	if fileErr != nil {
		Log.Errorln("标签文件写入失败")
		os.Exit(1)
	}
	wg.Done()
	<-this.goroutine_cnt
}

// 判断xml文件是否存在，存在的话就返回重新命名，因为图片文件有可能只是后缀不同
func rename(fileName string) (name string) {
	if _, err := os.Stat(AnnotationOutPath + "/" + fileName + ".xml"); os.IsNotExist(err) {
		Log.Println("原文件名:" + fileName)
		return fileName
	}
	name = getRandString(16)
	Log.Println("已存在，需要重命名,原文件名:" + fileName + "新文件名:" + name)
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
