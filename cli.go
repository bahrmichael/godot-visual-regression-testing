package main

import (
	"bytes"
	"os"
	"os/exec"
)

func executeCommandUnsafe(program string, args []string) (string, string, error) {

	cmd := exec.Command(program, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()

	return stdout.String(), stderr.String(), err
}
