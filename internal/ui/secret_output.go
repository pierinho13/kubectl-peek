package ui

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/lipgloss/v2"
	corev1 "k8s.io/api/core/v1"

	kube "github.com/pierinho13/kubectl-peek/internal/kubernetes"
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

func RenderSecret(
	secret *corev1.Secret,
	usages []kube.SecretUsage,
) string {
	var builder strings.Builder

	renderMetadataLine(&builder, "Secret", secret.Name)
	renderMetadataLine(&builder, "Namespace", secret.Namespace)
	renderMetadataLine(&builder, "Type", string(secret.Type))
	renderSecretUsages(&builder, usages)

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

func renderSecretUsages(
	builder *strings.Builder,
	usages []kube.SecretUsage,
) {
	builder.WriteString(secretMetadataLabelStyle.Render("Used by:"))
	builder.WriteString("\n")

	if len(usages) == 0 {
		builder.WriteString("  none")
		builder.WriteString("\n")
		return
	}

	for _, usage := range usages {
		builder.WriteString("  ")
		builder.WriteString(secretKeyStyle.Render(
			fmt.Sprintf("%s/%s", usage.Kind, usage.Name),
		))
		builder.WriteString("\n")

		for _, reference := range usage.References {
			builder.WriteString("    ")
			builder.WriteString(reference)
			builder.WriteString("\n")
		}
	}
}
