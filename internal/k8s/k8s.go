package k8s

import (
	"fmt"
	"infra/internal/helper"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetKubeConfig() (*rest.Config, error) {
	kubeConfigPath := os.Getenv("KUBECONFIG")
	if _, err := os.Stat("/etc/rancher/rke2/rke2.yaml"); err == nil {
		kubeConfigPath = "/etc/rancher/rke2/rke2.yaml"
	}

	if kubeConfigPath == "" {
		kubeConfigPath = fmt.Sprintf("%s/.kube/config", homedir.HomeDir())
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func NewCtrlClient(config *rest.Config) client.Client {
	if config == nil {
		config, _ = GetKubeConfig()
	}

	return helper.MustReturn(client.New(config, client.Options{}))
}
