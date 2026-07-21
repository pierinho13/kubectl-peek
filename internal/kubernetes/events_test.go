package kubernetes

import (
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestEventRecordsGroupsRepeatedEvents(t *testing.T) {
	t.Parallel()

	first := time.Date(2026, 7, 21, 8, 0, 0, 0, time.UTC)
	second := first.Add(2 * time.Minute)

	events := []corev1.Event{
		{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:         "test",
				CreationTimestamp: metav1.NewTime(first),
			},
			Type:    "Warning",
			Reason:  "BackOff",
			Message: "Back-off restarting failed container",
			Count:   2,
			InvolvedObject: corev1.ObjectReference{
				Kind: "Pod",
				Name: "api-123",
			},
			LastTimestamp: metav1.NewTime(first),
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:         "test",
				CreationTimestamp: metav1.NewTime(second),
			},
			Type:    "Warning",
			Reason:  "BackOff",
			Message: "Back-off restarting failed container",
			Count:   3,
			InvolvedObject: corev1.ObjectReference{
				Kind: "Pod",
				Name: "api-123",
			},
			LastTimestamp: metav1.NewTime(second),
		},
	}

	records := EventRecords(events, true)

	if len(records) != 1 {
		t.Fatalf("expected one grouped event, got %#v", records)
	}

	if records[0].Count != 5 {
		t.Fatalf("expected grouped count 5, got %d", records[0].Count)
	}

	if !records[0].LastSeen.Equal(second) {
		t.Fatalf("expected last seen %v, got %v", second, records[0].LastSeen)
	}
}

func TestEventRecordsWithoutGroupingPreservesItems(t *testing.T) {
	t.Parallel()

	events := []corev1.Event{
		{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test",
			},
			Type:   "Normal",
			Reason: "Pulled",
			InvolvedObject: corev1.ObjectReference{
				Kind: "Pod",
				Name: "api-123",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test",
			},
			Type:   "Normal",
			Reason: "Pulled",
			InvolvedObject: corev1.ObjectReference{
				Kind: "Pod",
				Name: "api-123",
			},
		},
	}

	records := EventRecords(events, false)

	if len(records) != 2 {
		t.Fatalf("expected two ungrouped events, got %#v", records)
	}
}

func TestEventRecordsUsesSeriesLastObservedTime(t *testing.T) {
	t.Parallel()

	observed := time.Date(2026, 7, 21, 9, 0, 0, 0, time.UTC)

	event := corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			CreationTimestamp: metav1.NewTime(observed.Add(-time.Hour)),
		},
		Series: &corev1.EventSeries{
			Count:            7,
			LastObservedTime: metav1.NewMicroTime(observed),
		},
	}

	records := EventRecords([]corev1.Event{event}, true)

	if !records[0].LastSeen.Equal(observed) {
		t.Fatalf("expected series timestamp %v, got %v", observed, records[0].LastSeen)
	}
}
