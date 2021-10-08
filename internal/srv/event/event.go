package event

import (
	"github.com/jypelle/vekigi/apimodel"
	"net/http"
)

// PopUp
type InternalEvent struct {
	Data interface{}
}

type InternalEventPopupHideData struct{}
type InternalEventAnimationTickData struct{}

// Ticker
type TickerEvent struct {
	Data interface{}
}

type TickerEventTickData struct{}
type TickerEventAlarmData struct{}

// Webradio
type WebradioEvent struct {
	ResponseWriter http.ResponseWriter
	Data           interface{}
}

type WebradioEventStopPlayingData struct{}

// Playlist
type PlaylistEvent struct {
	ResponseWriter http.ResponseWriter
	Data           interface{}
}

type PlaylistEventPlayingSongData struct{}

// Buttons
type ButtonId int

const (
	DIGIT1_BUTTON ButtonId = iota
	DIGIT2_BUTTON
	DIGIT3_BUTTON
	DIGIT4_BUTTON
	DIGIT5_BUTTON
	DIGIT6_BUTTON
	PLAYLIST_BUTTON
	ALARM_SETTING_BUTTON
	MORE_BUTTON
	LESS_BUTTON
	SNOOZE_BUTTON
	NEXT_POWEROFF_BUTTON
)

type ButtonEventType int

const (
	PRESS_EVENT_TYPE ButtonEventType = iota
	RELEASE_EVENT_TYPE
)

type ButtonEvent struct {
	ButtonId        ButtonId
	ButtonEventType ButtonEventType
	PressStepCount  int64
}

// Api
type ApiEvent struct {
	//	ResponseWriter http.ResponseWriter
	//	Request        *http.Request
	Result chan error
	Data   interface{}
}

type ApiEventWebradioPlayData struct {
	WebradioId apimodel.WebradioId
}

type ApiEventPlaylistPlayData struct {
	PlaylistId apimodel.PlaylistId
}

type ApiEventAudioVolumeData struct {
	Volume int64
}
