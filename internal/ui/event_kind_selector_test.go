package ui

import (
	"testing"

	kube "github.com/pierinho13/kubectl-peek/internal/kubernetes"
)

func TestEventKindOptionsAggregatesResourcesAndOccurrences(
	t *testing.T,
) {
	t.Parallel()

	options := eventKindOptions(
		[]kube.EventRecord{
			{
				Namespace: "test",
				Kind:      "Pod",
				Name:      "api",
				Count:     10,
			},
			{
				Namespace: "test",
				Kind:      "Pod",
				Name:      "api",
				Count:     20,
			},
			{
				Namespace: "test",
				Kind:      "Pod",
				Name:      "worker",
				Count:     5,
			},
			{
				Namespace: "test",
				Kind:      "Deployment",
				Name:      "web",
				Count:     2,
			},
		},
	)

	if len(options) != 2 {
		t.Fatalf("expected two kinds, got %#v", options)
	}

	if options[0].Kind != "Pod" {
		t.Fatalf("expected Pod first, got %#v", options)
	}

	if options[0].Resources != 2 ||
		options[0].Events != 3 ||
		options[0].Occurrences != 35 {
		t.Fatalf(
			"unexpected Pod aggregate: %#v",
			options[0],
		)
	}
}
