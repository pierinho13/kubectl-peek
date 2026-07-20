package kubernetes

import (
	"fmt"

	"k8s.io/client-go/tools/clientcmd"
)

func SetContextNamespace(
	kubeconfig string,
	contextName string,
	namespace string,
) (string, error) {
	loadingRules := newLoadingRules(kubeconfig)

	config, err := loadingRules.Load()
	if err != nil {
		return "", fmt.Errorf(
			"load Kubernetes configuration: %w",
			err,
		)
	}

	targetContext, err := resolveContextName(
		config.CurrentContext,
		config.Contexts,
		contextName,
	)
	if err != nil {
		return "", err
	}

	updatedContext := config.Contexts[targetContext].DeepCopy()
	updatedContext.Namespace = namespace
	config.Contexts[targetContext] = updatedContext

	if err := clientcmd.ModifyConfig(
		loadingRules,
		*config,
		false,
	); err != nil {
		return "", fmt.Errorf(
			"set namespace %q for context %q: %w",
			namespace,
			targetContext,
			err,
		)
	}

	return targetContext, nil
}

func newLoadingRules(
	kubeconfig string,
) *clientcmd.ClientConfigLoadingRules {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()

	if kubeconfig != "" {
		loadingRules.ExplicitPath = kubeconfig
	}

	return loadingRules
}
