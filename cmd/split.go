package cmd

import (
	"ai-dataset-tool/log"
	"ai-dataset-tool/sql"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var splitCmd = &cobra.Command{
	Use: "split",
	Run: splitCmdF,
}

func splitCmdF(c *cobra.Command, args []string) {
	if len(args) != 1 {
		fmt.Println("缺少项目id")
		log.Klog.Println("缺少项目id")
		os.Exit(1)
	}

	sql.GetClassesFromMysql(args[0])
	sql.GetLabelsFromMysql(args[0])
	// 简单抽取一个图片做测试
	var data sql.Data
	dataErr := json.Unmarshal([]byte(sql.Labels[0].Data), &data)
	i := NewImage(data.ImageWidth, data.ImageHeight, 500)
}
