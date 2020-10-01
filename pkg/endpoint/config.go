package endpoint

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/zizon/kbasectl/pkg/panichain"
	"gopkg.in/yaml.v2"
	"k8s.io/client-go/rest"
)

var (
	HOME, _           = os.UserHomeDir()
	DefaultConfigFile = path.Join(HOME, ".kbasectl", "config.yaml")
)

type Config struct {
	Rest rest.Config
	Ceph Ceph
}

type Ceph struct {
	Monitors []string
	User     string
	Token    string
}

func NewDefaultConfig() Config {
	config := Config{}

	raw, err := ioutil.ReadFile(DefaultConfigFile)
	panichain.Propogate(err)

	err = yaml.Unmarshal(raw, &config)
	panichain.Propogate(err)

	return config
}
