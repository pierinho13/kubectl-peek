package ui

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"
)

type containerSelectorModel struct {
	podName      string
	all          []string
	filtered     []string
	filter       []rune
	filtering    bool
	cursor       int
	page         int
	pageSize     int
	selected     string
	cancelled    bool
	windowHeight int
}

func newContainerSelectorModel(podName string, containers []string) containerSelectorModel {
	items := append([]string(nil), containers...)
	sort.Strings(items)
	return containerSelectorModel{podName: podName, all: items, filtered: items, pageSize: defaultPageSize}
}

func (m containerSelectorModel) Init() tea.Cmd { return nil }

func (m containerSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowHeight = msg.Height
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
		case "up", "k":
			m.move(-1)
		case "down", "j":
			m.move(1)
		case "left":
			m.previousPage()
		case "right":
			m.nextPage()
		case "enter":
			if len(m.filtered) > 0 {
				m.selected = m.filtered[m.cursor]
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m containerSelectorModel) updateFiltering(msg tea.KeyPressMsg, key string) (tea.Model, tea.Cmd) {
	switch key {
	case "ctrl+c":
		m.cancelled = true
		return m, tea.Quit
	case "esc":
		m.filtering = false
		return m, nil
	case "enter":
		if len(m.filtered) > 0 {
			m.selected = m.filtered[m.cursor]
			return m, tea.Quit
		}
	case "backspace":
		if len(m.filter) > 0 {
			m.filter = m.filter[:len(m.filter)-1]
			m.applyFilter()
		}
		return m, nil
	case "up":
		m.move(-1)
		return m, nil
	case "down":
		m.move(1)
		return m, nil
	}
	for _, r := range msg.Text {
		if unicode.IsPrint(r) {
			m.filter = append(m.filter, r)
		}
	}
	m.applyFilter()
	return m, nil
}

func (m *containerSelectorModel) applyFilter() {
	q := strings.ToLower(strings.TrimSpace(string(m.filter)))
	m.filtered = nil
	for _, item := range m.all {
		if q == "" || strings.Contains(strings.ToLower(item), q) {
			m.filtered = append(m.filtered, item)
		}
	}
	m.cursor, m.page = 0, 0
}

func (m *containerSelectorModel) move(delta int) {
	next := m.cursor + delta
	if next >= 0 && next < len(m.filtered) {
		m.cursor = next
	}
	if !m.hasFilter() {
		m.page = m.cursor / m.pageSize
	}
}
func (m containerSelectorModel) hasFilter() bool { return strings.TrimSpace(string(m.filter)) != "" }
func (m containerSelectorModel) totalPages() int {
	if len(m.filtered) == 0 {
		return 1
	}
	return (len(m.filtered) + m.pageSize - 1) / m.pageSize
}
func (m *containerSelectorModel) previousPage() {
	if !m.hasFilter() && m.page > 0 {
		m.page--
		m.cursor = m.page * m.pageSize
	}
}
func (m *containerSelectorModel) nextPage() {
	if !m.hasFilter() && m.page < m.totalPages()-1 {
		m.page++
		m.cursor = m.page * m.pageSize
	}
}
func (m containerSelectorModel) visible() ([]string, int) {
	if len(m.filtered) == 0 {
		return nil, 0
	}
	if !m.hasFilter() {
		s := m.page * m.pageSize
		e := s + m.pageSize
		if e > len(m.filtered) {
			e = len(m.filtered)
		}
		return m.filtered[s:e], s
	}
	max := m.windowHeight - 7
	if max < 5 {
		max = 5
	}
	if max > len(m.filtered) {
		max = len(m.filtered)
	}
	s := m.cursor - max/2
	if s < 0 {
		s = 0
	}
	if s > len(m.filtered)-max {
		s = len(m.filtered) - max
	}
	return m.filtered[s : s+max], s
}

func (m containerSelectorModel) View() tea.View {
	var b strings.Builder
	b.WriteString(titleStyle.Render(fmt.Sprintf("Select a container from Pod %q", m.podName)))
	b.WriteString("\n")
	if m.filtering || m.hasFilter() {
		b.WriteString(filterStyle.Render("/" + string(m.filter)))
		b.WriteString("\n")
	}
	b.WriteString(descriptionStyle.Render("Use ↑/↓ to move, ←/→ to change page, and / to filter."))
	b.WriteString("\n\n")
	items, offset := m.visible()
	if len(items) == 0 {
		b.WriteString(emptyStyle.Render("  No matching containers"))
		b.WriteString("\n")
	} else {
		for i, item := range items {
			if offset+i == m.cursor {
				b.WriteString(selectedStyle.Render("> " + item))
			} else {
				b.WriteString(normalStyle.Render("  " + item))
			}
			b.WriteString("\n")
		}
	}
	b.WriteString("\n")
	b.WriteString(footerStyle.Render(fmt.Sprintf("Page %d/%d · %d containers", m.page+1, m.totalPages(), len(m.filtered))))
	return tea.NewView(b.String())
}

func SelectContainer(podName string, containers []string) (string, error) {
	if len(containers) == 0 {
		return "", fmt.Errorf("Pod %q has no containers", podName)
	}
	final, err := tea.NewProgram(newContainerSelectorModel(podName, containers)).Run()
	if err != nil {
		return "", fmt.Errorf("run container selector: %w", err)
	}
	result, ok := final.(containerSelectorModel)
	if !ok {
		return "", fmt.Errorf("unexpected container selector model")
	}
	if result.cancelled {
		return "", fmt.Errorf("selection cancelled")
	}
	if result.selected == "" {
		return "", fmt.Errorf("no container selected")
	}
	return result.selected, nil
}
