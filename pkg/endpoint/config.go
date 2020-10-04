package endpoint

import (
	"github.com/spf13/viper"
	"github.com/zizon/kbasectl/pkg/panichain"
	"k8s.io/client-go/rest"
)

var (
	Namespace = "kbase"
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

	panichain.Propogate(viper.Unmarshal(&config))

	return config
}
