package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/shadinmd/anvil-cli/pkg/anvil"
)

type ListModel struct {
	allProjects []anvil.Project
	projects    []anvil.Project
	cursor      int
	Choice      *anvil.Project
	width       int
	searchQuery string
	quitting    bool
}

func NewListModel(projects []anvil.Project) *ListModel {
	return &ListModel{
		allProjects: projects,
		projects:    projects,
		width:       80,
	}
}

func (m *ListModel) Init() tea.Cmd {
	return nil
}

func (m *ListModel) filterProjects() {
	if m.searchQuery == "" {
		m.projects = m.allProjects
	} else {
		var filtered []anvil.Project
		for _, p := range m.allProjects {
			if fuzzyMatch(p.Name, m.searchQuery) || fuzzyMatch(p.Id, m.searchQuery) {
				filtered = append(filtered, p)
			}
		}
		m.projects = filtered
	}

	if m.cursor >= len(m.projects) {
		m.cursor = 0
		if len(m.projects) > 0 {
			m.cursor = len(m.projects) - 1
		}
	}
}

func (m *ListModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width

	case tea.KeyPressMsg:
		k := msg.String()
		switch k {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit

		case "up", "ctrl+p":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "ctrl+n":
			if m.cursor < len(m.projects)-1 {
				m.cursor++
			}

		case "enter":
			if len(m.projects) > 0 {
				m.Choice = &m.projects[m.cursor]
			}
			m.quitting = true
			return m, tea.Quit

		case "backspace":
			if len(m.searchQuery) > 0 {
				m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
				m.filterProjects()
			}

		case "space":
			m.searchQuery += " "
			m.filterProjects()

		default:
			if len(k) == 1 {
				m.searchQuery += k
				m.filterProjects()
			}
		}
	}

	return m, nil
}

func (m *ListModel) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}

	var s strings.Builder

	// Search prompt
	prompt := lipgloss.NewStyle().Foreground(Color.Primary).Bold(true).Render("?")
	label := lipgloss.NewStyle().Bold(true).Render(" Select project: ")
	searchText := m.searchQuery
	cursor := lipgloss.NewStyle().Foreground(Color.Primary).Blink(true).Render("|")

	s.WriteString(prompt + label + searchText + cursor + "\n")

	// Results
	maxResults := 10
	start := 0
	if m.cursor >= maxResults {
		start = m.cursor - maxResults + 1
	}

	for i := start; i < len(m.projects) && i < start+maxResults; i++ {
		p := m.projects[i]
		pointer := "  "
		style := lipgloss.NewStyle().Foreground(Color.Foreground)

		if i == m.cursor {
			pointer = lipgloss.NewStyle().Foreground(Color.Primary).Render("> ")
			style = lipgloss.NewStyle().Foreground(Color.Primary).Bold(true)
		}

		status := lipgloss.NewStyle().Foreground(Color.Secondary).Render("○")
		if p.Active {
			status = lipgloss.NewStyle().Foreground(Color.Success).Render("●")
		}

		s.WriteString(fmt.Sprintf("%s%s %s\n", pointer, status, style.Render(p.Name)))
	}

	if len(m.projects) == 0 {
		s.WriteString(helpStyle.Render("  No matches found") + "\n")
	}

	// Help line
	help := helpStyle.Render("  (type to filter, ↑/↓ to navigate, enter to select, esc to cancel)")
	s.WriteString(help)

	return tea.NewView(s.String())
}

func RunListUI(projects []anvil.Project) (*anvil.Project, error) {
	m := NewListModel(projects)
	p := tea.NewProgram(m)
	_, err := p.Run()
	if err != nil {
		return nil, err
	}
	return m.Choice, nil
}
