package config

import (
	"path/filepath"
)

const configCertFilename = "mifasolcert.pem"

type MifasolParam struct {
	ConfigDir        string `yaml:"-"`
	ServerHostname   string `yaml:"hostname"`
	ServerPort       int64  `yaml:"port"`
	ServerSsl        bool   `yaml:"ssl"`
	ServerSelfSigned bool   `yaml:"self_signed"`
	Username         string `yaml:"username"`
	Password         string `yaml:"password"`
	Timeout          int64  `yaml:"timeout"`
}

func (c MifasolParam) GetCompleteConfigCertFilename() string {
	return filepath.Join(c.ConfigDir, configCertFilename)
}

func (c MifasolParam) GetServerHostname() string {
	return c.ServerHostname
}

func (c MifasolParam) GetServerPort() int64 {
	return c.ServerPort
}

func (c MifasolParam) GetServerSsl() bool {
	return c.ServerSsl
}

func (c MifasolParam) GetServerSelfSigned() bool {
	return c.ServerSelfSigned
}

func (c MifasolParam) GetTimeout() int64 {
	return c.Timeout
}

func (c MifasolParam) GetUsername() string {
	return c.Username
}

func (c MifasolParam) GetPassword() string {
	return c.Password
}
