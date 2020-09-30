package endpoint

import (
	"io/ioutil"
	"os"
	"path"

	"gopkg.in/yaml.v2"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

var (
	HOME, nil         = os.UserHomeDir()
	DefaultConfigFile = path.Join(HOME, ".kbasectl", "config.yaml")
)

type Config struct {
	Rest rest.Config
}

func NewDefaultConfig() (Config, error) {
	config := Config{}

	raw, err := ioutil.ReadFile(DefaultConfigFile)
	if err != nil {
		klog.Errorf("fail to read default config: %s reason: %v", DefaultConfigFile, err)
		return config, err
	}

	if err := yaml.Unmarshal(raw, &config); err != nil {
		klog.Errorf("fail to decode config:%s reason: %v", DefaultConfigFile, err)
	}

	return config, nil
}
