package sql

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
)

type Lab struct {
	Id         int    `db:"id"`
	Project_id int    `db:"project_id"`
	Image_path string `db:"image_path"`
	Data       string `db:"data"`
	User_id    int    `db:"user_id"`
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

func NewLabelsFromMysql() (l []Lab, err error) {
	err = Db.Select(&l, viper.GetString("dataset.labelSql"))
	//err := cmd.Db.Select(&labels, "SELECT `id`, `project_id`, `image_path`, `data`, `user_id` FROM labels WHERE image_path regexp '/(12|13|15|19|21|23|33|34|38|39|72)/' AND project_id=13")
	if err != nil {
		panic(fmt.Errorf("mysql get labels err: %s \n", err))
	}
	return
}

func (l *Lab) JsonToStruct() (data Data, err error) {
	err = json.Unmarshal([]byte(l.Data), &data)
	return
}
