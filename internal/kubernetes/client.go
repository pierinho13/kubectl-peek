package kubernetes

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	Clientset kubernetes.Interface
	Namespace string
}

func NewClient(
	kubeconfig string,
	contextName string,
	namespace string,
) (*Client, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()

	if kubeconfig != "" {
		loadingRules.ExplicitPath = kubeconfig
	}

	overrides := &clientcmd.ConfigOverrides{}

	if contextName != "" {
		overrides.CurrentContext = contextName
	}

	if namespace != "" {
		overrides.Context.Namespace = namespace
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		overrides,
	)

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("load Kubernetes configuration: %w", err)
	}

	resolvedNamespace, _, err := clientConfig.Namespace()
	if err != nil {
		return nil, fmt.Errorf("resolve Kubernetes namespace: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("create Kubernetes client: %w", err)
	}

	return &Client{
		Clientset: clientset,
		Namespace: resolvedNamespace,
	}, nil
}
