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
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()

	if kubeconfig != "" {
		loadingRules.ExplicitPath = kubeconfig
	}

	config, err := loadingRules.Load()
	if err != nil {
		return "", fmt.Errorf(
			"load Kubernetes configuration: %w",
			err,
		)
	}

	targetContext := contextName
	if targetContext == "" {
		targetContext = config.CurrentContext
	}

	if targetContext == "" {
		return "", fmt.Errorf(
			"Kubernetes configuration has no current context",
		)
	}

	currentContext, ok := config.Contexts[targetContext]
	if !ok || currentContext == nil {
		return "", fmt.Errorf(
			"Kubernetes context %q not found",
			targetContext,
		)
	}

	updatedContext := currentContext.DeepCopy()
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
