package lib

import (
	"bytes"
	"os"
	"os/exec"
)

func executeCommandUnsafe(dir *string, program string, args []string) (string, string, error) {

	cmd := exec.Command(program, args...)
	if dir != nil {
		cmd.Dir = *dir
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()

	return stdout.String(), stderr.String(), err
}
