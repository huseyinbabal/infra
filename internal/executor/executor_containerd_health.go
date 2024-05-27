package executor

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"infra/internal/command"
	"os/exec"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

type ContainerdHealthExecutor struct {
	k8s             client.Client
	commandExecutor *command.Executor
}

func NewContainerdHealthExecutor(k8s client.Client, commandExecutor *command.Executor) *ContainerdHealthExecutor {
	return &ContainerdHealthExecutor{
		k8s:             k8s,
		commandExecutor: commandExecutor,
	}
}

func (e *ContainerdHealthExecutor) Run(ctx context.Context) error {
	criCtlPsOutputBuilder := new(strings.Builder)
	err := e.commandExecutor.ExecuteWith("", nil, criCtlPsOutputBuilder, "crictl", "ps")
	var errs *multierror.Error
	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("failed to list containers. %v", err))
	}

	if criCtlPsOutputBuilder.String() == "" {
		errs = multierror.Append(errs, errors.New("no containers found"))
	}

	tailCmd := exec.Command("tail", "-n", "30", "/var/lib/rancher/rke2/agent/containerd/containerd.log")
	tailOut, err := tailCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error creating tail command stdout pipe: %v", err)
	}

	if err := tailCmd.Start(); err != nil {
		return fmt.Errorf("error starting tail command: %v", err)
	}

	var errLines []string
	scanner := bufio.NewScanner(tailOut)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(strings.ToLower(line), "error") {
			errLines = append(errLines, line)
		}
	}

	// Check for scanner errors.
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading tail output: %v", err)
	}

	if err := tailCmd.Wait(); err != nil {
		return fmt.Errorf("error waiting for tail command: %v", err)
	}

	if len(errLines) > 0 {
		errs = multierror.Append(errs, fmt.Errorf("containerd logs contain errors. %v", errLines))
	}

	return errs.ErrorOrNil()
}

func (e *ContainerdHealthExecutor) Name() string {
	return "containerd_health"
}
