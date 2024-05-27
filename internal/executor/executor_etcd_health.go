package executor

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"infra/internal/command"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

type EtcdHealthExecutor struct {
	k8s             client.Client
	commandExecutor *command.Executor
}

func NewEtcdHealthExecutor(k8s client.Client, commandExecutor *command.Executor) *EtcdHealthExecutor {
	return &EtcdHealthExecutor{
		k8s:             k8s,
		commandExecutor: commandExecutor,
	}
}

func (e *EtcdHealthExecutor) Run(ctx context.Context) error {
	var etcdPods v1.PodList
	err := e.k8s.List(ctx, &etcdPods, &client.ListOptions{
		Namespace: "kube-system",
		LabelSelector: labels.SelectorFromSet(map[string]string{
			"component": "etcd",
		}),
	})
	if err != nil {
		return fmt.Errorf("failed to list etcd pods: %w", err)
	}

	var errs *multierror.Error
	for _, etcdPod := range etcdPods.Items {
		healthResultBuilder := new(strings.Builder)
		cmdParams := []string{
			"kubectl",
			"-n",
			"kube-system",
			"exec",
			etcdPod.Name,
			"--",
			"sh",
			"-c",
			"ETCDCTL_ENDPOINTS='https://127.0.0.1:2379' ETCDCTL_CACERT='/var/lib/rancher/rke2/server/tls/etcd/server-ca.crt' ETCDCTL_CERT='/var/lib/rancher/rke2/server/tls/etcd/server-client.crt' ETCDCTL_KEY='/var/lib/rancher/rke2/server/tls/etcd/server-client.key' ETCDCTL_API=3 etcdctl endpoint health",
		}
		err := e.commandExecutor.ExecuteWith("", os.Stdin, healthResultBuilder, cmdParams...)
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("failed to execute etcd health command on %s: %w", etcdPod.Name, err))
		}
		if !strings.Contains(healthResultBuilder.String(), "is healthy") {
			errs = multierror.Append(errs, fmt.Errorf("etcd pod %s is not healthy", etcdPod.Name))
		}
	}

	return errs.ErrorOrNil()
}

func (e *EtcdHealthExecutor) Name() string {
	return "etcd_health"
}
