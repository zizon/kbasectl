package generator

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/google/uuid"
	"github.com/zizon/kbasectl/pkg/endpoint"
	"github.com/zizon/kbasectl/pkg/panichain"
	"gopkg.in/yaml.v2"
)

func TestGenerationPrint(t *testing.T) {
	config := Config{
		Namespace: "test-namespae",
		Name:      "test-pod",

		Labels: map[string]string{
			"kbase-app": "some app",
		},
		Envs: map[string]string{
			"RUNTIME": "cri-o",
		},

		RunAS:      10086,
		Entrypoint: "/bin/bash /work/entrypoint",
		WorkDir:    "/work",

		Image:     "docker.io/golang:v15.2",
		CPU:       1,
		MemoryMb:  10,
		IngressMb: 1,
		EgressMb:  1,

		Replica: 1,

		CephBinds: []CephBind{
			{},
		},
		LocalBinds: []LocalBind{
			{
				Source:   "/etc",
				ReadOnly: true,

				MountTo: "/host/etc",
				SubPaths: []string{
					"hosts",
				},
			},
		},
		MemoryBinds: []MemoryBind{
			{
				CapacityMb: 12,
				MountTo:    "/dev/shm/writable",
			},
		},

		ConfigBind: ConfigMapBind{
			MountTo: "/configmap",
		},
	}

	cephConfig := map[string]CephConfig{}
	for i := 0; i < 10; i++ {
		cephConfig[fmt.Sprintf("10.116.100.%d", i)] = CephConfig{
			User:           fmt.Sprintf("user-%d", i%2),
			TokenName:      fmt.Sprintf("token-%d", i%2),
			TokenNamespace: fmt.Sprintf("ceph-ns-%d", i%2),
		}
	}
	config = RewriteWithCepConfig(config, cephConfig)

	//fill with some
	config = fillCephConfig(config)
	config = RewriteWithCepConfig(config, cephConfig)
	config = randomCutSomeCeph(config)

	// file configmap
	config = fillLocalCongfigFile(config)
	config = RewriteWithLocalConfigFiles(config, map[string]string{
		"/etc/kbasectl.yaml":         endpoint.DefaultConfigFile,
		"/home/nobody/kbasectl.yaml": endpoint.DefaultConfigFile,
	})

	docs := []interface{}{}
	for _, obj := range GenerateNamespace(config) {
		docs = append(docs, obj)
	}

	for _, obj := range GenerateSecret(config) {
		docs = append(docs, obj)
	}

	docs = append(docs, GenerateConfigMap(config))

	for _, obj := range GeneratePersistenVolume(config) {
		docs = append(docs, obj)
	}

	for _, obj := range GeneratePersistenVolumeClaim(config) {
		docs = append(docs, obj)
	}

	docs = append(docs, GenerateDeployment(config))

	t.Logf("\n%s", printYaml(docs))
}

func fillCephConfig(config Config) Config {
	binds := []CephBind{}
	for i := 0; i < 20; i++ {
		seed := []string{fmt.Sprintf("10.116.100.%d", i)}

		for j := 0; j < 5; j++ {
			seed = append(seed, uuid.New().String())
		}

		binds = append(binds, CephBind{
			Monitors:       seed,
			User:           fmt.Sprintf("user-%d", i%2),
			TokenName:      fmt.Sprintf("token-%d", i%2),
			TokenNamespace: fmt.Sprintf("ceph-ns-%d", i%2),

			Source: fmt.Sprintf("/ceph-root/%s", uuid.New().String()),

			ReadOnly:   i%3 == 0,
			CapacityMb: int(rand.Int31()),

			MountTo: fmt.Sprintf("/mnt/ceph/mounted-%s", uuid.New().String()),
			SubPaths: []string{
				"lib", "bin", "$(NAMESPCACE)/logs",
			},
		})
	}
	config.CephBinds = binds
	return config
}

func randomCutSomeCeph(config Config) Config {
	binds := config.CephBinds
	rand.Shuffle(len(binds), func(i, j int) {
		binds[i], binds[j] = binds[j], binds[i]
	})

	if len(binds) > 3 {
		config.CephBinds = binds[:3]
	}
	return config
}

func fillLocalCongfigFile(config Config) Config {
	config.ConfigBind = ConfigMapBind{
		MountTo: "/configmap",
		ConfigMap: map[string]string{
			"plaintext": "hello kitty",
		},
	}

	return config
}

func printYaml(obj interface{}) string {
	out, err := yaml.Marshal(obj)
	panichain.Propogate(err)
	return string(out)
}
