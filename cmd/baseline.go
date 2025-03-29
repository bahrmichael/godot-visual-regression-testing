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

	baselineCmd.Flags().StringVarP(&ProjectPath, "project", "p", "", "path to the project root (only required if you run godot-vrt from a different directory)")
	baselineCmd.Flags().IntVarP(&Frames, "frames", "f", 60, "number of frames to render (default 60)")

	// FPS only affects the fps of the video, but not the speed at which we render it. Speed
	// might only be affected by the hardware speed.
	//baselineCmd.Flags().IntVarP(&FPS, "fps", "", 10, "frames per second")
}

var baselineCmd = &cobra.Command{
	Use:   "baseline",
	Short: "Renders scenes and saves them as baseline .avi files",
	//Long:  `Baseline long description`,
	Args: func(cmd *cobra.Command, args []string) error {
		if Frames < 1 {
			fmt.Println("Frames must be greater than 0")
			os.Exit(1)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {

		if err := lib.Validate(GodotExecutable, ProjectPath); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		err := renderScenes(Verbose, ScenesGlob, GodotExecutable, Frames)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func renderScenes(verbose bool, scenes, godotBinary string, frames int) error {
	// list all sceneFiles at config.Scenes (that's a glob)
	sceneFiles, err := filepath.Glob(scenes)
	if err != nil {
		return fmt.Errorf("error listing sceneFiles: %v", err)
	}
	// for each file, render it
	defaultArgs := []string{
		"--quit-after",
		strconv.Itoa(frames),
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
