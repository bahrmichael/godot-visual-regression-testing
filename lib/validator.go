package lib

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func Validate(godotPath, projectPath string) error {
	if err := VerifyGodotInstallation(godotPath); err != nil {
		return err
	}
	if err := VerifyBinary("ffmpeg"); err != nil {
		return err
	}

	if projectPath != "" {
		err := os.Chdir(projectPath)
		if err != nil {
			return err
		}
	}
	if err := VerifyFileExists("project.godot"); err != nil {
		return err
	}
	return nil
}

func VerifyFileExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exist", path)
	}
	return nil
}

func VerifyBinary(binaryPath string) error {
	_, err := exec.LookPath(binaryPath)
	if err != nil {
		return fmt.Errorf("cannot find binary at %s", binaryPath)
	}
	return nil
}

var supportedVersions = []string{"4.4.1", "4.4", "4.3", "4.2.2", "4.1.4"}

func VerifyGodotInstallation(godotPath string) error {
	if err := VerifyBinary(godotPath); err != nil {
		return err
	}

	versionResult, stderr, err := executeCommandUnsafe(godotPath, []string{"--version", "--headless"})
	if err != nil {
		return fmt.Errorf("error executing Godot: %v %s", err, stderr)
	}
	fmt.Println("Godot version: " + versionResult)

	supportedVersion := false
	for _, v := range supportedVersions {
		if strings.HasPrefix(versionResult, v) && strings.Contains(versionResult, "stable") {
			supportedVersion = true
			break
		}
	}
	if !supportedVersion {
		return fmt.Errorf("godot version is currently not supported. Please install one of the stable versions %s and try again: %v %s", supportedVersions, err, stderr)
	}
	return nil
}
