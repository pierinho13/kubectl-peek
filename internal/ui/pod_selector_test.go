package ui

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestPodSelectorSortsAndFiltersPods(t *testing.T) {
	t.Parallel()

	model := newPodSelectorModel(
		"test",
		[]PodOption{
			{Name: "worker-2", Phase: corev1.PodRunning},
			{Name: "api-1", Phase: corev1.PodRunning},
			{Name: "worker-1", Phase: corev1.PodPending},
		},
	)

	wantSorted := []string{"api-1", "worker-1", "worker-2"}
	gotSorted := make([]string, 0, len(model.allPods))
	for _, pod := range model.allPods {
		gotSorted = append(gotSorted, pod.Name)
	}
	if !reflect.DeepEqual(gotSorted, wantSorted) {
		t.Fatalf("sorted Pods = %v, want %v", gotSorted, wantSorted)
	}

	model.filter = []rune("WORKER")
	model.applyFilter()

	wantFiltered := []string{"worker-1", "worker-2"}
	gotFiltered := make([]string, 0, len(model.filteredPods))
	for _, pod := range model.filteredPods {
		gotFiltered = append(gotFiltered, pod.Name)
	}
	if !reflect.DeepEqual(gotFiltered, wantFiltered) {
		t.Fatalf("filtered Pods = %v, want %v", gotFiltered, wantFiltered)
	}
}

func TestPodSelectorPagination(t *testing.T) {
	t.Parallel()

	model := newPodSelectorModel(
		"test",
		[]PodOption{{Name: "pod-1"}, {Name: "pod-2"}, {Name: "pod-3"}},
	)
	model.pageSize = 2
	model.nextPage()

	if model.page != 1 || model.cursor != 2 {
		t.Fatalf("page/cursor = %d/%d, want 1/2", model.page, model.cursor)
	}
}
