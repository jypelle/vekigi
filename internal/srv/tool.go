package srv

import (
	"github.com/hajimehoshi/bitmapfont/v2"
	"github.com/jypelle/vekigi/internal/images"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
)

var col = color.RGBA{255, 255, 255, 255}
var uniformImage = image.NewUniform(col)

func AddLabel(img *image.RGBA, x, y int, label string) {

	point := fixed.Point26_6{fixed.Int26_6((x + 4) * 64), fixed.Int26_6(y * 64)}

	d := &font.Drawer{
		Dst: img,
		Src: uniformImage,
		//		Face: basicfont.Face7x13,
		Face: bitmapfont.Face,
		Dot:  point,
	}
	d.DrawString(label)
}

func AddCenteredLabel(img *image.RGBA, y int, label string) {
	AddLabel(img, (128-len(label)*6)/2, y, label)
}

func AddNumber(img draw.Image, position image.Point, number int64) {
	draw.Draw(
		img,
		image.Rect(0, 0, 24, 36).Add(position),
		images.NumbersImage,
		images.NumbersImage.Bounds().Min.Add(image.Pt(24*int(number), 0)),
		draw.Src)
}
