package adapter

import (
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type SourceConfig struct {
	Sources map[string]SourceInfo `json:"sources"`
}

type SourceInfo struct {
	Host                string `json:"host"`
	Port                int    `json:"port"`
	Channel             string `json:"channel"`
	WorkerCount         *int   `json:"worker_count",omitempty`
	PingInterval        *int64 `json:"ping_interval",omitempty`
	MaxPingsOutstanding *int   `json:"max_pings_outstanding",omitempty`
	MaxReconnects       *int   `json:"max_reconnects",omitempty`
}

type SourceManager struct {
	adapter *Adapter
	sources map[string]*Source
}

func NewSourceManager(adapter *Adapter) *SourceManager {
	return &SourceManager{
		adapter: adapter,
		sources: make(map[string]*Source),
	}
}

func (sm *SourceManager) Initialize() error {

	config, err := sm.LoadSourceConfig(viper.GetString("source.config"))
	if err != nil {
		return err
	}

	// Initializing sources
	for name, info := range config.Sources {

		log.WithFields(log.Fields{
			"name": name,
			"host": info.Host,
			"port": info.Port,
		}).Info("Initializing source")

		source := NewSource(sm.adapter, name, &info)
		err := source.Init()
		if err != nil {
			log.Error(err)
			continue
		}

		sm.sources[name] = source
	}

	return nil
}

func (sm *SourceManager) Uninitialize() error {
	config, err := sm.LoadSourceConfig(viper.GetString("source.config"))
	if err != nil {
		return err
	}

	// Uninitializing sources
	for name, _ := range config.Sources {
		source := sm.sources[name]
		err := source.Uninit()
		if err != nil {
			log.Error(err)
		}
	}

	return nil

}

func (sm *SourceManager) LoadSourceConfig(filename string) (*SourceConfig, error) {

	// Open configuration file
	jsonFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer jsonFile.Close()

	// Read
	byteValue, _ := ioutil.ReadAll(jsonFile)

	var config SourceConfig

	json.Unmarshal(byteValue, &config)

	return &config, nil
}
