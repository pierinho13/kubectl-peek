package cmd

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/pierinho13/kubectl-peek/internal/kubernetes"
	"github.com/pierinho13/kubectl-peek/internal/ui"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	eventsNamespace     string
	eventsAllNamespaces bool
	eventsWarningsOnly  bool
	eventsNonNormalOnly bool
	eventsNoGroup       bool
	eventsBrowseByKind  bool
)

var eventsCmd = &cobra.Command{
	Use:     "events [pattern]",
	Aliases: []string{"event", "ev"},
	Short:   "Browse Kubernetes events ordered by last occurrence",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var pattern string
		if len(args) == 1 {
			pattern = args[0]
		}

		return runEvents(
			cmd.Context(),
			cmd.OutOrStdout(),
			pattern,
		)
	},
}

func init() {
	eventsCmd.Flags().StringVarP(
		&eventsNamespace,
		"namespace",
		"n",
		"",
		"Kubernetes namespace",
	)

	eventsCmd.Flags().BoolVarP(
		&eventsAllNamespaces,
		"all-namespaces",
		"A",
		false,
		"Show events from all namespaces",
	)

	eventsCmd.Flags().BoolVar(
		&eventsWarningsOnly,
		"warnings",
		false,
		"Show only Warning events",
	)

	eventsCmd.Flags().BoolVar(
		&eventsNonNormalOnly,
		"non-normal",
		false,
		"Show events whose type is not Normal",
	)

	eventsCmd.Flags().BoolVar(
		&eventsNoGroup,
		"no-group",
		false,
		"Show individual events without grouping repeated occurrences",
	)

	eventsCmd.Flags().BoolVar(
		&eventsBrowseByKind,
		"browse-by-kind",
		false,
		"Browse events by resource kind and resource",
	)

	eventsCmd.Flags().BoolVar(
		&eventsBrowseByKind,
		"browse",
		false,
		"Browse events by resource kind and resource",
	)
}

func runEvents(
	ctx context.Context,
	out io.Writer,
	pattern string,
) error {
	if eventsAllNamespaces && eventsNamespace != "" {
		return fmt.Errorf("--all-namespaces cannot be used with --namespace")
	}

	if eventsWarningsOnly && eventsNonNormalOnly {
		return fmt.Errorf("--warnings cannot be used with --non-normal")
	}

	client, err := kubernetes.NewClient(
		kubeconfig,
		contextName,
		eventsNamespace,
	)
	if err != nil {
		return err
	}

	namespace := client.Namespace
	selectorNamespace := namespace

	if eventsAllNamespaces {
		namespace = metav1.NamespaceAll
		selectorNamespace = ""
	}

	eventList, err := client.Clientset.
		CoreV1().
		Events(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		if eventsAllNamespaces {
			return fmt.Errorf("list Kubernetes events across all namespaces: %w", err)
		}

		return fmt.Errorf(
			"list Kubernetes events in namespace %q: %w",
			namespace,
			err,
		)
	}

	records := kubernetes.EventRecords(
		eventList.Items,
		!eventsNoGroup,
	)

	records = filterEventRecords(records, pattern)

	if len(records) == 0 {
		return fmt.Errorf("no Kubernetes events matched the selected filters")
	}

	if eventsBrowseByKind {
		selectedKind, err := ui.SelectEventKind(records)
		if err != nil {
			return err
		}

		selectedResource, err := ui.SelectEventResource(
			selectedKind,
			records,
			eventsAllNamespaces,
		)
		if err != nil {
			return err
		}

		records = filterEventRecordsByResource(
			records,
			selectedKind,
			selectedResource,
		)
	}

	selected, err := ui.SelectEvent(
		selectorNamespace,
		records,
		!eventsNoGroup,
	)
	if err != nil {
		return err
	}

	fmt.Fprintln(out)
	fmt.Fprintln(out, ui.RenderEventDetail(selected))

	return nil
}

func filterEventRecords(
	records []kubernetes.EventRecord,
	pattern string,
) []kubernetes.EventRecord {
	filtered := make([]kubernetes.EventRecord, 0, len(records))
	normalizedPattern := strings.ToLower(strings.TrimSpace(pattern))

	for _, record := range records {
		if eventsWarningsOnly && record.Type != "Warning" {
			continue
		}

		if eventsNonNormalOnly && record.Type == "Normal" {
			continue
		}

		if normalizedPattern != "" {
			searchable := strings.ToLower(strings.Join(
				[]string{
					record.Namespace,
					record.Type,
					record.Reason,
					record.Kind,
					record.Name,
					record.Message,
					record.Source,
				},
				" ",
			))

			if !strings.Contains(searchable, normalizedPattern) {
				continue
			}
		}

		filtered = append(filtered, record)
	}

	return filtered
}

func filterEventRecordsByResource(
	records []kubernetes.EventRecord,
	kind string,
	resource ui.EventResourceSelection,
) []kubernetes.EventRecord {
	filtered := make(
		[]kubernetes.EventRecord,
		0,
		len(records),
	)

	for _, record := range records {
		if record.Kind != kind ||
			record.Name != resource.Name {
			continue
		}

		if resource.Namespace != "" &&
			record.Namespace != resource.Namespace {
			continue
		}

		filtered = append(filtered, record)
	}

	return filtered
}
