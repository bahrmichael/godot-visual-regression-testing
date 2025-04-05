package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"godot-vrt/lib"

	"github.com/spf13/cobra"
)

var BaselineGlob string
var RetainAssets bool

func init() {
	RootCmd.AddCommand(testCmd)

	testCmd.Flags().StringVarP(&GodotExecutable, "godot", "g", "", "path to the godot executable (e.g. /usr/local/bin/godot)")
	testCmd.MarkFlagRequired("godot")

	testCmd.Flags().StringVarP(&ScenesGlob, "scenes", "s", "", "glob path to the .tscn files, relative from the godot project root (e.g. scenes-vrt/*.tscn)")
	testCmd.MarkFlagRequired("scenes")

	testCmd.Flags().StringVarP(&BaselineGlob, "baseline", "b", "", "glob path to the baseline .avi files, relative from the godot project root (e.g. scenes-vrt/*.avi)")
	testCmd.MarkFlagRequired("baseline")

	testCmd.Flags().StringVarP(&ProjectPath, "project", "p", ".", "path to the project root (only required if you run godot-vrt from a different directory)")
	testCmd.Flags().IntVarP(&Frames, "frames", "f", 60, "number of frames to render")
	testCmd.Flags().BoolVar(&RetainAssets, "retain-assets", false, "will keep the test videos around if set to true (useful for debugging why a test didn't fail)")
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Runs visual regression testing by rendering scenes and comparing them to their baselines",
	//Long:  `Test long description`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.Validate(GodotExecutable, ProjectPath); err != nil {
			fmt.Println(err)
			if !OmitExitCode {
				os.Exit(1)
			}
		}

		hasFailures, err := testScenes()
		if err != nil {
			fmt.Println(err)
			if !OmitExitCode {
				os.Exit(1)
			}
		}
		if hasFailures {
			fmt.Println("❌ One or more tests failed")
			// todo: document error codes (available range is 1-127: we use 1 for generic errors, and 50 for test failures)
			if !OmitExitCode {
				os.Exit(50)
			}
			return
		}
		fmt.Println("✅ All tests passed")
	},
}

func testScenes() (bool, error) {
	// list all sceneFiles at config.Scenes (that's a glob)
	sceneFiles, err := filepath.Glob(ProjectPath + ScenesGlob)
	if err != nil {
		return false, fmt.Errorf("error listing sceneFiles: %v", err)
	}
	baselineFiles, err := filepath.Glob(ProjectPath + BaselineGlob)
	if err != nil {
		return false, fmt.Errorf("error listing sceneFiles: %v", err)
	}

	tmpDir, cleanupTmpDir := lib.InitTmpDir()
	if !RetainAssets {
		defer cleanupTmpDir()
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

	foundDiff := false

	for _, file := range sceneFiles {
		sceneName := strings.Replace(file, ".tscn", "", 1)

		actualPathFile := fmt.Sprintf("%s%s%s", tmpDir, sceneName, "_actual.avi")
		renderedScene, err := lib.RenderScene(lib.RenderSceneArgs{
			SceneFileFromProjectRoot: strings.Replace(file, lib.WithFolderSuffix(ProjectPath), "", 1),
			OutputFile:               actualPathFile,
			GodotBinary:              GodotExecutable,
			Verbose:                  Verbose,
			Frames:                   Frames,
			ProjectPath:              ProjectPath,
		})
		if err != nil {
			return false, fmt.Errorf("error rendering file: %v", err)
		}

		baseline, err := filepath.Abs(fmt.Sprintf("%s%s", sceneName, ".avi"))
		if err != nil {
			return false, fmt.Errorf("error getting absolute path: %v", err)
		}
		diffOutFile := fmt.Sprintf("%s%s%s", tmpDir, sceneName, "_diff.avi")
		hasDiff, err := lib.HasDiff(renderedScene, baseline, diffOutFile, Verbose, Frames)
		if err != nil {
			return false, fmt.Errorf("error generating diff: %v", err)
		}
		if hasDiff {
			foundDiff = true
			d, err := lib.GenerateComparison(sceneName, renderedScene, baseline, "vrt-results/", Verbose)
			if err != nil {
				return false, fmt.Errorf("error generating comparison: %v", err)
			}
			fmt.Println(d)
		}
	}

	return foundDiff, nil
}
