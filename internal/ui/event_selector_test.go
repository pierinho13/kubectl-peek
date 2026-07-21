package ui

import (
	"strings"
	"testing"
	"time"

	kube "github.com/pierinho13/kubectl-peek/internal/kubernetes"
)

func TestEventSelectorFiltersAcrossEventFields(t *testing.T) {
	t.Parallel()

	model := newEventSelectorModel(
		"test",
		[]kube.EventRecord{
			{
				Type:    "Warning",
				Reason:  "BackOff",
				Kind:    "Pod",
				Name:    "api-123",
				Message: "Back-off restarting container",
			},
			{
				Type:   "Normal",
				Reason: "Pulled",
				Kind:   "Pod",
				Name:   "worker-123",
			},
		},
		true,
	)

	model.filter = []rune("backoff")
	model.applyFilter()

	if len(model.filteredEvents) != 1 {
		t.Fatalf("expected one matching event, got %#v", model.filteredEvents)
	}

	if model.filteredEvents[0].Name != "api-123" {
		t.Fatalf("expected api-123, got %q", model.filteredEvents[0].Name)
	}
}

func TestEventSelectorSortsNewestFirst(t *testing.T) {
	t.Parallel()

	now := time.Now()
	model := newEventSelectorModel(
		"test",
		[]kube.EventRecord{
			{Name: "older", LastSeen: now.Add(-time.Minute)},
			{Name: "newer", LastSeen: now},
		},
		true,
	)

	if model.allEvents[0].Name != "newer" {
		t.Fatalf("expected newest event first, got %q", model.allEvents[0].Name)
	}
}

func TestFormatEventCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		count int32
		want  string
	}{
		{count: 42, want: "42"},
		{count: 1_234, want: "1.2k"},
		{count: 111_671, want: "111.7k"},
		{count: 1_250_000, want: "1.2M"},
	}

	for _, test := range tests {
		if got := formatEventCount(test.count); got != test.want {
			t.Errorf("formatEventCount(%d) = %q, want %q", test.count, got, test.want)
		}
	}
}

func TestFormatEventColumnsKeepsOccurrencesAligned(t *testing.T) {
	t.Parallel()

	line := formatEventColumns(
		"1m",
		"Warning",
		"AmbiguousSelector",
		"HorizontalPodAutoscaler/iam-infinispan-with-a-very-long-name",
		"111.9k",
		36,
	)

	if !strings.HasSuffix(line, "      111.9k") {
		t.Fatalf("occurrences column is not right-aligned: %q", line)
	}

	if strings.Contains(
		line,
		"iam-infinispan-with-a-very-long-name",
	) {
		t.Fatalf("expected long object name to be truncated: %q", line)
	}
}

func TestEventObjectColumnWidthUsesTerminalWidth(t *testing.T) {
	t.Parallel()

	if got := eventObjectColumnWidth(80); got != eventMinObjectWidth {
		t.Fatalf(
			"eventObjectColumnWidth(80) = %d, want %d",
			got,
			eventMinObjectWidth,
		)
	}

	if got := eventObjectColumnWidth(200); got != eventMaxObjectWidth {
		t.Fatalf(
			"eventObjectColumnWidth(200) = %d, want %d",
			got,
			eventMaxObjectWidth,
		)
	}
}
