package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/shadinmd/anvil-cli/pkg/anvil"
)

type TemplatePickerModel struct {
	projects    []anvil.Project
	cursor      int
	Selected    *anvil.Project
	width       int
	Quitting    bool
	IsCancelled bool
}

func NewTemplatePickerModel(projects []anvil.Project) *TemplatePickerModel {
	return &TemplatePickerModel{
		projects: projects,
		width:    80, // Default fallback
	}
}

func (m *TemplatePickerModel) Init() tea.Cmd {
	return nil
}

func (m *TemplatePickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.projects) {
				m.cursor++
			}
		case "enter":
			if m.cursor > 0 {
				m.Selected = &m.projects[m.cursor-1]
			}
			m.Quitting = true
			return m, tea.Quit
		case "q", "esc", "ctrl+c":
			m.Quitting = true
			m.IsCancelled = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *TemplatePickerModel) View() tea.View {
	if m.Quitting {
		return tea.NewView("")
	}

	var s strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(Color.Primary).Underline(true)
	s.WriteString(titleStyle.Render("Select a Template:") + "\n\n")

	sidebarWidth := 30
	contentWidth := m.width - sidebarWidth - 4
	if contentWidth < 40 {
		contentWidth = 40
	}

	// Sidebar
	var sidebarLines []string
	
	// Blank Project Option
	blankStyle := lipgloss.NewStyle().Foreground(Color.Foreground)
	if m.cursor == 0 {
		blankStyle = lipgloss.NewStyle().Foreground(Color.Primary).Bold(true)
		sidebarLines = append(sidebarLines, blankStyle.Render("> [Blank Project]"))
	} else {
		sidebarLines = append(sidebarLines, blankStyle.Render("  [Blank Project]"))
	}

	for i, p := range m.projects {
		if i+1 == m.cursor {
			sidebarLines = append(sidebarLines, lipgloss.NewStyle().Foreground(Color.Primary).Bold(true).Render("> "+p.Name))
		} else {
			sidebarLines = append(sidebarLines, lipgloss.NewStyle().Foreground(Color.Foreground).Render("  "+p.Name))
		}
	}
	sidebarView := lipgloss.NewStyle().Width(sidebarWidth).Render(strings.Join(sidebarLines, "\n"))

	// Preview
	var previewView string
	if m.cursor == 0 {
		previewView = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Color.Secondary).
			Padding(1, 2).
			Width(contentWidth).
			Render("Start with a fresh workspace.\nContains one empty window.")
	} else if len(m.projects) > 0 {
		p := m.projects[m.cursor-1]
		var pb strings.Builder
		labelStyle := lipgloss.NewStyle().Foreground(Color.Primary).Bold(true)
		secondaryStyle := lipgloss.NewStyle().Foreground(Color.Secondary)

		pb.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("ID:"), p.Id))
		pb.WriteString(fmt.Sprintf("%s   %s\n\n", labelStyle.Render("Dir:"), p.Dir))

		if len(p.Windows) > 0 {
			pb.WriteString(labelStyle.Render("Windows & Panes:") + "\n")
			for i, w := range p.Windows {
				prefix := "├──"
				if i == len(p.Windows)-1 {
					prefix = "└──"
				}
				pb.WriteString(fmt.Sprintf(" %s %s\n", prefix, w.Name))
				for j, pane := range w.Panes {
					pPrefix := "│   ├──"
					if i == len(p.Windows)-1 {
						pPrefix = "    ├──"
					}
					if j == len(w.Panes)-1 {
						pPrefix = pPrefix[:len(pPrefix)-3] + "└──"
					}
					cmd := pane.Cmd
					if cmd == "" {
						cmd = "shell"
					}
					pb.WriteString(fmt.Sprintf(" %s %s\n", pPrefix, secondaryStyle.Render(cmd)))
				}
			}
		}
		previewView = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Color.Secondary).
			Padding(0, 1).
			Width(contentWidth).
			Render(pb.String())
	}

	mainView := lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, previewView)
	s.WriteString(mainView + "\n\n")
	s.WriteString(lipgloss.NewStyle().Foreground(Color.Secondary).Render("[↑/↓] Navigate • [enter] Select • [q] Cancel"))

	return tea.NewView(s.String())
}

func RunTemplatePicker(projects []anvil.Project) (*anvil.Project, bool, error) {
	m := NewTemplatePickerModel(projects)
	p := tea.NewProgram(m)
	_, err := p.Run()
	if err != nil {
		return nil, false, err
	}
	return m.Selected, m.IsCancelled, nil
}
