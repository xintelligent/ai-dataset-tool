package coco

import (
	"ai-dataset-tool/log"
	"ai-dataset-tool/transform"
	"encoding/json"
	"os"
	"strconv"
	"time"
)

type Coco struct {
	Info        Info         `json:"info"`
	Image       []Image      `json:"images"`
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

func WriteCocoFile(annotationOutPath string) error {
	setCoco()
	file, err := os.Create(annotationOutPath + "/coco.json")
	if err != nil {
		log.Klog.Println("创建文件失败:", err)
		return err
	}
	defer file.Close()
	jsonEncode := json.NewEncoder(file)
	writeErr := jsonEncode.Encode(CocoData)
	if writeErr != nil {
		log.Klog.Println("json写入失败:", writeErr)
		return err
	}
	return nil
}
func setCategories(c *[]Category) {
	for _, v := range transform.PreCategoriesData {
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
	for _, value := range transform.PreLabelsData.LabSlice {
		setImages(&images, &value)
		setAnnotation(&annotations, &value)
	}
	var categories []Category
	setCategories(&categories)
	CocoData.Image = images
	CocoData.Annotations = annotations
	CocoData.Categories = categories
}
func setImages(i *[]Image, l *transform.Label) {
	image := Image{
		Id:            l.Id,
		Width:         l.ImageWidth,
		Height:        l.ImageHeight,
		File_name:     l.Name + "." + l.Suffix,
		License:       0,
		Flickr_url:    "",
		Coco_url:      "",
		Date_captured: time.Now().Format("2006-01-02 15:04:05"),
	}
	*i = append(*i, image)
}
func setAnnotation(an *[]Annotation, l *transform.Label) {
	for _, v := range l.Rects {
		cid, err := strconv.Atoi(v.Category)
		if err != nil {
			log.Klog.Println(err)
		}
		annotation := Annotation{
			Id:           v.Index,
			Image_id:     l.Id,
			Category_id:  cid,
			Segmentation: [][]float64{{v.Xmin, v.Ymin, v.Xmax, v.Ymin, v.Xmax, v.Ymax, v.Xmin, v.Ymax}},
			Area:         (v.Xmax - v.Xmin) * (v.Ymax - v.Ymin),
			Bbox:         [4]float64{v.Xmin, v.Ymin, v.Xmax - v.Xmin, v.Ymax - v.Ymin},
			Iscrowd:      0,
		}
		*an = append(*an, annotation)
	}
}
