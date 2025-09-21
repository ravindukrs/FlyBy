package tui

import (
	"fmt"
	"strings"
	"time"

	"flyby/internal/concourse"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type resourcesState int

const (
	resourcesStateLoading resourcesState = iota
	resourcesStateList
	resourcesStateChecking
)

// formatTimeAgo returns a human-readable relative time string
func formatTimeAgo(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	
	duration := time.Since(t)
	
	if duration < time.Minute {
		return "just now"
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else if duration < 7*24*time.Hour {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	} else {
		weeks := int(duration.Hours() / (24 * 7))
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	}
}

// ResourcesViewModel represents the resources view
type ResourcesViewModel struct {
	client           *concourse.Client
	resources        []concourse.Resource
	filteredResources []concourse.Resource
	selected         int
	state            resourcesState
	err              error
	pipeline         string
	checkingResource string
	checkResult      string
	checkError       error
	searchQuery      string
	searchMode       bool
}

// ResourceCheckMsg represents a resource check result
type ResourceCheckMsg struct {
	Resource string
	Output   string
	Error    error
	Success  bool
}

// CheckResourceRequestMsg represents a request to check a resource
type CheckResourceRequestMsg struct {
	Pipeline string
	Resource string
}

// ReloadResourcesMsg represents a request to reload resources data
type ReloadResourcesMsg struct {
	Pipeline string
}

// NewResourcesViewModel creates a new resources view model
func NewResourcesViewModel() ResourcesViewModel {
	return ResourcesViewModel{
		selected:     0,
		state:        resourcesStateList,
		searchQuery:  "",
		searchMode:   false,
	}
}

// ResourcesLoadedMsg represents loaded resources
type ResourcesLoadedMsg struct {
	Resources []concourse.Resource
	Error     error
	Pipeline  string
	IsReload  bool // true when reloading after operations, false for initial load
}

// LoadResources loads resources for a specific pipeline
func (m ResourcesViewModel) LoadResources(client *concourse.Client, pipeline string) tea.Cmd {
	return func() tea.Msg {
		resources, err := client.GetResources(pipeline)
		if err != nil {
			return ResourcesLoadedMsg{Error: err, Pipeline: pipeline}
		}
		return ResourcesLoadedMsg{Resources: resources, Pipeline: pipeline}
	}
}

// filterResources filters resources based on the current search query
func (m *ResourcesViewModel) filterResources() {
	if m.searchQuery == "" {
		m.filteredResources = make([]concourse.Resource, len(m.resources))
		copy(m.filteredResources, m.resources)
	} else {
		m.filteredResources = nil
		query := strings.ToLower(m.searchQuery)
		for _, resource := range m.resources {
			if strings.Contains(strings.ToLower(resource.Name), query) ||
			   strings.Contains(strings.ToLower(resource.Type), query) ||
			   strings.Contains(strings.ToLower(resource.PipelineName), query) ||
			   strings.Contains(strings.ToLower(resource.TeamName), query) {
				m.filteredResources = append(m.filteredResources, resource)
			}
		}
	}
	
	// Reset selection and scroll if it's out of bounds
	if m.selected >= len(m.filteredResources) {
		m.selected = 0
	}
	if m.selected < 0 && len(m.filteredResources) > 0 {
		m.selected = 0
	}
}

// ReloadResources reloads resources data (used after successful operations)
func (m ResourcesViewModel) ReloadResources(client *concourse.Client) tea.Cmd {
	if m.pipeline == "" {
		return nil
	}
	
	return func() tea.Msg {
		resources, err := client.GetResources(m.pipeline)
		if err != nil {
			// Don't show error for background reload, just keep existing data
			return nil
		}
		return ResourcesLoadedMsg{Resources: resources, IsReload: true}
	}
}

// Update handles messages for the resources view
func (m ResourcesViewModel) Update(msg tea.KeyMsg) (ResourcesViewModel, tea.Cmd) {
	// Handle search mode
	if m.searchMode {
		switch msg.String() {
		case "enter":
			m.searchMode = false
		case "esc":
			m.searchMode = false
			m.searchQuery = ""
			m.filterResources()
		case "backspace":
			if len(m.searchQuery) > 0 {
				m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
				m.filterResources()
			}
		case "ctrl+u":
			m.searchQuery = ""
			m.filterResources()
		default:
			if len(msg.String()) == 1 {
				m.searchQuery += msg.String()
				m.filterResources()
			}
		}
		return m, nil
	}
	
	// Handle normal navigation mode
	switch msg.String() {
	case "f5":
		// Refresh resources
		if m.client != nil && m.pipeline != "" {
			m.state = resourcesStateLoading
			return m, m.LoadResources(m.client, m.pipeline)
		}
	case "up", "k":
		if m.selected > 0 {
			m.selected--
			// Clear check results when navigating
			m.checkResult = ""
			m.checkError = nil
		}
	case "down", "j":
		if m.selected < len(m.filteredResources)-1 {
			m.selected++
			// Clear check results when navigating
			m.checkResult = ""
			m.checkError = nil
		}
	case "enter", "c":
		if len(m.filteredResources) > 0 {
			resource := m.filteredResources[m.selected]
			return m, func() tea.Msg {
				return CheckResourceRequestMsg{
					Pipeline: resource.PipelineName,
					Resource: resource.Name,
				}
			}
		}
	case "x", "clear":
		// Clear check results
		m.checkResult = ""
		m.checkError = nil
		m.checkingResource = ""
	case "/", "s":
		m.searchMode = true
	}
	
	return m, nil
}

// checkResource checks the selected resource
func (m *ResourcesViewModel) checkResource(client *concourse.Client) tea.Cmd {
	if len(m.filteredResources) == 0 || client == nil {
		return nil
	}
	
	resource := m.filteredResources[m.selected]
	resourceName := fmt.Sprintf("%s/%s", resource.PipelineName, resource.Name)
	
	// Set checking state
	m.checkingResource = resourceName
	m.checkResult = ""
	m.checkError = nil
	
	return func() tea.Msg {
		success, output, err := client.CheckResourceWithOutput(resource.PipelineName, resource.Name)
		return ResourceCheckMsg{
			Resource: resourceName,
			Output:   output,
			Error:    err,
			Success:  success,
		}
	}
}

// HandleResourcesLoaded handles the resources loaded message
func (m ResourcesViewModel) HandleResourcesLoaded(msg ResourcesLoadedMsg) ResourcesViewModel {
	m.resources = msg.Resources
	m.err = msg.Error
	m.pipeline = msg.Pipeline
	m.state = resourcesStateList
	
	// For reloads, preserve the current selection; for initial loads, reset to 0
	if !msg.IsReload {
		m.selected = 0
	} else {
		// Ensure selection is still valid after reload
		if m.selected >= len(m.resources) {
			m.selected = 0
		}
	}
	
	m.filterResources() // Filter the loaded resources
	return m
}

// HandleResourceCheck handles the resource check result message
func (m ResourcesViewModel) HandleResourceCheck(msg ResourceCheckMsg) (ResourcesViewModel, tea.Cmd) {
	m.checkingResource = ""
	
	var cmd tea.Cmd
	
	if msg.Error != nil {
		// Actual command execution error
		m.checkError = msg.Error
		m.checkResult = ""
	} else if msg.Success {
		// Resource check succeeded - reload resources to get updated timestamps
		m.checkResult = msg.Output
		m.checkError = nil
		
		// Trigger resource reload
		cmd = func() tea.Msg {
			return ReloadResourcesMsg{Pipeline: m.pipeline}
		}
	} else {
		// Resource check failed (but fly command ran)
		m.checkResult = ""
		m.checkError = fmt.Errorf("Resource check failed: %s", msg.Output)
	}
	
	return m, cmd
}

// StartResourceCheck starts checking a resource
func (m ResourcesViewModel) StartResourceCheck(resourceName string) ResourcesViewModel {
	m.checkingResource = resourceName
	m.checkResult = ""
	m.checkError = nil
	return m
}

// View renders the resources view
func (m ResourcesViewModel) View(width, height int, target string) string {
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
	
	searchStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		MarginBottom(1)
	
	searchActiveStyle := searchStyle.Copy().
		BorderForeground(lipgloss.Color("205"))
	
	var content strings.Builder
	title := "Resources"
	if m.pipeline != "" {
		title = fmt.Sprintf("Resources - %s", m.pipeline)
	}
	content.WriteString(titleStyle.Render(title))
	content.WriteString("\n\n")
	
	if m.state == resourcesStateLoading {
		content.WriteString("Loading resources...\n")
		return content.String()
	}
	
	if m.err != nil {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		content.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		content.WriteString("\n")
		return content.String()
	}
	
	// Add search box
	searchPrompt := "Search: "
	searchText := m.searchQuery
	if m.searchMode {
		searchText += "█" // cursor
		content.WriteString(searchActiveStyle.Render(searchPrompt + searchText))
	} else {
		if m.searchQuery != "" {
			content.WriteString(searchStyle.Render(searchPrompt + searchText))
		} else {
			content.WriteString(searchStyle.Render(searchPrompt + "(/,s to search)"))
		}
	}
	content.WriteString("\n\n")
	
	if len(m.filteredResources) == 0 {
		if m.searchQuery != "" {
			content.WriteString("No resources match search query.\n")
		} else {
			content.WriteString("No resources found.\n")
		}
		return content.String()
	}
	
	// Show resources list
	for i, resource := range m.filteredResources {
		line := fmt.Sprintf("%s (%s)", resource.Name, resource.Type)
		
		if i == m.selected {
			content.WriteString(selectedStyle.Render("> " + line))
		} else {
			content.WriteString(itemStyle.Render("  " + line))
		}
		content.WriteString("\n")
	}
	
	// Show selected resource info
	if len(m.filteredResources) > 0 {
		content.WriteString("\n")
		infoStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1).
			MarginTop(1)
		
		resource := m.filteredResources[m.selected]
		info := fmt.Sprintf("Resource: %s\nType: %s\nPipeline: %s\nTeam: %s", 
			resource.Name, resource.Type, resource.PipelineName, resource.TeamName)
		
		lastChecked := resource.GetLastChecked()
		if !lastChecked.IsZero() {
			info += fmt.Sprintf("\nLast Checked: %s", formatTimeAgo(lastChecked))
		}
		
		// Show version information if available
		if len(resource.Version) > 0 {
			info += "\nVersion:"
			for key, value := range resource.Version {
				info += fmt.Sprintf("\n  %s: %v", key, value)
			}
		}
		
		// Show metadata if available
		if len(resource.Metadata) > 0 {
			info += "\nMetadata:"
			for _, metadata := range resource.Metadata {
				info += fmt.Sprintf("\n  %s: %s", metadata.Name, metadata.Value)
			}
		}
		
		content.WriteString(infoStyle.Render(info))
	}
	
	// Show resource checking status and results
	if m.checkingResource != "" {
		content.WriteString("\n")
		statusStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")).
			Bold(true).
			MarginTop(1)
		content.WriteString(statusStyle.Render(fmt.Sprintf("🔄 Checking resource: %s", m.checkingResource)))
		content.WriteString("\n")
		content.WriteString(fmt.Sprintf("Command: fly -t %s check-resource -r %s", target, m.checkingResource))
	} else if m.checkResult != "" || m.checkError != nil {
		content.WriteString("\n")
		
		if m.checkError != nil {
			errorStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")).
				Bold(true).
				MarginTop(1)
			content.WriteString(errorStyle.Render("❌ Resource check failed:"))
			content.WriteString("\n")
			content.WriteString(errorStyle.Render(m.checkError.Error()))
		} else {
			successStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("46")).
				Bold(true).
				MarginTop(1)
			content.WriteString(successStyle.Render("✅ Resource check completed successfully!"))
			content.WriteString("\n")
			
			if m.checkResult != "" {
				resultStyle := lipgloss.NewStyle().
					Border(lipgloss.RoundedBorder()).
					BorderForeground(lipgloss.Color("46")).
					Padding(1).
					MarginTop(1)
				content.WriteString(resultStyle.Render("Output:\n" + m.checkResult))
			}
		}
	}
	
	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true).
		MarginTop(1)
	
	var help string
	if m.searchMode {
		help = "Enter: finish search • Esc: cancel search • Ctrl+U: clear"
	} else {
		help = "↑/↓: navigate • Enter/c: check • /,s: search • x: clear • F5: refresh • Esc: back"
	}
	content.WriteString(helpStyle.Render(help))
	
	return content.String()
}