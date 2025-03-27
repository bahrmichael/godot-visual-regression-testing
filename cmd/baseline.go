package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"godot-vrt/lib"
)

func init() {
	rootCmd.AddCommand(baselineCmd)

	baselineCmd.Flags().StringVarP(&GodotExecutable, "godot", "g", "", "path to the godot executable (e.g. /usr/local/bin/godot)")
	baselineCmd.MarkFlagRequired("godot")

	baselineCmd.Flags().StringVarP(&ScenesGlob, "scenes", "s", "", "glob path to the .tscn files (e.g. scenes-vrt/*.tscn)")
	baselineCmd.MarkFlagRequired("scenes")

	baselineCmd.Flags().StringVarP(&ProjectPath, "project", "p", "", "path to the project root (only required if vrt is run from a different directory))")
}

var baselineCmd = &cobra.Command{
	Use:   "baseline",
	Short: "Renders scenes and saves them as baseline .avi files",
	//Long:  `Baseline long description`,
	Args: func(cmd *cobra.Command, args []string) error {

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {

		if err := lib.Validate(GodotExecutable, ProjectPath); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		err := renderScenes(Verbose, ScenesGlob, GodotExecutable, 60)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func renderScenes(verbose bool, scenes, godotBinary string, duration int) error {
	// list all sceneFiles at config.Scenes (that's a glob)
	sceneFiles, err := filepath.Glob(scenes)
	if err != nil {
		return fmt.Errorf("error listing sceneFiles: %v", err)
	}
	// for each file, render it
	defaultArgs := []string{
		"--quit-after",
		strconv.Itoa(duration),
	}
	for _, file := range sceneFiles {
		b, err := lib.RenderScene(file, strings.Replace(file, ".tscn", ".avi", 1), defaultArgs, verbose, godotBinary)
		if err != nil {
			return fmt.Errorf("error rendering file: %v", err)
		}
		fmt.Println("Rendered baseline: " + b)
	}
	return nil
}
