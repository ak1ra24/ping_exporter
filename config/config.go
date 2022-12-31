package config

import (
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type SafeConfig struct {
	sync.RWMutex
	C *Config
}

type Config struct {
	Targets []Target `yaml:"targets"`
}

type Host struct {
	IP          string `yaml:"ip"`
	Name        string `yaml:"name,omitempty"`
	Broadcast   bool   `yaml:"broadcast"`
	Description string `yaml:"description,omitempty"`
}

type Target struct {
	Hosts    []Host        `yaml:"hosts"`
	Interval time.Duration `yaml:"interval,omitempty"`
	Timeout  time.Duration `yaml:"timeout,omitempty"`
	Network  string        `yaml:"network,omitempty"`
	Protocol string        `yaml:"protocol,omitempty"`
	Size     int           `yaml:"size,omitempty"`
}

func (sc *SafeConfig) ReloadConfig(confFile string) (err error) {
	var c = &Config{}

	yamlReader, err := os.Open(confFile)
	if err != nil {
		return errors.Wrap(err, "Failed to read config file")
	}
	defer yamlReader.Close()
	decoder := yaml.NewDecoder(yamlReader)

	if err = decoder.Decode(c); err != nil {
		return errors.Wrap(err, "Failed to parse config file")
	}

	sc.Lock()
	sc.C = c
	sc.Unlock()

	return nil
}
