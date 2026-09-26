package device

import (
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/jypelle/vekigi/apimodel"
	"github.com/jypelle/vekigi/internal/srv/event"
)

type LocalPlaylistPlayer struct {
	lock           sync.RWMutex
	eventChannel   chan event.PlaylistEvent
	playlistFolder string

	currentPlaylist          *Playlist
	currentPlaylistSongFiles []string
	currentPlaylistPosition  int64
	currentPlaylistCmd       *exec.Cmd

	sendEvent bool
}

func NewLocalPlaylistPlayer(playlistFolder string) PlaylistPlayer {
	playlistPlayer := LocalPlaylistPlayer{
		playlistFolder: playlistFolder,
		eventChannel:   make(chan event.PlaylistEvent),
		sendEvent:      true,
	}

	return &playlistPlayer
}

func (d *LocalPlaylistPlayer) Start() {
	slog.Info("Start local playlist player device")
}

func (d *LocalPlaylistPlayer) StopSendingEvent() {
	slog.Info("Stop sending events for local playlist player device")

	d.lock.Lock()
	defer d.lock.Unlock()

	d.sendEvent = false
	//close(d.eventChannel)
}

func (d *LocalPlaylistPlayer) Stop() {
	slog.Info("Stop local playlist player device")

	d.lock.Lock()
	defer d.lock.Unlock()

	d.clear()
}

func (d *LocalPlaylistPlayer) EventChannel() chan event.PlaylistEvent {
	return d.eventChannel
}

func (d *LocalPlaylistPlayer) PlaylistCount() int64 {
	files, err := os.ReadDir(d.playlistFolder)
	if err != nil {
		slog.Warn("Unable to access local playlist folder", "error", err)
		return 0
	}
	len := int64(0)
	for _, file := range files {
		if file.IsDir() {
			len++
		}
	}
	return len
}

func (d *LocalPlaylistPlayer) GetPlaylist(playlistId apimodel.PlaylistId) *Playlist {
	d.lock.Lock()
	defer d.lock.Unlock()

	return d.getPlaylist(playlistId)
}

func (d *LocalPlaylistPlayer) getPlaylist(playlistId apimodel.PlaylistId) *Playlist {
	if playlistId < 1 {
		slog.Warn("Playlist is undefined", "playlistId", playlistId)
		return nil
	}
	files, err := os.ReadDir(d.playlistFolder)
	if err != nil {
		slog.Warn("Unable to access local playlist folder", "error", err)
		return nil
	}
	currPlaylistId := apimodel.PlaylistId(0)
	for _, file := range files {
		if file.IsDir() {
			currPlaylistId++
			if currPlaylistId == playlistId {
				return &Playlist{
					PlaylistId: playlistId,
					Name:       file.Name(),
				}
			}
		}
	}

	slog.Warn("Playlist is undefined", "playlistId", playlistId)
	return nil

}

func (d *LocalPlaylistPlayer) Play(playlistId apimodel.PlaylistId) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.currentPlaylist != nil && playlistId == d.currentPlaylist.PlaylistId {
		slog.Info("Already listening playlist", "playlistId", playlistId)
		return nil
	}
	playlist := d.getPlaylist(playlistId)
	if playlist == nil {
		return fmt.Errorf("Playlist %d is undefined", playlistId)
	}

	// Retrieve playlist content
	files, err := os.ReadDir(filepath.Join(d.playlistFolder, playlist.Name))
	if err != nil {
		return fmt.Errorf("Unable to parse playlist folder: %v", err)
	}
	var playlistSongFiles []string
	for _, file := range files {
		if !file.IsDir() {
			playlistSongFiles = append(playlistSongFiles, file.Name())
		}
	}
	rand.Shuffle(len(playlistSongFiles), func(i, j int) {
		playlistSongFiles[i], playlistSongFiles[j] = playlistSongFiles[j], playlistSongFiles[i]
	})

	// Clear actual playlist
	d.clear()

	d.currentPlaylist = playlist
	d.currentPlaylistSongFiles = playlistSongFiles
	slog.Info("Listening playlist", "playlistId", playlistId, "name", playlist.Name)

	d.playSong()

	return nil
}

func (d *LocalPlaylistPlayer) playSong() {
	if d.currentPlaylistCmd != nil {
		if err := d.currentPlaylistCmd.Process.Kill(); err != nil {
			slog.Error("Failed to kill process", "error", err)
		}
	}
	d.currentPlaylistCmd = nil

	if d.currentPlaylistPosition >= int64(len(d.currentPlaylistSongFiles)) {
		d.currentPlaylistCmd = nil
		d.currentPlaylist = nil
		d.currentPlaylistPosition = 0
		d.currentPlaylistSongFiles = nil
		return
	}

	d.currentPlaylistCmd = exec.Command("cvlc", "--aout=alsa", "--play-and-exit", filepath.Join(d.playlistFolder, d.currentPlaylist.Name, d.currentPlaylistSongFiles[d.currentPlaylistPosition]))
	err := d.currentPlaylistCmd.Start()
	if err != nil {
		slog.Warn("Unable to listen song on playlist", "position", d.currentPlaylistPosition, "playlist", d.currentPlaylist.Name, "error", err)
		d.clear()
		return
	}

	currentPlaylistCmd := d.currentPlaylistCmd
	go func() {
		currentPlaylistCmd.Wait()
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

func (d *LocalPlaylistPlayer) CurrentPlaylist() *Playlist {
	d.lock.Lock()
	defer d.lock.Unlock()

	return d.currentPlaylist
}

func (d *LocalPlaylistPlayer) CurrentSongName() string {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.currentPlaylistPosition < int64(len(d.currentPlaylistSongFiles)) {
		return d.currentPlaylistSongFiles[d.currentPlaylistPosition]
	} else {
		return ""
	}
}

func (d *LocalPlaylistPlayer) Clear() {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.clear()
}

func (d *LocalPlaylistPlayer) clear() {
	if d.currentPlaylist != nil {
		if d.currentPlaylistCmd != nil {
			if err := d.currentPlaylistCmd.Process.Kill(); err != nil {
				slog.Error("Failed to kill process", "error", err)
			}
		}
		d.currentPlaylistCmd = nil
		d.currentPlaylist = nil
		d.currentPlaylistPosition = 0
		d.currentPlaylistSongFiles = nil
	}
}

func (d *LocalPlaylistPlayer) NextSong() {
	d.lock.Lock()
	defer d.lock.Unlock()
	if d.currentPlaylist != nil {
		d.currentPlaylistPosition++
		d.playSong()
	}
}
