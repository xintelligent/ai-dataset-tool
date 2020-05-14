package cmd

import (
	"ai-dataset-tool/utils"
	"github.com/spf13/cobra"
)

var transferCmd = &cobra.Command{
	Use:  "transfer",
	Long: "传输服务器数据到本地",
	Run: func(c *cobra.Command, args []string) {

	},
}

func init() {
	transferCmd.Flags().StringVarP(&format, "Format", "f", "csv", "Format(csv or coco or voc)")
	transferCmd.Flags().StringVarP(&utils.DownloadIns.AnnotationOutPath, "AnnotationOutPath", "a", "./data", "label file out path")
	transferCmd.Flags().StringVarP(&utils.DownloadIns.ImageOutPath, "imageOutPath", "i", "./images", "images file out path")
	transferCmd.Flags().BoolVarP(&utils.DownloadIns.NeedDownloadImageFile, "needDownloadImageFile", "n", false, "need download image file")
}
