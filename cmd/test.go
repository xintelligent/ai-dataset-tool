package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"os"
)

type shape struct {
	x0 int
	y0 int
	x1 int
	y1 int
}

var testCmd = &cobra.Command{
	Use: "test",
	Run: func(cmd *cobra.Command, args []string) {
		reader, err := os.Open("./333.jpg")
		if err != nil {
			fmt.Println("图片没识别")
			log.Fatal(err)
		}
		defer reader.Close()
		clipErr := clip(reader, "new.jpg", shape{1000, 0, 1500, 500}, 100)
		if clipErr != nil {
			fmt.Println("没好用")
		}
	},
}

func init() {
	RootCmd.AddCommand(testCmd)
}
func saveImage(filename string, img image.Image) {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatalf("os.Create failed: %v", err)
	}
	defer f.Close()
	op := jpeg.Options{100}
	err = jpeg.Encode(f, img, &op)
	if err != nil {
		log.Fatalf("png.Encode failed: %v", err)
	}
}

/*
* 图片裁剪
* 入参:
* 规则:如果精度为0则精度保持不变
*
* 返回:error
 */
func clip(in io.Reader, outFileName string, shape shape, quality int) error {
	origin, fm, err := image.Decode(in)
	if err != nil {
		return err
	}
	outFile, err := os.Create(outFileName)
	if err != nil {
		log.Fatalf("os.Create failed: %v", err)
	}
	defer outFile.Close()

	re := image.Rect(shape.x0, shape.y0, shape.x1, shape.y1)
	switch fm {
	case "jpeg":
		img := origin.(*image.YCbCr)
		subImg := img.SubImage(re).(*image.YCbCr)
		return jpeg.Encode(outFile, subImg, &jpeg.Options{quality})
	case "png":
		switch origin.(type) {
		case *image.NRGBA:
			img := origin.(*image.NRGBA)
			subImg := img.SubImage(re).(*image.NRGBA)
			return png.Encode(outFile, subImg)
		case *image.RGBA:
			img := origin.(*image.RGBA)
			subImg := img.SubImage(re).(*image.RGBA)
			return png.Encode(outFile, subImg)
		}
	default:
		return errors.New("ERROR FORMAT")
	}
	return nil
}
