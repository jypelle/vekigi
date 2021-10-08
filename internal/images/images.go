package images

import (
	"bytes"
	_ "embed"
	"github.com/sirupsen/logrus"
	"image"
	_ "image/png"
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
		logrus.Panicf("Can't load intro image: %v", err)
	}

	AlarmImage, _, err = image.Decode(bytes.NewReader(AlarmImgFile))
	if err != nil {
		logrus.Fatalf("Can't load alarm image: %v", err)
	}

	SnoozeImage, _, err = image.Decode(bytes.NewReader(SnoozeImgFile))
	if err != nil {
		logrus.Fatalf("Can't load snooze image: %v", err)
	}

	NumbersImage, _, err = image.Decode(bytes.NewReader(NumbersImgFile))
	if err != nil {
		logrus.Fatalf("Can't load numbers image: %v", err)
	}

}
