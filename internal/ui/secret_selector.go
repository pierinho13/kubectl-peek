package ui

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const defaultPageSize = 10

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4"))

	descriptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#808080"))

	filterStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4"))

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4"))

	normalStyle = lipgloss.NewStyle()

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#808080"))

	emptyStyle = lipgloss.NewStyle().
			Italic(true).
			Foreground(lipgloss.Color("#808080"))
)

type secretSelectorModel struct {
	namespace string

	allSecrets      []string
	filteredSecrets []string

	filter    []rune
	filtering bool

	cursor   int
	page     int
	pageSize int

	selected  string
	cancelled bool

	windowWidth  int
	windowHeight int
}

func newSecretSelectorModel(
	namespace string,
	secretNames []string,
) secretSelectorModel {
	names := append([]string(nil), secretNames...)
	sort.Strings(names)

	return secretSelectorModel{
		namespace:       namespace,
		allSecrets:      names,
		filteredSecrets: names,
		pageSize:        defaultPageSize,
	}
}

func (m secretSelectorModel) Init() tea.Cmd {
	return nil
}

func (m secretSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if len(m.filteredSecrets) == 0 {
				return m, nil
			}

			m.selected = m.filteredSecrets[m.cursor]

			return m, tea.Quit
		}
	}

	return m, nil
}

func (m secretSelectorModel) updateFiltering(
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
		if len(m.filteredSecrets) == 0 {
			return m, nil
		}

		m.selected = m.filteredSecrets[m.cursor]

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

func (m *secretSelectorModel) applyFilter() {
	query := strings.ToLower(
		strings.TrimSpace(string(m.filter)),
	)

	if query == "" {
		m.filteredSecrets = m.allSecrets
	} else {
		m.filteredSecrets = make([]string, 0)

		for _, name := range m.allSecrets {
			if strings.Contains(strings.ToLower(name), query) {
				m.filteredSecrets = append(
					m.filteredSecrets,
					name,
				)
			}
		}
	}

	// Always move to the first result after changing the filter.
	m.cursor = 0
	m.page = 0
}

func (m *secretSelectorModel) moveUp() {
	if len(m.filteredSecrets) == 0 {
		return
	}

	if m.cursor > 0 {
		m.cursor--
	}

	if !m.hasFilter() {
		m.page = m.cursor / m.pageSize
	}
}

func (m *secretSelectorModel) moveDown() {
	if len(m.filteredSecrets) == 0 {
		return
	}

	if m.cursor < len(m.filteredSecrets)-1 {
		m.cursor++
	}

	if !m.hasFilter() {
		m.page = m.cursor / m.pageSize
	}
}

func (m *secretSelectorModel) previousPage() {
	if m.hasFilter() || m.page == 0 {
		return
	}

	m.page--
	m.cursor = m.page * m.pageSize
}

func (m *secretSelectorModel) nextPage() {
	if m.hasFilter() {
		return
	}

	if m.page >= m.totalPages()-1 {
		return
	}

	m.page++

	nextCursor := m.page * m.pageSize

	if nextCursor < len(m.filteredSecrets) {
		m.cursor = nextCursor
	}
}

func (m secretSelectorModel) hasFilter() bool {
	return strings.TrimSpace(string(m.filter)) != ""
}

func (m secretSelectorModel) totalPages() int {
	if len(m.filteredSecrets) == 0 {
		return 1
	}

	return (len(m.filteredSecrets) + m.pageSize - 1) / m.pageSize
}

func (m secretSelectorModel) visibleSecrets() ([]string, int) {
	if len(m.filteredSecrets) == 0 {
		return nil, 0
	}

	if !m.hasFilter() {
		start := m.page * m.pageSize

		if start >= len(m.filteredSecrets) {
			return nil, start
		}

		end := start + m.pageSize

		if end > len(m.filteredSecrets) {
			end = len(m.filteredSecrets)
		}

		return m.filteredSecrets[start:end], start
	}

	// Reserve space for title, filter, help, footer and blank lines.
	maxVisible := m.windowHeight - 7

	if maxVisible < 5 {
		maxVisible = 5
	}

	if maxVisible > len(m.filteredSecrets) {
		maxVisible = len(m.filteredSecrets)
	}

	start := m.cursor - maxVisible/2

	if start < 0 {
		start = 0
	}

	maxStart := len(m.filteredSecrets) - maxVisible

	if start > maxStart {
		start = maxStart
	}

	end := start + maxVisible

	return m.filteredSecrets[start:end], start
}

func (m secretSelectorModel) View() tea.View {
	var builder strings.Builder

	title := fmt.Sprintf(
		"Select a Secret from namespace %q",
		m.namespace,
	)

	builder.WriteString(titleStyle.Render(title))
	builder.WriteString("\n")

	if m.filtering || m.hasFilter() {
		filterText := fmt.Sprintf("/%s", string(m.filter))

		builder.WriteString(filterStyle.Render(filterText))
		builder.WriteString("\n")
	}

	builder.WriteString(
		descriptionStyle.Render(
			"Use ↑/↓ to move, ←/→ to change page, and / to filter.",
		),
	)
	builder.WriteString("\n\n")

	visibleSecrets, offset := m.visibleSecrets()

	if len(visibleSecrets) == 0 {
		builder.WriteString(
			emptyStyle.Render("  No matching Secrets"),
		)
		builder.WriteString("\n")
	} else {
		for index, name := range visibleSecrets {
			absoluteIndex := offset + index
			selected := absoluteIndex == m.cursor

			builder.WriteString(
				renderSecretLine(name, selected),
			)
			builder.WriteString("\n")
		}
	}

	builder.WriteString("\n")
	builder.WriteString(
		footerStyle.Render(m.footer()),
	)

	return tea.NewView(builder.String())
}

func (m secretSelectorModel) footer() string {
	if m.hasFilter() {
		if len(m.filteredSecrets) == 0 {
			return "0 matching Secrets"
		}

		return fmt.Sprintf(
			"%d matching Secrets · %d/%d selected",
			len(m.filteredSecrets),
			m.cursor+1,
			len(m.filteredSecrets),
		)
	}

	return fmt.Sprintf(
		"Page %d/%d · %d Secrets",
		m.page+1,
		m.totalPages(),
		len(m.filteredSecrets),
	)
}

func renderSecretLine(
	name string,
	selected bool,
) string {
	if selected {
		return selectedStyle.Render("> " + name)
	}

	return normalStyle.Render("  " + name)
}

func SelectSecret(
	namespace string,
	secretNames []string,
) (string, error) {
	if len(secretNames) == 0 {
		return "", fmt.Errorf(
			"no Secrets found in namespace %q",
			namespace,
		)
	}

	initialModel := newSecretSelectorModel(
		namespace,
		secretNames,
	)

	finalModel, err := tea.NewProgram(initialModel).Run()
	if err != nil {
		return "", fmt.Errorf(
			"run Secret selector: %w",
			err,
		)
	}

	result, ok := finalModel.(secretSelectorModel)
	if !ok {
		return "", fmt.Errorf(
			"unexpected Secret selector model",
		)
	}

	if result.cancelled {
		return "", fmt.Errorf("selection cancelled")
	}

	if result.selected == "" {
		return "", fmt.Errorf("no Secret selected")
	}

	return result.selected, nil
}
