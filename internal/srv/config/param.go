package config

import (
	_ "embed"
	"github.com/jypelle/vekigi/apimodel"
)

//go:embed param_default.yaml
var ParamDefaultFile []byte

type ServerParam struct {
	SnoozeDuration int64                 `yaml:"snooze_duration"`
	WebradioGroups map[int64][]*Webradio `yaml:"webradio_groups"`
	MifasolParam   *MifasolParam         `yaml:"mifasol,omitempty"`
	ApiParam       ApiParam              `yaml:"api"`
}

type Webradio struct {
	Name       string              `yaml:"name"`
	Url        string              `yaml:"url"`
	WebradioId apimodel.WebradioId `yaml:"-"`
}

type ApiParam struct {
	Enabled bool   `yaml:"enabled"`
	SslPort int64  `yaml:"ssl_port"`
	ApiKey  string `yaml:"api_key"`
}
