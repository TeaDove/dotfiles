package kube_supplier

import (
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

type Supplier struct {
	kc *kubernetes.Clientset
	mc *metrics.Clientset

	namespace string
}

func NewSupplier() (*Supplier, error) {
	r := Supplier{}

	config, err := clientcmd.NewDefaultClientConfigLoadingRules().Load()
	if err != nil {
		return nil, fmt.Errorf("error loading default config: %v", err)
	}

	if config.Contexts[config.CurrentContext] == nil {
		return nil, fmt.Errorf("no context found in config")
	}

	r.namespace = config.Contexts[config.CurrentContext].Namespace

	clientConfig, err := clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("error loading client config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build client set: %w", err)
	}

	mc, err := metrics.NewForConfig(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build metrics client: %w", err)
	}
	r.mc = mc

	r.kc = clientset
	return &r, nil
}
