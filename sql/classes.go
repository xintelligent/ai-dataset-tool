package sql

import (
	"fmt"
	"github.com/spf13/viper"
)

// 标签名称目录
var Classes []ClassName

type ClassName struct {
	Id    int    `json:"id"`
	Index int    `json:"index"`
	Name  string `json:"name"`
}

func GetClassesFromMysql() {
	err := Db.Select(&Classes, viper.GetString("dataset.classSql"))
	if err != nil {
		panic(fmt.Errorf("mysql get classes err: %s \n", err))
	}
}
