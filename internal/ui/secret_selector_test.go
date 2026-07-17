package ui

import (
	"reflect"
	"testing"
)

func TestNewSecretSelectorModelSortsSecrets(t *testing.T) {
	t.Parallel()

	model := newSecretSelectorModel(
		"default",
		[]string{
			"mongodb",
			"alpha-secret",
			"svc-api",
		},
	)

	expected := []string{
		"alpha-secret",
		"mongodb",
		"svc-api",
	}

	if !reflect.DeepEqual(model.allSecrets, expected) {
		t.Fatalf(
			"expected sorted Secrets %v, got %v",
			expected,
			model.allSecrets,
		)
	}

	if !reflect.DeepEqual(model.filteredSecrets, expected) {
		t.Fatalf(
			"expected filtered Secrets %v, got %v",
			expected,
			model.filteredSecrets,
		)
	}
}

func TestApplyFilterMatchesSecretsCaseInsensitively(t *testing.T) {
	t.Parallel()

	model := newSecretSelectorModel(
		"default",
		[]string{
			"growthbook-mongodb",
			"MongoDB-root-password",
			"postgres-credentials",
			"svc-webhook-mongodb",
		},
	)

	model.filter = []rune("mongo")
	model.cursor = 3
	model.page = 2
	model.applyFilter()

	expected := []string{
		"MongoDB-root-password",
		"growthbook-mongodb",
		"svc-webhook-mongodb",
	}

	if !reflect.DeepEqual(model.filteredSecrets, expected) {
		t.Fatalf(
			"expected filtered Secrets %v, got %v",
			expected,
			model.filteredSecrets,
		)
	}

	if model.cursor != 0 {
		t.Fatalf(
			"expected cursor to reset to 0, got %d",
			model.cursor,
		)
	}

	if model.page != 0 {
		t.Fatalf(
			"expected page to reset to 0, got %d",
			model.page,
		)
	}
}

func TestApplyFilterReturnsNoMatches(t *testing.T) {
	t.Parallel()

	model := newSecretSelectorModel(
		"default",
		[]string{
			"mongodb",
			"postgres",
		},
	)

	model.filter = []rune("redis")
	model.applyFilter()

	if len(model.filteredSecrets) != 0 {
		t.Fatalf(
			"expected no matching Secrets, got %v",
			model.filteredSecrets,
		)
	}

	if footer := model.footer(); footer != "0 matching Secrets" {
		t.Fatalf(
			"expected empty-filter footer, got %q",
			footer,
		)
	}
}

func TestEmptyFilterRestoresAllSecrets(t *testing.T) {
	t.Parallel()

	model := newSecretSelectorModel(
		"default",
		[]string{
			"mongodb",
			"postgres",
			"redis",
		},
	)

	model.filter = []rune("mongo")
	model.applyFilter()

	model.filter = nil
	model.applyFilter()

	if !reflect.DeepEqual(
		model.filteredSecrets,
		model.allSecrets,
	) {
		t.Fatalf(
			"expected all Secrets to be restored, got %v",
			model.filteredSecrets,
		)
	}
}

func TestPagination(t *testing.T) {
	t.Parallel()

	secrets := []string{
		"secret-01",
		"secret-02",
		"secret-03",
		"secret-04",
		"secret-05",
	}

	model := newSecretSelectorModel("default", secrets)
	model.pageSize = 2

	if pages := model.totalPages(); pages != 3 {
		t.Fatalf(
			"expected 3 pages, got %d",
			pages,
		)
	}

	visible, offset := model.visibleSecrets()

	expectedFirstPage := []string{
		"secret-01",
		"secret-02",
	}

	if !reflect.DeepEqual(visible, expectedFirstPage) {
		t.Fatalf(
			"expected first page %v, got %v",
			expectedFirstPage,
			visible,
		)
	}

	if offset != 0 {
		t.Fatalf(
			"expected first page offset 0, got %d",
			offset,
		)
	}

	model.nextPage()

	if model.page != 1 {
		t.Fatalf(
			"expected page 1 after moving forward, got %d",
			model.page,
		)
	}

	if model.cursor != 2 {
		t.Fatalf(
			"expected cursor 2 after moving forward, got %d",
			model.cursor,
		)
	}

	visible, offset = model.visibleSecrets()

	expectedSecondPage := []string{
		"secret-03",
		"secret-04",
	}

	if !reflect.DeepEqual(visible, expectedSecondPage) {
		t.Fatalf(
			"expected second page %v, got %v",
			expectedSecondPage,
			visible,
		)
	}

	if offset != 2 {
		t.Fatalf(
			"expected second page offset 2, got %d",
			offset,
		)
	}

	model.previousPage()

	if model.page != 0 {
		t.Fatalf(
			"expected page 0 after moving backward, got %d",
			model.page,
		)
	}

	if model.cursor != 0 {
		t.Fatalf(
			"expected cursor 0 after moving backward, got %d",
			model.cursor,
		)
	}
}

func TestPaginationDoesNotMoveBeyondLastPage(t *testing.T) {
	t.Parallel()

	model := newSecretSelectorModel(
		"default",
		[]string{
			"secret-01",
			"secret-02",
			"secret-03",
		},
	)

	model.pageSize = 2

	model.nextPage()
	model.nextPage()

	if model.page != 1 {
		t.Fatalf(
			"expected to remain on last page 1, got %d",
			model.page,
		)
	}
}

func TestFilteredResultsUseScrollableViewport(t *testing.T) {
	t.Parallel()

	model := newSecretSelectorModel(
		"default",
		[]string{
			"svc-01",
			"svc-02",
			"svc-03",
			"svc-04",
			"svc-05",
			"svc-06",
			"svc-07",
			"svc-08",
			"svc-09",
			"svc-10",
		},
	)

	model.filter = []rune("svc-")
	model.applyFilter()

	// windowHeight - 7 gives five visible items because the
	// implementation enforces a minimum viewport size of five.
	model.windowHeight = 10
	model.cursor = 6

	visible, offset := model.visibleSecrets()

	expected := []string{
		"svc-05",
		"svc-06",
		"svc-07",
		"svc-08",
		"svc-09",
	}

	if !reflect.DeepEqual(visible, expected) {
		t.Fatalf(
			"expected viewport %v, got %v",
			expected,
			visible,
		)
	}

	if offset != 4 {
		t.Fatalf(
			"expected viewport offset 4, got %d",
			offset,
		)
	}
}

func TestMoveDownAndMoveUp(t *testing.T) {
	t.Parallel()

	model := newSecretSelectorModel(
		"default",
		[]string{
			"secret-01",
			"secret-02",
			"secret-03",
		},
	)

	model.moveDown()

	if model.cursor != 1 {
		t.Fatalf(
			"expected cursor 1 after moving down, got %d",
			model.cursor,
		)
	}

	model.moveUp()

	if model.cursor != 0 {
		t.Fatalf(
			"expected cursor 0 after moving up, got %d",
			model.cursor,
		)
	}

	// Cursor must not move outside the available results.
	model.moveUp()

	if model.cursor != 0 {
		t.Fatalf(
			"expected cursor to remain at 0, got %d",
			model.cursor,
		)
	}

	model.moveDown()
	model.moveDown()
	model.moveDown()

	if model.cursor != 2 {
		t.Fatalf(
			"expected cursor to remain at last result 2, got %d",
			model.cursor,
		)
	}
}

func TestFilteredFooterShowsPosition(t *testing.T) {
	t.Parallel()

	model := newSecretSelectorModel(
		"default",
		[]string{
			"mongodb-01",
			"mongodb-02",
			"mongodb-03",
		},
	)

	model.filter = []rune("mongo")
	model.applyFilter()
	model.cursor = 1

	expected := "3 matching Secrets · 2/3 selected"

	if footer := model.footer(); footer != expected {
		t.Fatalf(
			"expected footer %q, got %q",
			expected,
			footer,
		)
	}
}

func TestUnfilteredFooterShowsPagination(t *testing.T) {
	t.Parallel()

	model := newSecretSelectorModel(
		"default",
		[]string{
			"secret-01",
			"secret-02",
			"secret-03",
		},
	)

	model.pageSize = 2
	model.nextPage()

	expected := "Page 2/2 · 3 Secrets"

	if footer := model.footer(); footer != expected {
		t.Fatalf(
			"expected footer %q, got %q",
			expected,
			footer,
		)
	}
}
