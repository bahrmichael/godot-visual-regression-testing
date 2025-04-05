package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"godot-vrt/lib"
)

func init() {
	RootCmd.AddCommand(baselineCmd)

	baselineCmd.Flags().StringVarP(&GodotExecutable, "godot", "g", "", "path to the godot executable (e.g. /usr/local/bin/godot)")
	baselineCmd.MarkFlagRequired("godot")

	baselineCmd.Flags().StringVarP(&ScenesGlob, "scenes", "s", "", "glob path to the .tscn files, relative from the godot project root (e.g. scenes-vrt/*.tscn)")
	baselineCmd.MarkFlagRequired("scenes")

	baselineCmd.Flags().StringVarP(&ProjectPath, "project", "p", ".", "path to the project root (only required if you run godot-vrt from a different directory)")
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
			if !OmitExitCode {
				os.Exit(1)
			}
		}

		err := renderScenes()
		if err != nil {
			fmt.Println(err)
			if !OmitExitCode {
				os.Exit(1)
			}
		}
	},
}

func renderScenes() error {
	// list all sceneFiles at config.Scenes (that's a glob)
	sceneFiles, err := filepath.Glob(ProjectPath + ScenesGlob)
	if err != nil {
		return fmt.Errorf("failed to list files at %s: %v", ProjectPath+ScenesGlob, err)
	}
	if len(sceneFiles) == 0 {
		return fmt.Errorf("search for files at %s yielded 0 results", ProjectPath+ScenesGlob)
	}

	for _, file := range sceneFiles {
		f, err := filepath.Abs(file)
		if err != nil {
			return fmt.Errorf("error getting absolute path: %v", err)
		}
		b, err := lib.RenderScene(lib.RenderSceneArgs{
			SceneFileFromProjectRoot: strings.Replace(file, ProjectPath, "", 1),
			OutputFile:               strings.Replace(f, ".tscn", ".avi", 1),
			GodotBinary:              GodotExecutable,
			Verbose:                  Verbose,
			Frames:                   Frames,
			ProjectPath:              ProjectPath,
		})
		if err != nil {
			return fmt.Errorf("error rendering file: %v", err)
		}
		fmt.Println("Rendered baseline: " + b)
	}
	return nil
}
