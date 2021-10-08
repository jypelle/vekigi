package srv

import (
	"github.com/jypelle/vekigi/internal/images"
	"github.com/jypelle/vekigi/internal/srv/event"
	"github.com/sirupsen/logrus"
	"image"
	"image/color"
	"image/draw"
	"time"
)

func (s *ServerApp) refreshDisplay(resetMode bool) {
	// Clear animation tick timer
	if s.animationTickTimer != nil {
		s.animationTickTimer.Stop()
		s.animationTickTimer = nil
	}

	if resetMode {
		s.animationTickCount = 0

		s.currentPopUp = NO_POPUP
		if s.popUpHideTimer != nil {
			s.popUpHideTimer.Stop()
			s.popUpHideTimer = nil
		}
	}

	var imgToDisplay image.Image

	if s.currentPopUp != NO_POPUP {
		switch s.currentPopUp {
		case VOLUME_POPUP:
			img := image.NewRGBA(image.Rect(0, 0, 128, 64))
			AddCenteredLabel(img, 20, "Volume")
			draw.Draw(img, image.Rect(13, 38, 115, 39), &image.Uniform{color.RGBA{255, 255, 255, 255}}, image.ZP, draw.Src)
			draw.Draw(img, image.Rect(13, 49, 115, 50), &image.Uniform{color.RGBA{255, 255, 255, 255}}, image.ZP, draw.Src)
			draw.Draw(img, image.Rect(12, 39, 13, 49), &image.Uniform{color.RGBA{255, 255, 255, 255}}, image.ZP, draw.Src)
			draw.Draw(img, image.Rect(115, 39, 116, 49), &image.Uniform{color.RGBA{255, 255, 255, 255}}, image.ZP, draw.Src)
			draw.Draw(img, image.Rect(14, 40, 14+int(s.Volume()), 48), &image.Uniform{color.RGBA{255, 255, 255, 255}}, image.ZP, draw.Src)
			imgToDisplay = img
		case SNOOZE_OFF_POPUP:
			img := image.NewRGBA(image.Rect(0, 0, 128, 64))
			AddCenteredLabel(img, 40, "Snooze off")
			imgToDisplay = img
		}
	} else {
		switch s.currentMode {
		case UNDEFINED_MODE:
			imgToDisplay = images.IntroImage
		case CLOCK_MODE:
			imgToDisplay = s.refreshClockDisplay()
		case ALARM_SETTING_MODE:
			imgToDisplay = s.refreshAlarmSettingsDisplay()
		case END_MODE:
			img := image.NewRGBA(image.Rect(0, 0, 128, 64))
			AddCenteredLabel(img, 40, "See you!")
			imgToDisplay = img
		}
	}
	s.displayDevice.ShowImage(imgToDisplay)
}

func (s *ServerApp) refreshClockDisplay() image.Image {
	logrus.Debugf("Display clock")
	currentWebradio := s.webradioPlayerDevice.CurrentWebRadio()
	currentPlaylist := s.playlistPlayerDevice.CurrentPlaylist()
	now := time.Now()

	img := image.NewRGBA(image.Rect(0, 0, 128, 64))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{0, 0, 0, 255}}, image.ZP, draw.Src)

	AddNumber(img, image.Pt(4, 14), int64(now.Hour())/10)
	AddNumber(img, image.Pt(4+1*24, 14), int64(now.Hour())%10)
	AddNumber(img, image.Pt(4+2*24, 14), 10)
	AddNumber(img, image.Pt(4+3*24, 14), int64(now.Minute())/10)
	AddNumber(img, image.Pt(4+4*24, 14), int64(now.Minute())%10)

	var name string
	if currentWebradio != nil {
		name = currentWebradio.Name
	} else if currentPlaylist != nil {
		name = currentPlaylist.Name + ":" + s.playlistPlayerDevice.CurrentSongName()
	}
	if len(name)*6-128 > 0 {
		deltaX := s.animationTickCount % (len(name)*6 + 20)
		AddLabel(img, 10-deltaX, 62, name)
		AddLabel(img, len(name)*6+20+10-deltaX, 62, name)
	} else {
		AddLabel(img, 0, 62, name)
	}

	if s.Alarm().Enabled {
		draw.Draw(
			img,
			images.AlarmImage.Bounds().Add(image.Pt(img.Bounds().Dx()-images.AlarmImage.Bounds().Dx(), 0)),
			images.AlarmImage,
			images.AlarmImage.Bounds().Min,
			draw.Src)
	}
	if s.clockDevice.IsAlarmRunning() {
		draw.Draw(
			img,
			images.SnoozeImage.Bounds().Add(image.Pt(img.Bounds().Dx()-images.AlarmImage.Bounds().Dx()-images.SnoozeImage.Bounds().Dx(), 0)),
			images.SnoozeImage,
			images.SnoozeImage.Bounds().Min,
			draw.Src)
	}

	if len(name)*6-128 > 0 {
		s.animationTickTimer = time.AfterFunc(100*time.Millisecond, func() {
			s.internalEventChannel <- event.InternalEvent{Data: event.InternalEventAnimationTickData{}}
		})
	}
	return img
}

func (s *ServerApp) refreshAlarmSettingsDisplay() image.Image {
	logrus.Debugf("Display alarm settings")

	img := image.NewRGBA(image.Rect(0, 0, 128, 64))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{0, 0, 0, 255}}, image.ZP, draw.Src)

	AddCenteredLabel(img, 9, "Alarm settings")
	alarmTime := s.Alarm()
	AddNumber(img, image.Pt(4, 14), alarmTime.Hour/10)
	AddNumber(img, image.Pt(4+1*24, 14), alarmTime.Hour%10)
	AddNumber(img, image.Pt(4+2*24, 14), 10)
	AddNumber(img, image.Pt(4+3*24, 14), alarmTime.Minute/10)
	AddNumber(img, image.Pt(4+4*24, 14), alarmTime.Minute%10)

	var name string
	if alarmTime.WebradioId != nil && s.webradioPlayerDevice.Webradio(*alarmTime.WebradioId) != nil {
		name = s.WebradioGroups[alarmTime.WebradioId.GroupId][alarmTime.WebradioId.IndexId-1].Name
	} else if alarmTime.PlaylistId != nil {
		playlist := s.playlistPlayerDevice.GetPlaylist(*alarmTime.PlaylistId)
		if playlist != nil {
			name = playlist.Name
		}
	}

	if len(name)*6-128 > 0 {
		deltaX := s.animationTickCount % (len(name)*6 + 20)
		AddLabel(img, 10-deltaX, 62, name)
		AddLabel(img, len(name)*6+20+10-deltaX, 62, name)
		s.animationTickTimer = time.AfterFunc(100*time.Millisecond, func() {
			s.internalEventChannel <- event.InternalEvent{Data: event.InternalEventAnimationTickData{}}
		})
	} else {
		AddLabel(img, 0, 62, name)
	}

	return img
}
