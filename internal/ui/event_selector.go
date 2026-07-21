package ui

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	kube "github.com/pierinho13/kubectl-peek/internal/kubernetes"
)

var (
	eventNormalStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6A9955"))

	eventWarningStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#D7A700"))

	eventNonNormalStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#D75F5F"))
)

const (
	eventLastWidth        = 8
	eventTypeWidth        = 10
	eventReasonWidth      = 22
	eventOccurrencesWidth = 12
	eventMinObjectWidth   = 24
	eventMaxObjectWidth   = 58
)

type eventSelectorModel struct {
	namespace string
	grouped   bool

	allEvents      []kube.EventRecord
	filteredEvents []kube.EventRecord

	filter    []rune
	filtering bool

	cursor   int
	page     int
	pageSize int

	selected  *kube.EventRecord
	cancelled bool

	windowWidth  int
	windowHeight int
}

func newEventSelectorModel(
	namespace string,
	events []kube.EventRecord,
	grouped bool,
) eventSelectorModel {
	items := append([]kube.EventRecord(nil), events...)

	sort.SliceStable(items, func(i, j int) bool {
		return items[i].LastSeen.After(items[j].LastSeen)
	})

	return eventSelectorModel{
		namespace:      namespace,
		grouped:        grouped,
		allEvents:      items,
		filteredEvents: items,
		pageSize:       defaultPageSize,
	}
}

func (m eventSelectorModel) Init() tea.Cmd {
	return nil
}

func (m eventSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
		return m, nil

	case tea.KeyPressMsg:
		keyName := msg.String()

		if m.filtering {
			return m.updateFiltering(msg, keyName)
		}

		switch keyName {
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
			if len(m.filteredEvents) == 0 {
				return m, nil
			}

			selected := m.filteredEvents[m.cursor]
			m.selected = &selected
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m eventSelectorModel) updateFiltering(
	msg tea.KeyPressMsg,
	keyName string,
) (tea.Model, tea.Cmd) {
	switch keyName {
	case "ctrl+c":
		m.cancelled = true
		return m, tea.Quit

	case "esc":
		m.filtering = false
		return m, nil

	case "enter":
		if len(m.filteredEvents) == 0 {
			return m, nil
		}

		selected := m.filteredEvents[m.cursor]
		m.selected = &selected
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

	if msg.Text == "" {
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

func (m *eventSelectorModel) applyFilter() {
	query := strings.ToLower(strings.TrimSpace(string(m.filter)))

	if query == "" {
		m.filteredEvents = m.allEvents
	} else {
		m.filteredEvents = make([]kube.EventRecord, 0)

		for _, event := range m.allEvents {
			searchable := strings.ToLower(strings.Join(
				[]string{
					event.Namespace,
					event.Type,
					event.Reason,
					event.Kind,
					event.Name,
					event.Message,
					event.Source,
				},
				" ",
			))

			if strings.Contains(searchable, query) {
				m.filteredEvents = append(m.filteredEvents, event)
			}
		}
	}

	m.cursor = 0
	m.page = 0
}

func (m *eventSelectorModel) moveUp() {
	if len(m.filteredEvents) == 0 {
		return
	}

	if m.cursor > 0 {
		m.cursor--
	}

	if !m.hasFilter() {
		m.page = m.cursor / m.pageSize
	}
}

func (m *eventSelectorModel) moveDown() {
	if len(m.filteredEvents) == 0 {
		return
	}

	if m.cursor < len(m.filteredEvents)-1 {
		m.cursor++
	}

	if !m.hasFilter() {
		m.page = m.cursor / m.pageSize
	}
}

func (m *eventSelectorModel) previousPage() {
	if m.hasFilter() || m.page == 0 {
		return
	}

	m.page--
	m.cursor = m.page * m.pageSize
}

func (m *eventSelectorModel) nextPage() {
	if m.hasFilter() || m.page >= m.totalPages()-1 {
		return
	}

	m.page++
	m.cursor = m.page * m.pageSize
}

func (m eventSelectorModel) hasFilter() bool {
	return strings.TrimSpace(string(m.filter)) != ""
}

func (m eventSelectorModel) totalPages() int {
	if len(m.filteredEvents) == 0 {
		return 1
	}

	return (len(m.filteredEvents) + m.pageSize - 1) / m.pageSize
}

func (m eventSelectorModel) visibleEvents() ([]kube.EventRecord, int) {
	if len(m.filteredEvents) == 0 {
		return nil, 0
	}

	if !m.hasFilter() {
		start := m.page * m.pageSize
		end := start + m.pageSize
		if end > len(m.filteredEvents) {
			end = len(m.filteredEvents)
		}

		return m.filteredEvents[start:end], start
	}

	maxVisible := m.windowHeight - 8
	if maxVisible < 5 {
		maxVisible = 5
	}
	if maxVisible > len(m.filteredEvents) {
		maxVisible = len(m.filteredEvents)
	}

	start := m.cursor - maxVisible/2
	if start < 0 {
		start = 0
	}

	maxStart := len(m.filteredEvents) - maxVisible
	if start > maxStart {
		start = maxStart
	}

	return m.filteredEvents[start : start+maxVisible], start
}

func (m eventSelectorModel) View() tea.View {
	var builder strings.Builder

	title := "Select a Kubernetes event"
	if m.namespace != "" {
		title = fmt.Sprintf("Select an event from namespace %q", m.namespace)
	}

	builder.WriteString(titleStyle.Render(title))
	builder.WriteString("\n")

	mode := "raw"
	if m.grouped {
		mode = "grouped"
	}

	builder.WriteString(descriptionStyle.Render(
		fmt.Sprintf("%d %s events", len(m.filteredEvents), mode),
	))
	builder.WriteString("\n")

	if m.filtering || m.hasFilter() {
		builder.WriteString(filterStyle.Render("/" + string(m.filter)))
		builder.WriteString("\n")
	}

	builder.WriteString(descriptionStyle.Render(
		"Use ↑/↓ to move, ←/→ to change page, / to filter, and Enter for details.",
	))
	builder.WriteString("\n\n")

	objectWidth := eventObjectColumnWidth(m.windowWidth)

	builder.WriteString(descriptionStyle.Render(
		formatEventColumns(
			"LAST",
			"TYPE",
			"REASON",
			"OBJECT",
			"OCCURRENCES",
			objectWidth,
		),
	))
	builder.WriteString("\n")

	visible, offset := m.visibleEvents()
	if len(visible) == 0 {
		builder.WriteString(emptyStyle.Render("  No matching events"))
		builder.WriteString("\n")
	} else {
		for index, event := range visible {
			absoluteIndex := offset + index
			selected := absoluteIndex == m.cursor

			builder.WriteString(
				renderEventLine(event, selected, objectWidth),
			)
			builder.WriteString("\n")

			if selected && event.Message != "" {
				messageWidth := m.windowWidth - 4
				if messageWidth < 40 {
					messageWidth = 40
				}

				builder.WriteString(descriptionStyle.Render(
					"    " + compactEventMessage(
						event.Message,
						messageWidth,
					),
				))
				builder.WriteString("\n")
			}
		}
	}

	builder.WriteString("\n")
	builder.WriteString(footerStyle.Render(m.footer()))

	view := tea.NewView(builder.String())
	view.AltScreen = true

	return view
}

func (m eventSelectorModel) footer() string {
	if m.hasFilter() {
		if len(m.filteredEvents) == 0 {
			return "0 matching events"
		}

		return fmt.Sprintf(
			"%d matching events · %d/%d selected",
			len(m.filteredEvents),
			m.cursor+1,
			len(m.filteredEvents),
		)
	}

	return fmt.Sprintf(
		"Page %d/%d · %d events",
		m.page+1,
		m.totalPages(),
		len(m.filteredEvents),
	)
}

func renderEventLine(
	event kube.EventRecord,
	selected bool,
	objectWidth int,
) string {
	marker := "  "
	if selected {
		marker = "> "
	}

	lastSeen := formatEventAge(event.LastSeen)
	object := event.Kind + "/" + event.Name

	line := formatEventColumns(
		lastSeen,
		event.Type,
		event.Reason,
		object,
		formatEventCount(event.Count),
		objectWidth,
	)

	if selected {
		return selectedStyle.Render(marker + line)
	}

	switch event.Type {
	case "Normal":
		return marker + eventNormalStyle.Render(line)
	case "Warning":
		return marker + eventWarningStyle.Render(line)
	default:
		return marker + eventNonNormalStyle.Render(line)
	}
}

func formatEventCount(count int32) string {
	switch {
	case count >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(count)/1_000_000)
	case count >= 1_000:
		return fmt.Sprintf("%.1fk", float64(count)/1_000)
	default:
		return fmt.Sprintf("%d", count)
	}
}

func formatEventColumns(
	lastSeen string,
	eventType string,
	reason string,
	object string,
	occurrences string,
	objectWidth int,
) string {
	return fmt.Sprintf(
		"%-*s %-*s %-*s %-*s %*s",
		eventLastWidth,
		truncateEventColumn(lastSeen, eventLastWidth),
		eventTypeWidth,
		truncateEventColumn(eventType, eventTypeWidth),
		eventReasonWidth,
		truncateEventColumn(reason, eventReasonWidth),
		objectWidth,
		truncateEventColumn(object, objectWidth),
		eventOccurrencesWidth,
		truncateEventColumn(occurrences, eventOccurrencesWidth),
	)
}

func eventObjectColumnWidth(windowWidth int) int {
	const fixedWidth = 2 +
		eventLastWidth + 1 +
		eventTypeWidth + 1 +
		eventReasonWidth + 1 +
		1 + eventOccurrencesWidth

	width := windowWidth - fixedWidth

	if width < eventMinObjectWidth {
		return eventMinObjectWidth
	}

	if width > eventMaxObjectWidth {
		return eventMaxObjectWidth
	}

	return width
}

func truncateEventColumn(value string, width int) string {
	if width <= 0 {
		return ""
	}

	runes := []rune(value)
	if len(runes) <= width {
		return value
	}

	if width == 1 {
		return "…"
	}

	return string(runes[:width-1]) + "…"
}

func compactEventMessage(message string, limit int) string {
	message = strings.Join(strings.Fields(message), " ")
	if len(message) <= limit {
		return message
	}

	if limit < 2 {
		return message[:limit]
	}

	return message[:limit-1] + "…"
}

func SelectEvent(
	namespace string,
	events []kube.EventRecord,
	grouped bool,
) (kube.EventRecord, error) {
	if len(events) == 0 {
		return kube.EventRecord{}, fmt.Errorf("no Kubernetes events found")
	}

	finalModel, err := tea.NewProgram(
		newEventSelectorModel(namespace, events, grouped),
	).Run()
	if err != nil {
		return kube.EventRecord{}, fmt.Errorf("run event selector: %w", err)
	}

	result, ok := finalModel.(eventSelectorModel)
	if !ok {
		return kube.EventRecord{}, fmt.Errorf("unexpected event selector model")
	}

	if result.cancelled {
		return kube.EventRecord{}, fmt.Errorf("selection cancelled")
	}

	if result.selected == nil {
		return kube.EventRecord{}, fmt.Errorf("no event selected")
	}

	return *result.selected, nil
}
