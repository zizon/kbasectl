package endpoint

import (
	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.kbasectl")
	viper.AddConfigPath(".")
}
