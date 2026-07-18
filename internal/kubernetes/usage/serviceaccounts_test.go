package usage

import (
	"context"
	"errors"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

func TestServiceAccountFinderFind(t *testing.T) {
	t.Parallel()

	const (
		namespace  = "test"
		secretName = "registry-credentials"
	)

	clientset := fake.NewSimpleClientset(
		&corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "application",
				Namespace: namespace,
			},
			ImagePullSecrets: []corev1.LocalObjectReference{
				{
					Name: secretName,
				},
			},
		},
		&corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "unrelated",
				Namespace: namespace,
			},
			ImagePullSecrets: []corev1.LocalObjectReference{
				{
					Name: "another-secret",
				},
			},
		},
	)

	finder := NewServiceAccountFinder(clientset)

	got := finder.Find(
		context.Background(),
		namespace,
		secretName,
	)

	want := Result{
		Usages: []Usage{
			{
				APIVersion: "v1",
				Kind:       "ServiceAccount",
				Namespace:  namespace,
				Name:       "application",
				Source:     SourceBuiltIn,
				References: []Reference{
					{
						Description: "image pull secret",
						Path:        "imagePullSecrets[0].name",
						Relation:    RelationUses,
					},
				},
			},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf(
			"ServiceAccountFinder.Find() mismatch\n\ngot:  %#v\n\nwant: %#v",
			got,
			want,
		)
	}
}

func TestServiceAccountFinderReturnsWarningOnForbidden(
	t *testing.T,
) {
	t.Parallel()

	const namespace = "test"

	clientset := fake.NewSimpleClientset()

	clientset.PrependReactor(
		"list",
		"serviceaccounts",
		func(
			action ktesting.Action,
		) (bool, runtime.Object, error) {
			return true, nil, apierrors.NewForbidden(
				schema.GroupResource{
					Resource: "serviceaccounts",
				},
				"",
				errors.New("access denied"),
			)
		},
	)

	finder := NewServiceAccountFinder(clientset)

	got := finder.Find(
		context.Background(),
		namespace,
		"registry-credentials",
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

	if got.Warnings[0].Resource != "serviceaccounts" {
		t.Errorf(
			"expected serviceaccounts warning, got %q",
			got.Warnings[0].Resource,
		)
	}
}
