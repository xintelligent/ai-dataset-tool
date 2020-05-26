package cmd

import (
	"ai-dataset-tool/transform/coco"
	"ai-dataset-tool/transform/csv"
	"ai-dataset-tool/transform/voc"
	"ai-dataset-tool/transform/yolo"
	"ai-dataset-tool/utils"
	"fmt"
	"github.com/spf13/cobra"
)

var format string

var transformCmd = &cobra.Command{
	Use: "transform",
	Run: func(cmd *cobra.Command, args []string) {
		prepare()
		baseOnFormat()
	},
}

func init() {
	transformCmd.Flags().StringVarP(&format, "Format", "f", "csv", "Format(csv or coco or voc or yolo)")
	RootCmd.AddCommand(transformCmd)
}
func baseOnFormat() {
	switch format {
	case "csv":
		csv.WriteCsvClassFile(utils.AnnotationOutPath)
		csv.WriteCsvLabelsFile(utils.AnnotationOutPath)
	case "coco":
		coco.WriteCocoFile(utils.AnnotationOutPath)
	case "voc":
		voc.WriteVocLabelsFile(utils.AnnotationOutPath, utils.ImageOutPath)
	case "yolo":
		fmt.Println("yolo")
		yolo.WriteYoloClassFile(utils.AnnotationOutPath)
		yolo.WriteYoloLabelsFile(utils.AnnotationOutPath, utils.ImageOutPath)
	}
}
