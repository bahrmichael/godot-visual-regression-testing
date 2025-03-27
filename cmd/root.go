package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var GodotExecutable string
var Verbose bool
var ScenesGlob string
var ProjectPath string

func init() {
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
}

var rootCmd = &cobra.Command{
	Use:   "godot-vrt",
	Short: "Godot Visual Regression Testing (VRT) helps you detect visual regressions in your Godot scenes",
	//Long:  `Long description`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
