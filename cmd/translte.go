package cmd

import (
	"ai-dataset-tool/sql"
	"ai-dataset-tool/transform/coco"
	"ai-dataset-tool/transform/csv"
	"ai-dataset-tool/transform/voc"
	"ai-dataset-tool/utils"
	"os"

	"github.com/spf13/cobra"
)

var format string

var translateCmd = &cobra.Command{
	Use: "translate",
	Run: func(cmd *cobra.Command, args []string) {
		sql.GetClassesFromMysql()
		sql.GetLabelsFromMysql()
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
	translateCmd.Flags().StringVarP(&utils.AnnotationOutPath, "AnnotationOutPath", "a", "./data", "label file out path")
	translateCmd.Flags().StringVarP(&utils.ImageOutPath, "imageOutPath", "i", "./images", "images file out path")
	translateCmd.Flags().BoolVarP(&utils.NeedDownloadImageFile, "needDownloadImageFile", "n", false, "need download imageTool file")
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
	}
}
