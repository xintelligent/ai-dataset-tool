package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var pid string
var fileName string

var makeClass = &cobra.Command{
	Use: "makeClass",
	Run: func(c *cobra.Command, args []string) {
		getClassesFromMysql(pid)
		if len(classes) == 0 {
			panic("没有数据")
		}
		var cl []string
		for _, v := range classes {
			cl = append(cl, v.Name)
		}
		classString := strings.Replace(strings.Trim(fmt.Sprintf("%q", cl), "[]"), " ", ",", -1)
		if classFile, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777); err == nil {
			WriteFile(classString, classFile)
		}
	},
}

func init() {
	makeClass.Flags().StringVarP(&pid, "pid", "p", "", "项目Id")
	makeClass.Flags().StringVarP(&fileName, "fileName", "f", "", "存储的文件名")
	makeClass.MarkFlagRequired("pid")
	makeClass.MarkFlagRequired("fileName")
	RootCmd.AddCommand(makeClass)
}
