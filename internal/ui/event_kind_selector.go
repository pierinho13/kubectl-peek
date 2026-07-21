package ui

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"

	kube "github.com/pierinho13/kubectl-peek/internal/kubernetes"
)

type EventKindOption struct {
	Kind        string
	Resources   int
	Events      int
	Occurrences int64
}

type eventKindSelectorModel struct {
	all      []EventKindOption
	filtered []EventKindOption

	filter    []rune
	filtering bool

	cursor   int
	page     int
	pageSize int

	selected  string
	cancelled bool

	windowHeight int
}

func eventKindOptions(
	events []kube.EventRecord,
) []EventKindOption {
	type aggregate struct {
		resources   map[string]struct{}
		events      int
		occurrences int64
	}

	byKind := make(map[string]*aggregate)

	for _, event := range events {
		kind := event.Kind
		if kind == "" {
			kind = "<unknown>"
		}

		current, found := byKind[kind]
		if !found {
			current = &aggregate{
				resources: make(map[string]struct{}),
			}
			byKind[kind] = current
		}

		resourceKey := event.Namespace + "\x00" + event.Name
		current.resources[resourceKey] = struct{}{}
		current.events++
		current.occurrences += int64(event.Count)
	}

	options := make([]EventKindOption, 0, len(byKind))

	for kind, aggregate := range byKind {
		options = append(
			options,
			EventKindOption{
				Kind:        kind,
				Resources:   len(aggregate.resources),
				Events:      aggregate.events,
				Occurrences: aggregate.occurrences,
			},
		)
	}

	sort.Slice(options, func(i, j int) bool {
		if options[i].Occurrences != options[j].Occurrences {
			return options[i].Occurrences > options[j].Occurrences
		}

		return options[i].Kind < options[j].Kind
	})

	return options
}

func newEventKindSelectorModel(
	events []kube.EventRecord,
) eventKindSelectorModel {
	options := eventKindOptions(events)

	return eventKindSelectorModel{
		all:      options,
		filtered: options,
		pageSize: defaultPageSize,
	}
}

func (m eventKindSelectorModel) Init() tea.Cmd {
	return nil
}

func (m eventKindSelectorModel) Update(
	msg tea.Msg,
) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowHeight = msg.Height
		return m, nil

	case tea.KeyPressMsg:
		key := msg.String()

		if m.filtering {
			return m.updateFiltering(msg, key)
		}

		switch key {
		case "ctrl+c", "esc", "q":
			m.cancelled = true
			return m, tea.Quit

		case "/":
			m.filtering = true
			return m, nil

		case "up", "k":
			m.moveUp()
			return m, nil

		case "down", "j":
			m.moveDown()
			return m, nil

		case "left":
			m.previousPage()
			return m, nil

		case "right":
			m.nextPage()
			return m, nil

		case "enter":
			if len(m.filtered) == 0 {
				return m, nil
			}

			m.selected = m.filtered[m.cursor].Kind
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m eventKindSelectorModel) updateFiltering(
	msg tea.KeyPressMsg,
	key string,
) (tea.Model, tea.Cmd) {
	switch key {
	case "ctrl+c":
		m.cancelled = true
		return m, tea.Quit

	case "esc":
		m.filtering = false
		return m, nil

	case "enter":
		if len(m.filtered) == 0 {
			return m, nil
		}

		m.selected = m.filtered[m.cursor].Kind
		return m, tea.Quit

	case "backspace":
		if len(m.filter) > 0 {
			m.filter = m.filter[:len(m.filter)-1]
			m.applyFilter()
		}
		return m, nil

	case "up":
		m.moveUp()
		return m, nil

	case "down":
		m.moveDown()
		return m, nil
	}

	for _, character := range msg.Text {
		if unicode.IsPrint(character) {
			m.filter = append(m.filter, character)
		}
	}

	m.applyFilter()
	return m, nil
}

func (m *eventKindSelectorModel) applyFilter() {
	query := strings.ToLower(
		strings.TrimSpace(string(m.filter)),
	)

	m.filtered = nil

	for _, option := range m.all {
		if query == "" ||
			strings.Contains(
				strings.ToLower(option.Kind),
				query,
			) {
			m.filtered = append(m.filtered, option)
		}
	}

	m.cursor = 0
	m.page = 0
}

func (m *eventKindSelectorModel) moveUp() {
	if m.cursor > 0 {
		m.cursor--
	}

	if !m.hasFilter() {
		m.page = m.cursor / m.pageSize
	}
}

func (m *eventKindSelectorModel) moveDown() {
	if m.cursor < len(m.filtered)-1 {
		m.cursor++
	}

	if !m.hasFilter() {
		m.page = m.cursor / m.pageSize
	}
}

func (m *eventKindSelectorModel) previousPage() {
	if m.hasFilter() || m.page == 0 {
		return
	}

	m.page--
	m.cursor = m.page * m.pageSize
}

func (m *eventKindSelectorModel) nextPage() {
	if m.hasFilter() || m.page >= m.totalPages()-1 {
		return
	}

	m.page++
	m.cursor = m.page * m.pageSize
}

func (m eventKindSelectorModel) hasFilter() bool {
	return strings.TrimSpace(string(m.filter)) != ""
}

func (m eventKindSelectorModel) totalPages() int {
	if len(m.filtered) == 0 {
		return 1
	}

	return (len(m.filtered) + m.pageSize - 1) /
		m.pageSize
}

func (m eventKindSelectorModel) visible() (
	[]EventKindOption,
	int,
) {
	if len(m.filtered) == 0 {
		return nil, 0
	}

	if !m.hasFilter() {
		start := m.page * m.pageSize
		end := start + m.pageSize

		if end > len(m.filtered) {
			end = len(m.filtered)
		}

		return m.filtered[start:end], start
	}

	maxVisible := m.windowHeight - 8
	if maxVisible < 5 {
		maxVisible = 5
	}
	if maxVisible > len(m.filtered) {
		maxVisible = len(m.filtered)
	}

	start := m.cursor - maxVisible/2
	if start < 0 {
		start = 0
	}

	maxStart := len(m.filtered) - maxVisible
	if start > maxStart {
		start = maxStart
	}

	return m.filtered[start : start+maxVisible], start
}

func (m eventKindSelectorModel) View() tea.View {
	var builder strings.Builder

	builder.WriteString(
		titleStyle.Render("Select a resource kind"),
	)
	builder.WriteString("\n")

	if m.filtering || m.hasFilter() {
		builder.WriteString(
			filterStyle.Render("/" + string(m.filter)),
		)
		builder.WriteString("\n")
	}

	builder.WriteString(descriptionStyle.Render(
		"Use ↑/↓ to move, ←/→ to change page, / to filter, and Enter to select.",
	))
	builder.WriteString("\n\n")

	builder.WriteString(descriptionStyle.Render(
		fmt.Sprintf(
			"  %-30s %10s %8s %12s",
			"KIND",
			"RESOURCES",
			"EVENTS",
			"OCCURRENCES",
		),
	))
	builder.WriteString("\n")

	visible, offset := m.visible()

	if len(visible) == 0 {
		builder.WriteString(
			emptyStyle.Render("  No matching resource kinds"),
		)
		builder.WriteString("\n")
	} else {
		for index, option := range visible {
			selected := offset+index == m.cursor
			marker := "  "
			style := normalStyle

			if selected {
				marker = "> "
				style = selectedStyle
			}

			line := fmt.Sprintf(
				"%-30s %10d %8d %12s",
				option.Kind,
				option.Resources,
				option.Events,
				formatEventCount(int32(option.Occurrences)),
			)

			builder.WriteString(style.Render(marker + line))
			builder.WriteString("\n")
		}
	}

	builder.WriteString("\n")
	builder.WriteString(footerStyle.Render(m.footer()))

	view := tea.NewView(builder.String())
	view.AltScreen = true

	return view
}

func (m eventKindSelectorModel) footer() string {
	if m.hasFilter() {
		if len(m.filtered) == 0 {
			return "0 matching resource kinds"
		}

		return fmt.Sprintf(
			"%d matching resource kinds · %d/%d selected",
			len(m.filtered),
			m.cursor+1,
			len(m.filtered),
		)
	}

	return fmt.Sprintf(
		"Page %d/%d · %d resource kinds",
		m.page+1,
		m.totalPages(),
		len(m.filtered),
	)
}

func SelectEventKind(
	events []kube.EventRecord,
) (string, error) {
	if len(events) == 0 {
		return "", fmt.Errorf(
			"no Kubernetes events found",
		)
	}

	finalModel, err := tea.NewProgram(
		newEventKindSelectorModel(events),
	).Run()
	if err != nil {
		return "", fmt.Errorf(
			"run event kind selector: %w",
			err,
		)
	}

	result, ok := finalModel.(eventKindSelectorModel)
	if !ok {
		return "", fmt.Errorf(
			"unexpected event kind selector model",
		)
	}

	if result.cancelled {
		return "", fmt.Errorf("selection cancelled")
	}

	if result.selected == "" {
		return "", fmt.Errorf(
			"no resource kind selected",
		)
	}

	return result.selected, nil
}
