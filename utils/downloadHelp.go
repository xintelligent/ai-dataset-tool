package utils

import (
	"ai-dataset-tool/log"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/spf13/viper"
)

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

func download(df DownloadFile) {
	log.Klog.Println(viper.GetString("alibucket.style"))
	derr := Bucket.GetObjectToFile(df.ObjectName, ImageOutPath+"/"+df.Name+"."+df.Suffix, oss.Process(viper.GetString("alibucket.style")), oss.Progress(&OssProgressListener{}))
	if derr != nil {
		handleError(derr, df.ObjectName)
	}
	return
}

func handleError(err error, objectName string) {
	log.Klog.Println("Error:", err)
	log.Klog.Info("下载失败：" + objectName)
}
