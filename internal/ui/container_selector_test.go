package ui

import (
	"reflect"
	"testing"
)

func TestContainerSelectorSortsAndFilters(t *testing.T) {
	t.Parallel()

	model := newContainerSelectorModel(
		"api-pod",
		[]string{"sidecar", "api", "metrics"},
	)

	want := []string{"api", "metrics", "sidecar"}
	if !reflect.DeepEqual(model.all, want) {
		t.Fatalf("containers = %v, want %v", model.all, want)
	}

	model.filter = []rune("MET")
	model.applyFilter()
	if !reflect.DeepEqual(model.filtered, []string{"metrics"}) {
		t.Fatalf("filtered containers = %v", model.filtered)
	}
}
