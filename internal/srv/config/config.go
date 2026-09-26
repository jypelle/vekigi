package config

import (
	"io/ioutil"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/jypelle/vekigi/apimodel"
	"github.com/jypelle/vekigi/internal/tool"
	"gopkg.in/yaml.v3"
)

const paramFilename = "param.yaml"
const stateFilename = "state.yaml"
const playlistFolder = "playlist"

type ServerConfig struct {
	ConfigDir      string
	DebugMode      bool
	SimulationMode bool

	*ServerParam
	*ServerState
}

func NewServerConfig(configDir string, debugMode bool) *ServerConfig {
	serverConfig := &ServerConfig{
		ConfigDir:      configDir,
		DebugMode:      debugMode,
		SimulationMode: simulationMode,
	}

	// Check Configuration folder
	_, err := os.Stat(configDir)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Info("Creation of config folder", "folder", configDir)
			err = os.Mkdir(configDir, 0770)
			if err != nil {
				tool.Fatal("Unable to create config folder", "error", err)
			}
		} else {
			tool.Fatal("Unable to access config folder", "folder", configDir)
		}
	}

	// Open param file
	rawConfig, err := ioutil.ReadFile(serverConfig.GetCompleteParamFilename())
	if err == nil {
		// Interpret param file
		serverConfig.ServerParam = &ServerParam{}
		err = yaml.Unmarshal(rawConfig, serverConfig.ServerParam)
		if err != nil {
			tool.Fatal("Unable to interpret config file", "error", err)
		}
	} else {
		// Create default param file
		slog.Info("Create default param file")
		serverConfig.ServerParam = &ServerParam{}

		err = yaml.Unmarshal(ParamDefaultFile, serverConfig.ServerParam)
		if err != nil {
			tool.Fatal("Unable to interpret config file", "error", err)
		}

		serverConfig.SaveParam()
	}

	if serverConfig.ServerParam.MifasolParam != nil {
		serverConfig.ServerParam.MifasolParam.ConfigDir = serverConfig.ConfigDir
	}

	for groupId, webradioList := range serverConfig.WebradioGroups {
		for pseudoIndexId, webradio := range webradioList {
			webradio.WebradioId = apimodel.WebradioId{
				GroupId: groupId,
				IndexId: int64(pseudoIndexId) + 1,
			}
		}
	}

	// Open state file
	serverConfig.ServerState = NewsServerState(serverConfig.GetCompleteStateFilename())

	return serverConfig
}

func (sc *ServerConfig) GetCompleteParamFilename() string {
	return filepath.Join(sc.ConfigDir, paramFilename)
}

func (sc *ServerConfig) GetCompleteStateFilename() string {
	return filepath.Join(sc.ConfigDir, stateFilename)
}

func (sc *ServerConfig) GetCompletePlaylistFolder() string {
	return filepath.Join(sc.ConfigDir, playlistFolder)
}

func (sc *ServerConfig) SaveParam() {
	slog.Debug("Save param file", "file", sc.GetCompleteParamFilename())
	rawConfig, err := yaml.Marshal(*sc.ServerParam)
	if err != nil {
		tool.Fatal("Unable to serialize param file", "error", err)
	}
	err = ioutil.WriteFile(sc.GetCompleteParamFilename(), rawConfig, 0660)
	if err != nil {
		tool.Fatal("Unable to save param file", "error", err)
	}
}
