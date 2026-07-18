package kubernetes

import (
	"context"
	"errors"
	"sort"

	"github.com/pierinho13/kubectl-peek/internal/kubernetes/usage"
)

type SecretUsageReference struct {
	Description string
	Path        string
	Key         string
	Relation    string
}

type SecretUsage struct {
	Kind       string
	Name       string
	References []SecretUsageReference
}

type SecretUsageWarning struct {
	Resource string
	Err      error
}

type SecretUsageResult struct {
	Usages   []SecretUsage
	Warnings []SecretUsageWarning
}

func FindSecretUsages(
	ctx context.Context,
	client *Client,
	namespace string,
	secretName string,
) ([]SecretUsage, error) {
	result, err := FindSecretUsagesDetailed(
		ctx,
		client,
		namespace,
		secretName,
	)
	if err != nil {
		return nil, err
	}

	return result.Usages, nil
}

func FindSecretUsagesDetailed(
	ctx context.Context,
	client *Client,
	namespace string,
	secretName string,
) (SecretUsageResult, error) {
	var result SecretUsageResult

	if client == nil {
		return result, errors.New("Kubernetes client is nil")
	}

	if client.Clientset == nil {
		return result, errors.New("Kubernetes typed client is nil")
	}

	typedFinders := []usage.Finder{
		usage.NewWorkloadFinder(client.Clientset),
		usage.NewServiceAccountFinder(client.Clientset),
		usage.NewIngressFinder(client.Clientset),
	}

	for _, finder := range typedFinders {
		finderResult := finder.Find(
			ctx,
			namespace,
			secretName,
		)

		result.Usages = append(
			result.Usages,
			convertDynamicUsages(
				finderResult.Usages,
			)...,
		)

		appendUsageWarnings(
			&result,
			finderResult.Warnings,
		)
	}

	if client.Discovery != nil && client.Dynamic != nil {
		gatewayResult := usage.FindGatewaySecretUsages(
			ctx,
			client.Discovery,
			client.Dynamic,
			namespace,
			secretName,
		)

		result.Usages = append(
			result.Usages,
			convertDynamicUsages(
				gatewayResult.Usages,
			)...,
		)

		appendUsageWarnings(
			&result,
			gatewayResult.Warnings,
		)

		if len(client.UsageRules) > 0 {
			ruleFinder := usage.NewRuleFinder(
				client.Discovery,
				client.Dynamic,
				client.UsageRules,
			)

			ruleResult := ruleFinder.Find(
				ctx,
				namespace,
				secretName,
			)

			result.Usages = append(
				result.Usages,
				convertDynamicUsages(
					ruleResult.Usages,
				)...,
			)

			appendUsageWarnings(
				&result,
				ruleResult.Warnings,
			)
		}
	}

	sortSecretUsages(result.Usages)

	return result, nil
}

func convertDynamicUsages(
	dynamicUsages []usage.Usage,
) []SecretUsage {
	usages := make(
		[]SecretUsage,
		0,
		len(dynamicUsages),
	)

	for _, dynamicUsage := range dynamicUsages {
		references := make(
			[]SecretUsageReference,
			0,
			len(dynamicUsage.References),
		)

		for _, reference := range dynamicUsage.References {
			references = append(
				references,
				SecretUsageReference{
					Description: reference.Description,
					Path:        reference.Path,
					Key:         reference.Key,
					Relation:    string(reference.Relation),
				},
			)
		}

		usages = append(
			usages,
			SecretUsage{
				Kind:       dynamicUsage.Kind,
				Name:       dynamicUsage.Name,
				References: references,
			},
		)
	}

	return usages
}

func sortSecretUsages(usages []SecretUsage) {
	sort.Slice(
		usages,
		func(i, j int) bool {
			if usages[i].Kind == usages[j].Kind {
				return usages[i].Name < usages[j].Name
			}

			return usages[i].Kind < usages[j].Kind
		},
	)
}

func appendUsageWarnings(
	result *SecretUsageResult,
	warnings []usage.Warning,
) {
	for _, warning := range warnings {
		result.Warnings = append(
			result.Warnings,
			SecretUsageWarning{
				Resource: warning.Resource,
				Err:      warning.Err,
			},
		)
	}
}
