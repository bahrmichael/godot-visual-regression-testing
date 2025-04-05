package e2e

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"godot-vrt/cmd"
	"godot-vrt/lib"
)

func TestBaseline(t *testing.T) {

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

	t.Run("should render baseline", func(t *testing.T) {

		sceneName := "baseline"
		scenes := sceneName + ".tscn"
		frames := 10

		deleteFiles(t, projectPath+sceneName+".avi")
		t.Cleanup(func() {
			deleteFiles(nil, projectPath+sceneName+".avi")
		})

		args := []string{
			"baseline",
			"--godot", godotExecutable,
			"--scenes", scenes,
			"--frames", strconv.Itoa(frames),
			"--project", projectPath,
			"--omit-exit-code", "true",
		}
		cmd.RootCmd.SetArgs(args)
		err := cmd.RootCmd.Execute()
		if err != nil {
			t.Fatal("Expected test to not fail because of omit-exit-code flag", err)
		}

		expectedBaselineFile := strings.Replace(projectPath+scenes, ".tscn", ".avi", 1)
		fi, err := os.Stat(expectedBaselineFile)
		if err != nil {
			t.Fatal(err)
		}

		// value captured with debugging. may change when scene changes.
		if fi.Size() != 204280 {
			t.Error("Baseline file does not have the expected size", expectedBaselineFile)
		}

	})
}

func deleteFiles(t *testing.T, glob string) {
	files, err := filepath.Glob(glob)
	if err != nil {
		t.Fatal("error listing files to be deleted", err, glob)
	}
	if len(files) > 0 {
		for _, f := range files {
			err = os.Remove(f)
			if err != nil {
				t.Fatal("error deleting file", err, f)
			}
		}
	}
}
