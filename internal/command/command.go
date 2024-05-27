package command

import (
	"io"
	"os"
	"os/exec"
)

type Executor struct {
	EnvironmentVariables map[string]string
}

func NewExecutor() *Executor {
	return &Executor{EnvironmentVariables: map[string]string{}}
}

func (c Executor) SilentExecute(dir string, params ...string) error {
	return c.ExecuteWith(dir, nil, os.Stdin, params...)
}

func (c Executor) Execute(dir string, params ...string) error {
	return c.ExecuteWith(dir, os.Stdout, os.Stdin, params...)
}

func (c Executor) ExecuteWith(dir string, stdIn io.Reader, stdOut io.Writer, params ...string) error {
	cmd := exec.Command(params[0], params[1:]...)
	cmd.Dir = dir
	cmd.Stdin = stdIn
	cmd.Stdout = stdOut
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
