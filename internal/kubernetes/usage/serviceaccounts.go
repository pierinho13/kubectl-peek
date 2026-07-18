package usage

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ServiceAccountFinder struct {
	client kubernetes.Interface
}

func NewServiceAccountFinder(
	client kubernetes.Interface,
) *ServiceAccountFinder {
	return &ServiceAccountFinder{
		client: client,
	}
}

func (finder *ServiceAccountFinder) Find(
	ctx context.Context,
	namespace string,
	secretName string,
) Result {
	var result Result

	if finder == nil || finder.client == nil {
		result.Warnings = append(
			result.Warnings,
			Warning{
				APIVersion: "v1",
				Resource:   "serviceaccounts",
				Err:        fmt.Errorf("Kubernetes typed client is nil"),
			},
		)

		return result
	}

	serviceAccounts, err := finder.client.
		CoreV1().
		ServiceAccounts(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		result.Warnings = append(
			result.Warnings,
			Warning{
				APIVersion: "v1",
				Resource:   "serviceaccounts",
				Err:        fmt.Errorf("list serviceaccounts: %w", err),
			},
		)

		return result
	}

	for i := range serviceAccounts.Items {
		serviceAccount := &serviceAccounts.Items[i]

		var references []Reference

		for refIndex, imagePullSecret := range serviceAccount.ImagePullSecrets {
			if imagePullSecret.Name != secretName {
				continue
			}

			references = append(
				references,
				Reference{
					Description: "image pull secret",
					Path: fmt.Sprintf(
						"imagePullSecrets[%d].name",
						refIndex,
					),
					Relation: RelationUses,
				},
			)
		}

		if len(references) == 0 {
			continue
		}

		result.Usages = append(
			result.Usages,
			Usage{
				APIVersion: "v1",
				Kind:       "ServiceAccount",
				Namespace:  serviceAccount.Namespace,
				Name:       serviceAccount.Name,
				Source:     SourceBuiltIn,
				References: references,
			},
		)
	}

	return result
}
