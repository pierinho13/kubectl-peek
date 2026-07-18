package usage

import (
	"context"
	"fmt"
	"sort"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type WorkloadFinder struct {
	client kubernetes.Interface
}

func NewWorkloadFinder(
	client kubernetes.Interface,
) *WorkloadFinder {
	return &WorkloadFinder{
		client: client,
	}
}

func (finder *WorkloadFinder) Find(
	ctx context.Context,
	namespace string,
	secretName string,
) Result {
	var result Result

	if finder == nil || finder.client == nil {
		result.Warnings = append(
			result.Warnings,
			Warning{
				Resource: "workloads",
				Err:      fmt.Errorf("Kubernetes typed client is nil"),
			},
		)

		return result
	}

	finder.findPods(
		ctx,
		namespace,
		secretName,
		&result,
	)

	finder.findDeployments(
		ctx,
		namespace,
		secretName,
		&result,
	)

	finder.findStatefulSets(
		ctx,
		namespace,
		secretName,
		&result,
	)

	finder.findDaemonSets(
		ctx,
		namespace,
		secretName,
		&result,
	)

	finder.findJobs(
		ctx,
		namespace,
		secretName,
		&result,
	)

	finder.findCronJobs(
		ctx,
		namespace,
		secretName,
		&result,
	)

	return result
}

func (finder *WorkloadFinder) findPods(
	ctx context.Context,
	namespace string,
	secretName string,
	result *Result,
) {
	pods, err := finder.client.
		CoreV1().
		Pods(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		appendWorkloadWarning(
			result,
			"v1",
			"pods",
			err,
		)

		return
	}

	for i := range pods.Items {
		pod := &pods.Items[i]

		if metav1.GetControllerOf(pod) != nil {
			continue
		}

		appendPodSpecUsage(
			result,
			"v1",
			"Pod",
			pod.Namespace,
			pod.Name,
			&pod.Spec,
			secretName,
		)
	}
}

func (finder *WorkloadFinder) findDeployments(
	ctx context.Context,
	namespace string,
	secretName string,
	result *Result,
) {
	deployments, err := finder.client.
		AppsV1().
		Deployments(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		appendWorkloadWarning(
			result,
			"apps/v1",
			"deployments",
			err,
		)

		return
	}

	for i := range deployments.Items {
		deployment := &deployments.Items[i]

		appendPodSpecUsage(
			result,
			"apps/v1",
			"Deployment",
			deployment.Namespace,
			deployment.Name,
			&deployment.Spec.Template.Spec,
			secretName,
		)
	}
}

func (finder *WorkloadFinder) findStatefulSets(
	ctx context.Context,
	namespace string,
	secretName string,
	result *Result,
) {
	statefulSets, err := finder.client.
		AppsV1().
		StatefulSets(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		appendWorkloadWarning(
			result,
			"apps/v1",
			"statefulsets",
			err,
		)

		return
	}

	for i := range statefulSets.Items {
		statefulSet := &statefulSets.Items[i]

		appendPodSpecUsage(
			result,
			"apps/v1",
			"StatefulSet",
			statefulSet.Namespace,
			statefulSet.Name,
			&statefulSet.Spec.Template.Spec,
			secretName,
		)
	}
}

func (finder *WorkloadFinder) findDaemonSets(
	ctx context.Context,
	namespace string,
	secretName string,
	result *Result,
) {
	daemonSets, err := finder.client.
		AppsV1().
		DaemonSets(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		appendWorkloadWarning(
			result,
			"apps/v1",
			"daemonsets",
			err,
		)

		return
	}

	for i := range daemonSets.Items {
		daemonSet := &daemonSets.Items[i]

		appendPodSpecUsage(
			result,
			"apps/v1",
			"DaemonSet",
			daemonSet.Namespace,
			daemonSet.Name,
			&daemonSet.Spec.Template.Spec,
			secretName,
		)
	}
}

func (finder *WorkloadFinder) findJobs(
	ctx context.Context,
	namespace string,
	secretName string,
	result *Result,
) {
	jobs, err := finder.client.
		BatchV1().
		Jobs(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		appendWorkloadWarning(
			result,
			"batch/v1",
			"jobs",
			err,
		)

		return
	}

	for i := range jobs.Items {
		job := &jobs.Items[i]

		appendPodSpecUsage(
			result,
			"batch/v1",
			"Job",
			job.Namespace,
			job.Name,
			&job.Spec.Template.Spec,
			secretName,
		)
	}
}

func (finder *WorkloadFinder) findCronJobs(
	ctx context.Context,
	namespace string,
	secretName string,
	result *Result,
) {
	cronJobs, err := finder.client.
		BatchV1().
		CronJobs(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		appendWorkloadWarning(
			result,
			"batch/v1",
			"cronjobs",
			err,
		)

		return
	}

	for i := range cronJobs.Items {
		cronJob := &cronJobs.Items[i]

		appendPodSpecUsage(
			result,
			"batch/v1",
			"CronJob",
			cronJob.Namespace,
			cronJob.Name,
			&cronJob.Spec.JobTemplate.Spec.Template.Spec,
			secretName,
		)
	}
}

func appendWorkloadWarning(
	result *Result,
	apiVersion string,
	resource string,
	err error,
) {
	result.Warnings = append(
		result.Warnings,
		Warning{
			APIVersion: apiVersion,
			Resource:   resource,
			Err: fmt.Errorf(
				"list %s: %w",
				resource,
				err,
			),
		},
	)
}

func appendPodSpecUsage(
	result *Result,
	apiVersion string,
	kind string,
	namespace string,
	name string,
	podSpec *corev1.PodSpec,
	secretName string,
) {
	references := findPodSpecSecretReferences(
		podSpec,
		secretName,
	)
	if len(references) == 0 {
		return
	}

	result.Usages = append(
		result.Usages,
		Usage{
			APIVersion: apiVersion,
			Kind:       kind,
			Namespace:  namespace,
			Name:       name,
			Source:     SourceBuiltIn,
			References: references,
		},
	)
}

func findPodSpecSecretReferences(
	podSpec *corev1.PodSpec,
	secretName string,
) []Reference {
	if podSpec == nil {
		return nil
	}

	referenceMap := make(map[string]Reference)

	for _, imagePullSecret := range podSpec.ImagePullSecrets {
		if imagePullSecret.Name != secretName {
			continue
		}

		referenceMap["imagePullSecret"] = Reference{
			Description: "image pull secret",
			Path:        "imagePullSecret",
			Relation:    RelationUses,
		}
	}

	for _, volume := range podSpec.Volumes {
		findVolumeSecretReferences(
			referenceMap,
			volume,
			secretName,
		)
	}

	findContainerSecretReferences(
		referenceMap,
		"container",
		podSpec.Containers,
		secretName,
	)

	findContainerSecretReferences(
		referenceMap,
		"initContainer",
		podSpec.InitContainers,
		secretName,
	)

	findContainerSecretReferences(
		referenceMap,
		"ephemeralContainer",
		ephemeralContainersToContainers(
			podSpec.EphemeralContainers,
		),
		secretName,
	)

	paths := make(
		[]string,
		0,
		len(referenceMap),
	)

	for path := range referenceMap {
		paths = append(paths, path)
	}

	sort.Strings(paths)

	references := make(
		[]Reference,
		0,
		len(paths),
	)

	for _, path := range paths {
		references = append(
			references,
			referenceMap[path],
		)
	}

	return references
}

func findVolumeSecretReferences(
	references map[string]Reference,
	volume corev1.Volume,
	secretName string,
) {
	if volume.Secret != nil &&
		volume.Secret.SecretName == secretName {
		path := fmt.Sprintf(
			"volume/%s",
			volume.Name,
		)

		references[path] = Reference{
			Description: "Secret volume",
			Path:        path,
			Relation:    RelationUses,
		}
	}

	if volume.Projected == nil {
		return
	}

	for sourceIndex, source := range volume.Projected.Sources {
		if source.Secret == nil ||
			source.Secret.Name != secretName {
			continue
		}

		path := fmt.Sprintf(
			"volume/%s projected[%d]",
			volume.Name,
			sourceIndex,
		)

		references[path] = Reference{
			Description: "projected Secret volume",
			Path:        path,
			Relation:    RelationUses,
		}
	}
}

func findContainerSecretReferences(
	references map[string]Reference,
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

			path := fmt.Sprintf(
				"%s/%s envFrom",
				containerType,
				container.Name,
			)

			references[path] = Reference{
				Description: "container environment",
				Path:        path,
				Relation:    RelationUses,
			}
		}

		for _, env := range container.Env {
			if env.ValueFrom == nil ||
				env.ValueFrom.SecretKeyRef == nil ||
				env.ValueFrom.SecretKeyRef.Name != secretName {
				continue
			}

			secretKeyRef := env.ValueFrom.SecretKeyRef

			path := fmt.Sprintf(
				"%s/%s env/%s -> %s",
				containerType,
				container.Name,
				env.Name,
				secretKeyRef.Key,
			)

			references[path] = Reference{
				Description: "container environment variable",
				Path:        path,
				Key:         secretKeyRef.Key,
				Relation:    RelationUses,
			}
		}
	}
}

func ephemeralContainersToContainers(
	ephemeralContainers []corev1.EphemeralContainer,
) []corev1.Container {
	containers := make(
		[]corev1.Container,
		0,
		len(ephemeralContainers),
	)

	for _, ephemeralContainer := range ephemeralContainers {
		containers = append(
			containers,
			corev1.Container(
				ephemeralContainer.EphemeralContainerCommon,
			),
		)
	}

	return containers
}
