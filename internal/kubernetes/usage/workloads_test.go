package usage

import (
	"context"
	"errors"
	"reflect"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

func TestWorkloadFinderFind(t *testing.T) {
	t.Parallel()

	const (
		namespace  = "test"
		secretName = "database-credentials"
	)

	controller := true

	clientset := fake.NewSimpleClientset(
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "standalone",
				Namespace: namespace,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name: "application",
						Env: []corev1.EnvVar{
							{
								Name: "PASSWORD",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: secretName,
										},
										Key: "password",
									},
								},
							},
						},
					},
				},
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "managed",
				Namespace: namespace,
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: "apps/v1",
						Kind:       "ReplicaSet",
						Name:       "backend-123",
						Controller: &controller,
					},
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name: "application",
						EnvFrom: []corev1.EnvFromSource{
							{
								SecretRef: &corev1.SecretEnvSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: secretName,
									},
								},
							},
						},
					},
				},
			},
		},
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "backend",
				Namespace: namespace,
			},
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "backend",
								EnvFrom: []corev1.EnvFromSource{
									{
										SecretRef: &corev1.SecretEnvSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: secretName,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	)

	finder := NewWorkloadFinder(clientset)

	got := finder.Find(
		context.Background(),
		namespace,
		secretName,
	)

	want := Result{
		Usages: []Usage{
			{
				APIVersion: "v1",
				Kind:       "Pod",
				Namespace:  namespace,
				Name:       "standalone",
				Source:     SourceBuiltIn,
				References: []Reference{
					{
						Description: "container environment variable",
						Path:        "container/application env/PASSWORD -> password",
						Key:         "password",
						Relation:    RelationUses,
					},
				},
			},
			{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Namespace:  namespace,
				Name:       "backend",
				Source:     SourceBuiltIn,
				References: []Reference{
					{
						Description: "container environment",
						Path:        "container/backend envFrom",
						Relation:    RelationUses,
					},
				},
			},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf(
			"WorkloadFinder.Find() mismatch\n\ngot:  %#v\n\nwant: %#v",
			got,
			want,
		)
	}
}

func TestWorkloadFinderContinuesAfterForbiddenPods(
	t *testing.T,
) {
	t.Parallel()

	const (
		namespace  = "test"
		secretName = "database-credentials"
	)

	clientset := fake.NewSimpleClientset(
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "backend",
				Namespace: namespace,
			},
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "backend",
								EnvFrom: []corev1.EnvFromSource{
									{
										SecretRef: &corev1.SecretEnvSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: secretName,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	)

	clientset.PrependReactor(
		"list",
		"pods",
		func(
			action ktesting.Action,
		) (bool, runtime.Object, error) {
			return true, nil, apierrors.NewForbidden(
				schema.GroupResource{
					Resource: "pods",
				},
				"",
				errors.New("access denied"),
			)
		},
	)

	finder := NewWorkloadFinder(clientset)

	got := finder.Find(
		context.Background(),
		namespace,
		secretName,
	)

	if len(got.Warnings) != 1 {
		t.Fatalf(
			"expected one warning, got %#v",
			got.Warnings,
		)
	}

	if got.Warnings[0].Resource != "pods" {
		t.Errorf(
			"expected pods warning, got %q",
			got.Warnings[0].Resource,
		)
	}

	if !apierrors.IsForbidden(got.Warnings[0].Err) {
		t.Errorf(
			"expected forbidden warning, got %v",
			got.Warnings[0].Err,
		)
	}

	if len(got.Usages) != 1 {
		t.Fatalf(
			"expected Deployment usage despite Pods failure, got %#v",
			got.Usages,
		)
	}

	if got.Usages[0].Kind != "Deployment" ||
		got.Usages[0].Name != "backend" {
		t.Errorf(
			"expected Deployment/backend, got %#v",
			got.Usages[0],
		)
	}
}
