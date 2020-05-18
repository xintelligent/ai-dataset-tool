package cmd

import (
	"math"
)

// 标签列表，从标签集合对象复制，
// 按照裁剪目标，初始化原图像，后分割图像
// 计算图片交集
// 保留标签的比例
var ratio = 0.25
var label []labelRect

type labelRect struct {
	x0       int
	y0       int
	x1       int
	y1       int
	area     int
	Category int
}

type labelGroup struct {
	labelMembers []labelMember
}

type labelMember struct {
	splitIndex int
	labelRect
}

// 是一个面积大于真实图片的尺寸
type imageObject struct {
	width  int
	height int
}

func NewImage(width int, height int, splitSize int) (i *imageObject) {
	i.width = int(math.Ceil(float64(width)/float64(splitSize))) * splitSize
	i.height = int(math.Ceil(float64(height)/float64(splitSize))) * splitSize
	return
}

func (lg *labelGroup) set(ss int) {

}
