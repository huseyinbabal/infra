package executor

import (
	"context"
	"errors"
	"fmt"
	"infra/internal/command"
	"infra/internal/helper"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

type RkeSystemdHealthExecutor struct {
	k8s             client.Client
	commandExecutor *command.Executor
}

func NewRkeSystemdHealthExecutor(k8s client.Client, commandExecutor *command.Executor) *RkeSystemdHealthExecutor {
	return &RkeSystemdHealthExecutor{
		k8s:             k8s,
		commandExecutor: commandExecutor,
	}
}

func (e *RkeSystemdHealthExecutor) Run(ctx context.Context) error {
	if !helper.CommandExists("rke2") {
		return errors.New("this is not a rancher environment")
	}

	systemCtlListServicesOutputBuilder := new(strings.Builder)
	err := e.commandExecutor.ExecuteWith("", nil, systemCtlListServicesOutputBuilder, "systemctl", "list-units", "--type=service")
	if err != nil {
		return fmt.Errorf("failed to list systemd services. %v", err)
	}
	rke2NodeType := ""
	if strings.Contains(systemCtlListServicesOutputBuilder.String(), "rke2-server.service") {
		rke2NodeType = "rke2-server.service"
	} else if strings.Contains(systemCtlListServicesOutputBuilder.String(), "rke2-agent.service") {
		rke2NodeType = "rke2-agent.service"
	} else {
		return errors.New("failed to determine rke2 node type")
	}

	systemdServiceStatusBuilder := new(strings.Builder)
	err = e.commandExecutor.ExecuteWith("", nil, systemdServiceStatusBuilder, "systemctl", "is-active", rke2NodeType)
	if err != nil {
		return fmt.Errorf("failed to check rke2 service status. %v", err)
	}
	if strings.Contains(systemdServiceStatusBuilder.String(), "inactive") {
		journalLogsOutputBuilder := new(strings.Builder)
		err = e.commandExecutor.ExecuteWith("", nil, journalLogsOutputBuilder, "journalctl", "-u", rke2NodeType, "--no-pager", "-n", "10")
		if err != nil {
			return fmt.Errorf("failed to get rke2 journal logs. %v", err)
		}
		return fmt.Errorf("rke2 service is not active. %v", journalLogsOutputBuilder.String())
	}
	return nil
}

func (e *RkeSystemdHealthExecutor) Name() string {
	return "rke2_systemd_health"
}
