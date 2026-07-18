package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLoadUsageRules(t *testing.T) {
	t.Parallel()

	path := writeUsageRulesFile(
		t,
		`
rules:
  - apiVersions:
      - example.io/v1
      - example.io/v1beta1
    kind: Application
    resource: applications
    references:
      - path: spec.credentials.secretName
        description: application credentials
        relation: uses
      - path: spec.generatedSecret.name
        description: generated Secret
        relation: produces
`,
	)

	got, err := LoadUsageRules(path)
	if err != nil {
		t.Fatalf(
			"LoadUsageRules() error = %v",
			err,
		)
	}

	want := []UsageRule{
		{
			APIVersions: []string{
				"example.io/v1",
				"example.io/v1beta1",
			},
			Kind:     "Application",
			Resource: "applications",
			References: []SecretReferenceRule{
				{
					Path:        "spec.credentials.secretName",
					Description: "application credentials",
					Relation:    RelationUses,
				},
				{
					Path:        "spec.generatedSecret.name",
					Description: "generated Secret",
					Relation:    RelationProduces,
				},
			},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf(
			"LoadUsageRules() mismatch\n\ngot:  %#v\n\nwant: %#v",
			got,
			want,
		)
	}
}

func TestLoadUsageRulesReturnsNilForEmptyPath(
	t *testing.T,
) {
	t.Parallel()

	got, err := LoadUsageRules("")
	if err != nil {
		t.Fatalf(
			"LoadUsageRules() error = %v",
			err,
		)
	}

	if got != nil {
		t.Errorf(
			"expected nil rules, got %#v",
			got,
		)
	}
}

func TestLoadUsageRulesReturnsErrorForUnknownField(
	t *testing.T,
) {
	t.Parallel()

	path := writeUsageRulesFile(
		t,
		`
rules:
  - apiVersion:
      - example.io/v1
    kind: Application
    resource: applications
    references:
      - path: spec.secretName
        relation: uses
`,
	)

	_, err := LoadUsageRules(path)
	if err == nil {
		t.Fatal(
			"expected unknown field error",
		)
	}

	if !strings.Contains(
		err.Error(),
		"apiVersion",
	) {
		t.Errorf(
			"expected apiVersion error, got %v",
			err,
		)
	}
}

func TestLoadUsageRulesReturnsErrorForEmptyRules(
	t *testing.T,
) {
	t.Parallel()

	path := writeUsageRulesFile(
		t,
		`
rules: []
`,
	)

	_, err := LoadUsageRules(path)
	if err == nil {
		t.Fatal(
			"expected empty rules error",
		)
	}

	if !strings.Contains(
		err.Error(),
		"contains no rules",
	) {
		t.Errorf(
			"unexpected error: %v",
			err,
		)
	}
}

func TestLoadUsageRulesReturnsErrorForMissingFile(
	t *testing.T,
) {
	t.Parallel()

	path := filepath.Join(
		t.TempDir(),
		"missing.yaml",
	)

	_, err := LoadUsageRules(path)
	if err == nil {
		t.Fatal(
			"expected missing file error",
		)
	}

	if !strings.Contains(
		err.Error(),
		"read usage rules file",
	) {
		t.Errorf(
			"unexpected error: %v",
			err,
		)
	}
}

func writeUsageRulesFile(
	t *testing.T,
	content string,
) string {
	t.Helper()

	path := filepath.Join(
		t.TempDir(),
		"rules.yaml",
	)

	if err := os.WriteFile(
		path,
		[]byte(content),
		0o600,
	); err != nil {
		t.Fatalf(
			"write usage rules file: %v",
			err,
		)
	}

	return path
}
