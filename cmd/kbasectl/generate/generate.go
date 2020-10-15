package generate

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zizon/kbasectl/cmd/kbasectl/defaults"
	"github.com/zizon/kbasectl/pkg/context"
	"github.com/zizon/kbasectl/pkg/endpoint"
	"github.com/zizon/kbasectl/pkg/generator"
	"github.com/zizon/kbasectl/pkg/panichain"
	"gopkg.in/yaml.v2"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog"
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

	UseContainerNetwork bool

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
	from    string
	minimal bool
)

func NewGenerateComand() *cobra.Command {
	command := &cobra.Command{
		Use: "gen",

		Short: "generate necessary kubernates configs to actiave deployment specified by kbasectl deploy config",
		Run: func(cmd *cobra.Command, args []string) {
			generate()
		},
	}

	command.Flags().StringVarP(&from, "from", "f", "", "deployment config that used to gen kubernetes config")
	command.MarkFlagFilename("from", "yaml")
	command.MarkFlagRequired("from")

	command.Flags().BoolVarP(&minimal, "minimal", "m", false, "whether or not to output configs that already exists in kubernets cluster")

	return command
}

func generate() {
	// load kbasectl config
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

	client := endpoint.NewAPIClient(endpoint.NewDefaultConfig())
	batcher := context.NewBatcher()
	// collect pv
	for _, pv := range generator.GeneratePersistenVolume(generatorConfig) {
		local := pv
		batcher.Add(func() interface{} {
			if existsInKubernetes(client, &local) {
				return nil
			}

			return local
		})
	}

	// collect pbv
	for _, pvc := range generator.GeneratePersistenVolumeClaim(generatorConfig) {
		local := pvc
		batcher.Add(func() interface{} {
			if existsInKubernetes(client, &local) {
				return nil
			}

			return local
		})
	}

	// collect deployment
	objects = append(objects, func(deployment appv1.Deployment) runtime.Object {
		return &deployment
	}(generator.GenerateDeployment(generatorConfig)))

	filtered := objects
	if minimal {
		filtered = []runtime.Object{}
		for _, obj := range objects {
			local := obj
			batcher.Add(func() interface{} {
				if existsInKubernetes(client, local) {
					return nil
				}
				return local
			})
		}
	}

	for obj := range batcher.Join() {
		if obj != nil {
			filtered = append(filtered, obj.(runtime.Object))
		}
	}

	os.Stdout.Write(defaults.KubernatesObjectsToYaml(filtered))
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

		HostNetwork: !config.UseContainerNetwork,

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

func existsInKubernetes(client endpoint.Client, obj runtime.Object) bool {
	value := reflect.ValueOf(obj).Elem()
	metaField := value.FieldByName("ObjectMeta").Interface().(metav1.ObjectMeta)

	gvk := obj.GetObjectKind().GroupVersionKind()

	urlPath := path.Join("/",
		apiPath(gvk),
		gvk.Group,
		gvk.Version)
	if metaField.Namespace != "" {
		urlPath = path.Join(urlPath,
			"namespaces",
			metaField.Namespace,
			apiKind(gvk),
			metaField.Name,
		)
	} else {
		urlPath = path.Join(urlPath,
			apiKind(gvk),
			metaField.Name,
		)
	}

	resp := client.Do(http.Request{
		Method: "GET",
		URL: &url.URL{
			Path: urlPath,
		},
	})
	if resp.StatusCode == http.StatusOK {
		return true
	}
	klog.Errorf("bad response:%v", resp)

	return false
}

func apiKind(gvk schema.GroupVersionKind) string {
	return fmt.Sprintf("%ss", strings.ToLower(gvk.Kind))
}

func apiPath(gvk schema.GroupVersionKind) string {
	switch gvk.Kind {
	case "Deployment":
		return "apis"
	default:
		return "api"
	}
}
