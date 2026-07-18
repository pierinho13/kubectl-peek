package usage

import (
	"context"
	"errors"
	"reflect"
	"testing"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	discoveryfake "k8s.io/client-go/discovery/fake"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	kubernetesfake "k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

func TestFindGatewaySecretUsages(t *testing.T) {
	t.Parallel()

	const (
		namespace  = "test"
		secretName = "gateway-tls"
	)

	ctx := context.Background()

	gvr := schema.GroupVersionResource{
		Group:    "gateway.networking.k8s.io",
		Version:  "v1",
		Resource: "gateways",
	}

	gateway := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "gateway.networking.k8s.io/v1",
			"kind":       "Gateway",
			"metadata": map[string]interface{}{
				"name":      "external",
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"listeners": []interface{}{
					map[string]interface{}{
						"name":     "https",
						"protocol": "HTTPS",
						"port":     int64(443),
						"tls": map[string]interface{}{
							"certificateRefs": []interface{}{
								map[string]interface{}{
									"group": "",
									"kind":  "Secret",
									"name":  secretName,
								},
							},
						},
					},
				},
			},
		},
	}

	scheme := runtime.NewScheme()

	dynamicClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(
		scheme,
		map[schema.GroupVersionResource]string{
			gvr: "GatewayList",
		},
	)

	_, err := dynamicClient.
		Resource(gvr).
		Namespace(namespace).
		Create(
			ctx,
			gateway,
			metav1.CreateOptions{},
		)
	if err != nil {
		t.Fatalf("create fake Gateway: %v", err)
	}

	clientset := kubernetesfake.NewSimpleClientset()

	discoveryClient := clientset.Discovery().(*discoveryfake.FakeDiscovery)

	discoveryClient.Resources = []*metav1.APIResourceList{
		{
			GroupVersion: "gateway.networking.k8s.io/v1",
			APIResources: []metav1.APIResource{
				{
					Name:       "gateways",
					Kind:       "Gateway",
					Namespaced: true,
					Verbs: metav1.Verbs{
						"get",
						"list",
					},
				},
			},
		},
	}

	got := FindGatewaySecretUsages(
		ctx,
		discoveryClient,
		dynamicClient,
		namespace,
		secretName,
	)

	want := Result{
		Usages: []Usage{
			{
				APIVersion: "gateway.networking.k8s.io/v1",
				Kind:       "Gateway",
				Namespace:  namespace,
				Name:       "external",
				Source:     SourceBuiltIn,
				References: []Reference{
					{
						Description: "TLS certificate",
						Path: "spec.listeners[0].tls." +
							"certificateRefs[0].name",
						Relation: RelationUses,
					},
				},
			},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf(
			"FindGatewaySecretUsages() mismatch\n\ngot:  %#v\n\nwant: %#v",
			got,
			want,
		)
	}
}
func TestFindGatewaySecretUsagesResourceUnavailable(t *testing.T) {
	t.Parallel()

	const (
		namespace  = "test"
		secretName = "gateway-tls"
	)

	clientset := kubernetesfake.NewSimpleClientset()

	discoveryClient := clientset.Discovery().(*discoveryfake.FakeDiscovery)

	dynamicClient := dynamicfake.NewSimpleDynamicClient(
		runtime.NewScheme(),
	)

	got := FindGatewaySecretUsages(
		context.Background(),
		discoveryClient,
		dynamicClient,
		namespace,
		secretName,
	)

	want := Result{}

	if !reflect.DeepEqual(got, want) {
		t.Errorf(
			"FindGatewaySecretUsages() mismatch\n\ngot:  %#v\n\nwant: %#v",
			got,
			want,
		)
	}
}
func TestFindGatewaySecretUsagesReturnsWarningOnForbidden(
	t *testing.T,
) {
	t.Parallel()

	const (
		namespace  = "test"
		secretName = "gateway-tls"
	)

	gvr := schema.GroupVersionResource{
		Group:    "gateway.networking.k8s.io",
		Version:  "v1",
		Resource: "gateways",
	}

	clientset := kubernetesfake.NewSimpleClientset()

	discoveryClient := clientset.Discovery().(*discoveryfake.FakeDiscovery)

	discoveryClient.Resources = []*metav1.APIResourceList{
		{
			GroupVersion: "gateway.networking.k8s.io/v1",
			APIResources: []metav1.APIResource{
				{
					Name:       "gateways",
					Kind:       "Gateway",
					Namespaced: true,
					Verbs: metav1.Verbs{
						"get",
						"list",
					},
				},
			},
		},
	}

	dynamicClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(
		runtime.NewScheme(),
		map[schema.GroupVersionResource]string{
			gvr: "GatewayList",
		},
	)

	dynamicClient.PrependReactor(
		"list",
		"gateways",
		func(
			action ktesting.Action,
		) (bool, runtime.Object, error) {
			return true, nil, apierrors.NewForbidden(
				schema.GroupResource{
					Group:    "gateway.networking.k8s.io",
					Resource: "gateways",
				},
				"",
				errors.New("access denied"),
			)
		},
	)

	got := FindGatewaySecretUsages(
		context.Background(),
		discoveryClient,
		dynamicClient,
		namespace,
		secretName,
	)

	if len(got.Usages) != 0 {
		t.Fatalf(
			"expected no usages, got %#v",
			got.Usages,
		)
	}

	if len(got.Warnings) != 1 {
		t.Fatalf(
			"expected one warning, got %#v",
			got.Warnings,
		)
	}

	warning := got.Warnings[0]

	if warning.APIVersion != "gateway.networking.k8s.io/v1" {
		t.Errorf(
			"expected Gateway API version, got %q",
			warning.APIVersion,
		)
	}

	if warning.Resource != "gateways" {
		t.Errorf(
			"expected gateways resource, got %q",
			warning.Resource,
		)
	}

	if warning.Err == nil {
		t.Fatal("expected warning error")
	}

	if !apierrors.IsForbidden(warning.Err) {
		t.Errorf(
			"expected forbidden error, got %v",
			warning.Err,
		)
	}
}
func TestFindGatewaySecretUsagesWithNilClients(
	t *testing.T,
) {
	t.Parallel()

	got := FindGatewaySecretUsages(
		context.Background(),
		nil,
		nil,
		"test",
		"application-secret",
	)

	if len(got.Usages) != 0 {
		t.Errorf(
			"expected no usages, got %#v",
			got.Usages,
		)
	}

	if len(got.Warnings) != 0 {
		t.Errorf(
			"expected no warnings, got %#v",
			got.Warnings,
		)
	}
}
