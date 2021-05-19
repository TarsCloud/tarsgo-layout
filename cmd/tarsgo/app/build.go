package app

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.tarsctl.yaml)")
}

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the server",
	Long:  `Build the server`,
	Run: func(cmd *cobra.Command, args []string) {
		// check tars2go
		// gen tars file
		// go build
		// do tgz
	},
}
