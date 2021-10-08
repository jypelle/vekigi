package device

import (
	"github.com/jypelle/vekigi/apimodel"
	"github.com/jypelle/vekigi/internal/srv/event"
)

type Playlist struct {
	PlaylistId apimodel.PlaylistId
	Name       string
}

type PlaylistPlayer interface {
	Start()
	StopSendingEvent()
	Stop()
	EventChannel() chan event.PlaylistEvent
	PlaylistCount() int64
	GetPlaylist(playlistId apimodel.PlaylistId) *Playlist
	Play(playlistId apimodel.PlaylistId) error
	CurrentPlaylist() *Playlist
	CurrentSongName() string
	Clear()
	NextSong()
}
