package main

import (
	"github.com/spf13/cobra"
	"github.com/zizon/kbasectl/cmd/kbasectl/generate"
	"github.com/zizon/kbasectl/pkg/panichain"
)

func main() {
	rootCmd := &cobra.Command{
		Use: "kbasectl",
	}

	rootCmd.AddCommand(
		generate.NewGenerateComand(),
		generate.NewTemplateCommand(),
	)

	panichain.Propogate(rootCmd.Execute())
}
