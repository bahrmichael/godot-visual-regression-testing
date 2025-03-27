package lib

import (
	"fmt"
	"os"
	"slices"
)

func RenderScene(scene, outputFile string, args []string, verbose bool, godotBinary string) (string, error) {
	args = append(args, "--write-movie", outputFile, scene)
	if verbose {
		args = slices.Insert(args, 0, "--verbose")
	}
	stdout, stderr, err := executeCommandUnsafe(godotBinary, args)
	if verbose {
		fmt.Println(stdout)
		fmt.Println(stderr)
	}
	if err != nil {
		return "", fmt.Errorf("error rendering scene: %v %s", err, stderr)
	}
	fileInfo, err := os.Stat(outputFile)
	if err != nil {
		return "", fmt.Errorf("error getting rendered file info: %v", err)
	}
	if fileInfo.Size() == 0 {
		return "", fmt.Errorf("error: rendered file is empty")
	}

	return outputFile, nil
}
