package usage

import (
	"context"
	"errors"
	"reflect"
	"testing"

	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

func TestIngressFinderFind(t *testing.T) {
	t.Parallel()

	const (
		namespace  = "test"
		secretName = "web-tls"
	)

	clientset := fake.NewSimpleClientset(
		&networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "web",
				Namespace: namespace,
			},
			Spec: networkingv1.IngressSpec{
				TLS: []networkingv1.IngressTLS{
					{
						Hosts: []string{
							"example.com",
						},
						SecretName: secretName,
					},
				},
			},
		},
		&networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "unrelated",
				Namespace: namespace,
			},
			Spec: networkingv1.IngressSpec{
				TLS: []networkingv1.IngressTLS{
					{
						SecretName: "another-secret",
					},
				},
			},
		},
	)

	finder := NewIngressFinder(clientset)

	got := finder.Find(
		context.Background(),
		namespace,
		secretName,
	)

	want := Result{
		Usages: []Usage{
			{
				APIVersion: "networking.k8s.io/v1",
				Kind:       "Ingress",
				Namespace:  namespace,
				Name:       "web",
				Source:     SourceBuiltIn,
				References: []Reference{
					{
						Description: "TLS certificate",
						Path:        "spec.tls[0].secretName",
						Relation:    RelationUses,
					},
				},
			},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf(
			"IngressFinder.Find() mismatch\n\ngot:  %#v\n\nwant: %#v",
			got,
			want,
		)
	}
}

func TestIngressFinderReturnsWarningOnForbidden(
	t *testing.T,
) {
	t.Parallel()

	const namespace = "test"

	clientset := fake.NewSimpleClientset()

	clientset.PrependReactor(
		"list",
		"ingresses",
		func(
			action ktesting.Action,
		) (bool, runtime.Object, error) {
			return true, nil, apierrors.NewForbidden(
				schema.GroupResource{
					Group:    "networking.k8s.io",
					Resource: "ingresses",
				},
				"",
				errors.New("access denied"),
			)
		},
	)

	finder := NewIngressFinder(clientset)

	got := finder.Find(
		context.Background(),
		namespace,
		"web-tls",
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

	if got.Warnings[0].Resource != "ingresses" {
		t.Errorf(
			"expected ingresses warning, got %q",
			got.Warnings[0].Resource,
		)
	}
}
