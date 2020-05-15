package cmd

import (
	"ai-dataset-tool/log"
	"ai-dataset-tool/sql"
	"ai-dataset-tool/transform/coco"
	"ai-dataset-tool/transform/csv"
	"ai-dataset-tool/transform/voc"
	"ai-dataset-tool/utils"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var format string

var translateCmd = &cobra.Command{
	Use: "translate",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("缺少项目id")
			log.Klog.Println("缺少项目id")
			os.Exit(1)
		}

		sql.GetClassesFromMysql(args[0])
		sql.GetLabelsFromMysql(args[0])
		os.MkdirAll(utils.DownloadIns.AnnotationOutPath, 0777)
		os.MkdirAll(utils.DownloadIns.ImageOutPath, 0777)
		if utils.DownloadIns.NeedDownloadImageFile {
			utils.InitDownload()
			defer close(utils.DownloadIns.Goroutine_cnt)
		}
		utils.InitOss()
		baseOnFormat()
	},
}

func init() {
	translateCmd.Flags().StringVarP(&format, "Format", "f", "csv", "Format(csv or coco or voc)")
	//translateCmd.Flags().StringVarP(&utils.DownloadIns.AnnotationOutPath, "AnnotationOutPath", "a", "./data", "label file out path")
	//translateCmd.Flags().StringVarP(&utils.DownloadIns.ImageOutPath, "imageOutPath", "i", "./images", "images file out path")
	//translateCmd.Flags().BoolVarP(&utils.DownloadIns.NeedDownloadImageFile, "needDownloadImageFile", "n", false, "need download imageTool file")
}
func baseOnFormat() {
	switch format {
	case "csv":
		csv.WriteCsvClassFile()
		csv.WriteCsvLabelsFile()
	case "coco":
		coco.WriteCocoFile()
	case "voc":
		voc.WriteVocLabelsFile()
	}
}
func init() {
	RootCmd.AddCommand(translateCmd)
}
