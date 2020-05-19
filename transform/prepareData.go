package transform

var LabelsData = Labels{}

type Labels struct {
	labSlice []Lab
}
type Lab struct {
	Image_path   string // oss 存储路径加文件名
	ImageOutPath string // 本地存储路径
	Name         string // 本地文件名，不包含后缀
	Suffix       string // 文件后缀
	Data
}
type Data struct {
	Label       []Shape
	ImageWidth  int
	ImageHeight int
}
type Shape struct {
	Xmax     float64
	Xmin     float64
	Ymax     float64
	Ymin     float64
	Category string
}

func (ls *Labels) Push(l Lab) {
	ls.labSlice = append(ls.labSlice, l)
}
