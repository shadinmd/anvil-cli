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
	// Styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Color.Primary).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Color.Foreground).
			Align(lipgloss.Center)

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Color.Foreground).
			Padding(0, 1)

	focusedBorderStyle = borderStyle.Copy().
				BorderForeground(Color.Primary).
				BorderStyle(lipgloss.ThickBorder())

	sidebarItemStyle = lipgloss.NewStyle().
				PaddingLeft(1).
				Foreground(Color.Foreground)

	selectedSidebarItemStyle = sidebarItemStyle.Copy().
					Foreground(Color.Primary).
					Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(Color.Secondary)

	statusStyle = lipgloss.NewStyle().
			Inherit(helpStyle)

	errorStyle = lipgloss.NewStyle().
			Foreground(Color.Error).
			Bold(true)

	labelStyle = lipgloss.NewStyle().
			Foreground(Color.Primary).
			Bold(true)
)

type DashboardModel struct {
	allProjects  []anvil.Project
	projects     []anvil.Project
	cursor       int
	offset       int
	width        int
	height       int
	err          error
	form         *FormModel
	confirm      *ConfirmModel
	searchQuery  string
	showArchived bool
}

func NewDashboardModel() *DashboardModel {
	projects := anvil.GetAllProjects()
	m := &DashboardModel{
		allProjects: projects,
	}
	m.filterProjects()
	return m
}

func (m *DashboardModel) Init() tea.Cmd {
	return nil
}

func (m *DashboardModel) scrollToCursor() {
	headerHeight := 2
	footerHeight := 3
	if m.err != nil {
		footerHeight = 4
	}
	mainHeight := m.height - headerHeight - footerHeight
	if mainHeight < 5 {
		mainHeight = 5
	}
	pageSize := mainHeight - 5
	if pageSize < 1 {
		pageSize = 1
	}

	if m.cursor < m.offset {
		m.offset = m.cursor
	} else if m.cursor >= m.offset+pageSize {
		m.offset = m.cursor - pageSize + 1
	}
}

func (m *DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.confirm != nil {
		newConfirm, cmd := m.confirm.Update(msg)
		m.confirm = newConfirm.(*ConfirmModel)
		if m.confirm.Done {
			if m.confirm.Confirm {
				if len(m.projects) > 0 {
					p := m.projects[m.cursor]
					if err := anvil.DeleteProject(p.Id); err != nil {
						m.err = err
						logger.Logf("Error deleting project: %v", err)
					} else {
						m.allProjects = anvil.GetAllProjects()
						m.filterProjects()
					}
				}
			}
			m.confirm = nil
		}
		return m, cmd
	}

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
		m.scrollToCursor()

	case tea.KeyPressMsg:
		k := msg.String()

		switch k {
		case "ctrl+c", "ctrl+q", "esc":
			return m, tea.Quit

		case "up", "ctrl+p":
			if m.cursor > 0 {
				m.cursor--
				m.scrollToCursor()
			}

		case "down", "ctrl+n":
			if m.cursor < len(m.projects)-1 {
				m.cursor++
				m.scrollToCursor()
			}

		case "ctrl+a":
			m.form = NewNewProjectForm()

		case "ctrl+t":
			if len(m.projects) > 0 {
				m.form = NewCopyProjectForm(m.projects[m.cursor])
			}

		case "ctrl+e":
			if len(m.projects) > 0 {
				m.form = NewEditProjectForm(m.projects[m.cursor])
			}

		case "ctrl+d":
			if len(m.projects) > 0 {
				p := m.projects[m.cursor]
				m.confirm = NewConfirmModel("Delete Project", fmt.Sprintf("Are you sure you want to delete '%s'?", p.Name))
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

		case "ctrl+s":
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

		case "ctrl+x":
			if len(m.projects) > 0 {
				p := m.projects[m.cursor]
				var err error
				if p.Archived {
					err = anvil.UnarchiveProject(p.Id)
				} else {
					if p.Active {
						_ = anvil.StopProject(&p)
					}
					err = anvil.ArchiveProject(p.Id)
				}
				if err != nil {
					m.err = err
					logger.Logf("Dashboard error toggling archive: %v", err)
				} else {
					m.allProjects = anvil.GetAllProjects()
					m.filterProjects()
				}
			}

		case "ctrl+g":
			m.showArchived = !m.showArchived
			m.cursor = 0
			m.filterProjects()

		case "ctrl+r":
			m.allProjects = anvil.GetAllProjects()
			m.filterProjects()

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

func fuzzyMatch(s, query string) bool {
	s = strings.ToLower(s)
	query = strings.ToLower(query)
	if len(query) == 0 {
		return true
	}
	if len(query) > len(s) {
		return false
	}

	qi := 0
	for i := 0; i < len(s); i++ {
		if s[i] == query[qi] {
			qi++
		}
		if qi == len(query) {
			return true
		}
	}
	return false
}

func (m *DashboardModel) filterProjects() {
	var base []anvil.Project
	for _, p := range m.allProjects {
		if p.Archived == m.showArchived {
			base = append(base, p)
		}
	}

	if m.searchQuery == "" {
		m.projects = base
	} else {
		var filtered []anvil.Project
		for _, p := range base {
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
	m.scrollToCursor()
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

	if m.confirm != nil {
		m.confirm.width = m.width
		m.confirm.height = m.height
		return m.confirm.View()
	}

	var s strings.Builder

	// Calculate layout
	headerHeight := 2
	footerHeight := 3
	if m.err != nil {
		footerHeight = 4
	}
	sidebarWidth := 25
	contentWidth := m.width - sidebarWidth - 1 // Match header width
	mainHeight := m.height - headerHeight - footerHeight
	if mainHeight < 5 {
		mainHeight = 5
	}

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(Color.Primary).
		Align(lipgloss.Center).
		Width(m.width)

	var modeName string
	if m.showArchived {
		modeName = "ARCHIVED PROJECTS"
	} else {
		modeName = "ACTIVE PROJECTS"
	}
	s.WriteString(headerStyle.Render("ANVIL - "+modeName) + "\n\n")

	// Search box
	searchBoxStyle := focusedBorderStyle
	searchText := m.searchQuery
	if searchText == "" {
		if m.showArchived {
			searchText = helpStyle.Render("Search archived...")
		} else {
			searchText = helpStyle.Render("Type to search...")
		}
	} else {
		searchText += lipgloss.NewStyle().Foreground(Color.Primary).Blink(true).Render("|")
	}
	searchBox := searchBoxStyle.Width(sidebarWidth).Render(searchText)

	// Sidebar List
	pageSize := mainHeight - 5
	if pageSize < 1 {
		pageSize = 1
	}

	var sidebarLines []string
	for i := m.offset; i < len(m.projects) && i < m.offset+pageSize; i++ {
		p := m.projects[i]
		status := lipgloss.NewStyle().Foreground(Color.Secondary).Render("○")
		if p.Active {
			status = lipgloss.NewStyle().Foreground(Color.Success).Render("●")
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
		statusText := lipgloss.NewStyle().Foreground(Color.Secondary).Render("Inactive")
		if p.Active {
			statusText = lipgloss.NewStyle().Foreground(Color.Success).Bold(true).Render("Active")
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

					paneLine := fmt.Sprintf(" %s %s", panePrefix, cmd)
					cb.WriteString(paneLine + "\n")
				}
			}
			cb.WriteString("\n")
		}

		cb.WriteString(helpStyle.Render("Press [enter] to start or switch to this project."))
		content = cb.String()
	} else {
		if m.showArchived {
			content = "No archived projects found."
		} else {
			content = "No active projects found.\nPress [ctrl+a] to add your first project."
		}
	}

	contentView := borderStyle.
		Width(contentWidth).
		Height(mainHeight).
		Render(content)

	// Join Sidebar and Content
	mainView := lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, contentView)
	s.WriteString(mainView + "\n")

	// Footer / Help
	helpLine1 := "[ctrl+n/p] Nav • [enter] Start/Switch • [ctrl+s] Stop • [ctrl+x] "
	if m.showArchived {
		helpLine1 += "Restore"
	} else {
		helpLine1 += "Archive"
	}
	helpLine1 += " • [ctrl+g] "
	if m.showArchived {
		helpLine1 += "Show Active"
	} else {
		helpLine1 += "Show Archived"
	}

	helpLine2 := "[ctrl+a] New • [ctrl+t] Copy • [ctrl+e] Edit • [ctrl+d] Delete • [ctrl+r] Refresh • [ctrl+q] Quit"

	if m.err != nil {
		s.WriteString(errorStyle.Render(fmt.Sprintf("!! Error: %v", m.err)) + "\n")
	}
	s.WriteString(statusStyle.Render(helpLine1) + "\n")
	s.WriteString(statusStyle.Render(helpLine2))

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
