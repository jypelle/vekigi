package images

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	_ "image/png"

	"github.com/jypelle/vekigi/internal/tool"
)

//go:embed intro.png
var IntroImgFile []byte

var IntroImage image.Image

//go:embed alarm.png
var AlarmImgFile []byte

var AlarmImage image.Image

//go:embed snooze.png
var SnoozeImgFile []byte

var SnoozeImage image.Image

//go:embed numbers.png
var NumbersImgFile []byte

var NumbersImage image.Image

func init() {
	// Load images
	var err error

	IntroImage, _, err = image.Decode(bytes.NewReader(IntroImgFile))
	if err != nil {
		panic(fmt.Errorf("can't load intro image: %w", err))
	}

	AlarmImage, _, err = image.Decode(bytes.NewReader(AlarmImgFile))
	if err != nil {
		tool.Fatal("Can't load alarm image", "error", err)
	}

	SnoozeImage, _, err = image.Decode(bytes.NewReader(SnoozeImgFile))
	if err != nil {
		tool.Fatal("Can't load snooze image", "error", err)
	}

	NumbersImage, _, err = image.Decode(bytes.NewReader(NumbersImgFile))
	if err != nil {
		tool.Fatal("Can't load numbers image", "error", err)
	}

}
