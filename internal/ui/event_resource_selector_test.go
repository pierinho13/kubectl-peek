package ui

import (
	"testing"

	kube "github.com/pierinho13/kubectl-peek/internal/kubernetes"
)

func TestEventResourceOptionsAggregatesByNamespaceAndName(
	t *testing.T,
) {
	t.Parallel()

	options := eventResourceOptions(
		"Pod",
		[]kube.EventRecord{
			{
				Namespace: "first",
				Kind:      "Pod",
				Name:      "api",
				Count:     10,
			},
			{
				Namespace: "first",
				Kind:      "Pod",
				Name:      "api",
				Count:     20,
			},
			{
				Namespace: "second",
				Kind:      "Pod",
				Name:      "api",
				Count:     5,
			},
			{
				Namespace: "first",
				Kind:      "Deployment",
				Name:      "api",
				Count:     100,
			},
		},
	)

	if len(options) != 2 {
		t.Fatalf(
			"expected two Pod resources, got %#v",
			options,
		)
	}

	if options[0].Namespace != "first" ||
		options[0].Name != "api" ||
		options[0].Events != 2 ||
		options[0].Occurrences != 30 {
		t.Fatalf(
			"unexpected first resource aggregate: %#v",
			options[0],
		)
	}
}
