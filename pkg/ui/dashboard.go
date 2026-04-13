package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/shadinmd/anvil-cli/pkg/anvil"
	"github.com/shadinmd/anvil-cli/pkg/logger"
)

var (
	// Colors
	cyan  = lipgloss.Color("#89dceb") // Sky
	green = lipgloss.Color("#a6e3a1") // Green
	red   = lipgloss.Color("#f38ba8") // Red
	grey  = lipgloss.Color("#6c7086") // Overlay0
	white = lipgloss.Color("#cdd6f4") // Text

	// Styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(cyan).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(white).
			Align(lipgloss.Center)

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(white).
			Padding(0, 1)

	focusedBorderStyle = borderStyle.Copy().
				BorderForeground(cyan).
				BorderStyle(lipgloss.ThickBorder())

	sidebarItemStyle = lipgloss.NewStyle().
				PaddingLeft(1).
				Foreground(white)

	selectedSidebarItemStyle = sidebarItemStyle.Copy().
					Foreground(cyan).
					Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(grey)

	statusStyle = lipgloss.NewStyle().
			Inherit(helpStyle)

	errorStyle = lipgloss.NewStyle().
			Foreground(red).
			Bold(true)

	labelStyle = lipgloss.NewStyle().
			Foreground(cyan).
			Bold(true)
)

type DashboardModel struct {
	allProjects []anvil.Project
	projects    []anvil.Project
	cursor      int
	width       int
	height      int
	err         error
	form        *FormModel
	searchQuery string
	searching   bool
}

func NewDashboardModel() *DashboardModel {
	projects := anvil.GetAllProjects()
	return &DashboardModel{
		allProjects: projects,
		projects:    projects,
	}
}

func (m *DashboardModel) Init() tea.Cmd {
	return nil
}

func (m *DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.form != nil {
		newForm, cmd := m.form.Update(msg)
		m.form = newForm.(*FormModel)
		if m.form.Done {
			name, dir, windows := m.form.GetProjectData()
			if name != "" && dir != "" {
				var err error
				if m.form.isEdit {
					err = anvil.UpdateProject(m.form.targetId, name, dir, windows)
				} else {
					err = anvil.CreateProject(name, dir, windows)
				}

				if err != nil {
					m.err = err
					logger.Logf("Error processing project form: %v", err)
				} else {
					m.allProjects = anvil.GetAllProjects()
					m.filterProjects()
				}
			}
			m.form = nil
		} else if m.form.Cancel {
			m.form = nil
		}
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyPressMsg:
		k := msg.String()
		
		if k == "ctrl+c" {
			return m, tea.Quit
		}

		if m.searching {
			switch k {
			case "esc":
				m.searching = false
				return m, nil
			case "backspace":
				if len(m.searchQuery) > 0 {
					m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
					m.filterProjects()
				}
				return m, nil
			case "enter":
				m.searching = false
				if len(m.projects) > 0 {
					// Proceed to select first result
				} else {
					return m, nil
				}
			case "space":
				m.searchQuery += " "
				m.filterProjects()
				return m, nil
			default:
				if len(k) == 1 {
					m.searchQuery += k
					m.filterProjects()
					return m, nil
				}
			}
		}

		switch k {
		case "q":
			return m, tea.Quit

		case "/":
			m.searching = true
			m.cursor = 0
			return m, nil

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.projects)-1 {
				m.cursor++
			}

		case "n":
			m.form = NewNewProjectForm()

		case "e":
			if len(m.projects) > 0 {
				m.form = NewEditProjectForm(m.projects[m.cursor])
			}

		case "enter":
			if len(m.projects) > 0 {
				p := &m.projects[m.cursor]
				var err error
				if p.Active {
					err = anvil.SwitchProject(p)
				} else {
					err = anvil.StartProject(p)
				}
				if err != nil {
					m.err = err
					logger.Logf("Dashboard error: %v", err)
				} else {
					return m, tea.Quit
				}
			}

		case "s":
			if len(m.projects) > 0 {
				p := &m.projects[m.cursor]
				if p.Active {
					if err := anvil.StopProject(p); err != nil {
						m.err = err
						logger.Logf("Dashboard error stopping: %v", err)
					} else {
						m.allProjects = anvil.GetAllProjects()
						m.filterProjects()
					}
				}
			}

		case "r":
			m.allProjects = anvil.GetAllProjects()
			m.filterProjects()
		}
	}

	return m, nil
}

func (m *DashboardModel) filterProjects() {
	if m.searchQuery == "" {
		m.projects = m.allProjects
	} else {
		var filtered []anvil.Project
		query := strings.ToLower(m.searchQuery)
		for _, p := range m.allProjects {
			if strings.Contains(strings.ToLower(p.Name), query) || strings.Contains(strings.ToLower(p.Id), query) {
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

func (m *DashboardModel) View() tea.View {
	if m.width == 0 {
		return tea.NewView("Loading...")
	}

	if m.form != nil {
		// Pass size to form
		m.form.width = m.width
		m.form.height = m.height
		return m.form.View()
	}

	var s strings.Builder

	// Header
	header := titleStyle.Width(m.width - 2).Render("Anvil")
	s.WriteString(header + "\n")

	// Calculate layout
	sidebarWidth := 26
	contentWidth := m.width - sidebarWidth - 2 // Match header width
	mainHeight := m.height - 6

	// Search box
	searchStyle := borderStyle
	searchText := m.searchQuery
	if m.searching {
		searchStyle = focusedBorderStyle
		searchText += lipgloss.NewStyle().Foreground(cyan).Blink(true).Render("|")
	} else if searchText == "" {
		searchText = helpStyle.Render("Search...")
	}
	searchBox := searchStyle.Width(sidebarWidth).Render(" / " + searchText)

	// Sidebar List
	var sidebarLines []string
	for i, p := range m.projects {
		status := lipgloss.NewStyle().Foreground(grey).Render("○")
		if p.Active {
			status = lipgloss.NewStyle().Foreground(green).Render("●")
		}

		line := fmt.Sprintf("%s %s", status, p.Name)
		if i == m.cursor {
			sidebarLines = append(sidebarLines, selectedSidebarItemStyle.Render("> "+line))
		} else {
			sidebarLines = append(sidebarLines, sidebarItemStyle.Render("  "+line))
		}
	}

	listView := borderStyle.
		Width(sidebarWidth).
		Height(mainHeight - 3).
		Render(strings.Join(sidebarLines, "\n"))

	sidebarView := lipgloss.JoinVertical(lipgloss.Left, searchBox, listView)

	// Content Area
	var content string
	if len(m.projects) > 0 {
		p := m.projects[m.cursor]
		statusText := lipgloss.NewStyle().Foreground(grey).Render("Inactive")
		if p.Active {
			statusText = lipgloss.NewStyle().Foreground(green).Bold(true).Render("Active")
		}

		var cb strings.Builder
		cb.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Project:"), p.Name))
		cb.WriteString(fmt.Sprintf("%s      %s\n", labelStyle.Render("ID:"), p.Id))
		cb.WriteString(fmt.Sprintf("%s    %s\n", labelStyle.Render("Path:"), p.Dir))
		cb.WriteString(fmt.Sprintf("%s  %s\n\n", labelStyle.Render("Status:"), statusText))

		if len(p.Windows) > 0 {
			cb.WriteString(labelStyle.Render("Windows & Panes:") + "\n")
			for i, w := range p.Windows {
				isLastWindow := i == len(p.Windows)-1
				windowPrefix := "├──"
				if isLastWindow {
					windowPrefix = "└──"
				}

				layout := w.Layout
				if layout == "" {
					layout = "default"
				}
				
				windowLine := fmt.Sprintf(" %s %s [%s]", windowPrefix, w.Name, helpStyle.Render(layout))
				cb.WriteString(windowLine + "\n")

				for j, pane := range w.Panes {
					isLastPane := j == len(w.Panes)-1
					
					var panePrefix string
					if isLastWindow {
						panePrefix = "    "
					} else {
						panePrefix = "│   "
					}
					
					if isLastPane {
						panePrefix += "└──"
					} else {
						panePrefix += "├──"
					}

					cmd := pane.Cmd
					if cmd == "" {
						cmd = "default shell"
					}
					
					split := "vertical"
					if pane.Split == "h" {
						split = "horizontal"
					}
					
					paneLine := fmt.Sprintf(" %s %s (%s)", panePrefix, cmd, helpStyle.Render(split))
					cb.WriteString(paneLine + "\n")
				}
			}
			cb.WriteString("\n")
		}

		cb.WriteString(helpStyle.Render("Press [enter] to start or switch to this project."))
		content = cb.String()
	} else {
		content = "No projects found.\nPress [n] to add your first project."
	}

	contentView := borderStyle.
		Width(contentWidth).
		Height(mainHeight).
		Render(content)

	// Join Sidebar and Content
	mainView := lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, contentView)
	s.WriteString(mainView + "\n")

	// Footer / Help
	help := "[/] Search • [enter] Start/Switch • [s] Stop • [n] New • [e] Edit • [r] Refresh • [q] Quit"
	if m.err != nil {
		s.WriteString(errorStyle.Render(fmt.Sprintf("!! Error: %v", m.err)) + "\n")
	}
	s.WriteString(statusStyle.Render(help))

	v := tea.NewView(s.String())
	v.AltScreen = true
	return v
}

func RunDashboard() error {
	m := NewDashboardModel()
	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}
