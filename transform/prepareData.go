package transform

// 如果拆分图片，变量会被拆分结果覆盖
var PreLabelsData = Labels{}

type Labels struct {
	LabSlice []Label
}
type Label struct {
	Id           int
	Image_path   string // oss 存储路径加文件名
	ImageOutPath string // 本地存储路径
	Name         string // 本地文件名，不包含后缀
	Suffix       string // 文件后缀
	ImageWidth   int
	ImageHeight  int
	Rects        []Rect
}

type Rect struct {
	Index    int
	Xmax     float64
	Xmin     float64
	Ymax     float64
	Ymin     float64
	Category string
}

func (ls *Labels) Push(l Label) {
	ls.LabSlice = append(ls.LabSlice, l)
}

var PreCategoriesData = []Category{}

type Category struct {
	Id    int
	Index int
	Name  string
}
