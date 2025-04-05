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
var Frames int
var OmitExitCode bool

func init() {
	RootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	RootCmd.PersistentFlags().BoolVar(&OmitExitCode, "omit-exit-code", false, "always exit with 0 despite failures")
}

var RootCmd = &cobra.Command{
	Use:   "godot-vrt",
	Short: "Godot Visual Regression Testing (VRT) helps you detect visual regressions in your Godot scenes",
	//Long:  `Long description`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		if !OmitExitCode {
			os.Exit(1)
		}
	}
}
