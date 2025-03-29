package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"godot-vrt/lib"

	"github.com/spf13/cobra"
)

var BaselineGlob string

func init() {
	rootCmd.AddCommand(testCmd)

	testCmd.Flags().StringVarP(&GodotExecutable, "godot", "g", "", "path to the godot executable (e.g. /usr/local/bin/godot)")
	testCmd.MarkFlagRequired("godot")

	testCmd.Flags().StringVarP(&ScenesGlob, "scenes", "s", "", "glob path to the .tscn files (e.g. scenes-vrt/*.tscn)")
	testCmd.MarkFlagRequired("scenes")

	testCmd.Flags().StringVarP(&BaselineGlob, "baseline", "b", "", "glob path to the baseline .avi files (e.g. scenes-vrt/*.avi)")
	testCmd.MarkFlagRequired("baseline")

	testCmd.Flags().StringVarP(&ProjectPath, "project", "p", "", "path to the project root (only required if you run godot-vrt from a different directory)")
	testCmd.Flags().IntVarP(&Frames, "frames", "f", 60, "number of frames to render")
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Runs visual regression testing by rendering scenes and comparing them to their baselines",
	//Long:  `Test long description`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.Validate(GodotExecutable, ProjectPath); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		hasFailures, err := testScenes(ScenesGlob, BaselineGlob, GodotExecutable, "vrt-results", Frames, Verbose)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if hasFailures {
			fmt.Println("❌ One or more tests failed")
			// todo: document error codes (available range is 1-127: we use 1 for generic errors, and 50 for test failures)
			os.Exit(50)
		}
		fmt.Println("✅ All tests passed")
	},
}

func testScenes(scenes, baseline, godotBinary, resultDir string, frames int, verbose bool) (bool, error) {
	// list all sceneFiles at config.Scenes (that's a glob)
	sceneFiles, err := filepath.Glob(scenes)
	if err != nil {
		return false, fmt.Errorf("error listing sceneFiles: %v", err)
	}
	baselineFiles, err := filepath.Glob(baseline)
	if err != nil {
		return false, fmt.Errorf("error listing sceneFiles: %v", err)
	}

	// there must be a baseline for each scene if we're in test mode
	var missingBaselines []string
	for _, file := range sceneFiles {
		target := strings.Replace(file, ".tscn", ".avi", 1)
		if !slices.Contains(baselineFiles, target) {
			missingBaselines = append(missingBaselines, file)
		}
	}
	if len(missingBaselines) > 0 {
		return false, fmt.Errorf("missing baselines for scenes: %v", missingBaselines)
	}

	//renderDir, err := os.MkdirTemp(config.TmpDir, "renders_")
	renderDir, err := os.MkdirTemp("", ".vrt")
	if err != nil {
		return false, fmt.Errorf("error creating temp dir: %v", err)
	}
	if !strings.HasSuffix(renderDir, "/") {
		renderDir = renderDir + "/"
	}
	//if !config.RetainRenderDir {
	defer os.RemoveAll(renderDir)
	//}

	// for each file, render it
	defaultArgs := []string{
		"--quit-after",
		strconv.Itoa(frames),
		//"--fixed-fps",
		//strconv.Itoa(fps),
	}

	foundDiff := false

	for _, file := range sceneFiles {
		sceneName := strings.Replace(file, ".tscn", "", 1)

		actualPathFile := fmt.Sprintf("%s%s%s", renderDir, sceneName, "_actual.avi")
		err := os.MkdirAll(filepath.Dir(actualPathFile), 0755)
		if err != nil {
			return false, fmt.Errorf("error creating dir: %v", err)
		}
		renderedScene, err := lib.RenderScene(file, actualPathFile, defaultArgs, verbose, godotBinary)
		if err != nil {
			return false, fmt.Errorf("error rendering file: %v", err)
		}

		baseline := fmt.Sprintf("%s%s", sceneName, ".avi")
		diffOutFile := fmt.Sprintf("%s%s%s", renderDir, sceneName, "_diff.avi")
		err = os.MkdirAll(filepath.Dir(diffOutFile), 0755)
		if err != nil {
			return false, fmt.Errorf("error creating dir: %v", err)
		}
		h, err := lib.HasDiff(renderedScene, baseline, diffOutFile, verbose, frames)
		if h {
			foundDiff = true
			d, err := lib.GenerateComparison(sceneName, renderedScene, baseline, resultDir, verbose)
			if err != nil {
				return false, fmt.Errorf("error generating comparison: %v", err)
			}
			fmt.Println(d)
		}
	}

	return foundDiff, nil
}
