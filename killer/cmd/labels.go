package cmd

import "fmt"

// 标签数据
var labels []lab

type lab struct {
	Id         int    `json:"id"`
	Project_id int    `json:"project_id"`
	Image_path string `json:"image_path"`
	Data       string `json:"data"`
	User_id    int    `json:"user_id"`
}
type data struct {
	Label       []shape `json:"label"`
	ImageWidth  int     `json:"imageWidth"`
	ImageHeight int     `json:"imageHeight"`
}
type shape struct {
	Xmax     interface{} `json:"xmax"`
	Xmin     interface{} `json:"xmin"`
	Ymax     interface{} `json:"ymax"`
	Ymin     interface{} `json:"ymin"`
	Category string      `json:"category"`
}

func getLabelsFromMysql(pid string) {
	err := Db.Select(&labels, "select `id`, `project_id`, `image_path`, `data`, `user_id` from labels where project_id="+pid)
	if err != nil {
		panic(fmt.Errorf("mysql get labels err: %s \n", err))
	}
}
