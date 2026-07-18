package usage

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
)

var gatewayGVRs = []schema.GroupVersionResource{
	{
		Group:    "gateway.networking.k8s.io",
		Version:  "v1",
		Resource: "gateways",
	},
	{
		Group:    "gateway.networking.k8s.io",
		Version:  "v1beta1",
		Resource: "gateways",
	},
}

func FindGatewaySecretUsages(
	ctx context.Context,
	discoveryClient discovery.DiscoveryInterface,
	dynamicClient dynamic.Interface,
	namespace string,
	secretName string,
) Result {
	var result Result

	if discoveryClient == nil || dynamicClient == nil {
		return result
	}

	gvr, available, err := firstAvailableResource(
		discoveryClient,
		gatewayGVRs,
	)
	if err != nil {
		result.Warnings = append(
			result.Warnings,
			Warning{
				APIVersion: "gateway.networking.k8s.io",
				Resource:   "gateways",
				Err:        err,
			},
		)

		return result
	}

	if !available {
		return result
	}

	gateways, err := dynamicClient.
		Resource(gvr).
		Namespace(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		result.Warnings = append(
			result.Warnings,
			Warning{
				APIVersion: gvr.GroupVersion().String(),
				Resource:   gvr.Resource,
				Err:        err,
			},
		)

		return result
	}

	for i := range gateways.Items {
		gateway := &gateways.Items[i]

		references := findGatewaySecretReferences(
			gateway,
			namespace,
			secretName,
		)
		if len(references) == 0 {
			continue
		}

		result.Usages = append(
			result.Usages,
			Usage{
				APIVersion: gateway.GetAPIVersion(),
				Kind:       gateway.GetKind(),
				Namespace:  gateway.GetNamespace(),
				Name:       gateway.GetName(),
				Source:     SourceBuiltIn,
				References: references,
			},
		)
	}

	return result
}

func firstAvailableResource(
	discoveryClient discovery.DiscoveryInterface,
	resources []schema.GroupVersionResource,
) (
	schema.GroupVersionResource,
	bool,
	error,
) {
	if discoveryClient == nil {
		return schema.GroupVersionResource{}, false, nil
	}

	for _, resource := range resources {
		available, err := ResourceAvailable(
			discoveryClient,
			resource,
		)
		if err != nil {
			return schema.GroupVersionResource{}, false, err
		}

		if available {
			return resource, true, nil
		}
	}

	return schema.GroupVersionResource{}, false, nil
}

func findGatewaySecretReferences(
	gateway *unstructured.Unstructured,
	gatewayNamespace string,
	secretName string,
) []Reference {
	listeners, found, err := unstructured.NestedSlice(
		gateway.Object,
		"spec",
		"listeners",
	)
	if err != nil || !found {
		return nil
	}

	var references []Reference

	for listenerIndex, listenerValue := range listeners {
		listener, ok := listenerValue.(map[string]interface{})
		if !ok {
			continue
		}

		tls, found, err := unstructured.NestedMap(
			listener,
			"tls",
		)
		if err != nil || !found {
			continue
		}

		certificateRefs, found, err := unstructured.NestedSlice(
			tls,
			"certificateRefs",
		)
		if err != nil || !found {
			continue
		}

		for refIndex, refValue := range certificateRefs {
			ref, ok := refValue.(map[string]interface{})
			if !ok {
				continue
			}

			if !isGatewaySecretReference(
				ref,
				gatewayNamespace,
				secretName,
			) {
				continue
			}

			path := fmt.Sprintf(
				"spec.listeners[%d].tls.certificateRefs[%d].name",
				listenerIndex,
				refIndex,
			)

			references = append(
				references,
				Reference{
					Description: "TLS certificate",
					Path:        path,
					Relation:    RelationUses,
				},
			)
		}
	}

	return references
}

func isGatewaySecretReference(
	ref map[string]interface{},
	gatewayNamespace string,
	secretName string,
) bool {
	name, found, err := unstructured.NestedString(
		ref,
		"name",
	)
	if err != nil || !found || name != secretName {
		return false
	}

	kind, found, err := unstructured.NestedString(
		ref,
		"kind",
	)
	if err != nil {
		return false
	}

	if found && kind != "" && kind != "Secret" {
		return false
	}

	group, found, err := unstructured.NestedString(
		ref,
		"group",
	)
	if err != nil {
		return false
	}

	if found && group != "" {
		return false
	}

	referenceNamespace, found, err := unstructured.NestedString(
		ref,
		"namespace",
	)
	if err != nil {
		return false
	}

	if found &&
		referenceNamespace != "" &&
		referenceNamespace != gatewayNamespace {
		return false
	}

	return true
}
