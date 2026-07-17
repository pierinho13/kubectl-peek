package kubernetes

import (
	"context"
	"reflect"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
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
							{Name: secretName},
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
	)

	got, err := FindSecretUsages(
		context.Background(),
		clientset,
		namespace,
		secretName,
	)
	if err != nil {
		t.Fatalf("FindSecretUsages() error = %v", err)
	}

	want := []SecretUsage{
		{
			Kind:       "CronJob",
			Name:       "backup",
			References: []string{"container/backup env/BACKUP_PASSWORD -> password"},
		},
		{
			Kind:       "DaemonSet",
			Name:       "agent",
			References: []string{"imagePullSecret"},
		},
		{
			Kind:       "Deployment",
			Name:       "backend",
			References: []string{"container/backend envFrom"},
		},
		{
			Kind:       "Job",
			Name:       "migration",
			References: []string{"initContainer/prepare envFrom"},
		},
		{
			Kind:       "Pod",
			Name:       "api-pod",
			References: []string{"container/api env/DATABASE_PASSWORD -> password"},
		},
		{
			Kind:       "StatefulSet",
			Name:       "database",
			References: []string{"volume/credentials"},
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
