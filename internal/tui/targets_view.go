package tui

import (
	"fmt"
	"strings"

	"flyby/internal/config"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TargetsViewModel represents the targets management view
type TargetsViewModel struct {
	configManager *config.ConfigManager
	targets       []config.Target
	selected      int
	showingDetail bool
	scrollOffset  int
	maxVisible    int
}

// NewTargetsViewModel creates a new targets view model
func NewTargetsViewModel(configManager *config.ConfigManager) TargetsViewModel {
	vm := TargetsViewModel{
		configManager: configManager,
		selected:      0,
		showingDetail: false,
		scrollOffset:  0,
		maxVisible:    10, // Show max 10 items at once
	}
	vm.loadTargets()
	return vm
}

// loadTargets loads targets from configuration
func (m *TargetsViewModel) loadTargets() {
	m.targets = nil
	for name, target := range m.configManager.GetTargets() {
		// Ensure the target has its name set correctly
		target.Name = name
		m.targets = append(m.targets, target)
	}
}

// Update handles messages for the targets view
func (m TargetsViewModel) Update(msg tea.KeyMsg) (TargetsViewModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selected > 0 {
			m.selected--
			// Adjust scroll if needed
			if m.selected < m.scrollOffset {
				m.scrollOffset = m.selected
			}
		}
	case "down", "j":
		if m.selected < len(m.targets)-1 {
			m.selected++
			// Adjust scroll if needed
			if m.selected >= m.scrollOffset+m.maxVisible {
				m.scrollOffset = m.selected - m.maxVisible + 1
			}
		}
	case "enter":
		if len(m.targets) > 0 {
			return m, m.selectTarget()
		}
	case "a":
		return m, func() tea.Msg {
			return SwitchViewMsg{View: ViewAddTarget}
		}
	case "d":
		if len(m.targets) > 0 {
			return m, m.deleteTarget()
		}
	case "i":
		m.showingDetail = !m.showingDetail
	}
	
	return m, nil
}

// selectTarget selects a target and switches to pipelines view
func (m TargetsViewModel) selectTarget() tea.Cmd {
	if len(m.targets) == 0 {
		return nil
	}
	
	target := m.targets[m.selected]
	return func() tea.Msg {
		return SwitchViewMsg{View: ViewPipelines, Target: target.Name}
	}
}

// deleteTarget deletes the selected target
func (m TargetsViewModel) deleteTarget() tea.Cmd {
	if len(m.targets) == 0 {
		return nil
	}
	
	target := m.targets[m.selected]
	err := m.configManager.RemoveTarget(target.Name)
	if err == nil {
		m.loadTargets()
		// Adjust selected and scroll position
		if m.selected >= len(m.targets) && len(m.targets) > 0 {
			m.selected = len(m.targets) - 1
		}
		// Adjust scroll offset if needed
		if m.scrollOffset > 0 && m.selected < m.scrollOffset {
			m.scrollOffset = max(0, m.scrollOffset-1)
		}
	}
	
	return nil
}

// max returns the larger of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// View renders the targets view
func (m TargetsViewModel) View(width, height int) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		MarginBottom(1)
	
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
	content.WriteString(titleStyle.Render("Manage Targets"))
	content.WriteString("\n\n")
	
	if len(m.targets) == 0 {
		content.WriteString("No targets configured. Press 'a' to add a new target.\n")
		return content.String()
	}

	// Calculate visible range
	maxVisible := m.maxVisible
	if height-6 > 0 { // Account for title, header, footer, details
		maxVisible = min(height-6, len(m.targets))
	}
	
	// Adjust maxVisible if showing details
	if m.showingDetail {
		maxVisible = min(maxVisible-6, len(m.targets)) // Leave space for details
	}
	
	start := m.scrollOffset
	end := min(start+maxVisible, len(m.targets))
	
	// Add scroll indicator at top
	if start > 0 {
		content.WriteString(itemStyle.Render("  ↑ (more above)"))
		content.WriteString("\n")
	}
	
	// Show visible targets only
	for i := start; i < end; i++ {
		target := m.targets[i]
		var line string
		if m.showingDetail {
			line = fmt.Sprintf("%s (%s - %s)", target.Name, target.Team, target.GetURL())
		} else {
			line = fmt.Sprintf("%s (%s)", target.Name, target.Team)
		}
		
		if i == m.selected {
			content.WriteString(selectedStyle.Render("> " + line))
		} else {
			content.WriteString(itemStyle.Render("  " + line))
		}
		content.WriteString("\n")
	}
	
	// Add scroll indicator at bottom
	if end < len(m.targets) {
		content.WriteString(itemStyle.Render("  ↓ (more below)"))
		content.WriteString("\n")
	}
	
	// Show details if enabled
	if m.showingDetail && len(m.targets) > 0 {
		content.WriteString("\n")
		detailStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1).
			MarginTop(1)
		
		target := m.targets[m.selected]
		details := fmt.Sprintf("Target: %s\nTeam: %s\nAPI: %s\nToken: %s", 
			target.Name, target.Team, target.GetURL(), 
			func() string {
				if target.HasToken() {
					return "Present"
				}
				return "Not set"
			}())
		
		content.WriteString(detailStyle.Render(details))
	}
	
	return content.String()
}