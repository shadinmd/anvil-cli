package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type InputModel struct {
	title       string
	message     string
	value       string
	placeholder string
	Done        bool
	Cancel      bool
	width       int
	height      int
}

func NewInputModel(title, message, placeholder string) *InputModel {
	return &InputModel{
		title:       title,
		message:     message,
		placeholder: placeholder,
	}
}

func (m *InputModel) Value() string {
	return strings.TrimSpace(m.value)
}

func (m *InputModel) Init() tea.Cmd {
	return nil
}

func (m *InputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyPressMsg:
		k := msg.String()
		switch k {
		case "esc", "ctrl+c":
			m.Cancel = true
			m.Done = true
			return m, nil

		case "enter":
			if strings.TrimSpace(m.value) != "" {
				m.Done = true
			}
			return m, nil

		case "backspace":
			if len(m.value) > 0 {
				m.value = m.value[:len(m.value)-1]
			}
			return m, nil

		case "space":
			m.value += "-"
			return m, nil

		default:
			if len(k) == 1 {
				m.value += k
			}
		}
	}
	return m, nil
}

func (m *InputModel) View() tea.View {
	var s strings.Builder

	modalWidth := 60
	if m.width < modalWidth+4 {
		modalWidth = m.width - 4
	}

	title := titleStyle.Width(modalWidth).Render(m.title)

	messageStyle := lipgloss.NewStyle().
		Width(modalWidth).
		Align(lipgloss.Left).
		Padding(1, 1)

	message := messageStyle.Render(m.message)

	inputBoxStyle := focusedBorderStyle.Width(modalWidth - 4)
	displayVal := m.value
	if displayVal == "" && m.placeholder != "" {
		displayVal = helpStyle.Render(m.placeholder)
	}
	cursor := lipgloss.NewStyle().Foreground(Color.Primary).Blink(true).Render("|")
	inputBox := inputBoxStyle.Render(displayVal + cursor)

	helpText := helpStyle.Render(" [enter] Confirm • [esc] Cancel")

	content := lipgloss.JoinVertical(lipgloss.Center, title, message, inputBox, "\n"+helpText)

	modal := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Color.Primary).
		Padding(1, 2).
		Render(content)

	vCenter := (m.height - lipgloss.Height(modal)) / 2
	if vCenter < 0 {
		vCenter = 0
	}

	hCenter := (m.width - lipgloss.Width(modal)) / 2
	if hCenter < 0 {
		hCenter = 0
	}

	s.WriteString(strings.Repeat("\n", vCenter))
	lines := strings.Split(modal, "\n")
	for _, line := range lines {
		s.WriteString(strings.Repeat(" ", hCenter) + line + "\n")
	}

	v := tea.NewView(s.String())
	v.AltScreen = true
	return v
}
