package helper

import (
	"github.com/fatih/color"
	"os"
	"os/exec"
)

func Must(e error) {
	if e != nil {
		color.Red(e.Error())
		os.Exit(1)
	}
}

func MustReturn[T any](t T, err error) T {
	Must(err)
	return t
}

func CommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
