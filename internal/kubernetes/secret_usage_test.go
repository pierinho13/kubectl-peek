package kubernetes

import (
	"context"
	"reflect"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"errors"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	discoveryfake "k8s.io/client-go/discovery/fake"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	ktesting "k8s.io/client-go/testing"
)

func TestFindSecretUsages(t *testing.T) {
	t.Parallel()

	const (
		namespace  = "test"
		secretName = "database-credentials"
	)

	clientset := fake.NewSimpleClientset(
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "api-pod",
				Namespace: namespace,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name: "api",
						Env: []corev1.EnvVar{
							{
								Name: "DATABASE_PASSWORD",
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
		&appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "database",
				Namespace: namespace,
			},
			Spec: appsv1.StatefulSetSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "database",
								Image: "postgres",
							},
						},
						Volumes: []corev1.Volume{
							{
								Name: "credentials",
								VolumeSource: corev1.VolumeSource{
									Secret: &corev1.SecretVolumeSource{
										SecretName: secretName,
									},
								},
							},
						},
					},
				},
			},
		},
		&appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "agent",
				Namespace: namespace,
			},
			Spec: appsv1.DaemonSetSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "agent",
								Image: "agent",
							},
						},
						ImagePullSecrets: []corev1.LocalObjectReference{
							{
								Name: secretName,
							},
						},
					},
				},
			},
		},
		&batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "migration",
				Namespace: namespace,
			},
			Spec: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						RestartPolicy: corev1.RestartPolicyNever,
						InitContainers: []corev1.Container{
							{
								Name: "prepare",
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
						Containers: []corev1.Container{
							{
								Name:  "migration",
								Image: "migration",
							},
						},
					},
				},
			},
		},
		&batchv1.CronJob{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "backup",
				Namespace: namespace,
			},
			Spec: batchv1.CronJobSpec{
				JobTemplate: batchv1.JobTemplateSpec{
					Spec: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								RestartPolicy: corev1.RestartPolicyNever,
								Containers: []corev1.Container{
									{
										Name: "backup",
										Env: []corev1.EnvVar{
											{
												Name: "BACKUP_PASSWORD",
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
					},
				},
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "projected-pod",
				Namespace: namespace,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "projected",
						Image: "busybox",
					},
				},
				Volumes: []corev1.Volume{
					{
						Name: "projected-credentials",
						VolumeSource: corev1.VolumeSource{
							Projected: &corev1.ProjectedVolumeSource{
								Sources: []corev1.VolumeProjection{
									{
										Secret: &corev1.SecretProjection{
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
		&corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "registry-user",
				Namespace: namespace,
			},
			ImagePullSecrets: []corev1.LocalObjectReference{
				{
					Name: secretName,
				},
			},
		},
		&networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "web",
				Namespace: namespace,
			},
			Spec: networkingv1.IngressSpec{
				TLS: []networkingv1.IngressTLS{
					{
						SecretName: secretName,
						Hosts: []string{
							"example.com",
						},
					},
				},
			},
		},

		// Este recurso referencia otro Secret y no debe aparecer.
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "unrelated",
				Namespace: namespace,
			},
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "unrelated",
								EnvFrom: []corev1.EnvFromSource{
									{
										SecretRef: &corev1.SecretEnvSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "another-secret",
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
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "backend-managed-pod",
				Namespace: namespace,
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: "apps/v1",
						Kind:       "ReplicaSet",
						Name:       "backend-7c6f8f9d7",
						Controller: boolPointer(true),
					},
				},
			},
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
	)

	client := &Client{
		Clientset: clientset,
	}

	got, err := FindSecretUsages(
		context.Background(),
		client,
		namespace,
		secretName,
	)
	if err != nil {
		t.Fatalf("FindSecretUsages() error = %v", err)
	}

	want := []SecretUsage{
		{
			Kind: "CronJob",
			Name: "backup",
			References: []SecretUsageReference{
				{
					Description: "container environment variable",
					Path:        "container/backup env/BACKUP_PASSWORD -> password",
					Key:         "password",
					Relation:    "uses",
				},
			},
		},
		{
			Kind: "DaemonSet",
			Name: "agent",
			References: []SecretUsageReference{
				{
					Description: "image pull secret",
					Path:        "imagePullSecret",
					Relation:    "uses",
				},
			},
		},
		{
			Kind: "Deployment",
			Name: "backend",
			References: []SecretUsageReference{
				{
					Description: "container environment",
					Path:        "container/backend envFrom",
					Relation:    "uses",
				},
			},
		},
		{
			Kind: "Ingress",
			Name: "web",
			References: []SecretUsageReference{
				{
					Description: "TLS certificate",
					Path:        "spec.tls[0].secretName",
					Relation:    "uses",
				},
			},
		},
		{
			Kind: "Job",
			Name: "migration",
			References: []SecretUsageReference{
				{
					Description: "container environment",
					Path:        "initContainer/prepare envFrom",
					Relation:    "uses",
				},
			},
		},
		{
			Kind: "Pod",
			Name: "api-pod",
			References: []SecretUsageReference{
				{
					Description: "container environment variable",
					Path:        "container/api env/DATABASE_PASSWORD -> password",
					Key:         "password",
					Relation:    "uses",
				},
			},
		},
		{
			Kind: "Pod",
			Name: "projected-pod",
			References: []SecretUsageReference{
				{
					Description: "projected Secret volume",
					Path:        "volume/projected-credentials projected[0]",
					Relation:    "uses",
				},
			},
		},
		{
			Kind: "ServiceAccount",
			Name: "registry-user",
			References: []SecretUsageReference{
				{
					Description: "image pull secret",
					Path:        "imagePullSecrets[0].name",
					Relation:    "uses",
				},
			},
		},
		{
			Kind: "StatefulSet",
			Name: "database",
			References: []SecretUsageReference{
				{
					Description: "Secret volume",
					Path:        "volume/credentials",
					Relation:    "uses",
				},
			},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf(
			"FindSecretUsages() mismatch\n\ngot:  %#v\n\nwant: %#v",
			got,
			want,
		)
	}
}
func boolPointer(value bool) *bool {
	return &value
}
func TestFindSecretUsagesDetailedPreservesGatewayWarning(
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

	clientset := fake.NewSimpleClientset()

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

	client := &Client{
		Clientset: clientset,
		Discovery: discoveryClient,
		Dynamic:   dynamicClient,
		Namespace: namespace,
	}

	got, err := FindSecretUsagesDetailed(
		context.Background(),
		client,
		namespace,
		secretName,
	)
	if err != nil {
		t.Fatalf(
			"FindSecretUsagesDetailed() error = %v",
			err,
		)
	}

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

	if got.Warnings[0].Resource != "gateways" {
		t.Errorf(
			"expected gateways warning, got %q",
			got.Warnings[0].Resource,
		)
	}

	if !apierrors.IsForbidden(got.Warnings[0].Err) {
		t.Errorf(
			"expected forbidden warning, got %v",
			got.Warnings[0].Err,
		)
	}
}
