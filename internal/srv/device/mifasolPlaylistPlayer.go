package device

import (
	"fmt"
	"github.com/jypelle/mifasol/restApiV1"
	"github.com/jypelle/mifasol/restClientV1"
	"github.com/jypelle/vekigi/apimodel"
	"github.com/jypelle/vekigi/internal/srv/config"
	"github.com/jypelle/vekigi/internal/srv/event"
	"github.com/sirupsen/logrus"
	"math/rand"
	"os/exec"
	"sync"
)

type MifasolPlaylistPlayer struct {
	lock          sync.RWMutex
	eventChannel  chan event.PlaylistEvent
	mifasolClient *restClientV1.RestClient

	currentPlaylistId       apimodel.PlaylistId
	currentPlaylistPosition int64
	currentSongName         string
	currentPlaylistCmd      *exec.Cmd

	sendEvent bool

	mifasolPlaylistList []restApiV1.Playlist
}

func NewMifasolPlaylistPlayer(mifasolParam *config.MifasolParam) PlaylistPlayer {
	playlistPlayer := MifasolPlaylistPlayer{
		eventChannel: make(chan event.PlaylistEvent),
		sendEvent:    true,
	}

	var err error
	playlistPlayer.mifasolClient, err = restClientV1.NewRestClient(mifasolParam, false)
	if err != nil {
		logrus.Warningf("Failed to create mifasol client: %v", err)
	}

	userId := playlistPlayer.mifasolClient.UserId()
	playlistFilterOrder := restApiV1.PlaylistFilterOrderByName
	playlistPlayer.mifasolPlaylistList, err = playlistPlayer.mifasolClient.ReadPlaylists(&restApiV1.PlaylistFilter{
		FavoriteUserId: &userId,
		OrderBy:        &playlistFilterOrder,
	})

	return &playlistPlayer
}

func (d *MifasolPlaylistPlayer) Start() {
	logrus.Infof("Start mifasol playlist player device")
}

func (d *MifasolPlaylistPlayer) StopSendingEvent() {
	logrus.Infof("Stop sending events for mifasol playlist player device")

	d.lock.Lock()
	defer d.lock.Unlock()

	d.sendEvent = false
	//close(d.eventChannel)
}

func (d *MifasolPlaylistPlayer) Stop() {
	logrus.Infof("Stop playlist mifasol player device")

	d.lock.Lock()
	defer d.lock.Unlock()

	d.clear()
}

func (d *MifasolPlaylistPlayer) EventChannel() chan event.PlaylistEvent {
	return d.eventChannel
}

func (d *MifasolPlaylistPlayer) PlaylistCount() int64 {
	d.lock.Lock()
	defer d.lock.Unlock()

	return int64(len(d.mifasolPlaylistList))
}

func (d *MifasolPlaylistPlayer) GetPlaylist(playlistId apimodel.PlaylistId) *Playlist {
	d.lock.Lock()
	defer d.lock.Unlock()

	mifasolPlaylist := d.getMifasolPlaylist(playlistId)

	if mifasolPlaylist != nil {
		return &Playlist{
			PlaylistId: playlistId,
			Name:       mifasolPlaylist.Name,
		}
	} else {
		return nil
	}
}

func (d *MifasolPlaylistPlayer) getMifasolPlaylist(playlistId apimodel.PlaylistId) *restApiV1.Playlist {
	if playlistId < 1 {
		return nil
	}

	if int(playlistId) <= len(d.mifasolPlaylistList) {
		return &d.mifasolPlaylistList[playlistId-1]
	}

	logrus.Warnf("Mifasol playlist %d is undefined", playlistId)
	return nil

}

func (d *MifasolPlaylistPlayer) Play(playlistId apimodel.PlaylistId) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	mifasolPlaylist := d.getMifasolPlaylist(playlistId)
	if mifasolPlaylist == nil {
		return fmt.Errorf("Playlist %d is undefined", playlistId)
	}

	if d.currentPlaylistId > 0 && d.currentPlaylistId == playlistId {
		logrus.Infof("Already listening playlist %d", playlistId)
		return nil
	}

	// Shuffle playlist content
	rand.Shuffle(len(mifasolPlaylist.SongIds), func(i, j int) {
		mifasolPlaylist.SongIds[i], mifasolPlaylist.SongIds[j] = mifasolPlaylist.SongIds[j], mifasolPlaylist.SongIds[i]
	})

	// Clear actual playlist
	d.clear()

	d.currentPlaylistId = playlistId
	logrus.Infof("Listening Playlist %d: \"%s\" ", playlistId, mifasolPlaylist.Name)

	d.playSong()

	return nil
}

func (d *MifasolPlaylistPlayer) playSong() {
	if d.currentPlaylistCmd != nil {
		if err := d.currentPlaylistCmd.Process.Kill(); err != nil {
			logrus.Errorf("Failed to kill process: %v", err)
		}
	}
	d.currentPlaylistCmd = nil

	currentMifasolPlaylist := d.getMifasolPlaylist(d.currentPlaylistId)
	if currentMifasolPlaylist == nil || d.currentPlaylistPosition >= int64(len(currentMifasolPlaylist.SongIds)) {
		d.currentPlaylistCmd = nil
		d.currentPlaylistId = 0
		d.currentPlaylistPosition = 0
		d.currentSongName = ""
		return
	}

	song, cliErr := d.mifasolClient.ReadSong(currentMifasolPlaylist.SongIds[d.currentPlaylistPosition])
	if cliErr != nil {
		logrus.Warnf("Unknown song %d on playlist %s", d.currentPlaylistPosition, currentMifasolPlaylist.Name)
		d.clear()
		return
	}
	songContent, _, cliErr := d.mifasolClient.ReadSongContent(currentMifasolPlaylist.SongIds[d.currentPlaylistPosition])
	if cliErr != nil {
		logrus.Warnf("Unable to read %d on playlist %s", d.currentPlaylistPosition, currentMifasolPlaylist.Name)
		d.clear()
		return
	}

	d.currentSongName = song.Name
	d.currentPlaylistCmd = exec.Command("cvlc", "--aout=alsa", "--play-and-exit", "-")
	d.currentPlaylistCmd.Stdin = songContent
	err := d.currentPlaylistCmd.Start()
	if err != nil {
		logrus.Warnf("Unable to listen song %d on playlist %s", d.currentPlaylistPosition, currentMifasolPlaylist.Name)
		d.clear()
		return
	}

	currentPlaylistCmd := d.currentPlaylistCmd
	go func() {
		currentPlaylistCmd.Wait()
		songContent.Close()
		d.lock.Lock()
		defer d.lock.Unlock()

		if d.currentPlaylistCmd == currentPlaylistCmd {
			d.currentPlaylistCmd = nil
			d.currentPlaylistPosition++
			d.playSong()
			if d.sendEvent {
				go func() { d.eventChannel <- event.PlaylistEvent{Data: event.PlaylistEventPlayingSongData{}} }()
			}
		}
	}()
}

func (d *MifasolPlaylistPlayer) CurrentPlaylist() *Playlist {
	d.lock.Lock()
	defer d.lock.Unlock()

	currentMifasolPlaylist := d.getMifasolPlaylist(d.currentPlaylistId)

	if currentMifasolPlaylist != nil {
		return &Playlist{
			PlaylistId: d.currentPlaylistId,
			Name:       currentMifasolPlaylist.Name,
		}
	} else {
		return nil
	}
}

func (d *MifasolPlaylistPlayer) CurrentSongName() string {
	d.lock.Lock()
	defer d.lock.Unlock()

	currentMifasolPlaylist := d.getMifasolPlaylist(d.currentPlaylistId)

	if currentMifasolPlaylist != nil && d.currentPlaylistPosition < int64(len(currentMifasolPlaylist.SongIds)) {
		return d.currentSongName
	} else {
		return ""
	}
}

func (d *MifasolPlaylistPlayer) Clear() {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.clear()
}

func (d *MifasolPlaylistPlayer) clear() {
	if d.currentPlaylistId > 0 {
		if d.currentPlaylistCmd != nil {
			if err := d.currentPlaylistCmd.Process.Kill(); err != nil {
				logrus.Errorf("Failed to kill process: %v", err)
			}
		}
		d.currentPlaylistCmd = nil
		d.currentPlaylistId = 0
		d.currentPlaylistPosition = 0
		d.currentSongName = ""
	}
}

func (d *MifasolPlaylistPlayer) NextSong() {
	d.lock.Lock()
	defer d.lock.Unlock()
	if d.currentPlaylistId > 0 {
		d.currentPlaylistPosition++
		d.playSong()
	}
}
