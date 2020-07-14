package utils

import (
	"ai-dataset-tool/log"
	"encoding/json"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/spf13/viper"
	"io/ioutil"
)

var Bucket *oss.Bucket

func InitOss() {
	client, err := oss.New(viper.GetString("alibucket.endpoint"), viper.GetString("alibucket.accessKeyId"), viper.GetString("alibucket.accessKeySecret"))
	if err != nil {
		log.Klog.Println(err)
	}
	var berr error
	Bucket, berr = client.Bucket(viper.GetString("alibucket.bucketName"))
	if berr != nil {
		log.Klog.Println(berr)
	}
}

/**
图片的基本信息
{
  "FileSize": {"value": "21839"},
  "Format": {"value": "jpg"},
  "ImageHeight": {"value": "267"},
  "ImageWidth": {"value": "400"}
}
{
    "FileSize": {"value": "1054784"},
    "Format": {"value": "jpg"},
    "ImageHeight": {"value": "2160"},
    "ImageWidth": {"value": "2880"},
    "ResolutionUnit": {"value": "1"},
    "XResolution": {"value": "1/1"},
    "YResolution": {"value": "1/1"}}
*/
type ImageInfo struct {
	FileSize    FileSize    `json:"FileSize"`
	Format      Format      `json:"format"`
	ImageHeight ImageHeight `json:"ImageHeight"`
	ImageWidth  ImageWidth  `json:"ImageWidth"`
}
type FileSize struct {
	Value string `json:"value"`
}
type Format struct {
	Value string `json:"value"`
}
type ImageHeight struct {
	Value string `json:"value"`
}
type ImageWidth struct {
	Value string `json:"value"`
}

func GetImageInfoFromOSS(imageN string) (i ImageInfo) {
	resp, err := Bucket.GetObject(imageN, oss.Process("image/info"))
	if err != nil {
		log.Klog.Println("查询图片信息失败: "+imageN, err)
	}
	body, IOErr := ioutil.ReadAll(resp)
	if IOErr != nil {
		fmt.Println("读取响应内容失败: ", IOErr)
	}
	jsonErr := json.Unmarshal(body, &i)
	if jsonErr != nil {
		log.Klog.Println(jsonErr)
	}
	_ = resp.Close()
	return
}
