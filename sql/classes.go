package sql

import (
	"encoding/json"
	"fmt"
	"github.com/aliyun/aliyun-tablestore-go-sdk/tablestore"
	"github.com/spf13/viper"
	"strconv"
)

type ClassName struct {
	Index int
	Name  string
}

func NewClassesFromMysql() (err error) {
	getRowRequest := new(tablestore.GetRowRequest)
	criteria := new(tablestore.SingleRowQueryCriteria)
	putPk := new(tablestore.PrimaryKey)
	putPk.AddPrimaryKeyColumn("id", viper.GetString("ali-table-store.projectId"))

	criteria.PrimaryKey = putPk
	getRowRequest.SingleRowQueryCriteria = criteria
	getRowRequest.SingleRowQueryCriteria.TableName = viper.GetString("ali-table-store.projectTableName")
	getRowRequest.SingleRowQueryCriteria.MaxVersion = 1
	getResp, err := DB.GetRow(getRowRequest)
	if err != nil {
		fmt.Println("getrow failed with error:", err)
		return err
	} else {
		type class struct {
			Index string `json:"index"`
			Name  string `json:"name"`
		}
		var c []class
		if err := json.Unmarshal([]byte(getResp.Columns[0].Value.(string)), &c); err != nil {
			return err
		}
		for _, v := range c {
			index, err := strconv.Atoi(v.Index)
			if err != nil {
				return err
			}
			Classes = append(Classes, ClassName{
				Index: index,
				Name:  v.Name,
			})
		}
	}
	return nil
}
