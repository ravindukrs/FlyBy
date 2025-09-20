package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MainViewModel represents the main menu
type MainViewModel struct {
	choices  []string
	selected int
}

// NewMainViewModel creates a new main view model
func NewMainViewModel() MainViewModel {
	return MainViewModel{
		choices: []string{
			"Manage Targets",
			"Exit",
		},
		selected: 0,
	}
}

// Update handles messages for the main view
func (m MainViewModel) Update(msg tea.KeyMsg) (MainViewModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}
	case "down", "j":
		if m.selected < len(m.choices)-1 {
			m.selected++
		}
	case "enter":
		return m, m.handleSelection()
	}
	
	return m, nil
}

// handleSelection handles menu selection
func (m MainViewModel) handleSelection() tea.Cmd {
	switch m.selected {
	case 0: // Manage Targets
		return func() tea.Msg {
			return SwitchViewMsg{View: ViewTargets}
		}
	case 1: // Exit
		return tea.Quit
	}
	return nil
}

// View renders the main view
func (m MainViewModel) View(width, height int) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		MarginBottom(2)
	
	itemStyle := lipgloss.NewStyle().
		PaddingLeft(2).
		MarginBottom(1)
		
	selectedStyle := itemStyle.Copy().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		PaddingLeft(1).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.Color("205"))
	
	var content strings.Builder
	content.WriteString(titleStyle.Render("Welcome to FlyBy"))
	content.WriteString("\n")
	content.WriteString("Select an option:\n\n")
	
	for i, choice := range m.choices {
		if i == m.selected {
			content.WriteString(selectedStyle.Render("> " + choice))
		} else {
			content.WriteString(itemStyle.Render("  " + choice))
		}
		content.WriteString("\n")
	}
	
	return content.String()
}