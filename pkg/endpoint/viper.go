package endpoint

import (
	"github.com/spf13/viper"
	"github.com/zizon/kbasectl/pkg/panichain"
)

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.kbasectl")
	viper.AddConfigPath(".")

	panichain.Propogate(viper.ReadInConfig())
}
