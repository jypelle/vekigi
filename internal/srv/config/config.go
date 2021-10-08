package config

import (
	"github.com/jypelle/vekigi/apimodel"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
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

func NewServerConfig(configDir string, debugMode bool, simulationMode bool) *ServerConfig {
	serverConfig := &ServerConfig{
		ConfigDir:      configDir,
		DebugMode:      debugMode,
		SimulationMode: simulationMode,
	}

	// Check Configuration folder
	_, err := os.Stat(configDir)
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Printf("Creation of config folder: %s", configDir)
			err = os.Mkdir(configDir, 0770)
			if err != nil {
				logrus.Fatalf("Unable to create config folder: %v\n", err)
			}
		} else {
			logrus.Fatalf("Unable to access config folder: %s", configDir)
		}
	}

	// Open param file
	rawConfig, err := ioutil.ReadFile(serverConfig.GetCompleteParamFilename())
	if err == nil {
		// Interpret param file
		serverConfig.ServerParam = &ServerParam{}
		err = yaml.Unmarshal(rawConfig, serverConfig.ServerParam)
		if err != nil {
			logrus.Fatalf("Unable to interpret config file: %v\n", err)
		}
	} else {
		// Create default param file
		logrus.Infof("Create default param file")
		serverConfig.ServerParam = &ServerParam{}

		err = yaml.Unmarshal(ParamDefaultFile, serverConfig.ServerParam)
		if err != nil {
			logrus.Fatalf("Unable to interpret config file: %v\n", err)
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
	logrus.Debugf("Save param file: %s", sc.GetCompleteParamFilename())
	rawConfig, err := yaml.Marshal(*sc.ServerParam)
	if err != nil {
		logrus.Fatalf("Unable to serialize param file: %v\n", err)
	}
	err = ioutil.WriteFile(sc.GetCompleteParamFilename(), rawConfig, 0660)
	if err != nil {
		logrus.Fatalf("Unable to save param file: %v\n", err)
	}
}
