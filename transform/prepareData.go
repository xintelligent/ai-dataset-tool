package transform

import "sort"

// 如果拆分图片，变量会被拆分结果覆盖
var PreLabelsData = Labels{}

type Labels struct {
	LabSlice []Label
}
type Label struct {
	Id           int
	Index        int
	Image_path   string // oss 存储路径加文件名
	ImageOutPath string // 本地存储路径
	Name         string // 本地文件名，不包含后缀
	Suffix       string // 文件后缀
	ImageWidth   int
	ImageHeight  int
	Polygons     []Polygon
}

type Polygon struct {
	Id        string
	Index     int
	Point     [][]float64
	Category  string
	LabelType int
}

func (ls *Labels) Push(l Label) {
	ls.LabSlice = append(ls.LabSlice, l)
}

var PreCategoriesData = []Category{}

type Category struct {
	Index int
	Name  string
}

/**
边界
*/
type Bound struct {
	Xmin float64
	Xmax float64
	Ymin float64
	Ymax float64
}

func GetBound(ps [][]float64) (b Bound) {
	var x []float64
	var y []float64
	for _, v := range ps {
		x = append(x, v[0])
		y = append(y, v[1])
	}
	sort.Float64s(x)
	sort.Float64s(y)
	b.Xmin = x[0]
	b.Xmax = x[len(x)-1]
	b.Ymin = y[0]
	b.Ymax = y[len(y)-1]
	return
}
