package ui

import (
	"reflect"
	"testing"
)

func TestNamespaceSelectorSortsNamespaces(t *testing.T) {
	t.Parallel()

	model := newNamespaceSelectorModel(
		[]string{"monitoring", "default", "kube-system"},
	)

	want := []string{"default", "kube-system", "monitoring"}

	if !reflect.DeepEqual(model.allNamespaces, want) {
		t.Fatalf(
			"expected sorted namespaces %v, got %v",
			want,
			model.allNamespaces,
		)
	}
}

func TestNamespaceSelectorFiltersCaseInsensitively(
	t *testing.T,
) {
	t.Parallel()

	model := newNamespaceSelectorModel(
		[]string{
			"monitoring",
			"devbox-sre-3",
			"devbox-platform",
		},
	)

	model.filter = []rune("DEVBOX")
	model.applyFilter()

	want := []string{
		"devbox-platform",
		"devbox-sre-3",
	}

	if !reflect.DeepEqual(model.filteredNamespaces, want) {
		t.Fatalf(
			"expected namespaces %v, got %v",
			want,
			model.filteredNamespaces,
		)
	}
}

func TestNamespaceSelectorPagination(t *testing.T) {
	t.Parallel()

	model := newNamespaceSelectorModel(
		[]string{"ns-01", "ns-02", "ns-03"},
	)
	model.pageSize = 2
	model.nextPage()

	if model.page != 1 {
		t.Fatalf("expected page 1, got %d", model.page)
	}

	if model.cursor != 2 {
		t.Fatalf("expected cursor 2, got %d", model.cursor)
	}
}
