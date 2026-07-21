package ui

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"
	corev1 "k8s.io/api/core/v1"
)

type PodOption struct {
	Name       string
	Phase      corev1.PodPhase
	Ready      int
	Containers int
}

type podSelectorModel struct {
	namespace string

	allPods      []PodOption
	filteredPods []PodOption

	filter    []rune
	filtering bool
	cursor    int
	page      int
	pageSize  int
	selected  string
	cancelled bool

	windowHeight int
}

func newPodSelectorModel(
	namespace string,
	pods []PodOption,
) podSelectorModel {
	options := append([]PodOption(nil), pods...)
	sort.Slice(options, func(i, j int) bool {
		return options[i].Name < options[j].Name
	})

	return podSelectorModel{
		namespace:    namespace,
		allPods:      options,
		filteredPods: options,
		pageSize:     defaultPageSize,
	}
}

func (m podSelectorModel) Init() tea.Cmd { return nil }

func (m podSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if len(m.filteredPods) == 0 {
				return m, nil
			}
			m.selected = m.filteredPods[m.cursor].Name
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m podSelectorModel) updateFiltering(
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
		if len(m.filteredPods) == 0 {
			return m, nil
		}
		m.selected = m.filteredPods[m.cursor].Name
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

func (m *podSelectorModel) applyFilter() {
	query := strings.ToLower(strings.TrimSpace(string(m.filter)))
	if query == "" {
		m.filteredPods = m.allPods
	} else {
		m.filteredPods = make([]PodOption, 0)
		for _, pod := range m.allPods {
			if strings.Contains(strings.ToLower(pod.Name), query) {
				m.filteredPods = append(m.filteredPods, pod)
			}
		}
	}
	m.cursor = 0
	m.page = 0
}

func (m *podSelectorModel) moveUp() {
	if m.cursor > 0 {
		m.cursor--
	}
	if !m.hasFilter() {
		m.page = m.cursor / m.pageSize
	}
}

func (m *podSelectorModel) moveDown() {
	if m.cursor < len(m.filteredPods)-1 {
		m.cursor++
	}
	if !m.hasFilter() {
		m.page = m.cursor / m.pageSize
	}
}

func (m *podSelectorModel) previousPage() {
	if m.hasFilter() || m.page == 0 {
		return
	}
	m.page--
	m.cursor = m.page * m.pageSize
}

func (m *podSelectorModel) nextPage() {
	if m.hasFilter() || m.page >= m.totalPages()-1 {
		return
	}
	m.page++
	m.cursor = m.page * m.pageSize
}

func (m podSelectorModel) hasFilter() bool {
	return strings.TrimSpace(string(m.filter)) != ""
}

func (m podSelectorModel) totalPages() int {
	if len(m.filteredPods) == 0 {
		return 1
	}
	return (len(m.filteredPods) + m.pageSize - 1) / m.pageSize
}

func (m podSelectorModel) visiblePods() ([]PodOption, int) {
	if len(m.filteredPods) == 0 {
		return nil, 0
	}

	if !m.hasFilter() {
		start := m.page * m.pageSize
		end := start + m.pageSize
		if end > len(m.filteredPods) {
			end = len(m.filteredPods)
		}
		return m.filteredPods[start:end], start
	}

	maxVisible := m.windowHeight - 7
	if maxVisible < 5 {
		maxVisible = 5
	}
	if maxVisible > len(m.filteredPods) {
		maxVisible = len(m.filteredPods)
	}

	start := m.cursor - maxVisible/2
	if start < 0 {
		start = 0
	}
	maxStart := len(m.filteredPods) - maxVisible
	if start > maxStart {
		start = maxStart
	}
	return m.filteredPods[start : start+maxVisible], start
}

func (m podSelectorModel) View() tea.View {
	var builder strings.Builder
	builder.WriteString(titleStyle.Render(
		fmt.Sprintf("Select a Pod from namespace %q", m.namespace),
	))
	builder.WriteString("\n")

	if m.filtering || m.hasFilter() {
		builder.WriteString(filterStyle.Render(
			fmt.Sprintf("/%s", string(m.filter)),
		))
		builder.WriteString("\n")
	}

	builder.WriteString(descriptionStyle.Render(
		"Use ↑/↓ to move, ←/→ to change page, / to filter, and Enter to exec.",
	))
	builder.WriteString("\n\n")

	visible, offset := m.visiblePods()
	if len(visible) == 0 {
		builder.WriteString(emptyStyle.Render("  No matching Pods"))
		builder.WriteString("\n")
	} else {
		for index, pod := range visible {
			prefix := "  "
			style := normalStyle
			if offset+index == m.cursor {
				prefix = "> "
				style = selectedStyle
			}
			line := fmt.Sprintf(
				"%-55s %d/%d  %s",
				pod.Name,
				pod.Ready,
				pod.Containers,
				pod.Phase,
			)
			builder.WriteString(style.Render(prefix + line))
			builder.WriteString("\n")
		}
	}

	builder.WriteString("\n")
	builder.WriteString(footerStyle.Render(m.footer()))
	return tea.NewView(builder.String())
}

func (m podSelectorModel) footer() string {
	if m.hasFilter() {
		if len(m.filteredPods) == 0 {
			return "0 matching Pods"
		}
		return fmt.Sprintf(
			"%d matching Pods · %d/%d selected",
			len(m.filteredPods),
			m.cursor+1,
			len(m.filteredPods),
		)
	}
	return fmt.Sprintf(
		"Page %d/%d · %d Pods",
		m.page+1,
		m.totalPages(),
		len(m.filteredPods),
	)
}

func SelectPod(
	namespace string,
	pods []PodOption,
) (string, error) {
	if len(pods) == 0 {
		return "", fmt.Errorf("no Pods found in namespace %q", namespace)
	}

	finalModel, err := tea.NewProgram(
		newPodSelectorModel(namespace, pods),
	).Run()
	if err != nil {
		return "", fmt.Errorf("run Pod selector: %w", err)
	}

	result, ok := finalModel.(podSelectorModel)
	if !ok {
		return "", fmt.Errorf("unexpected Pod selector model")
	}
	if result.cancelled {
		return "", fmt.Errorf("selection cancelled")
	}
	if result.selected == "" {
		return "", fmt.Errorf("no Pod selected")
	}
	return result.selected, nil
}
