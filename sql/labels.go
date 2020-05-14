package sql

import (
	"fmt"
	"github.com/spf13/viper"
)

// 标签数据
var Labels []Lab

type Lab struct {
	Id         int    `json:"id"`
	Project_id int    `json:"project_id"`
	Image_path string `json:"image_path"`
	Data       string `json:"data"`
	User_id    int    `json:"user_id"`
}
type Data struct {
	Label       []Shape `json:"label"`
	ImageWidth  int     `json:"imageWidth"`
	ImageHeight int     `json:"imageHeight"`
}
type Shape struct {
	Xmax     interface{} `json:"xmax"`
	Xmin     interface{} `json:"xmin"`
	Ymax     interface{} `json:"ymax"`
	Ymin     interface{} `json:"ymin"`
	Category string      `json:"category"`
}

func GetLabelsFromMysql(pid string) {
	err := Db.Select(&Labels, viper.GetString("dataset.labelSql"))
	//err := cmd.Db.Select(&labels, "SELECT `id`, `project_id`, `image_path`, `data`, `user_id` FROM labels WHERE image_path regexp '/(12|13|15|19|21|23|33|34|38|39|72)/' AND project_id=13")
	if err != nil {
		panic(fmt.Errorf("mysql get labels err: %s \n", err))
	}
}
