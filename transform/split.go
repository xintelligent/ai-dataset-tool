package transform

import (
	"ai-dataset-tool/imageTool"
	"ai-dataset-tool/log"
	"ai-dataset-tool/sql"
	"ai-dataset-tool/utils"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"image"
	"image/jpeg"
	"math"
	"os"
)

var splitLabelsData = Labels{}

/**
输入一个图片，产出多个切分后的图片，和对应的标签
*/
type rect struct {
	Xmax     int
	Xmin     int
	Ymax     int
	Ymin     int
	Category string
}

// 图片分割
type cellGroup struct {
	cellMembers []cellMember
}

type cellMember struct {
	row         int
	column      int
	x0          int
	y0          int
	x1          int
	y1          int
	ImagePath   string
	ImageName   string
	rects       []rect
	ImageWidth  int
	ImageHeight int
}

// 是一个面积大于真实图片的尺寸
type imageObject struct {
	width     int
	height    int
	splitSize int
	cellGroup
	data      sql.Data
	imageName string
}

var imageFile *os.File

func NewImage(l sql.Lab, splitSize int, df utils.DownloadFile) (i imageObject) {
	var data sql.Data
	dataErr := json.Unmarshal([]byte(l.Data), &data)
	if dataErr != nil {
		log.Klog.Println(dataErr)
	}
	i.width = int(math.Ceil(float64(data.ImageWidth)/float64(splitSize))) * splitSize
	i.height = int(math.Ceil(float64(data.ImageHeight)/float64(splitSize))) * splitSize
	i.splitSize = splitSize
	i.data = data
	i.imageName = df.Name + "." + df.Suffix
	return
}

func (i *imageObject) initCellGroup(imageOutPath string) {
	var imageErr error
	if imageFile, imageErr = os.OpenFile(imageOutPath+"/"+i.imageName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777); imageErr != nil {
		log.Klog.Println("没有能打开图片文件", imageErr)
		os.Exit(1)
	}
	origin, _, err := image.Decode(imageFile)
	if err != nil {
		log.Klog.Println("没有能解码图片文件", err)
	}
	rowCount := i.height / i.splitSize
	columnCount := i.width / i.splitSize
	for row := 0; row < rowCount; row++ {
		for column := 0; column < columnCount; column++ {
			newImageName := utils.RandStringBytesMaskImprSrcUnsafe(16) + "-" + utils.PxToString(row) + "-" + utils.PxToString(column)
			c := cellMember{
				column:      column,
				row:         row,
				x0:          column * i.splitSize,
				y0:          row * i.splitSize,
				x1:          (column + 1) * i.splitSize,
				y1:          (row + 1) * i.splitSize,
				ImagePath:   newImageName,
				ImageName:   imageOutPath,
				ImageWidth:  i.splitSize,
				ImageHeight: i.splitSize,
			}
			cropCell(&origin, &c, newImageName+".jpg")
			i.cellGroup.cellMembers = append(i.cellGroup.cellMembers, c)

			splitLabelsData.Push(Lab{
				"",
				viper.GetString("dataset.splitImageOutPath"),
				newImageName,
				"jpg",
				Data{
					labels,
					c.ImageWidth,
					c.ImageHeight,
				},
			})
		}
	}
}

// 计算有那个shape 在当前cellMember里面
func (i *imageObject) getSubLabels(c *cellMember) (s []Shape) {
	return
}
func cropCell(o *image.Image, c *cellMember, newFile string) {
	r := image.Rect(c.x0, c.y0, c.x1, c.y1)
	crop := imageTool.NewCrop(r)
	dst := image.NewRGBA(crop.Bounds(r))
	crop.Draw(dst, *o, &imageTool.Options{Parallelization: true})
	saveImage(newFile, dst)
	return
}
func saveImage(filename string, img image.Image) {
	f, err := os.Create(viper.GetString("dataset.splitImageOutPath") + "/" + filename)
	if err != nil {
		log.Klog.Fatalf("os.Create failed: %v", err)
	}
	defer f.Close()
	op := jpeg.Options{100}
	err = jpeg.Encode(f, img, &op)
	if err != nil {
		log.Klog.Fatalf("png.Encode failed: %v", err)
	}
}
func (i *imageObject) dispatchLabel() {
	for _, value := range i.data.Label {
		xmin := utils.ToFloat64(value.Xmin)
		xmax := utils.ToFloat64(value.Xmax)
		ymin := utils.ToFloat64(value.Ymin)
		ymax := utils.ToFloat64(value.Ymax)
		rx := (xmax-xmin)/2 + xmin
		ry := (ymax-ymin)/2 + ymin
		column := int(math.Ceil(rx / utils.ToFloat64(i.splitSize)))
		row := int(math.Ceil(ry / utils.ToFloat64(i.splitSize)))
		index := (row-1)*(i.height/i.splitSize) + column
		appendLabelToCell(
			index,
			rect{
				int(xmax) % i.splitSize,
				int(xmin) % i.splitSize,
				int(ymax) % i.splitSize,
				int(ymin) % i.splitSize,
				value.Category,
			},
			i)
	}
}
func appendLabelToCell(index int, r rect, i *imageObject) {
	i.cellGroup.cellMembers[index].rects = append(i.cellGroup.cellMembers[index].rects, r)
}
func SlitImage(imageOutPath string) {
	for _, lab := range sql.Labels {
		df := utils.TransformFile(lab.Image_path)
		fmt.Println(df.Name + "." + df.Suffix)
		utils.DownloadIns.GDownload(df)
		imageObj := NewImage(lab, 500, df)
		imageObj.initCellGroup()
		imageObj.dispatchLabel()
		break
	}
}
