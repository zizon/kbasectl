package generate

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/zizon/kbasectl/pkg/panichain"
	"gopkg.in/yaml.v2"
)

func TestGenerate(t *testing.T) {
	configDir, err := ioutil.TempDir("", "test-generate-cmd")
	panichain.Propogate(err)
	defer os.RemoveAll(configDir)

	configFile := path.Join(configDir, "deploy.yaml")
	t.Logf("config file:%s", configFile)
	config, err := yaml.Marshal(Config{
		Namespace: "test-nampespace",
		Name:      "test-pod",

		Labels: map[string]string{
			"kbase-app": "test",
		},
		Envs: map[string]string{
			"hello": "kitty",
		},

		RunAS:      10086,
		Entrypoint: "/bin/bash /work/start.sh",
		WorkDir:    "/work",

		Image:     "docker.io/golang:v15.2",
		CPU:       1,
		MemoryMb:  10,
		IngressMb: 1,
		EgressMb:  1,

		Replica: 1,

		ConfigFiles: []FileMap{
			{
				From:     configFile,
				MapToKey: "/work/config/origin.yaml",
			},

			{
				From:     configFile,
				MapToKey: "/work/origin.yaml",
			},
		},
		CephBind: []CephBind{
			{
				From: "/template",
				To:   "/work",

				ReadOnly:   false,
				CapacityMb: 1024,

				Filter: []string{
					"bin", "sbin",
				},
			},

			{
				From: "/template/logs",
				To:   "/work/logs",

				ReadOnly:   true,
				CapacityMb: 1024,

				Filter: []string{
					"local", "$(CONTAINER_NAMESPACE)", "$(CONTAINER_ID)",
				},
			},
		},
	})
	panichain.Propogate(err)
	panichain.Propogate(ioutil.WriteFile(configFile, config, os.FileMode(777)))

	cmd := NewGenerateComand()

	cmd.SetArgs([]string{
		"-f", configFile,
	})

	panichain.Propogate(cmd.Execute())
}
