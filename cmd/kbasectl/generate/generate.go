package generate

import (
	"io/ioutil"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	from string
)

type Config struct {
}

func NewGenerateComand() *cobra.Command {
	command := &cobra.Command{
		Use: "gen",

		RunE: run,
	}

	command.Flags().StringVarP(&from, "from", "f", "", "deployment config that used to gen kubernetes config")
	command.MarkFlagFilename("from", "yaml")
	command.MarkFlagRequired("from")

	return command
}

func run(cmd *cobra.Command, args []string) error {
	raw, err := ioutil.ReadFile(from)
	if err != nil {
		return err
	}

	config := Config{}
	if err := yaml.Unmarshal(raw, &config); err != nil {
		return err
	}

	return nil
}
