package ui

import (
	"fmt"
	"strings"
	"time"

	kube "github.com/pierinho13/kubectl-peek/internal/kubernetes"
)

func RenderEventDetail(event kube.EventRecord) string {
	var builder strings.Builder

	builder.WriteString(titleStyle.Render("Event details"))
	builder.WriteString("\n")
	builder.WriteString(secretSeparatorStyle.Render(strings.Repeat("─", 60)))
	builder.WriteString("\n\n")

	renderEventDetailLine(&builder, "Type", event.Type)
	renderEventDetailLine(&builder, "Reason", event.Reason)
	renderEventDetailLine(&builder, "Object", event.Kind+"/"+event.Name)

	if event.Namespace != "" {
		renderEventDetailLine(&builder, "Namespace", event.Namespace)
	}

	renderEventDetailLine(
		&builder,
		"Occurrences",
		fmt.Sprintf("%d", event.Count),
	)

	if event.EventObjects > 1 {
		renderEventDetailLine(
			&builder,
			"Grouped Event objects",
			fmt.Sprintf("%d", event.EventObjects),
		)
	}
	renderEventDetailLine(&builder, "First seen", formatEventTimestamp(event.FirstSeen))
	renderEventDetailLine(&builder, "Last seen", formatEventTimestamp(event.LastSeen))

	if event.Source != "" {
		renderEventDetailLine(&builder, "Source", event.Source)
	}

	builder.WriteString("\n")
	builder.WriteString(secretMetadataLabelStyle.Render("Message:"))
	builder.WriteString("\n")
	builder.WriteString(event.Message)

	return strings.TrimRight(builder.String(), "\n")
}

func renderEventDetailLine(
	builder *strings.Builder,
	label string,
	value string,
) {
	builder.WriteString(secretMetadataLabelStyle.Render(label + ":"))
	builder.WriteString(" ")
	builder.WriteString(value)
	builder.WriteString("\n")
}

func formatEventTimestamp(value time.Time) string {
	if value.IsZero() {
		return "unknown"
	}

	return fmt.Sprintf(
		"%s (%s ago)",
		value.Local().Format("2006-01-02 15:04:05 MST"),
		formatEventAge(value),
	)
}

func formatEventAge(value time.Time) string {
	if value.IsZero() {
		return "unknown"
	}

	duration := time.Since(value)
	if duration < 0 {
		duration = 0
	}

	switch {
	case duration < time.Minute:
		return fmt.Sprintf("%ds", int(duration.Seconds()))
	case duration < time.Hour:
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	case duration < 24*time.Hour:
		return fmt.Sprintf("%dh", int(duration.Hours()))
	default:
		return fmt.Sprintf("%dd", int(duration.Hours()/24))
	}
}
