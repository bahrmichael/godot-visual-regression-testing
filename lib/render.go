package lib

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
)

type RenderSceneArgs struct {
	SceneFileFromProjectRoot string
	OutputFile               string
	GodotBinary              string
	Verbose                  bool
	Frames                   int
	ProjectPath              string
}

func RenderScene(args RenderSceneArgs) (string, error) {
	if err := os.MkdirAll(filepath.Dir(args.OutputFile), 0755); err != nil {
		return "", fmt.Errorf("error creating dir: %v", err)
	}

	a := []string{
		"--quit-after",
		strconv.Itoa(args.Frames),
		"--write-movie", args.OutputFile,
		args.SceneFileFromProjectRoot,
	}
	if args.Verbose {
		a = slices.Insert(a, 0, "--verbose")
	}
	stdout, stderr, err := executeCommandUnsafe(&args.ProjectPath, args.GodotBinary, a)
	if args.Verbose {
		fmt.Println(stdout)
		fmt.Println(stderr)
	}
	if err != nil {
		return "", fmt.Errorf("error rendering scene: %v %s", err, stderr)
	}
	fileInfo, err := os.Stat(args.OutputFile)
	if err != nil {
		return "", fmt.Errorf("error getting rendered file info: %v", err)
	}
	if fileInfo.Size() == 0 {
		return "", fmt.Errorf("error: rendered file is empty")
	}

	return args.OutputFile, nil
}
