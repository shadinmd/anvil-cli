package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/shadinmd/anvil-cli/pkg/anvil"
)

type ListModel struct {
	projects []anvil.Project
	cursor   int
	Choice   *anvil.Project
	width    int
	height   int
}

func NewListModel(projects []anvil.Project) *ListModel {
	return &ListModel{
		projects: projects,
	}
}

func (m *ListModel) Init() tea.Cmd {
	return nil
}

func (m *ListModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.projects)-1 {
				m.cursor++
			}

		case "enter":
			if len(m.projects) > 0 {
				m.Choice = &m.projects[m.cursor]
			}
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m *ListModel) View() tea.View {
	var s strings.Builder

	s.WriteString(titleStyle.Width(m.width - 2).Render("Select Project") + "\n")

	var items []string
	for i, p := range m.projects {
		status := lipgloss.NewStyle().Foreground(grey).Render("○")
		if p.Active {
			status = lipgloss.NewStyle().Foreground(green).Render("●")
		}

		line := fmt.Sprintf("%s %s", status, p.Name)
		if i == m.cursor {
			items = append(items, selectedSidebarItemStyle.Render("> "+line))
		} else {
			items = append(items, sidebarItemStyle.Render("  "+line))
		}
	}

	listView := borderStyle.
		Width(m.width - 2).
		Height(m.height - 6).
		Render(strings.Join(items, "\n"))

	s.WriteString(listView + "\n")
	s.WriteString(statusStyle.Render("[↑/↓] Navigate • [enter] Select • [q] Quit"))

	v := tea.NewView(s.String())
	v.AltScreen = true
	return v
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
