package imageTool

import (
	"image"
	"image/draw"
)

// 图像拆分

type Crop struct {
	rect image.Rectangle
}

func (p *Crop) Bounds(srcBounds image.Rectangle) (dstBounds image.Rectangle) {
	b := srcBounds.Intersect(p.rect)
	return b.Sub(b.Min)
}

func (p *Crop) Draw(dst draw.Image, src image.Image, options *Options) {
	srcb := src.Bounds().Intersect(p.rect)
	dstb := dst.Bounds()
	pixGetter := newPixelGetter(src)
	pixSetter := newPixelSetter(dst)

	parallelize(options.Parallelization, srcb.Min.Y, srcb.Max.Y, func(start, stop int) {
		for srcy := start; srcy < stop; srcy++ {
			for srcx := srcb.Min.X; srcx < srcb.Max.X; srcx++ {
				dstx := dstb.Min.X + srcx - srcb.Min.X
				dsty := dstb.Min.Y + srcy - srcb.Min.Y
				pixSetter.setPixel(dstx, dsty, pixGetter.getPixel(srcx, srcy))
			}
		}
	})
}
func TDraw(p image.Rectangle, dst draw.Image, src image.Image, options *Options) {
	srcb := src.Bounds().Intersect(p)
	dstb := dst.Bounds()
	pixGetter := newPixelGetter(src)
	pixSetter := newPixelSetter(dst)

	parallelize(options.Parallelization, srcb.Min.Y, srcb.Max.Y, func(start, stop int) {
		for srcy := start; srcy < stop; srcy++ {
			for srcx := srcb.Min.X; srcx < srcb.Max.X; srcx++ {
				dstx := dstb.Min.X + srcx - srcb.Min.X
				dsty := dstb.Min.Y + srcy - srcb.Min.Y
				pixSetter.setPixel(dstx, dsty, pixGetter.getPixel(srcx, srcy))
			}
		}
	})
}
