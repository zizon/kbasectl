package main

import (
	"github.com/spf13/cobra"
	"github.com/zizon/kbasectl/cmd/kbasectl/generate"
	"github.com/zizon/kbasectl/pkg/endpoint"
	"github.com/zizon/kbasectl/pkg/panichain"
)

func main() {
	var (
		logLevel = 0
	)
	rootCmd := &cobra.Command{
		Use: "kbasectl",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			endpoint.SetLogLevel(logLevel)
		},
	}

	rootCmd.PersistentFlags().IntVarP(&logLevel, "verbose", "v", 0, "log verbose level")

	rootCmd.AddCommand(
		generate.NewGenerateComand(),
		generate.NewTemplateCommand(),
	)

	panichain.Propogate(rootCmd.Execute())
}
