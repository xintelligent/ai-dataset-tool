package cmd

import "fmt"

// 标签名称目录
var classes []className

type className struct {
	Id    int    `json:"id"`
	Index int    `json:"index"`
	Name  string `json:"name"`
}

func getClassesFromMysql(pid string) {
	err := Db.Select(&classes, "SELECT b.id AS id, b.`index` AS `index`, b.`name` AS `name` from project_label_categories as a JOIN label_categories as b ON a.project_id="+pid+" and a.label_category_id=b.id")
	if err != nil {
		panic(fmt.Errorf("mysql get classes err: %s \n", err))
	}
}
