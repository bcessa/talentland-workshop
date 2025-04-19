package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/bcessa/echo-service/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.bryk.io/pkg/cli"
	viperUtils "go.bryk.io/pkg/cli/viper"
)

var versionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"info"},
	Short:   "Show version information",
	Run: func(_ *cobra.Command, _ []string) {
		details := internal.BuildDetails()

		// print as JSON
		if viper.GetBool("version.json") {
			js, _ := json.MarshalIndent(details, "", "  ")
			fmt.Printf("%s\n", js)
			return
		}

		for k, v := range details.Values() {
			if v != "" {
				fmt.Printf("\033[1m%s\033[0m: %s\n", k, v)
			}
		}
	},
}

func init() {
	params := []cli.Param{
		{
			Name:      "json",
			Usage:     "output version information in JSON format",
			FlagKey:   "version.json",
			ByDefault: false,
		},
	}
	if err := cli.SetupCommandParams(versionCmd, params); err != nil {
		panic(err)
	}
	if err := viperUtils.BindFlags(versionCmd, params, viper.GetViper()); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(versionCmd)
}
