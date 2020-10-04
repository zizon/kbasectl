package generate

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zizon/kbasectl/pkg/endpoint"
	"github.com/zizon/kbasectl/pkg/panichain"
	"gopkg.in/yaml.v2"
)

func NewTemplateCommand() *cobra.Command {
	command := &cobra.Command{
		Use: "template",

		Short: "generate a template deploy config",

		Run: func(cmd *cobra.Command, args []string) {
			template()
		},
	}

	return command
}

func template() {
	// load kbasectl config
	panichain.Propogate(viper.ReadInConfig())
	clientConfig := endpoint.Config{}
	if err := viper.Unmarshal(&clientConfig); err != nil {
		panichain.Propogate(err)
		return
	}

	if out, err := yaml.Marshal(&Config{
		Labels: map[string]string{
			"label-key": "label-value",
		},
		Envs: map[string]string{
			"env-key": "env-value",
		},

		ConfigFiles: []FileMap{
			{
				From:     "/file/path/from/local/fs.file",
				MapToKey: "filename.that.mounted.into.container",
			},
		},
		CephBind: []CephBind{
			{
				From:       "/cephfs/path",
				To:         "/mounted/to/container/path",
				ReadOnly:   true,
				CapacityMb: 65536,
				Filter: []string{
					"only/this/files/under/from/will/be/mounted/inside/container",
					"only/this/directory/under/from/will/be/mounted/inside/container",
				},
			},
		},
	}); err != nil {
		panichain.Propogate(err)
	} else if _, err := os.Stdout.Write(out); err != nil {
		panichain.Propogate(err)
	}
}
