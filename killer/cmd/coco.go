package cmd

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

type Coco struct {
	Info        Info         `json:"info"`
	Image       []Image      `json:"image"`
	Annotations []Annotation `json:"annotations"`
	Categories  []Category   `json:"categories"`
}

type Info struct {
	Year         int    `json:"year"`
	Version      string `json:"version"`
	Description  string `json:"description"`
	Contributor  string `json:"contributor"`
	Url          string `json:"url"`
	Date_created string `json:"date_created"`
}
type Image struct {
	Id            int    `json:"id"`
	Width         int    `json:"width"`
	Height        int    `json:"height"`
	File_name     string `json:"file_name"`
	License       int    `json:"license"`
	Flickr_url    string `json:"flickr_url"`
	Coco_url      string `json:"coco_url"`
	Date_captured string `json:"date_captured"`
}
type License struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

// ------- 对象识别需要
type Annotation struct {
	Id           int         `json:"id"`
	Image_id     int         `json:"image_id"`
	Category_id  int         `json:"category_id"`
	Segmentation [][]float64 `json:"segmentation"`
	Area         float64     `json:"area"`
	Bbox         [4]float64  `json:"bbox"`
	Iscrowd      int         `json:"iscrowd"`
}
type Category struct {
	Id            int    `json:"id"`
	Name          string `json:"name"`
	Supercategory string `json:"supercategory"`
}
type Box struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// -------

var CocoData Coco

func WriteCocoFile() error {
	setCoco()
	file, err := os.Create(AnnotationOutPath + "/coco.json")
	if err != nil {
		fmt.Println("创建文件失败:", err)
		return err
	}
	defer file.Close()
	jsonEncode := json.NewEncoder(file)
	writeErr := jsonEncode.Encode(CocoData)
	if writeErr != nil {
		fmt.Println("json写入失败:", writeErr)
		return err
	}
	return nil
}
func setCategories(c *[]Category) {
	for _, v := range classes {
		category := Category{
			Id:            v.Index,
			Name:          v.Name,
			Supercategory: v.Name,
		}
		*c = append(*c, category)
	}
}
func setCoco() {
	CocoData.Info = Info{
		Year:         time.Now().Year(),
		Version:      "v1.0",
		Description:  "v1",
		Contributor:  "zoomicro",
		Url:          "",
		Date_created: time.Now().Format("2006-01-02 15:04:05"),
	}
	var images []Image
	var annotations []Annotation

	labelData := data{}
	for _, value := range labels {
		err := json.Unmarshal([]byte(value.Data), &labelData)
		if err != nil {
			fmt.Println("Unmarshal err", err)
		}
		subIndex := strings.LastIndex(value.Image_path, "/")
		if NeedDownloadImageFile {
			var file = downloadFile{objectName: value.Image_path, downloadedFileName: ImageOutPath + "/" + value.Image_path[subIndex+1:]}
			DownloadPoolIns.goroutine_cnt <- 1
			go DownloadPoolIns.DGoroutine(file)
		}
		setImages(&images, &value, &labelData)
		var bili float64
		if labelData.ImageWidth > 1024 {
			bili = float64(1024) / float64(labelData.ImageWidth)
		} else {
			bili = 1
		}
		setAnnotation(&annotations, &labelData, value.Id, value.Id, bili)
	}
	var categories []Category
	setCategories(&categories)
	CocoData.Image = images
	CocoData.Annotations = annotations
	CocoData.Categories = categories
}
func setImages(i *[]Image, l *lab, ld *data) {
	subIndex := strings.LastIndex(l.Image_path, "/")
	image := Image{
		Id:            l.Id,
		Width:         ld.ImageWidth,
		Height:        ld.ImageHeight,
		File_name:     l.Image_path[subIndex+1:],
		License:       0,
		Flickr_url:    "",
		Coco_url:      "",
		Date_captured: time.Now().Format("2006-01-02 15:04:05"),
	}
	*i = append(*i, image)
}
func setAnnotation(an *[]Annotation, ld *data, id int, imageId int, bili float64) {
	for _, v := range ld.Label {
		cid, err := strconv.Atoi(v.Category)
		if err != nil {
			fmt.Println(err)
		}
		xmin := math.Floor(toFloat64(v.Xmin)*bili + 0.5)
		ymin := math.Floor(toFloat64(v.Ymin)*bili + 0.5)
		xmax := math.Floor(toFloat64(v.Xmax)*bili + 0.5)
		ymax := math.Floor(toFloat64(v.Ymax)*bili + 0.5)
		annotation := Annotation{
			Id:           id,
			Image_id:     imageId,
			Category_id:  cid,
			Segmentation: [][]float64{{xmin, ymin, xmax, ymin, xmax, ymax, xmin, ymax}},
			Area:         (xmax - xmin) * (ymax - ymin),
			Bbox:         [4]float64{xmin, ymin, xmax - xmin, ymax - ymin},
			Iscrowd:      0,
		}
		*an = append(*an, annotation)
	}
}
