package usage

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type IngressFinder struct {
	client kubernetes.Interface
}

func NewIngressFinder(
	client kubernetes.Interface,
) *IngressFinder {
	return &IngressFinder{
		client: client,
	}
}

func (finder *IngressFinder) Find(
	ctx context.Context,
	namespace string,
	secretName string,
) Result {
	var result Result

	if finder == nil || finder.client == nil {
		result.Warnings = append(
			result.Warnings,
			Warning{
				APIVersion: "networking.k8s.io/v1",
				Resource:   "ingresses",
				Err:        fmt.Errorf("Kubernetes typed client is nil"),
			},
		)

		return result
	}

	ingresses, err := finder.client.
		NetworkingV1().
		Ingresses(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		result.Warnings = append(
			result.Warnings,
			Warning{
				APIVersion: "networking.k8s.io/v1",
				Resource:   "ingresses",
				Err:        fmt.Errorf("list ingresses: %w", err),
			},
		)

		return result
	}

	for i := range ingresses.Items {
		ingress := &ingresses.Items[i]

		var references []Reference

		for tlsIndex, tls := range ingress.Spec.TLS {
			if tls.SecretName != secretName {
				continue
			}

			references = append(
				references,
				Reference{
					Description: "TLS certificate",
					Path: fmt.Sprintf(
						"spec.tls[%d].secretName",
						tlsIndex,
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
				APIVersion: "networking.k8s.io/v1",
				Kind:       "Ingress",
				Namespace:  ingress.Namespace,
				Name:       ingress.Name,
				Source:     SourceBuiltIn,
				References: references,
			},
		)
	}

	return result
}
