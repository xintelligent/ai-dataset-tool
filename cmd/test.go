package cmd

import (
	"ai-dataset-tool/imageTool"
	"github.com/spf13/cobra"
	"image"
	"log"
	"os"
)

var testCmd = &cobra.Command{
	Use: "test",
	Run: func(cmd *cobra.Command, args []string) {
		reader, err := os.Open("111.jpg")
		if err != nil {
			log.Fatal(err)
		}
		defer reader.Close()
		m, _, err := image.Decode(reader)
		if err != nil {
			log.Fatal(err)
		}
		b := m.Bounds()
		dst := image.NewRGBA(b)
		op := imageTool.Options{}
		re := image.Rect(100, 100, 200, 200)
		// 3. Use the Draw func to apply the filters to src and store the result in dst.
		imageTool.TDraw(re, dst, m, &op)
	},
}

func init() {
	RootCmd.AddCommand(testCmd)
}
