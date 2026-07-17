package ui

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/lipgloss/v2"
	corev1 "k8s.io/api/core/v1"
)

var (
	secretMetadataLabelStyle = lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("#7D56F4"))

	secretKeyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4"))

	secretSeparatorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#555555"))
)

// RenderSecret renders every Secret key using the same simple format.
//
// Secret.Data values are already decoded by client-go.
// Values are rendered completely, without truncation or wrapping.
func RenderSecret(secret *corev1.Secret) string {
	var builder strings.Builder

	renderMetadataLine(&builder, "Secret", secret.Name)
	renderMetadataLine(&builder, "Namespace", secret.Namespace)
	renderMetadataLine(&builder, "Type", string(secret.Type))

	keys := make([]string, 0, len(secret.Data))

	for key := range secret.Data {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		value := strings.TrimSuffix(
			string(secret.Data[key]),
			"\n",
		)

		builder.WriteString("\n")
		builder.WriteString(secretKeyStyle.Render(key + ":"))
		builder.WriteString("\n")
		builder.WriteString(
			secretSeparatorStyle.Render(
				strings.Repeat("─", 60),
			),
		)
		builder.WriteString("\n")
		builder.WriteString(value)
		builder.WriteString("\n")
	}

	return strings.TrimRight(builder.String(), "\n")
}

func renderMetadataLine(
	builder *strings.Builder,
	label string,
	value string,
) {
	builder.WriteString(
		secretMetadataLabelStyle.Render(label + ":"),
	)
	builder.WriteString(" ")
	builder.WriteString(value)
	builder.WriteString("\n")
}

func RenderEmptySecretError(secretName string) error {
	return fmt.Errorf(
		"Secret %q contains no data",
		secretName,
	)
}
