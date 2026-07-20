package ui

import (
	"reflect"
	"testing"
)

func TestContextSelectorSortsContexts(t *testing.T) {
	t.Parallel()

	model := newContextSelectorModel(
		[]string{"staging", "operations", "production"},
	)

	want := []string{
		"operations",
		"production",
		"staging",
	}

	if !reflect.DeepEqual(model.allContexts, want) {
		t.Fatalf(
			"expected sorted contexts %v, got %v",
			want,
			model.allContexts,
		)
	}
}

func TestContextSelectorFiltersCaseInsensitively(
	t *testing.T,
) {
	t.Parallel()

	model := newContextSelectorModel(
		[]string{
			"operations-admin",
			"operations-readonly",
			"staging",
		},
	)

	model.filter = []rune("OPERATIONS")
	model.applyFilter()

	want := []string{
		"operations-admin",
		"operations-readonly",
	}

	if !reflect.DeepEqual(model.filteredContexts, want) {
		t.Fatalf(
			"expected contexts %v, got %v",
			want,
			model.filteredContexts,
		)
	}
}
