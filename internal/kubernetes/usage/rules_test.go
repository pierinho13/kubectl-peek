package usage

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/pierinho13/kubectl-peek/internal/config"
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

func TestRuleFinderFindsProducedSecretUsingAvailableVersion(
	t *testing.T,
) {
	t.Parallel()

	const (
		namespace  = "test"
		secretName = "database-credentials"
	)

	gvr := schema.GroupVersionResource{
		Group:    "example.io",
		Version:  "v1beta1",
		Resource: "secretproducers",
	}

	clientset := kubernetesfake.NewSimpleClientset()

	discoveryClient := clientset.
		Discovery().(*discoveryfake.FakeDiscovery)

	discoveryClient.Resources = []*metav1.APIResourceList{
		{
			GroupVersion: "example.io/v1beta1",
			APIResources: []metav1.APIResource{
				{
					Name:       "secretproducers",
					Kind:       "SecretProducer",
					Namespaced: true,
					Verbs: metav1.Verbs{
						"get",
						"list",
					},
				},
			},
		},
	}

	dynamicClient :=
		dynamicfake.NewSimpleDynamicClientWithCustomListKinds(
			runtime.NewScheme(),
			map[schema.GroupVersionResource]string{
				gvr: "SecretProducerList",
			},
		)

	_, err := dynamicClient.
		Resource(gvr).
		Namespace(namespace).
		Create(
			context.Background(),
			&unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "example.io/v1beta1",
					"kind":       "SecretProducer",
					"metadata": map[string]interface{}{
						"name":      "database",
						"namespace": namespace,
					},
					"spec": map[string]interface{}{
						"target": map[string]interface{}{
							"name": secretName,
						},
					},
				},
			},
			metav1.CreateOptions{},
		)
	if err != nil {
		t.Fatalf(
			"create SecretProducer: %v",
			err,
		)
	}

	finder := NewRuleFinder(
		discoveryClient,
		dynamicClient,
		[]config.UsageRule{
			{
				APIVersions: []string{
					"example.io/v1",
					"example.io/v1beta1",
				},
				Kind:     "SecretProducer",
				Resource: "secretproducers",
				References: []config.SecretReferenceRule{
					{
						Path:        "spec.target.name",
						Description: "generated Secret",
						Relation:    config.RelationProduces,
					},
				},
			},
		},
	)

	got := finder.Find(
		context.Background(),
		namespace,
		secretName,
	)

	want := Result{
		Usages: []Usage{
			{
				APIVersion: "example.io/v1beta1",
				Kind:       "SecretProducer",
				Namespace:  namespace,
				Name:       "database",
				Source:     SourceRule,
				References: []Reference{
					{
						Description: "generated Secret",
						Path:        "spec.target.name",
						Relation:    RelationProduces,
					},
				},
			},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf(
			"RuleFinder.Find() mismatch\n\ngot:  %#v\n\nwant: %#v",
			got,
			want,
		)
	}
}

func TestRuleFinderSupportsWildcardPaths(t *testing.T) {
	t.Parallel()

	const (
		namespace  = "test"
		secretName = "database-credentials"
	)

	gvr := schema.GroupVersionResource{
		Group:    "example.io",
		Version:  "v1",
		Resource: "applications",
	}

	clientset := kubernetesfake.NewSimpleClientset()

	discoveryClient := clientset.
		Discovery().(*discoveryfake.FakeDiscovery)

	discoveryClient.Resources = []*metav1.APIResourceList{
		{
			GroupVersion: "example.io/v1",
			APIResources: []metav1.APIResource{
				{
					Name:       "applications",
					Kind:       "Application",
					Namespaced: true,
					Verbs: metav1.Verbs{
						"get",
						"list",
					},
				},
			},
		},
	}

	dynamicClient :=
		dynamicfake.NewSimpleDynamicClientWithCustomListKinds(
			runtime.NewScheme(),
			map[schema.GroupVersionResource]string{
				gvr: "ApplicationList",
			},
		)

	_, err := dynamicClient.
		Resource(gvr).
		Namespace(namespace).
		Create(
			context.Background(),
			&unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "example.io/v1",
					"kind":       "Application",
					"metadata": map[string]interface{}{
						"name":      "backend",
						"namespace": namespace,
					},
					"spec": map[string]interface{}{
						"credentials": []interface{}{
							map[string]interface{}{
								"secretRef": map[string]interface{}{
									"name": "redis-credentials",
								},
							},
							map[string]interface{}{
								"secretRef": map[string]interface{}{
									"name": secretName,
								},
							},
						},
					},
				},
			},
			metav1.CreateOptions{},
		)
	if err != nil {
		t.Fatalf(
			"create Application: %v",
			err,
		)
	}

	finder := NewRuleFinder(
		discoveryClient,
		dynamicClient,
		[]config.UsageRule{
			{
				APIVersions: []string{
					"example.io/v1",
				},
				Kind:     "Application",
				Resource: "applications",
				References: []config.SecretReferenceRule{
					{
						Path: "spec.credentials[*]." +
							"secretRef.name",
						Description: "application credentials",
						Relation:    config.RelationUses,
					},
				},
			},
		},
	)

	got := finder.Find(
		context.Background(),
		namespace,
		secretName,
	)

	if len(got.Usages) != 1 {
		t.Fatalf(
			"expected one usage, got %#v",
			got.Usages,
		)
	}

	if len(got.Usages[0].References) != 1 {
		t.Fatalf(
			"expected one reference, got %#v",
			got.Usages[0].References,
		)
	}

	reference := got.Usages[0].References[0]

	if reference.Path !=
		"spec.credentials[*].secretRef.name" {
		t.Errorf(
			"unexpected reference path %q",
			reference.Path,
		)
	}

	if reference.Relation != RelationUses {
		t.Errorf(
			"expected uses relation, got %q",
			reference.Relation,
		)
	}
}

func TestRuleFinderIgnoresUnavailableResource(
	t *testing.T,
) {
	t.Parallel()

	clientset := kubernetesfake.NewSimpleClientset()

	dynamicClient := dynamicfake.NewSimpleDynamicClient(
		runtime.NewScheme(),
	)

	finder := NewRuleFinder(
		clientset.Discovery(),
		dynamicClient,
		[]config.UsageRule{
			{
				APIVersions: []string{
					"example.io/v1",
					"example.io/v1beta1",
				},
				Kind:     "Application",
				Resource: "applications",
				References: []config.SecretReferenceRule{
					{
						Path:        "spec.secretName",
						Description: "application Secret",
						Relation:    config.RelationUses,
					},
				},
			},
		},
	)

	got := finder.Find(
		context.Background(),
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

func TestRuleFinderReturnsWarningOnForbiddenList(
	t *testing.T,
) {
	t.Parallel()

	const namespace = "test"

	gvr := schema.GroupVersionResource{
		Group:    "example.io",
		Version:  "v1",
		Resource: "applications",
	}

	clientset := kubernetesfake.NewSimpleClientset()

	discoveryClient := clientset.
		Discovery().(*discoveryfake.FakeDiscovery)

	discoveryClient.Resources = []*metav1.APIResourceList{
		{
			GroupVersion: "example.io/v1",
			APIResources: []metav1.APIResource{
				{
					Name:       "applications",
					Kind:       "Application",
					Namespaced: true,
					Verbs: metav1.Verbs{
						"get",
						"list",
					},
				},
			},
		},
	}

	dynamicClient :=
		dynamicfake.NewSimpleDynamicClientWithCustomListKinds(
			runtime.NewScheme(),
			map[schema.GroupVersionResource]string{
				gvr: "ApplicationList",
			},
		)

	dynamicClient.PrependReactor(
		"list",
		"applications",
		func(
			action ktesting.Action,
		) (bool, runtime.Object, error) {
			return true, nil, apierrors.NewForbidden(
				schema.GroupResource{
					Group:    "example.io",
					Resource: "applications",
				},
				"",
				errors.New("access denied"),
			)
		},
	)

	finder := NewRuleFinder(
		discoveryClient,
		dynamicClient,
		[]config.UsageRule{
			{
				APIVersions: []string{
					"example.io/v1",
				},
				Kind:     "Application",
				Resource: "applications",
				References: []config.SecretReferenceRule{
					{
						Path:        "spec.secretName",
						Description: "application Secret",
						Relation:    config.RelationUses,
					},
				},
			},
		},
	)

	got := finder.Find(
		context.Background(),
		namespace,
		"application-secret",
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

	if warning.Resource != "applications" {
		t.Errorf(
			"unexpected warning resource %q",
			warning.Resource,
		)
	}

	if warning.Err == nil {
		t.Fatal("expected warning error")
	}

	if !strings.Contains(
		warning.Err.Error(),
		"access denied",
	) {
		t.Errorf(
			"expected access denied warning, got %v",
			warning.Err,
		)
	}
}

func TestRuleFinderReturnsWarningForInvalidRule(
	t *testing.T,
) {
	t.Parallel()

	clientset := kubernetesfake.NewSimpleClientset()

	finder := NewRuleFinder(
		clientset.Discovery(),
		dynamicfake.NewSimpleDynamicClient(
			runtime.NewScheme(),
		),
		[]config.UsageRule{
			{
				APIVersions: []string{
					"example.io/v1",
				},
				Kind:     "Application",
				Resource: "applications",
				References: []config.SecretReferenceRule{
					{
						Path:     "spec.data[0].name",
						Relation: config.RelationUses,
					},
				},
			},
		},
	)

	got := finder.Find(
		context.Background(),
		"test",
		"application-secret",
	)

	if len(got.Usages) != 0 {
		t.Errorf(
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

	if got.Warnings[0].Resource != "applications" {
		t.Errorf(
			"unexpected warning resource %q",
			got.Warnings[0].Resource,
		)
	}
}

func TestRuleFinderWithNilClients(t *testing.T) {
	t.Parallel()

	finder := NewRuleFinder(
		nil,
		nil,
		[]config.UsageRule{
			{
				APIVersions: []string{
					"example.io/v1",
				},
				Kind:     "Application",
				Resource: "applications",
				References: []config.SecretReferenceRule{
					{
						Path:     "spec.secretName",
						Relation: config.RelationUses,
					},
				},
			},
		},
	)

	got := finder.Find(
		context.Background(),
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
