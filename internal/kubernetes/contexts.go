package kubernetes

import (
	"fmt"
	"sort"
)

func ContextNames(
	kubeconfig string,
) ([]string, error) {
	loadingRules := newLoadingRules(kubeconfig)

	config, err := loadingRules.Load()
	if err != nil {
		return nil, fmt.Errorf(
			"load Kubernetes configuration: %w",
			err,
		)
	}

	contextNames := make(
		[]string,
		0,
		len(config.Contexts),
	)

	for name, context := range config.Contexts {
		if context == nil {
			continue
		}

		contextNames = append(contextNames, name)
	}

	if len(contextNames) == 0 {
		return nil, fmt.Errorf(
			"Kubernetes configuration contains no contexts",
		)
	}

	sort.Strings(contextNames)

	return contextNames, nil
}
