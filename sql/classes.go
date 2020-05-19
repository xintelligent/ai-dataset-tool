package sql

import (
	"github.com/spf13/viper"
)

type ClassName struct {
	Id    int
	Index int
	Name  string
}

func NewClassesFromMysql() (c []ClassName, err error) {
	err = Db.Select(&c, viper.GetString("dataset.classSql"))
	return
}
