package sql

import (
	"fmt"
	"github.com/aliyun/aliyun-tablestore-go-sdk/tablestore"
	"github.com/spf13/viper"
	"math"
	"strconv"
	"strings"
)

/**
Label 是一个图片和对应标记的对象
*/
type Label struct {
	ProjectId    string
	ImagePath    string
	LabelPolygon []Polygon
	ImageWidth   int
	ImageHeight  int
}

type Polygon struct {
	Id        string
	Point     [][]float64
	Category  string
	LabelType int
}

func NewLabelsFromMysql() (err error) {
	// 先查询所有该项目下的图片
	getRangeRequest := &tablestore.GetRangeRequest{}
	rangeRowQueryCriteria := &tablestore.RangeRowQueryCriteria{}
	rangeRowQueryCriteria.TableName = viper.GetString("ali-table-store.imageTableName")

	startPK := new(tablestore.PrimaryKey)
	startPK.AddPrimaryKeyColumn("project_id", viper.GetString("ali-table-store.projectId"))
	startPK.AddPrimaryKeyColumnWithMinValue("is_labeled")
	startPK.AddPrimaryKeyColumnWithMinValue("id")
	endPK := new(tablestore.PrimaryKey)
	endPK.AddPrimaryKeyColumn("project_id", viper.GetString("ali-table-store.projectId"))
	endPK.AddPrimaryKeyColumnWithMaxValue("is_labeled")
	endPK.AddPrimaryKeyColumnWithMaxValue("id")

	rangeRowQueryCriteria.StartPrimaryKey = startPK
	rangeRowQueryCriteria.EndPrimaryKey = endPK
	rangeRowQueryCriteria.Direction = tablestore.FORWARD
	rangeRowQueryCriteria.MaxVersion = 1
	//rangeRowQueryCriteria.Limit = 10
	getRangeRequest.RangeRowQueryCriteria = rangeRowQueryCriteria
	getRangeResp, err := DB.GetRange(getRangeRequest)

	for {
		if err != nil {
			fmt.Println("get range failed with error:", err)
		}
		if len(getRangeResp.Rows) > 0 {
			// 遍历读取数据内容，配合阿里oss获取图片信息
			for _, row := range getRangeResp.Rows {
				//fmt.Println("图片数据: ", row)
				//fmt.Println(row.PrimaryKey.PrimaryKeys[2].Value)
				//fmt.Println(row.Columns[0].Value)
				imageRows := getLabelByImageId(row.PrimaryKey.PrimaryKeys[2].Value.(string))
				// 将数据填充到变量，准备接受后续处理
				Labels = append(Labels, Label{
					ProjectId:    row.Columns[0].Value.(string),
					ImagePath:    row.Columns[0].Value.(string),
					LabelPolygon: imageRows,
					ImageHeight:  0,
					ImageWidth:   0,
				})
			}
			if getRangeResp.NextStartPrimaryKey == nil {
				break
			} else {
				getRangeRequest.RangeRowQueryCriteria.StartPrimaryKey = getRangeResp.NextStartPrimaryKey
				getRangeResp, err = DB.GetRange(getRangeRequest)
			}
		} else {
			break
		}
	}
	return
}
func getLabelByImageId(imageId string) (imageRows []Polygon) {
	// 根据图片的id去查找对应的标签数据
	getRangeRequest := &tablestore.GetRangeRequest{}
	rangeRowQueryCriteria := &tablestore.RangeRowQueryCriteria{}
	rangeRowQueryCriteria.TableName = viper.GetString("ali-table-store.labelTableName")

	startPK := new(tablestore.PrimaryKey)
	startPK.AddPrimaryKeyColumn("image_id", imageId)
	startPK.AddPrimaryKeyColumnWithMinValue("id")
	endPK := new(tablestore.PrimaryKey)
	endPK.AddPrimaryKeyColumn("image_id", imageId)
	endPK.AddPrimaryKeyColumnWithMaxValue("id")

	rangeRowQueryCriteria.StartPrimaryKey = startPK
	rangeRowQueryCriteria.EndPrimaryKey = endPK
	rangeRowQueryCriteria.Direction = tablestore.FORWARD
	rangeRowQueryCriteria.MaxVersion = 1
	//rangeRowQueryCriteria.Limit = 10
	getRangeRequest.RangeRowQueryCriteria = rangeRowQueryCriteria
	getRangeResp, err := DB.GetRange(getRangeRequest)

	for {
		if err != nil {
			fmt.Println("get range failed with error:", err)
		}
		if len(getRangeResp.Rows) > 0 {
			for _, row := range getRangeResp.Rows {
				imageRows = append(imageRows, Polygon{
					Id:        row.PrimaryKey.PrimaryKeys[0].Value.(string),
					Category:  row.Columns[0].Value.(string),
					Point:     polygonDataToSlice(row.Columns[2].Value.(string)),
					LabelType: int(row.Columns[4].Value.(int64)),
				})
			}
			if getRangeResp.NextStartPrimaryKey == nil {
				break
			} else {
				getRangeRequest.RangeRowQueryCriteria.StartPrimaryKey = getRangeResp.NextStartPrimaryKey
				getRangeResp, err = DB.GetRange(getRangeRequest)
			}
		} else {
			break
		}
	}
	return
}
func polygonDataToSlice(data string) (result [][]float64) {
	coordinateSlice := strings.Split(strings.TrimRight(strings.TrimLeft(data, "["), "]"), "],[")
	for _, v := range coordinateSlice {
		coordinatePoint := strings.Split(v, ",")
		point := make([]float64, 2)
		point[0] = ToFloat64(coordinatePoint[0])
		point[1] = ToFloat64(coordinatePoint[1])
		result = append(result, point)
	}
	return
}

func ToFloat64(unk interface{}) float64 {
	switch i := unk.(type) {
	case float64:
		return i
	case float32:
		return float64(i)
	case int64:
		return float64(i)
	case int:
		return float64(i)
	case string:
		result, _ := strconv.ParseFloat(i, 64)
		return result
	default:
		return math.NaN()
	}
}
