package usage

import (
	"reflect"
	"testing"
)

func TestFindStringValuesAtPathSimplePath(t *testing.T) {
	t.Parallel()

	object := map[string]interface{}{
		"spec": map[string]interface{}{
			"target": map[string]interface{}{
				"name": "application-secret",
			},
		},
	}

	got, err := findStringValuesAtPath(
		object,
		"spec.target.name",
	)
	if err != nil {
		t.Fatalf(
			"findStringValuesAtPath() error = %v",
			err,
		)
	}

	want := []string{
		"application-secret",
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf(
			"findStringValuesAtPath() mismatch\n\ngot:  %#v\n\nwant: %#v",
			got,
			want,
		)
	}
}

func TestFindStringValuesAtPathMetadataName(t *testing.T) {
	t.Parallel()

	object := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name": "sealed-credentials",
		},
	}

	got, err := findStringValuesAtPath(
		object,
		"metadata.name",
	)
	if err != nil {
		t.Fatalf(
			"findStringValuesAtPath() error = %v",
			err,
		)
	}

	want := []string{
		"sealed-credentials",
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf(
			"findStringValuesAtPath() mismatch\n\ngot:  %#v\n\nwant: %#v",
			got,
			want,
		)
	}
}

func TestFindStringValuesAtPathWildcard(t *testing.T) {
	t.Parallel()

	object := map[string]interface{}{
		"spec": map[string]interface{}{
			"data": []interface{}{
				map[string]interface{}{
					"secretKeyRef": map[string]interface{}{
						"name": "first-secret",
					},
				},
				map[string]interface{}{
					"secretKeyRef": map[string]interface{}{
						"name": "second-secret",
					},
				},
				map[string]interface{}{
					"secretKeyRef": map[string]interface{}{
						"name": "third-secret",
					},
				},
			},
		},
	}

	got, err := findStringValuesAtPath(
		object,
		"spec.data[*].secretKeyRef.name",
	)
	if err != nil {
		t.Fatalf(
			"findStringValuesAtPath() error = %v",
			err,
		)
	}

	want := []string{
		"first-secret",
		"second-secret",
		"third-secret",
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf(
			"findStringValuesAtPath() mismatch\n\ngot:  %#v\n\nwant: %#v",
			got,
			want,
		)
	}
}

func TestFindStringValuesAtPathNestedWildcards(t *testing.T) {
	t.Parallel()

	object := map[string]interface{}{
		"spec": map[string]interface{}{
			"groups": []interface{}{
				map[string]interface{}{
					"references": []interface{}{
						map[string]interface{}{
							"name": "first-secret",
						},
						map[string]interface{}{
							"name": "second-secret",
						},
					},
				},
				map[string]interface{}{
					"references": []interface{}{
						map[string]interface{}{
							"name": "third-secret",
						},
					},
				},
			},
		},
	}

	got, err := findStringValuesAtPath(
		object,
		"spec.groups[*].references[*].name",
	)
	if err != nil {
		t.Fatalf(
			"findStringValuesAtPath() error = %v",
			err,
		)
	}

	want := []string{
		"first-secret",
		"second-secret",
		"third-secret",
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf(
			"findStringValuesAtPath() mismatch\n\ngot:  %#v\n\nwant: %#v",
			got,
			want,
		)
	}
}

func TestFindStringValuesAtPathMissingField(t *testing.T) {
	t.Parallel()

	object := map[string]interface{}{
		"spec": map[string]interface{}{},
	}

	got, err := findStringValuesAtPath(
		object,
		"spec.target.name",
	)
	if err != nil {
		t.Fatalf(
			"findStringValuesAtPath() error = %v",
			err,
		)
	}

	if len(got) != 0 {
		t.Errorf(
			"expected no values, got %#v",
			got,
		)
	}
}

func TestFindStringValuesAtPathIgnoresNonStringValues(
	t *testing.T,
) {
	t.Parallel()

	object := map[string]interface{}{
		"spec": map[string]interface{}{
			"data": []interface{}{
				map[string]interface{}{
					"name": "valid-secret",
				},
				map[string]interface{}{
					"name": int64(42),
				},
				map[string]interface{}{
					"name": "",
				},
			},
		},
	}

	got, err := findStringValuesAtPath(
		object,
		"spec.data[*].name",
	)
	if err != nil {
		t.Fatalf(
			"findStringValuesAtPath() error = %v",
			err,
		)
	}

	want := []string{
		"valid-secret",
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf(
			"findStringValuesAtPath() mismatch\n\ngot:  %#v\n\nwant: %#v",
			got,
			want,
		)
	}
}

func TestFindStringValuesAtPathInvalidPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
	}{
		{
			name: "empty",
			path: "",
		},
		{
			name: "empty segment",
			path: "spec..name",
		},
		{
			name: "wildcard without field",
			path: "spec.[*].name",
		},
		{
			name: "numeric array index",
			path: "spec.data[0].name",
		},
		{
			name: "invalid wildcard syntax",
			path: "spec.data[].name",
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := findStringValuesAtPath(
				map[string]interface{}{},
				test.path,
			)
			if err == nil {
				t.Fatalf(
					"expected error for path %q",
					test.path,
				)
			}
		})
	}
}
