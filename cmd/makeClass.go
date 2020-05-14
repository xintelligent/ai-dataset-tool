package cmd

import (
	"ai-dataset-tool/sql"
	"github.com/spf13/cobra"
)

var pid string
var fileName string

var makeClassCmd = &cobra.Command{
	Use: "makeClass",
	Run: func(c *cobra.Command, args []string) {
		sql.GetClassesFromMysql(pid)
		//if len(sql.Classes) == 0 {
		//	panic("没有数据")
		//}
		//var cl []string
		//for _, v := range sql.Classes {
		//	cl = append(cl, v.Name)
		//}
		//classString := strings.Replace(strings.Trim(fmt.Sprintf("%q", cl), "[]"), " ", ",", -1)
		//if classFile, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777); err == nil {
		//	utils.WriteFile(classString, classFile)
		//}
	},
}

func init() {
	makeClassCmd.Flags().StringVarP(&pid, "pid", "p", "", "项目Id")
	makeClassCmd.Flags().StringVarP(&fileName, "fileName", "f", "", "存储的文件名")
	makeClassCmd.MarkFlagRequired("pid")
	makeClassCmd.MarkFlagRequired("fileName")
	RootCmd.AddCommand(makeClassCmd)
}
