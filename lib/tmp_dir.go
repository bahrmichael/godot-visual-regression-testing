package lib

import (
	"fmt"
	"os"
	"path/filepath"
)

func InitTmpDir() (string, func()) {
	tmpDir, err := os.MkdirTemp(".", ".vrt_")
	if err != nil {
		fmt.Println("Failed to create temp dir", err)
		os.Exit(1)
	}
	tmpDir, err = filepath.Abs(tmpDir)
	if err != nil {
		fmt.Println("Failed to get absolute path of temp dir", err)
		os.Exit(1)
	}
	tmpDir = WithFolderSuffix(tmpDir)

	return tmpDir, func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			fmt.Printf("error removing temp dir: %v", err)
		}
	}
}
