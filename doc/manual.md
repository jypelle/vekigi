# User manual

Your Vekigi is built and operational? Great! Let's see how to use and tune it.

## Buttons

Most buttons react differently to a short press and to a long press (the button being held down).

| Button                      | Short press                                              | Long press                                                                                                        |
|-----------------------------|----------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------|
| `1` to `6`                  | Play the first webradio of the group, then the next ones | Cycle through the webradios of the group (about every half second)                                               |
| `PLAYLIST`                  | Play the next playlist                                   | Cycle through the playlists (about every half second)                                                            |
| `ALARM SETTING`             | Enter / leave the alarm settings                         | ~1 s: enable / disable the alarm                                                                                  |
| `-` / `+`                   | Decrease / increase the volume                           | Keep changing the volume                                                                                          |
| `Display On/Off - Snooze`   | Switch the display on / off                              | ~0.6 s: stop the sound (snooze if the alarm is ringing)<br>~2 s: stop the sound and cancel the snooze ("Snooze off") |
| `Next song - Power off`     | Play the next song of the current playlist               | ~3 s: power off the Vekigi ("See you!")                                                                           |

In the alarm settings, the `1` to `6`, `PLAYLIST` and `-` / `+` buttons change the alarm instead (see below).

## Display

The display shows the current time and, at the bottom, the name of the webradio or of the playlist and song being played (scrolling when too long).

In the top right corner:

- the alarm icon means the alarm is enabled
- the snooze icon means the alarm has rung and will ring again after the snooze time

A volume gauge pops up when the volume changes.

## Alarm

### Set the alarm

- Press `ALARM SETTING`: the display shows *Alarm settings* with the alarm time and sound
- Use `-` / `+` to change the time: 1 minute at a time, then 5 minutes after holding the button ~3 s, then 30 minutes after ~5 s
- Choose the sound to wake up with:
  - `1` to `6`: a webradio of the group (press again to select the next webradio of the group)
  - `PLAYLIST`: a playlist (press again to select the next one)
- Press `ALARM SETTING` again to go back to the clock

The alarm is saved automatically.

### Enable or disable the alarm

From the clock, hold `ALARM SETTING` for about 1 second: the alarm icon appears or disappears.

### When the alarm rings

The selected webradio or playlist starts playing and the snooze icon appears.

- Hold `Display On/Off - Snooze` for ~0.6 s to stop the sound: the alarm will ring again after the snooze time (10 minutes by default)
- Keep holding it for ~2 s to also cancel the snooze: *Snooze off* is displayed and the alarm won't ring again before the next day

Playing a webradio or a playlist also cancels the snooze.

### Weekends

To skip the alarm on Saturdays and Sundays, set `no_alarm_on_weekends` to `true` in the `state.yaml` file (see [State file](#state-file)).

## Webradios

Webradios are organized in 6 groups, one for each `1` to `6` button. Pressing a digit button plays the first webradio of its group, pressing it again plays the next one, and so on.

The list of webradios is set in the config file (see [Config file](#config-file)).

## Playlists

The `PLAYLIST` button plays the playlists one after another. Songs of a playlist are played in random order, and the playback stops after the last song.

Playlists come from one of these sources:

- **Local playlists** (default): each subfolder of `~/.config/vekigi/playlist` is a playlist (sorted by name), containing music files playable by [VLC](https://www.videolan.org). For instance, to copy a folder from your computer:

  ```bash
  ssh pi@vekigi "mkdir -p .config/vekigi/playlist"
  scp -r "My playlist" pi@vekigi:.config/vekigi/playlist/
  ```

- **Remote playlists**: when the `mifasol` section is set in the config file, the playlists are the favorite playlists of the Mifasol user (sorted by name), played from your [Mifasol music server](https://github.com/jypelle/mifasol).

## Configuration

The Vekigi stores its configuration in `~/.config/vekigi` (another folder can be used with the `-c` option of `vekigisrv`). Both files are created with default values at the first start.

To edit them:

- Connect to your Vekigi through SSH: `ssh pi@vekigi`
- Stop the Vekigi server: `sudo systemctl stop vekigisrv.service`
- Edit the file, for instance: `nano .config/vekigi/param.yaml`
- Start the Vekigi server: `sudo systemctl start vekigisrv.service`

NB: the server must be stopped before editing `state.yaml`, otherwise your changes would be overwritten.

### Config file

`param.yaml`:

```yaml
# Snooze time, in seconds
snooze_duration: 600

# Webradios of each group (1 to 6 for the digit buttons), in order
webradio_groups:
  1:
    - name: France info
      url: http://direct.franceinfo.fr/live/franceinfo-hifi.aac
    - name: France bleu nord
      url: http://direct.francebleu.fr/live/fbnord-midfi.mp3
  2:
    - name: FIP Jazz
      url: http://icecast.radiofrance.fr/fipjazz-hifi.aac
  # ...

# Optional: use the favorite playlists of a Mifasol user instead of the local playlists
mifasol:
  hostname: localhost
  port: 6620
  ssl: true
  self_signed: true
  username: mifasol
  password: mifasol
  timeout: 600

# REST API
api:
  enabled: true # Set to false to disable the REST API
  ssl_port: 6650
  api_key: timesup
```

Any audio stream playable by [VLC](https://www.videolan.org) can be used as a webradio `url`.

Don't forget to change the default `api_key`.

### State file

`state.yaml` holds what the Vekigi remembers between restarts. It is saved automatically a few seconds after each change made with the buttons or the API.

```yaml
# Volume, from 0 to 100
volume: 40
alarm:
  hour: 8
  minute: 0
  # Alarm sound: either a webradio or a playlist (the other one being null)
  webradio_id:
    group_id: 1 # Group (digit button)
    index_id: 1 # Position of the webradio in the group, starting from 1
  playlist_id: null # Position of the playlist, starting from 1
  enabled: true
  # Set to true to skip the alarm on Saturdays and Sundays
  no_alarm_on_weekends: false
```

## REST API

The Vekigi can be controlled from your home automation system with a REST API, served over HTTPS on the `ssl_port` of the config file.

A self-signed certificate is generated at the first start (`cert.pem` and `key.pem` in the config folder): with curl, use `-k` to accept it.

Each request must send the `api_key` of the config file in the `x-api-key` header.

| Request                                           | Description                                                     |
|---------------------------------------------------|-----------------------------------------------------------------|
| `GET /api/is_alive`                               | Check that the Vekigi is up                                     |
| `POST /api/webradio/play/{group_id}/{index_id}`   | Play the webradio at position `index_id` (from 1) of the group  |
| `POST /api/playlist/play/{playlist_id}`           | Play the playlist at position `playlist_id` (from 1)            |
| `POST /api/audio/volume/{volume}`                 | Set the volume (0 to 100)                                       |

For instance, to play the second webradio of the first group:

```bash
curl -k -X POST -H "x-api-key: timesup" https://vekigi:6650/api/webradio/play/1/2
```

The API answers with a `200` status on success, `403` for a wrong API key or an unknown webradio / playlist, and `400` for invalid parameters.

## Power off

Hold `Next song - Power off` for about 3 seconds: *See you!* is displayed and the Vekigi shuts down. Wait a few seconds before unplugging it.
