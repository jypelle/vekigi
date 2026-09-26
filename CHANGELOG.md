## v1.1.0

*2026-09-26*

- New features:
  - **Simulator buttons**: the 12 buttons of the Vekigi can now be pressed from the keyboard in the simulator (same auto-repeat as the real buttons when held down)
    - `1`-`6` (top row or numpad): digit buttons
    - `P`: playlist
    - `Enter`: alarm setting
    - `Up` / `Down` (or numpad `+` / `-`): more / less
    - `Space`: snooze
    - `Right`: next / power off
- Enhancements:
  - **Simulator** rewritten with [Ebitengine](https://ebitengine.org) (replaces Gio):
    - Pixel-perfect rendering of the OLED display, which now also shows a black screen when the display is switched off
    - No longer requires cgo nor any system library to build
    - **Breaking change**: simulation mode is now enabled at build time with `-tags simulator` (the `-s` flag is removed), use `task simulator` to run it
  - Structured logs with `log/slog`
  - Build with Go 1.27 and [Task](https://taskfile.dev) (`task build`, `task deploy`, `task simulator`) instead of make
  - Update dependencies
- Bug fixes:
  - The `api.enabled` param is now taken into account: set it to `false` to disable the REST API
  - The ARM binary is now explicitly built for ARMv6, as required by the Raspberry Pi Zero (recent Go versions target ARMv7 by default)
  - A stop signal received while the server was busy could be missed
  - Some log messages were missing the error details (popping/clicking cleaner, volume, playlist songs)

## v1.0.0

*2021-10-08*

- Initial release:
  - Alarm clock with adjustable snooze time
  - REST API
  - Play webradios, local playlists or remote playlists (through [Mifasol music server](https://github.com/jypelle/mifasol))
  - `no_alarm_on_weekends` param to skip the alarm on saturdays and sundays (added on 2022-01-08)
