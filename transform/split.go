package transform

import (
	"ai-dataset-tool/imageTool"
	"ai-dataset-tool/log"
	"ai-dataset-tool/utils"
	"github.com/spf13/viper"
	"image"
	"image/jpeg"
	"math"
	"os"
)

var splitLabelsData = Labels{}

// 图片分割
type cellGroup struct {
	cellMembers []cellMember
}

type cellMember struct {
	row          int
	column       int
	x0           int
	y0           int
	x1           int
	y1           int
	ImageOutPath string
	ImageName    string
	Suffix       string
	rects        []Rect
	ImageWidth   int
	ImageHeight  int
}

// 是一个面积大于真实图片的尺寸
type imageObject struct {
	width     int
	height    int
	splitSize int
	cellGroup
	imageName string
	rects     []Rect
}

var imageFile *os.File

func NewImage(l Label, splitSize int) (i imageObject) {
	i.width = int(math.Ceil(float64(l.ImageWidth)/float64(splitSize))) * splitSize
	i.height = int(math.Ceil(float64(l.ImageHeight)/float64(splitSize))) * splitSize
	i.splitSize = splitSize
	i.imageName = l.Name + "." + l.Suffix
	i.rects = l.Rects
	return
}

func (i *imageObject) createCellGroup(imageOutPath string) {

	rowCount := i.height / i.splitSize
	columnCount := i.width / i.splitSize
	for row := 0; row < rowCount; row++ {
		for column := 0; column < columnCount; column++ {
			newImageName := utils.RandStringBytesMaskImprSrcUnsafe(16) + "-" + utils.PxToString(row) + "-" + utils.PxToString(column)
			c := cellMember{
				column:       column,
				row:          row,
				x0:           column * i.splitSize,
				y0:           row * i.splitSize,
				x1:           (column + 1) * i.splitSize,
				y1:           (row + 1) * i.splitSize,
				ImageOutPath: imageOutPath,
				Suffix:       ".jpg",
				ImageName:    newImageName,
				ImageWidth:   i.splitSize,
				ImageHeight:  i.splitSize,
			}
			i.cellGroup.cellMembers = append(i.cellGroup.cellMembers, c)
		}
	}
}

func cropCell(o *image.Image, c *cellMember) {
	r := image.Rect(c.x0, c.y0, c.x1, c.y1)
	crop := imageTool.NewCrop(r)
	dst := image.NewRGBA(crop.Bounds(r))
	crop.Draw(dst, *o, &imageTool.Options{Parallelization: true})
	saveImage(c.ImageName+".jpg", dst)
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
	for _, value := range i.rects {
		rx := (value.Xmax-value.Xmin)/2 + value.Xmin
		ry := (value.Ymax-value.Ymin)/2 + value.Ymin
		column := int(math.Ceil(rx / utils.ToFloat64(i.splitSize)))
		row := int(math.Ceil(ry / utils.ToFloat64(i.splitSize)))
		index := (row-1)*(i.width/i.splitSize) + column - 1
		appendLabelToCell(
			index,
			Rect{
				utils.ToFloat64(int(value.Xmax) % i.splitSize),
				utils.ToFloat64(int(value.Xmin) % i.splitSize),
				utils.ToFloat64(int(value.Ymax) % i.splitSize),
				utils.ToFloat64(int(value.Ymin) % i.splitSize),
				value.Category,
			},
			i)
	}
}

func appendLabelToCell(index int, r Rect, i *imageObject) {
	i.cellGroup.cellMembers[index].rects = append(i.cellGroup.cellMembers[index].rects, r)
}

func SlitImage(imageOutPath string) {
	// 遍历进行拆分

	for _, lab := range PreLabelsData.LabSlice {
		// 初始化
		imageObj := NewImage(lab, 500)
		// 创建 500*500 的subImage
		imageObj.createCellGroup(imageOutPath)
		// 分配原始标签到每个cellMember
		imageObj.dispatchLabel()

		var imageErr error
		if imageFile, imageErr = os.OpenFile(imageOutPath+"/"+lab.Name+"."+lab.Suffix, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777); imageErr != nil {
			log.Klog.Println("没有能打开图片文件", imageErr)
			os.Exit(1)
		}
		origin, _, err := image.Decode(imageFile)
		if err != nil {
			log.Klog.Println("没有能解码图片文件", err)
		}

		for _, cm := range imageObj.cellMembers {
			log.Klog.Println(cm)
			if len(cm.rects) == 0 {
				// cellMember 中没有标签的情况下，就排除
				continue
			}
			cropCell(&origin, &cm)

			splitLabelsData.LabSlice = append(splitLabelsData.LabSlice, Label{
				"",
				cm.ImageOutPath,
				cm.ImageName,
				cm.Suffix,
				cm.ImageWidth,
				cm.ImageHeight,
				cm.rects,
			})
		}
	}
	PreLabelsData = splitLabelsData
}
