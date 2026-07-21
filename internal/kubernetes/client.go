package kubernetes

import (
	"fmt"

	"github.com/pierinho13/kubectl-peek/internal/config"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	RESTConfig *rest.Config
	Clientset  kubernetes.Interface
	Discovery  discovery.DiscoveryInterface
	Dynamic    dynamic.Interface
	Namespace  string
	UsageRules []config.UsageRule
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
		return nil, fmt.Errorf(
			"load Kubernetes configuration: %w",
			err,
		)
	}

	resolvedNamespace, _, err := clientConfig.Namespace()
	if err != nil {
		return nil, fmt.Errorf(
			"resolve Kubernetes namespace: %w",
			err,
		)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf(
			"create Kubernetes client: %w",
			err,
		)
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf(
			"create Kubernetes discovery client: %w",
			err,
		)
	}

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf(
			"create Kubernetes dynamic client: %w",
			err,
		)
	}

	return &Client{
		RESTConfig: restConfig,
		Clientset:  clientset,
		Discovery:  discoveryClient,
		Dynamic:    dynamicClient,
		Namespace:  resolvedNamespace,
	}, nil
}
