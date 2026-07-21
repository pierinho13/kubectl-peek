package kubernetes

import (
	"sort"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
)

type EventRecord struct {
	Namespace string
	Type      string
	Reason    string
	Kind      string
	Name      string
	Message   string
	Source    string
	Count        int32
	EventObjects int
	FirstSeen    time.Time
	LastSeen  time.Time
}

func EventRecords(
	events []corev1.Event,
	group bool,
) []EventRecord {
	if !group {
		records := make([]EventRecord, 0, len(events))

		for i := range events {
			records = append(records, eventRecord(&events[i]))
		}

		sortEventRecords(records)
		return records
	}

	grouped := make(map[string]EventRecord)

	for i := range events {
		record := eventRecord(&events[i])
		key := eventGroupKey(record)

		existing, found := grouped[key]
		if !found {
			grouped[key] = record
			continue
		}

		existing.Count += record.Count
		existing.EventObjects += record.EventObjects

		if existing.FirstSeen.IsZero() ||
			(!record.FirstSeen.IsZero() && record.FirstSeen.Before(existing.FirstSeen)) {
			existing.FirstSeen = record.FirstSeen
		}

		if record.LastSeen.After(existing.LastSeen) {
			existing.LastSeen = record.LastSeen
		}

		if existing.Source == "" {
			existing.Source = record.Source
		}

		grouped[key] = existing
	}

	records := make([]EventRecord, 0, len(grouped))

	for _, record := range grouped {
		records = append(records, record)
	}

	sortEventRecords(records)
	return records
}

func eventRecord(event *corev1.Event) EventRecord {
	count := event.Count
	if event.Series != nil && event.Series.Count > count {
		count = event.Series.Count
	}
	if count < 1 {
		count = 1
	}

	firstSeen := event.FirstTimestamp.Time
	if firstSeen.IsZero() {
		firstSeen = event.CreationTimestamp.Time
	}

	lastSeen := eventLastSeen(event)
	if lastSeen.IsZero() {
		lastSeen = firstSeen
	}

	source := event.Source.Component
	if event.Source.Host != "" {
		if source != "" {
			source += "/"
		}
		source += event.Source.Host
	}

	return EventRecord{
		Namespace: event.Namespace,
		Type:      event.Type,
		Reason:    event.Reason,
		Kind:      event.InvolvedObject.Kind,
		Name:      event.InvolvedObject.Name,
		Message:   event.Message,
		Source:    source,
		Count:        count,
		EventObjects: 1,
		FirstSeen:    firstSeen,
		LastSeen:  lastSeen,
	}
}

func eventLastSeen(event *corev1.Event) time.Time {
	if event.Series != nil && !event.Series.LastObservedTime.IsZero() {
		return event.Series.LastObservedTime.Time
	}

	if !event.EventTime.IsZero() {
		return event.EventTime.Time
	}

	if !event.LastTimestamp.IsZero() {
		return event.LastTimestamp.Time
	}

	return event.CreationTimestamp.Time
}

func eventGroupKey(record EventRecord) string {
	return strings.Join(
		[]string{
			record.Namespace,
			record.Type,
			record.Reason,
			record.Kind,
			record.Name,
			record.Message,
		},
		"\x00",
	)
}

func sortEventRecords(records []EventRecord) {
	sort.SliceStable(records, func(i, j int) bool {
		if records[i].LastSeen.Equal(records[j].LastSeen) {
			if records[i].Type != records[j].Type {
				return records[i].Type < records[j].Type
			}

			if records[i].Kind != records[j].Kind {
				return records[i].Kind < records[j].Kind
			}

			return records[i].Name < records[j].Name
		}

		return records[i].LastSeen.After(records[j].LastSeen)
	})
}
