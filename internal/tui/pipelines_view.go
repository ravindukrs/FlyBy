package tui

import (
	"fmt"
	"strings"

	"flyby/internal/concourse"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type pipelinesState int

const (
	pipelinesStateLoading pipelinesState = iota
	pipelinesStateList
)

// PipelinesViewModel represents the pipelines view
type PipelinesViewModel struct {
	client       *concourse.Client
	pipelines    []concourse.Pipeline
	selected     int
	state        pipelinesState
	err          error
	scrollOffset int
	maxVisible   int
}

// NewPipelinesViewModel creates a new pipelines view model
func NewPipelinesViewModel() PipelinesViewModel {
	return PipelinesViewModel{
		selected:     0,
		state:        pipelinesStateList,
		scrollOffset: 0,
		maxVisible:   10,
	}
}

// PipelinesLoadedMsg represents loaded pipelines
type PipelinesLoadedMsg struct {
	Pipelines []concourse.Pipeline
	Error     error
}

// LoadPipelines loads pipelines from Concourse
func (m *PipelinesViewModel) LoadPipelines(client *concourse.Client) tea.Cmd {
	m.client = client
	m.state = pipelinesStateLoading
	return func() tea.Msg {
		pipelines, err := client.GetPipelines()
		return PipelinesLoadedMsg{Pipelines: pipelines, Error: err}
	}
}

// Update handles messages for the pipelines view
func (m PipelinesViewModel) Update(msg tea.KeyMsg) (PipelinesViewModel, tea.Cmd) {
	switch msg.String() {
	case "f5":
		// Refresh pipelines
		if m.client != nil {
			m.state = pipelinesStateLoading
			return m, m.LoadPipelines(m.client)
		}
	case "up", "k":
		if m.selected > 0 {
			m.selected--
			// Adjust scroll if needed
			if m.selected < m.scrollOffset {
				m.scrollOffset = m.selected
			}
		}
	case "down":
		if m.selected < len(m.pipelines)-1 {
			m.selected++
			// Adjust scroll if needed
			if m.selected >= m.scrollOffset+m.maxVisible {
				m.scrollOffset = m.selected - m.maxVisible + 1
			}
		}
	case "j":
		if len(m.pipelines) > 0 {
			return m, func() tea.Msg {
				return SwitchViewMsg{View: ViewJobs}
			}
		}
	case "r":
		if len(m.pipelines) > 0 {
			return m, func() tea.Msg {
				return SwitchViewMsg{View: ViewResources}
			}
		}
	case "p":
		if len(m.pipelines) > 0 {
			return m, m.togglePipeline()
		}
	case "enter":
		if len(m.pipelines) > 0 {
			return m, func() tea.Msg {
				return SwitchViewMsg{View: ViewJobs}
			}
		}
	}
	
	return m, nil
}

// togglePipeline pauses or unpauses the selected pipeline
func (m PipelinesViewModel) togglePipeline() tea.Cmd {
	if len(m.pipelines) == 0 {
		return nil
	}
	
	pipeline := m.pipelines[m.selected]
	return func() tea.Msg {
		// This would need to be implemented with proper client integration
		// For now, return a message indicating the action
		action := "paused"
		if pipeline.Paused {
			action = "unpaused"
		}
		return fmt.Sprintf("Pipeline %s %s", pipeline.Name, action)
	}
}

// GetSelectedPipeline returns the currently selected pipeline name
func (m PipelinesViewModel) GetSelectedPipeline() string {
	if len(m.pipelines) == 0 || m.selected >= len(m.pipelines) {
		return ""
	}
	return m.pipelines[m.selected].Name
}

// HandlePipelinesLoaded handles the pipelines loaded message
func (m PipelinesViewModel) HandlePipelinesLoaded(msg PipelinesLoadedMsg) PipelinesViewModel {
	m.pipelines = msg.Pipelines
	m.err = msg.Error
	m.state = pipelinesStateList
	
	// Reset selection and scroll to top when loading new data
	if msg.Error == nil {
		m.selected = 0
		m.scrollOffset = 0
	}
	
	return m
}

// View renders the pipelines view
func (m PipelinesViewModel) View(width, height int) string {
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
	content.WriteString(titleStyle.Render("Pipelines"))
	content.WriteString("\n\n")
	
	if m.state == pipelinesStateLoading {
		content.WriteString("Loading pipelines...\n")
		return content.String()
	}
	
	if m.err != nil {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		content.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		content.WriteString("\n")
		return content.String()
	}
	
	if len(m.pipelines) == 0 {
		content.WriteString("No pipelines found.\n")
		return content.String()
	}
	
	// Show pipelines list
	for i, pipeline := range m.pipelines {
		status := ""
		if pipeline.Paused {
			status = " [PAUSED]"
		}
		if pipeline.Archived {
			status += " [ARCHIVED]"
		}
		
		line := fmt.Sprintf("%s%s", pipeline.Name, status)
		
		if i == m.selected {
			content.WriteString(selectedStyle.Render("> " + line))
		} else {
			content.WriteString(itemStyle.Render("  " + line))
		}
		content.WriteString("\n")
	}
	
	// Show selected pipeline info
	if len(m.pipelines) > 0 {
		content.WriteString("\n")
		infoStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1).
			MarginTop(1)
		
		pipeline := m.pipelines[m.selected]
		info := fmt.Sprintf("Pipeline: %s\nTeam: %s\nStatus: %s\nPublic: %v", 
			pipeline.Name, pipeline.TeamName,
			func() string {
				if pipeline.Paused {
					return "Paused"
				}
				return "Running"
			}(), pipeline.Public)
		
		content.WriteString(infoStyle.Render(info))
	}
	
	return content.String()
}