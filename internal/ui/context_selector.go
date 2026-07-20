package ui

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"
)

type contextSelectorModel struct {
	allContexts      []string
	filteredContexts []string

	filter    []rune
	filtering bool

	cursor   int
	page     int
	pageSize int

	selected  string
	cancelled bool

	windowHeight int
}

func newContextSelectorModel(
	contextNames []string,
) contextSelectorModel {
	names := append([]string(nil), contextNames...)
	sort.Strings(names)

	return contextSelectorModel{
		allContexts:      names,
		filteredContexts: names,
		pageSize:         defaultPageSize,
	}
}

func (m contextSelectorModel) Init() tea.Cmd {
	return nil
}

func (m contextSelectorModel) Update(
	msg tea.Msg,
) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
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
			if len(m.filteredContexts) == 0 {
				return m, nil
			}

			m.selected = m.filteredContexts[m.cursor]
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m contextSelectorModel) updateFiltering(
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
		if len(m.filteredContexts) == 0 {
			return m, nil
		}

		m.selected = m.filteredContexts[m.cursor]
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

func (m *contextSelectorModel) applyFilter() {
	query := strings.ToLower(
		strings.TrimSpace(string(m.filter)),
	)

	if query == "" {
		m.filteredContexts = m.allContexts
	} else {
		m.filteredContexts = make([]string, 0)

		for _, name := range m.allContexts {
			if strings.Contains(strings.ToLower(name), query) {
				m.filteredContexts = append(
					m.filteredContexts,
					name,
				)
			}
		}
	}

	m.cursor = 0
	m.page = 0
}

func (m *contextSelectorModel) moveUp() {
	if len(m.filteredContexts) == 0 {
		return
	}

	if m.cursor > 0 {
		m.cursor--
	}

	if !m.hasFilter() {
		m.page = m.cursor / m.pageSize
	}
}

func (m *contextSelectorModel) moveDown() {
	if len(m.filteredContexts) == 0 {
		return
	}

	if m.cursor < len(m.filteredContexts)-1 {
		m.cursor++
	}

	if !m.hasFilter() {
		m.page = m.cursor / m.pageSize
	}
}

func (m *contextSelectorModel) previousPage() {
	if m.hasFilter() || m.page == 0 {
		return
	}

	m.page--
	m.cursor = m.page * m.pageSize
}

func (m *contextSelectorModel) nextPage() {
	if m.hasFilter() || m.page >= m.totalPages()-1 {
		return
	}

	m.page++

	nextCursor := m.page * m.pageSize
	if nextCursor < len(m.filteredContexts) {
		m.cursor = nextCursor
	}
}

func (m contextSelectorModel) hasFilter() bool {
	return strings.TrimSpace(string(m.filter)) != ""
}

func (m contextSelectorModel) totalPages() int {
	if len(m.filteredContexts) == 0 {
		return 1
	}

	return (len(m.filteredContexts) + m.pageSize - 1) /
		m.pageSize
}

func (m contextSelectorModel) visibleContexts() (
	[]string,
	int,
) {
	if len(m.filteredContexts) == 0 {
		return nil, 0
	}

	if !m.hasFilter() {
		start := m.page * m.pageSize
		end := start + m.pageSize

		if end > len(m.filteredContexts) {
			end = len(m.filteredContexts)
		}

		return m.filteredContexts[start:end], start
	}

	maxVisible := m.windowHeight - 7
	if maxVisible < 5 {
		maxVisible = 5
	}

	if maxVisible > len(m.filteredContexts) {
		maxVisible = len(m.filteredContexts)
	}

	start := m.cursor - maxVisible/2
	if start < 0 {
		start = 0
	}

	maxStart := len(m.filteredContexts) - maxVisible
	if start > maxStart {
		start = maxStart
	}

	return m.filteredContexts[start : start+maxVisible], start
}

func (m contextSelectorModel) View() tea.View {
	var builder strings.Builder

	builder.WriteString(
		titleStyle.Render("Select a Kubernetes context"),
	)
	builder.WriteString("\n")

	if m.filtering || m.hasFilter() {
		builder.WriteString(
			filterStyle.Render(
				fmt.Sprintf("/%s", string(m.filter)),
			),
		)
		builder.WriteString("\n")
	}

	builder.WriteString(
		descriptionStyle.Render(
			"Use ↑/↓ to move, ←/→ to change page, and / to filter.",
		),
	)
	builder.WriteString("\n\n")

	visibleContexts, offset := m.visibleContexts()

	if len(visibleContexts) == 0 {
		builder.WriteString(
			emptyStyle.Render("  No matching contexts"),
		)
		builder.WriteString("\n")
	} else {
		for index, name := range visibleContexts {
			absoluteIndex := offset + index
			selected := absoluteIndex == m.cursor

			builder.WriteString(
				renderContextLine(name, selected),
			)
			builder.WriteString("\n")
		}
	}

	builder.WriteString("\n")
	builder.WriteString(footerStyle.Render(m.footer()))

	return tea.NewView(builder.String())
}

func (m contextSelectorModel) footer() string {
	if m.hasFilter() {
		if len(m.filteredContexts) == 0 {
			return "0 matching contexts"
		}

		return fmt.Sprintf(
			"%d matching contexts · %d/%d selected",
			len(m.filteredContexts),
			m.cursor+1,
			len(m.filteredContexts),
		)
	}

	return fmt.Sprintf(
		"Page %d/%d · %d contexts",
		m.page+1,
		m.totalPages(),
		len(m.filteredContexts),
	)
}

func renderContextLine(
	name string,
	selected bool,
) string {
	if selected {
		return selectedStyle.Render("> " + name)
	}

	return normalStyle.Render("  " + name)
}

func SelectContext(
	contextNames []string,
) (string, error) {
	if len(contextNames) == 0 {
		return "", fmt.Errorf("no Kubernetes contexts found")
	}

	finalModel, err := tea.NewProgram(
		newContextSelectorModel(contextNames),
	).Run()
	if err != nil {
		return "", fmt.Errorf(
			"run context selector: %w",
			err,
		)
	}

	result, ok := finalModel.(contextSelectorModel)
	if !ok {
		return "", fmt.Errorf(
			"unexpected context selector model",
		)
	}

	if result.cancelled {
		return "", fmt.Errorf("selection cancelled")
	}

	if result.selected == "" {
		return "", fmt.Errorf("no context selected")
	}

	return result.selected, nil
}
