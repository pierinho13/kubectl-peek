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

	secretWarningStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#D7A700"))
)

func RenderSecret(
	secret *corev1.Secret,
	result kube.SecretUsageResult,
	showUsage bool,
	showValues bool,
) string {
	var builder strings.Builder

	renderMetadataLine(&builder, "Secret", secret.Name)
	renderMetadataLine(&builder, "Namespace", secret.Namespace)
	renderMetadataLine(&builder, "Type", string(secret.Type))

	if showUsage {
		renderSecretUsages(&builder, result.Usages)
		renderSecretWarnings(&builder, result.Warnings)
	}

	keys := make([]string, 0, len(secret.Data))

	for key := range secret.Data {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		value := renderSecretValue(
			secret.Data[key],
			showValues,
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

func renderSecretValue(
	value []byte,
	showValue bool,
) string {
	if !showValue {
		return fmt.Sprintf(
			"<redacted: %d bytes>",
			len(value),
		)
	}

	return strings.TrimSuffix(
		string(value),
		"\n",
	)
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
	builder.WriteString(
		secretMetadataLabelStyle.Render("Used by:"),
	)
	builder.WriteString("\n")

	if len(usages) == 0 {
		builder.WriteString(
			"  No references were found among the supported " +
				"built-in resources and configured usage rules.",
		)
		builder.WriteString("\n")
		builder.WriteString(
			descriptionStyle.Render(
				"  This does not guarantee that the Secret is unused; " +
					"unsupported resources, external systems, or " +
					"unconfigured custom resources may still reference it.",
			),
		)
		builder.WriteString("\n")

		return
	}

	for _, usage := range usages {
		builder.WriteString("  ")
		builder.WriteString(
			secretKeyStyle.Render(
				fmt.Sprintf(
					"%s/%s",
					usage.Kind,
					usage.Name,
				),
			),
		)
		builder.WriteString("\n")

		for _, reference := range usage.References {
			builder.WriteString("    ")
			builder.WriteString(renderSecretUsageReference(reference))
			builder.WriteString("\n")
		}
	}
}

func renderSecretWarnings(
	builder *strings.Builder,
	warnings []kube.SecretUsageWarning,
) {
	if len(warnings) == 0 {
		return
	}

	builder.WriteString(
		secretWarningStyle.Render("Warnings:"),
	)
	builder.WriteString("\n")

	for _, warning := range warnings {
		builder.WriteString("  ")

		if warning.Resource != "" {
			builder.WriteString(warning.Resource)
			builder.WriteString(": ")
		}

		builder.WriteString(warning.Err.Error())
		builder.WriteString("\n")
	}
}

func renderSecretUsageReference(
	reference kube.SecretUsageReference,
) string {
	var builder strings.Builder

	if reference.Relation != "" {
		builder.WriteString(reference.Relation)
		builder.WriteString(": ")
	}

	if reference.Description != "" {
		builder.WriteString(reference.Description)
	}

	detail := reference.Path

	if reference.Key != "" &&
		!strings.Contains(detail, "-> "+reference.Key) {
		if detail != "" {
			detail += " -> "
		}

		detail += reference.Key
	}

	if detail != "" {
		if reference.Description != "" {
			builder.WriteString(" (")
			builder.WriteString(detail)
			builder.WriteString(")")
		} else {
			builder.WriteString(detail)
		}
	}

	return builder.String()
}
