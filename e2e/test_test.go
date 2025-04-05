package e2e

import (
	"os"
	"strconv"
	"strings"
	"testing"

	"godot-vrt/cmd"
	"godot-vrt/lib"
)

func TestTest(t *testing.T) {

	godotExecutable := "/Applications/Godot.app/Contents/MacOS/Godot"
	env := os.Environ()
	for _, e := range env {
		if strings.HasPrefix("GODOT_EXECUTABLE", e) {
			godotExecutable = strings.Split(e, "=")[1]
		}
	}

	projectPath := "test-project/"
	if err := lib.Validate(godotExecutable, projectPath); err != nil {
		t.Fatal("Test prerequisites not met", err)
	}

	t.Run("should find no changes", func(t *testing.T) {

		sceneName := "test_no_changes"
		scenes := sceneName + ".tscn"
		baseline := sceneName + ".avi"
		frames := 10

		if _, err := os.Stat(projectPath + scenes); err != nil {
			t.Fatal("Expected scenes file to exist", err)
		}
		if _, err := os.Stat(projectPath + baseline); err != nil {
			t.Fatal("Expected baseline file to exist", err)
		}

		args := []string{
			"test",
			"--godot", godotExecutable,
			"--scenes", scenes,
			"--baseline", baseline,
			"--frames", strconv.Itoa(frames),
			"--project", projectPath,
		}
		cmd.RootCmd.SetArgs(args)
		err := cmd.RootCmd.Execute()
		if err != nil {
			t.Fatal("Expected test to not find differences and not fail", err)
		}
	})

	t.Run("should find changes", func(t *testing.T) {

		sceneName := "test_with_changes"
		scenes := sceneName + ".tscn"
		baseline := sceneName + ".avi"
		frames := 10

		if _, err := os.Stat(projectPath + scenes); err != nil {
			t.Fatal("Expected scenes file to exist", err)
		}
		if _, err := os.Stat(projectPath + baseline); err != nil {
			t.Fatal("Expected baseline file to exist", err)
		}

		expectedComparisonFile := strings.Replace("vrt-results/"+projectPath+scenes, ".tscn", ".avi", 1)

		deleteFiles(t, expectedComparisonFile)
		t.Cleanup(func() {
			deleteFiles(nil, expectedComparisonFile)
		})

		args := []string{
			"test",
			"--godot", godotExecutable,
			"--scenes", scenes,
			"--baseline", baseline,
			"--frames", strconv.Itoa(frames),
			"--project", projectPath,
			"--omit-exit-code", "true",
		}
		cmd.RootCmd.SetArgs(args)
		err := cmd.RootCmd.Execute()
		if err != nil {
			t.Fatal("Expected test to not fail because of omit-exit-code flag", err)
		}

		fi, err := os.Stat(expectedComparisonFile)
		if err != nil {
			t.Fatal(err)
		}

		// value captured with debugging. may change when scene changes.
		if fi.Size() != 56642 {
			t.Error("Baseline file does not have the expected size", expectedComparisonFile)
		}
	})

}
