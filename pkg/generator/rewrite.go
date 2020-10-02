package generator

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/zizon/kbasectl/pkg/panichain"
)

type CephConfig struct {
	User           string
	TokenName      string
	TokenNamespace string
}

func RewriteWithCepConfig(config Config, monitorToConfig map[string]CephConfig) Config {
	configToMonitors := map[string][]string{}
	for monitor, monitorConfig := range monitorToConfig {
		lookupKey := fmt.Sprintf("namesapce:%s name:%s user:%s",
			monitorConfig.TokenNamespace,
			monitorConfig.TokenName,
			monitorConfig.User,
		)
		if monitors, exists := configToMonitors[lookupKey]; exists {
			configToMonitors[lookupKey] = append(monitors, monitor)
		} else {
			configToMonitors[lookupKey] = []string{monitor}
		}
	}

	newBinds := []CephBind{}
	for _, bindConfig := range config.CephBinds {
		configs := map[string]CephConfig{}

		// refresh monitor configs
		for _, monitor := range bindConfig.Monitors {
			standardName := strings.TrimSpace(monitor)
			if monitorConfig, exits := monitorToConfig[standardName]; exits {
				lookupKey := fmt.Sprintf("namesapce:%s name:%s user:%s",
					monitorConfig.TokenNamespace,
					monitorConfig.TokenName,
					monitorConfig.User,
				)
				configs[lookupKey] = monitorConfig
			}
		}

		// expect only one config
		switch len(configs) {
		case 0:
			continue
		case 1:
			// refresh with new one.
			// may be a copy?
			for lookupKey, monitorConfig := range configs {
				bindConfig.Monitors = configToMonitors[lookupKey]
				bindConfig.TokenName = monitorConfig.TokenName
				bindConfig.TokenNamespace = monitorConfig.TokenNamespace
				bindConfig.User = monitorConfig.User
				newBinds = append(newBinds, bindConfig)
				continue
			}
		default:
			panichain.Propogate(fmt.Errorf("monitors:%v had different configs:%v", bindConfig.Monitors, configs))
		}
	}

	config.CephBinds = newBinds
	return config
}

func RewriteWithLocalConfigFiles(config Config, localFileMap map[string]string) Config {
	if config.ConfigBind.MountTo == "" {
		config.ConfigBind.MountTo = "/"
	}

	if config.ConfigBind.ConfigMap == nil {
		config.ConfigBind.ConfigMap = map[string]string{}
	}

	fileToContent := config.ConfigBind.ConfigMap
	for mapTo, local := range localFileMap {
		content, err := ioutil.ReadFile(local)
		panichain.Propogate(err)

		fileToContent[mapTo] = string(content)
	}

	return config
}
