package kubernetes

import (
	"context"
	"fmt"
	"sort"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type SecretUsage struct {
	Kind       string
	Name       string
	References []string
}

func FindSecretUsages(
	ctx context.Context,
	clientset kubernetes.Interface,
	namespace string,
	secretName string,
) ([]SecretUsage, error) {
	var usages []SecretUsage

	pods, err := clientset.CoreV1().
		Pods(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list pods: %w", err)
	}

	for i := range pods.Items {
		pod := &pods.Items[i]

		references := findSecretReferences(&pod.Spec, secretName)
		if len(references) == 0 {
			continue
		}

		usages = append(usages, SecretUsage{
			Kind:       "Pod",
			Name:       pod.Name,
			References: references,
		})
	}

	deployments, err := clientset.AppsV1().
		Deployments(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list deployments: %w", err)
	}

	for i := range deployments.Items {
		deployment := &deployments.Items[i]
		appendWorkloadUsage(
			&usages,
			"Deployment",
			deployment.Name,
			&deployment.Spec.Template.Spec,
			secretName,
		)
	}

	statefulSets, err := clientset.AppsV1().
		StatefulSets(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list statefulsets: %w", err)
	}

	for i := range statefulSets.Items {
		statefulSet := &statefulSets.Items[i]
		appendWorkloadUsage(
			&usages,
			"StatefulSet",
			statefulSet.Name,
			&statefulSet.Spec.Template.Spec,
			secretName,
		)
	}

	daemonSets, err := clientset.AppsV1().
		DaemonSets(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list daemonsets: %w", err)
	}

	for i := range daemonSets.Items {
		daemonSet := &daemonSets.Items[i]
		appendWorkloadUsage(
			&usages,
			"DaemonSet",
			daemonSet.Name,
			&daemonSet.Spec.Template.Spec,
			secretName,
		)
	}

	jobs, err := clientset.BatchV1().
		Jobs(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list jobs: %w", err)
	}

	for i := range jobs.Items {
		job := &jobs.Items[i]
		appendWorkloadUsage(
			&usages,
			"Job",
			job.Name,
			&job.Spec.Template.Spec,
			secretName,
		)
	}

	cronJobs, err := clientset.BatchV1().
		CronJobs(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list cronjobs: %w", err)
	}

	for i := range cronJobs.Items {
		cronJob := &cronJobs.Items[i]
		appendWorkloadUsage(
			&usages,
			"CronJob",
			cronJob.Name,
			&cronJob.Spec.JobTemplate.Spec.Template.Spec,
			secretName,
		)
	}

	sort.Slice(usages, func(i, j int) bool {
		if usages[i].Kind == usages[j].Kind {
			return usages[i].Name < usages[j].Name
		}

		return usages[i].Kind < usages[j].Kind
	})

	return usages, nil
}

func appendWorkloadUsage(
	usages *[]SecretUsage,
	kind string,
	name string,
	podSpec *corev1.PodSpec,
	secretName string,
) {
	references := findSecretReferences(podSpec, secretName)
	if len(references) == 0 {
		return
	}

	*usages = append(*usages, SecretUsage{
		Kind:       kind,
		Name:       name,
		References: references,
	})
}

func findSecretReferences(
	podSpec *corev1.PodSpec,
	secretName string,
) []string {
	references := make(map[string]struct{})

	for _, imagePullSecret := range podSpec.ImagePullSecrets {
		if imagePullSecret.Name == secretName {
			references["imagePullSecret"] = struct{}{}
		}
	}

	for _, volume := range podSpec.Volumes {
		if volume.Secret != nil && volume.Secret.SecretName == secretName {
			references[fmt.Sprintf("volume/%s", volume.Name)] = struct{}{}
		}
	}

	findContainerSecretReferences(
		references,
		"container",
		podSpec.Containers,
		secretName,
	)

	findContainerSecretReferences(
		references,
		"initContainer",
		podSpec.InitContainers,
		secretName,
	)

	findContainerSecretReferences(
		references,
		"ephemeralContainer",
		ephemeralContainersToContainers(podSpec.EphemeralContainers),
		secretName,
	)

	result := make([]string, 0, len(references))
	for reference := range references {
		result = append(result, reference)
	}

	sort.Strings(result)

	return result
}

func findContainerSecretReferences(
	references map[string]struct{},
	containerType string,
	containers []corev1.Container,
	secretName string,
) {
	for _, container := range containers {
		for _, envFrom := range container.EnvFrom {
			if envFrom.SecretRef == nil ||
				envFrom.SecretRef.Name != secretName {
				continue
			}

			references[fmt.Sprintf(
				"%s/%s envFrom",
				containerType,
				container.Name,
			)] = struct{}{}
		}

		for _, env := range container.Env {
			if env.ValueFrom == nil ||
				env.ValueFrom.SecretKeyRef == nil ||
				env.ValueFrom.SecretKeyRef.Name != secretName {
				continue
			}

			references[fmt.Sprintf(
				"%s/%s env/%s -> %s",
				containerType,
				container.Name,
				env.Name,
				env.ValueFrom.SecretKeyRef.Key,
			)] = struct{}{}
		}
	}
}

func ephemeralContainersToContainers(
	ephemeralContainers []corev1.EphemeralContainer,
) []corev1.Container {
	containers := make([]corev1.Container, 0, len(ephemeralContainers))

	for _, ephemeralContainer := range ephemeralContainers {
		containers = append(
			containers,
			corev1.Container(ephemeralContainer.EphemeralContainerCommon),
		)
	}

	return containers
}

// These declarations ensure the Kubernetes API packages remain explicit
// dependencies of this file and make the supported workload types obvious.
var (
	_ appsv1.Deployment
	_ batchv1.Job
)
