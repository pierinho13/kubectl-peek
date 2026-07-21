package ui

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"

	kube "github.com/pierinho13/kubectl-peek/internal/kubernetes"
)

type EventResourceSelection struct {
	Namespace string
	Name      string
}

type EventResourceOption struct {
	Namespace   string
	Name        string
	Events      int
	Occurrences int64
}

type eventResourceSelectorModel struct {
	kind          string
	allNamespaces bool

	all      []EventResourceOption
	filtered []EventResourceOption

	filter    []rune
	filtering bool

	cursor   int
	page     int
	pageSize int

	selected  *EventResourceSelection
	cancelled bool

	windowHeight int
}

func eventResourceOptions(
	kind string,
	events []kube.EventRecord,
) []EventResourceOption {
	optionsByKey := make(map[string]EventResourceOption)

	for _, event := range events {
		if event.Kind != kind {
			continue
		}

		key := event.Namespace + "\x00" + event.Name
		option := optionsByKey[key]

		option.Namespace = event.Namespace
		option.Name = event.Name
		option.Events++
		option.Occurrences += int64(event.Count)

		optionsByKey[key] = option
	}

	options := make(
		[]EventResourceOption,
		0,
		len(optionsByKey),
	)

	for _, option := range optionsByKey {
		options = append(options, option)
	}

	sort.Slice(options, func(i, j int) bool {
		if options[i].Occurrences != options[j].Occurrences {
			return options[i].Occurrences >
				options[j].Occurrences
		}

		if options[i].Namespace != options[j].Namespace {
			return options[i].Namespace <
				options[j].Namespace
		}

		return options[i].Name < options[j].Name
	})

	return options
}

func newEventResourceSelectorModel(
	kind string,
	events []kube.EventRecord,
	allNamespaces bool,
) eventResourceSelectorModel {
	options := eventResourceOptions(kind, events)

	return eventResourceSelectorModel{
		kind:          kind,
		allNamespaces: allNamespaces,
		all:           options,
		filtered:      options,
		pageSize:      defaultPageSize,
	}
}

func (m eventResourceSelectorModel) Init() tea.Cmd {
	return nil
}

func (m eventResourceSelectorModel) Update(
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
			return m.selectCurrent()
		}
	}

	return m, nil
}

func (m eventResourceSelectorModel) updateFiltering(
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
		return m.selectCurrent()

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

func (m eventResourceSelectorModel) selectCurrent() (
	tea.Model,
	tea.Cmd,
) {
	if len(m.filtered) == 0 {
		return m, nil
	}

	option := m.filtered[m.cursor]
	m.selected = &EventResourceSelection{
		Namespace: option.Namespace,
		Name:      option.Name,
	}

	return m, tea.Quit
}

func (m *eventResourceSelectorModel) applyFilter() {
	query := strings.ToLower(
		strings.TrimSpace(string(m.filter)),
	)

	m.filtered = nil

	for _, option := range m.all {
		searchable := strings.ToLower(
			option.Namespace + " " + option.Name,
		)

		if query == "" ||
			strings.Contains(searchable, query) {
			m.filtered = append(m.filtered, option)
		}
	}

	m.cursor = 0
	m.page = 0
}

func (m *eventResourceSelectorModel) moveUp() {
	if m.cursor > 0 {
		m.cursor--
	}

	if !m.hasFilter() {
		m.page = m.cursor / m.pageSize
	}
}

func (m *eventResourceSelectorModel) moveDown() {
	if m.cursor < len(m.filtered)-1 {
		m.cursor++
	}

	if !m.hasFilter() {
		m.page = m.cursor / m.pageSize
	}
}

func (m *eventResourceSelectorModel) previousPage() {
	if m.hasFilter() || m.page == 0 {
		return
	}

	m.page--
	m.cursor = m.page * m.pageSize
}

func (m *eventResourceSelectorModel) nextPage() {
	if m.hasFilter() || m.page >= m.totalPages()-1 {
		return
	}

	m.page++
	m.cursor = m.page * m.pageSize
}

func (m eventResourceSelectorModel) hasFilter() bool {
	return strings.TrimSpace(string(m.filter)) != ""
}

func (m eventResourceSelectorModel) totalPages() int {
	if len(m.filtered) == 0 {
		return 1
	}

	return (len(m.filtered) + m.pageSize - 1) /
		m.pageSize
}

func (m eventResourceSelectorModel) visible() (
	[]EventResourceOption,
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

func (m eventResourceSelectorModel) View() tea.View {
	var builder strings.Builder

	builder.WriteString(
		titleStyle.Render(
			fmt.Sprintf("Select a %s", m.kind),
		),
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
			"  %-58s %8s %12s",
			"RESOURCE",
			"EVENTS",
			"OCCURRENCES",
		),
	))
	builder.WriteString("\n")

	visible, offset := m.visible()

	if len(visible) == 0 {
		builder.WriteString(
			emptyStyle.Render("  No matching resources"),
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

			resource := option.Name
			if m.allNamespaces {
				resource = option.Namespace + "/" + option.Name
			}

			resource = truncateEventColumn(resource, 58)

			line := fmt.Sprintf(
				"%-58s %8d %12s",
				resource,
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

func (m eventResourceSelectorModel) footer() string {
	if m.hasFilter() {
		if len(m.filtered) == 0 {
			return "0 matching resources"
		}

		return fmt.Sprintf(
			"%d matching resources · %d/%d selected",
			len(m.filtered),
			m.cursor+1,
			len(m.filtered),
		)
	}

	return fmt.Sprintf(
		"Page %d/%d · %d resources",
		m.page+1,
		m.totalPages(),
		len(m.filtered),
	)
}

func SelectEventResource(
	kind string,
	events []kube.EventRecord,
	allNamespaces bool,
) (EventResourceSelection, error) {
	options := eventResourceOptions(kind, events)
	if len(options) == 0 {
		return EventResourceSelection{}, fmt.Errorf(
			"no resources found for kind %q",
			kind,
		)
	}

	finalModel, err := tea.NewProgram(
		newEventResourceSelectorModel(
			kind,
			events,
			allNamespaces,
		),
	).Run()
	if err != nil {
		return EventResourceSelection{}, fmt.Errorf(
			"run event resource selector: %w",
			err,
		)
	}

	result, ok := finalModel.(eventResourceSelectorModel)
	if !ok {
		return EventResourceSelection{}, fmt.Errorf(
			"unexpected event resource selector model",
		)
	}

	if result.cancelled {
		return EventResourceSelection{}, fmt.Errorf(
			"selection cancelled",
		)
	}

	if result.selected == nil {
		return EventResourceSelection{}, fmt.Errorf(
			"no resource selected",
		)
	}

	return *result.selected, nil
}
