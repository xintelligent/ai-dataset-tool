package cmd

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/spf13/viper"
)

type downloadPool struct {
	goroutine_cnt chan int
}

type downloadFile struct {
	objectName         string
	downloadedFileName string
}

// 定义进度变更事件处理函数。
func (listener *OssProgressListener) ProgressChanged(event *oss.ProgressEvent) {
	switch event.EventType {
	case oss.TransferStartedEvent:
		Log.Printf("Transfer Started, ConsumedBytes: %d, TotalBytes %d.\n",
			event.ConsumedBytes, event.TotalBytes)
	case oss.TransferDataEvent:
		Log.Printf("\rTransfer Data, ConsumedBytes: %d, TotalBytes %d, %d%%.",
			event.ConsumedBytes, event.TotalBytes, event.ConsumedBytes*100/event.TotalBytes)
	case oss.TransferCompletedEvent:
		Log.Printf("\nTransfer Completed, ConsumedBytes: %d, TotalBytes %d.\n",
			event.ConsumedBytes, event.TotalBytes)
	case oss.TransferFailedEvent:
		Log.Printf("\nTransfer Failed, ConsumedBytes: %d, TotalBytes %d.\n",
			event.ConsumedBytes, event.TotalBytes)
	default:
	}
}

// 定义进度条监听器。
type OssProgressListener struct {
}

func download(objectName string, downloadedFileName string) {
	derr := Bucket.GetObjectToFile(objectName, downloadedFileName, oss.Process(viper.GetString("alibucket.style")), oss.Progress(&OssProgressListener{}))
	if derr != nil {
		handleError(derr, objectName)
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
	download(file.objectName, file.downloadedFileName)
	<-this.goroutine_cnt
}
