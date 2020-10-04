package generate

import (
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zizon/kbasectl/cmd/kbasectl/defaults"
	"github.com/zizon/kbasectl/pkg/endpoint"
	"github.com/zizon/kbasectl/pkg/generator"
	"github.com/zizon/kbasectl/pkg/panichain"
	"gopkg.in/yaml.v2"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Config struct {
	Namespace string
	Name      string

	Labels map[string]string
	Envs   map[string]string

	RunAS      int
	Entrypoint string
	WorkDir    string

	Image     string
	CPU       int
	MemoryMb  int
	IngressMb int
	EgressMb  int

	ConfigFiles []FileMap

	Replica int

	CephBind []CephBind
}

type FileMap struct {
	From     string
	MapToKey string
}

type CephBind struct {
	From string
	To   string

	ReadOnly   bool
	CapacityMb int

	Filter []string
}

var (
	from string
)

func NewGenerateComand() *cobra.Command {
	command := &cobra.Command{
		Use: "gen",

		Short: "generate necessary kubernates configs to actiave deployment specified by kbasectl deploy config",
		Run: func(cmd *cobra.Command, args []string) {
			run()
		},
	}

	command.Flags().StringVarP(&from, "from", "f", "", "deployment config that used to gen kubernetes config")
	command.MarkFlagFilename("from", "yaml")
	command.MarkFlagRequired("from")

	return command
}

func run() {
	// load kbasectl config
	panichain.Propogate(viper.ReadInConfig())
	clientConfig := endpoint.Config{}
	if err := viper.Unmarshal(&clientConfig); err != nil {
		panichain.Propogate(err)
		return
	}

	// load deployment config
	raw, err := ioutil.ReadFile(from)
	if err != nil {
		panichain.Propogate(err)
		return
	}
	fromConfig := Config{}
	if err := yaml.Unmarshal(raw, &fromConfig); err != nil {
		panichain.Propogate(err)
		return
	}

	// generator config
	generatorConfig := convertToGeneratorConfig(fromConfig, clientConfig)

	// collect namespace
	objects := []runtime.Object{}
	for _, ns := range generator.GenerateNamespace(generatorConfig) {
		objects = append(objects, func(ns v1.Namespace) runtime.Object {
			return &ns
		}(ns))
	}

	// collect secrets
	for _, secret := range generator.GenerateSecret(generatorConfig) {
		objects = append(objects, func(secret v1.Secret) runtime.Object {
			secret.StringData[clientConfig.Ceph.User] = clientConfig.Ceph.Token
			return &secret
		}(secret))
	}

	// collect config map
	objects = append(objects, func(configMap v1.ConfigMap) runtime.Object {
		return &configMap
	}(generator.GenerateConfigMap(generatorConfig)))

	// collect pv
	for _, pv := range generator.GeneratePersistenVolume(generatorConfig) {
		objects = append(objects, func(secret v1.PersistentVolume) runtime.Object {
			return &pv
		}(pv))
	}

	// collect pbv
	for _, pvc := range generator.GeneratePersistenVolumeClaim(generatorConfig) {
		objects = append(objects, func(secret v1.PersistentVolumeClaim) runtime.Object {
			return &pvc
		}(pvc))
	}

	// collect deployment
	objects = append(objects, func(deployment appv1.Deployment) runtime.Object {
		return &deployment
	}(generator.GenerateDeployment(generatorConfig)))

	os.Stdout.Write(defaults.KubernatesObjectsToYaml(objects))
	return
}

func convertToGeneratorConfig(config Config, clientConfig endpoint.Config) generator.Config {
	generatorConfig := generator.Config{
		Namespace: config.Namespace,
		Name:      config.Name,

		Labels: config.Labels,
		Envs:   config.Envs,

		RunAS:      config.RunAS,
		Entrypoint: config.Entrypoint,
		WorkDir:    config.WorkDir,

		Image:     config.Image,
		CPU:       config.CPU,
		MemoryMb:  config.MemoryMb,
		IngressMb: config.IngressMb,
		EgressMb:  config.EgressMb,

		Replica: config.Replica,

		CephBinds: convertCephBind(config.CephBind, clientConfig.Ceph),
	}

	// rewrite ceph
	ctlCephConfig := map[string]generator.CephConfig{}
	for _, monitor := range clientConfig.Ceph.Monitors {
		ctlCephConfig[monitor] = generator.CephConfig{
			User:           clientConfig.Ceph.User,
			TokenName:      defaults.KbaseCephTokenName,
			TokenNamespace: defaults.KbaseCTLNamespace,
		}
	}
	generatorConfig = generator.RewriteWithCepConfig(generatorConfig, ctlCephConfig)

	// rewrite config files
	fileMaps := map[string]string{}
	for _, fileMap := range config.ConfigFiles {
		fileMaps[fileMap.MapToKey] = fileMap.From
	}
	generatorConfig = generator.RewriteWithLocalConfigFiles(generatorConfig, fileMaps)

	return generatorConfig
}

func convertCephBind(binds []CephBind, ceph endpoint.Ceph) []generator.CephBind {
	generatorCephBinds := []generator.CephBind{}
	for _, bind := range binds {
		generatorCephBinds = append(generatorCephBinds, generator.CephBind{
			Monitors: ceph.Monitors,
			User:     ceph.User,

			TokenName:      defaults.KbaseCephTokenName,
			TokenNamespace: defaults.KbaseCTLNamespace,

			Source: bind.From,

			ReadOnly:   bind.ReadOnly,
			CapacityMb: bind.CapacityMb,

			MountTo:  bind.To,
			SubPaths: bind.Filter,
		})
	}

	return generatorCephBinds
}
