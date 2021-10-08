package srv

import (
	"github.com/jypelle/vekigi/internal/srv/config"
	"github.com/jypelle/vekigi/internal/srv/device"
	"github.com/jypelle/vekigi/internal/srv/event"
	"github.com/sirupsen/logrus"
	"syscall"
	"time"
)

func (s *ServerApp) eventLoop() {
	for loop := true; loop; {
		select {
		case ev := <-s.internalEventChannel:
			switch ev.Data.(type) {
			case event.InternalEventPopupHideData:
				if s.popUpHideTimer != nil {
					s.refreshDisplay(true)
				}
			case event.InternalEventAnimationTickData:
				if s.animationTickTimer != nil {
					s.animationTickCount++
					s.refreshDisplay(false)
				}
			}
		case ev := <-s.clockDevice.EventChannel():
			switch ev.Data.(type) {
			case event.TickerEventTickData:
				logrus.Debugf("Receive Ticker tick event")
				if s.currentMode == CLOCK_MODE && s.currentPopUp == NO_POPUP {
					s.refreshDisplay(false)
				}
			case event.TickerEventAlarmData:
				logrus.Infof("Receive Ticker alarm event")
				alarmTime := s.Alarm()
				if alarmTime.WebradioId != nil {
					s.playlistPlayerDevice.Clear()
					err := s.webradioPlayerDevice.Play(*alarmTime.WebradioId)
					if err != nil {
						logrus.Warn(err)
					}
				} else if alarmTime.PlaylistId != nil {
					s.webradioPlayerDevice.Clear()
					err := s.playlistPlayerDevice.Play(*alarmTime.PlaylistId)
					if err != nil {
						logrus.Warn(err)
					}
				}
				s.refreshDisplay(true)
			}
		case ev := <-s.apiDevice.EventChannel():
			switch data := ev.Data.(type) {
			case event.ApiEventWebradioPlayData:
				s.clockDevice.ClearAlarm()
				s.playlistPlayerDevice.Clear()
				err := s.webradioPlayerDevice.Play(data.WebradioId)
				ev.Result <- err
				s.refreshDisplay(true)
			case event.ApiEventPlaylistPlayData:
				s.clockDevice.ClearAlarm()
				s.webradioPlayerDevice.Clear()
				err := s.playlistPlayerDevice.Play(data.PlaylistId)
				ev.Result <- err
				s.refreshDisplay(true)
			case event.ApiEventAudioVolumeData:
				err := s.audioDevice.SetVolume(data.Volume)
				ev.Result <- err
			}
		case ev := <-s.webradioPlayerDevice.EventChannel():
			switch ev.Data.(type) {
			case event.WebradioEventStopPlayingData:
				logrus.Debugf("Receive webradioStopPlaying event")
				s.refreshDisplay(true)
			}
		case ev := <-s.playlistPlayerDevice.EventChannel():
			switch ev.Data.(type) {
			case event.PlaylistEventPlayingSongData:
				logrus.Infof("Receive playlistPlayingSong event")
				s.refreshDisplay(true)
			}
		case ev := <-s.buttonsDevice.EventChannel():
			logrus.Debugf("Receive button event: %d, %d, %d", ev.ButtonId, ev.ButtonEventType, ev.PressStepCount)
			switch ev.ButtonId {
			case event.DIGIT1_BUTTON:
				fallthrough
			case event.DIGIT2_BUTTON:
				fallthrough
			case event.DIGIT3_BUTTON:
				fallthrough
			case event.DIGIT4_BUTTON:
				fallthrough
			case event.DIGIT5_BUTTON:
				fallthrough
			case event.DIGIT6_BUTTON:
				if ev.ButtonEventType == event.PRESS_EVENT_TYPE {
					logrus.Debugf("Receive button digit press event")
					if (ev.PressStepCount-1)%3 == 0 {
						groupId := int64(ev.ButtonId) + 1
						webradioList, ok := s.WebradioGroups[groupId]
						if ok {
							if s.currentMode == CLOCK_MODE {
								currentWebradio := s.webradioPlayerDevice.CurrentWebRadio()
								var nextWebradio *config.Webradio
								if currentWebradio != nil && currentWebradio.WebradioId.GroupId == groupId {
									nextWebradio = webradioList[int(currentWebradio.WebradioId.IndexId)%len(webradioList)]
								} else {
									nextWebradio = webradioList[0]
								}
								s.clockDevice.ClearAlarm()
								s.playlistPlayerDevice.Clear()
								err := s.webradioPlayerDevice.Play(nextWebradio.WebradioId)
								if err != nil {
									logrus.Warn(err)
								}
							} else if s.currentMode == ALARM_SETTING_MODE {
								alarmTime := s.Alarm()
								var nextWebradio *config.Webradio
								if alarmTime.WebradioId != nil && alarmTime.WebradioId.GroupId == groupId {
									nextWebradio = webradioList[int(alarmTime.WebradioId.IndexId)%len(webradioList)]
								} else {
									nextWebradio = webradioList[0]
								}
								alarmTime.WebradioId = &nextWebradio.WebradioId
								alarmTime.PlaylistId = nil
								s.SetAlarm(alarmTime)
							}
							s.refreshDisplay(true)
						}
					}
				}
			case event.PLAYLIST_BUTTON:
				if ev.ButtonEventType == event.PRESS_EVENT_TYPE {
					logrus.Debugf("Receive button playlist press event")
					if (ev.PressStepCount-1)%3 == 0 {
						if s.currentMode == CLOCK_MODE {
							currentPlaylist := s.playlistPlayerDevice.CurrentPlaylist()
							var nextPlaylist *device.Playlist
							if currentPlaylist != nil {
								nextPlaylist = s.playlistPlayerDevice.GetPlaylist(currentPlaylist.PlaylistId + 1)
							}
							if nextPlaylist == nil {
								nextPlaylist = s.playlistPlayerDevice.GetPlaylist(1)
							}
							if nextPlaylist != nil {
								s.clockDevice.ClearAlarm()
								s.webradioPlayerDevice.Clear()
								err := s.playlistPlayerDevice.Play(nextPlaylist.PlaylistId)
								if err != nil {
									logrus.Warn(err)
								}
							}
						} else if s.currentMode == ALARM_SETTING_MODE {
							alarmTime := s.Alarm()
							var nextPlaylist *device.Playlist
							if alarmTime.PlaylistId != nil {
								nextPlaylist = s.playlistPlayerDevice.GetPlaylist(*alarmTime.PlaylistId + 1)
							}
							if nextPlaylist == nil {
								nextPlaylist = s.playlistPlayerDevice.GetPlaylist(1)
							}
							if nextPlaylist != nil {
								alarmTime.WebradioId = nil
								alarmTime.PlaylistId = &nextPlaylist.PlaylistId
								s.SetAlarm(alarmTime)
							}
						}
						s.refreshDisplay(true)

					}
				}
			case event.ALARM_SETTING_BUTTON:
				if ev.ButtonEventType == event.RELEASE_EVENT_TYPE && ev.PressStepCount < 6 {
					logrus.Debugf("Switch alarm setting mode")
					if s.currentMode == CLOCK_MODE {
						s.currentMode = ALARM_SETTING_MODE
					} else if s.currentMode == ALARM_SETTING_MODE {
						s.currentMode = CLOCK_MODE
					}
					s.refreshDisplay(true)
				} else if ev.ButtonEventType == event.PRESS_EVENT_TYPE && ev.PressStepCount == 6 {
					if s.currentMode == CLOCK_MODE {
						logrus.Debugf("Switch alarm enabled state")
						alarmTime := s.Alarm()
						alarmTime.Enabled = !alarmTime.Enabled
						s.SetAlarm(alarmTime)
						s.refreshDisplay(true)
					}
				}
			case event.LESS_BUTTON:
				if ev.ButtonEventType == event.PRESS_EVENT_TYPE {
					logrus.Debugf("Receive button less event")
					if s.currentMode == CLOCK_MODE {
						if s.popUpHideTimer != nil {
							s.popUpHideTimer.Stop()
						}
						s.audioDevice.DecreaseVolume()
						s.currentPopUp = VOLUME_POPUP
						s.popUpHideTimer = time.AfterFunc(1200*time.Millisecond, func() {
							s.internalEventChannel <- event.InternalEvent{Data: event.InternalEventPopupHideData{}}
						})
						s.refreshDisplay(false)
					} else if s.currentMode == ALARM_SETTING_MODE {
						alarmTime := s.Alarm()
						if ev.PressStepCount <= 20 {
							alarmTime.AddMinute(-1)
						} else if ev.PressStepCount <= 30 {
							alarmTime.AddMinute(-5)
						} else {
							alarmTime.AddMinute(-30)
						}
						s.SetAlarm(alarmTime)
						s.refreshDisplay(true)
					}
				}
			case event.MORE_BUTTON:
				if ev.ButtonEventType == event.PRESS_EVENT_TYPE {
					logrus.Debugf("Receive button more event")
					if s.currentMode == CLOCK_MODE {
						if s.popUpHideTimer != nil {
							s.popUpHideTimer.Stop()
						}
						s.audioDevice.IncreaseVolume()
						s.currentPopUp = VOLUME_POPUP
						s.popUpHideTimer = time.AfterFunc(1200*time.Millisecond, func() {
							s.internalEventChannel <- event.InternalEvent{Data: event.InternalEventPopupHideData{}}
						})
						s.refreshDisplay(false)
					} else if s.currentMode == ALARM_SETTING_MODE {
						alarmTime := s.Alarm()
						if ev.PressStepCount <= 20 {
							alarmTime.AddMinute(1)
						} else if ev.PressStepCount <= 30 {
							alarmTime.AddMinute(5)
						} else {
							alarmTime.AddMinute(30)
						}
						s.SetAlarm(alarmTime)
						s.refreshDisplay(true)
					}
				}
			case event.SNOOZE_BUTTON:
				if ev.ButtonEventType == event.RELEASE_EVENT_TYPE && ev.PressStepCount < 5 {
					logrus.Debugf("Switch display on/off")
					s.displayDevice.Switch()
				} else if ev.ButtonEventType == event.PRESS_EVENT_TYPE {
					if ev.PressStepCount == 5 {
						logrus.Debugf("Stop playing sound")
						s.clockDevice.Snooze()
						s.webradioPlayerDevice.Clear()
						s.playlistPlayerDevice.Clear()
						s.refreshDisplay(true)
					} else if ev.PressStepCount == 15 {
						if s.clockDevice.IsAlarmRunning() {
							logrus.Debugf("Snooze off")
							s.currentPopUp = SNOOZE_OFF_POPUP
							s.popUpHideTimer = time.AfterFunc(1200*time.Millisecond, func() {
								s.internalEventChannel <- event.InternalEvent{Data: event.InternalEventPopupHideData{}}
							})
							s.clockDevice.ClearAlarm()
							s.refreshDisplay(false)
						}
					}
				}
			case event.NEXT_POWEROFF_BUTTON:
				if ev.ButtonEventType == event.RELEASE_EVENT_TYPE && ev.PressStepCount < 20 {
					if s.playlistPlayerDevice.CurrentPlaylist() != nil {
						logrus.Debugf("Next song in playlist")
						s.playlistPlayerDevice.NextSong()
						s.refreshDisplay(true)
					}
				} else if ev.ButtonEventType == event.PRESS_EVENT_TYPE && ev.PressStepCount == 20 {
					logrus.Debugf("See you!")
					s.clockDevice.ClearAlarm()
					syscall.Kill(syscall.Getpid(), syscall.SIGUSR1)

				}
			}
		case <-s.eventLoopAskDone:
			loop = false
		}
	}
	s.eventLoopDone <- true
}
