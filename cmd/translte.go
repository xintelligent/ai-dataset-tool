package cmd

import (
	"ai-dataset-tool/sql"
	"ai-dataset-tool/transform"
	"ai-dataset-tool/transform/coco"
	"ai-dataset-tool/transform/csv"
	"ai-dataset-tool/transform/voc"
	"ai-dataset-tool/utils"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var format string

var translateCmd = &cobra.Command{
	Use: "translate",
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Println(sql.Labels[0].Data)
		os.MkdirAll(utils.AnnotationOutPath, 0777)
		os.MkdirAll(utils.ImageOutPath, 0777)
		utils.InitDownload()
		defer close(utils.DownloadIns.Goroutine_cnt)
		utils.InitOss()
		baseOnFormat()
	},
}

func init() {
	translateCmd.Flags().StringVarP(&format, "Format", "f", "csv", "Format(csv or coco or voc)")

	RootCmd.AddCommand(translateCmd)
}
func baseOnFormat() {
	switch format {
	case "csv":
		csv.WriteCsvClassFile(utils.AnnotationOutPath)
		csv.WriteCsvLabelsFile(utils.AnnotationOutPath, utils.NeedDownloadImageFile)
	case "coco":
		coco.WriteCocoFile(utils.AnnotationOutPath, utils.NeedDownloadImageFile)
	case "voc":
		voc.WriteVocLabelsFile(utils.AnnotationOutPath, utils.ImageOutPath)
	case "split":
		transform.Test(utils.AnnotationOutPath, utils.ImageOutPath)
	}
}
