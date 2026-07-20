package ui

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"
)

type namespaceSelectorModel struct {
	allNamespaces      []string
	filteredNamespaces []string

	filter    []rune
	filtering bool

	cursor   int
	page     int
	pageSize int

	selected  string
	cancelled bool

	windowHeight int
}

func newNamespaceSelectorModel(
	namespaceNames []string,
) namespaceSelectorModel {
	names := append([]string(nil), namespaceNames...)
	sort.Strings(names)

	return namespaceSelectorModel{
		allNamespaces:      names,
		filteredNamespaces: names,
		pageSize:           defaultPageSize,
	}
}

func (m namespaceSelectorModel) Init() tea.Cmd {
	return nil
}

func (m namespaceSelectorModel) Update(
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
			if len(m.filteredNamespaces) == 0 {
				return m, nil
			}

			m.selected = m.filteredNamespaces[m.cursor]
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m namespaceSelectorModel) updateFiltering(
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
		if len(m.filteredNamespaces) == 0 {
			return m, nil
		}

		m.selected = m.filteredNamespaces[m.cursor]
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

func (m *namespaceSelectorModel) applyFilter() {
	query := strings.ToLower(
		strings.TrimSpace(string(m.filter)),
	)

	if query == "" {
		m.filteredNamespaces = m.allNamespaces
	} else {
		m.filteredNamespaces = make([]string, 0)

		for _, name := range m.allNamespaces {
			if strings.Contains(strings.ToLower(name), query) {
				m.filteredNamespaces = append(
					m.filteredNamespaces,
					name,
				)
			}
		}
	}

	m.cursor = 0
	m.page = 0
}

func (m *namespaceSelectorModel) moveUp() {
	if len(m.filteredNamespaces) == 0 {
		return
	}

	if m.cursor > 0 {
		m.cursor--
	}

	if !m.hasFilter() {
		m.page = m.cursor / m.pageSize
	}
}

func (m *namespaceSelectorModel) moveDown() {
	if len(m.filteredNamespaces) == 0 {
		return
	}

	if m.cursor < len(m.filteredNamespaces)-1 {
		m.cursor++
	}

	if !m.hasFilter() {
		m.page = m.cursor / m.pageSize
	}
}

func (m *namespaceSelectorModel) previousPage() {
	if m.hasFilter() || m.page == 0 {
		return
	}

	m.page--
	m.cursor = m.page * m.pageSize
}

func (m *namespaceSelectorModel) nextPage() {
	if m.hasFilter() || m.page >= m.totalPages()-1 {
		return
	}

	m.page++

	nextCursor := m.page * m.pageSize
	if nextCursor < len(m.filteredNamespaces) {
		m.cursor = nextCursor
	}
}

func (m namespaceSelectorModel) hasFilter() bool {
	return strings.TrimSpace(string(m.filter)) != ""
}

func (m namespaceSelectorModel) totalPages() int {
	if len(m.filteredNamespaces) == 0 {
		return 1
	}

	return (len(m.filteredNamespaces) + m.pageSize - 1) /
		m.pageSize
}

func (m namespaceSelectorModel) visibleNamespaces() (
	[]string,
	int,
) {
	if len(m.filteredNamespaces) == 0 {
		return nil, 0
	}

	if !m.hasFilter() {
		start := m.page * m.pageSize
		end := start + m.pageSize

		if end > len(m.filteredNamespaces) {
			end = len(m.filteredNamespaces)
		}

		return m.filteredNamespaces[start:end], start
	}

	maxVisible := m.windowHeight - 7
	if maxVisible < 5 {
		maxVisible = 5
	}
	if maxVisible > len(m.filteredNamespaces) {
		maxVisible = len(m.filteredNamespaces)
	}

	start := m.cursor - maxVisible/2
	if start < 0 {
		start = 0
	}

	maxStart := len(m.filteredNamespaces) - maxVisible
	if start > maxStart {
		start = maxStart
	}

	return m.filteredNamespaces[start : start+maxVisible], start
}

func (m namespaceSelectorModel) View() tea.View {
	var builder strings.Builder

	builder.WriteString(
		titleStyle.Render("Select a namespace"),
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

	visibleNamespaces, offset := m.visibleNamespaces()

	if len(visibleNamespaces) == 0 {
		builder.WriteString(
			emptyStyle.Render("  No matching namespaces"),
		)
		builder.WriteString("\n")
	} else {
		for index, name := range visibleNamespaces {
			absoluteIndex := offset + index
			selected := absoluteIndex == m.cursor

			builder.WriteString(
				renderNamespaceLine(name, selected),
			)
			builder.WriteString("\n")
		}
	}

	builder.WriteString("\n")
	builder.WriteString(footerStyle.Render(m.footer()))

	return tea.NewView(builder.String())
}

func (m namespaceSelectorModel) footer() string {
	if m.hasFilter() {
		if len(m.filteredNamespaces) == 0 {
			return "0 matching namespaces"
		}

		return fmt.Sprintf(
			"%d matching namespaces · %d/%d selected",
			len(m.filteredNamespaces),
			m.cursor+1,
			len(m.filteredNamespaces),
		)
	}

	return fmt.Sprintf(
		"Page %d/%d · %d namespaces",
		m.page+1,
		m.totalPages(),
		len(m.filteredNamespaces),
	)
}

func renderNamespaceLine(
	name string,
	selected bool,
) string {
	if selected {
		return selectedStyle.Render("> " + name)
	}

	return normalStyle.Render("  " + name)
}

func SelectNamespace(
	namespaceNames []string,
) (string, error) {
	if len(namespaceNames) == 0 {
		return "", fmt.Errorf("no namespaces found")
	}

	finalModel, err := tea.NewProgram(
		newNamespaceSelectorModel(namespaceNames),
	).Run()
	if err != nil {
		return "", fmt.Errorf(
			"run namespace selector: %w",
			err,
		)
	}

	result, ok := finalModel.(namespaceSelectorModel)
	if !ok {
		return "", fmt.Errorf(
			"unexpected namespace selector model",
		)
	}

	if result.cancelled {
		return "", fmt.Errorf("selection cancelled")
	}

	if result.selected == "" {
		return "", fmt.Errorf("no namespace selected")
	}

	return result.selected, nil
}
