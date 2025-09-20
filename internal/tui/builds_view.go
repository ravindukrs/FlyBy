package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"flyby/internal/concourse"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type buildsState int

const (
	buildsStateLoading buildsState = iota
	buildsStateList
	buildsStateRerunning
)

// BuildsViewModel represents the builds view
type BuildsViewModel struct {
	client       *concourse.Client
	builds       []concourse.Build
	cursor       int
	state        buildsState
	err          error
	job          string
	pipeline     string
	rerunMessage string
}

// NewBuildsViewModel creates a new builds view model
func NewBuildsViewModel(client *concourse.Client) BuildsViewModel {
	return BuildsViewModel{
		client: client,
		cursor: 0,
		state:  buildsStateLoading,
	}
}

// BuildsLoadedMsg represents loaded builds
type BuildsLoadedMsg struct {
	Builds   []concourse.Build
	Error    error
	Job      string
	Pipeline string
}

// BuildRerunResultMsg represents the result of a build rerun operation
type BuildRerunResultMsg struct {
	Success bool
	Output  string
	Error   error
	Build   int
}

// BuildRerunTickMsg for animation during rerunning
type BuildRerunTickMsg struct{}

// ClearRerunMessageMsg to clear rerun messages
type ClearRerunMessageMsg struct{}

func (m BuildsViewModel) Init() tea.Cmd {
	return nil
}

func (m BuildsViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case buildsStateLoading:
			if msg.String() == "q" || msg.String() == "esc" {
				// Go back to jobs view
				return m, func() tea.Msg {
					return SwitchViewMsg{View: ViewJobs}
				}
			}
		case buildsStateList:
			switch msg.String() {
			case "f5":
				// Refresh builds
				if m.client != nil && m.pipeline != "" && m.job != "" {
					m.state = buildsStateLoading
					return m, m.LoadBuilds(m.pipeline, m.job)
				}
			case "q", "esc":
				// Go back to jobs view
				return m, func() tea.Msg {
					return SwitchViewMsg{View: ViewJobs}
				}
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.builds)-1 {
					m.cursor++
				}
			case "enter":
				if len(m.builds) > 0 {
					selected := m.builds[m.cursor]
					// Convert build name (string) to integer
					buildNum, err := strconv.Atoi(selected.Name)
					if err != nil {
						m.rerunMessage = fmt.Sprintf("Error: Invalid build number %s", selected.Name)
						return m, tea.Tick(5*time.Second, func(time.Time) tea.Msg {
							return ClearRerunMessageMsg{}
						})
					}
					
					// Start rerunning the selected build
					m.state = buildsStateRerunning
					m.rerunMessage = fmt.Sprintf("Rerunning build %s/%s #%d...", m.pipeline, m.job, buildNum)
					
					return m, tea.Batch(
						func() tea.Msg {
							success, output, err := m.client.RerunBuildWithOutput(m.pipeline, m.job, buildNum)
							return BuildRerunResultMsg{
								Success: success,
								Output:  output,
								Error:   err,
								Build:   buildNum,
							}
						},
						tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
							return BuildRerunTickMsg{}
						}),
					)
				}
			}
		case buildsStateRerunning:
			// Only allow quitting during rerunning state
			if msg.String() == "q" || msg.String() == "esc" {
				return m, func() tea.Msg {
					return SwitchViewMsg{View: ViewJobs}
				}
			}
		}
	case BuildRerunResultMsg:
		if msg.Error != nil {
			m.state = buildsStateList
			m.rerunMessage = fmt.Sprintf("Error: %v", msg.Error)
		} else if msg.Success {
			m.state = buildsStateList
			m.rerunMessage = fmt.Sprintf("✓ Successfully reran build %s/%s #%d: %s", m.pipeline, m.job, msg.Build, msg.Output)
			// Reload builds after successful rerun to show the new build
			return m, tea.Batch(
				tea.Tick(5*time.Second, func(time.Time) tea.Msg {
					return ClearRerunMessageMsg{}
				}),
				tea.Tick(2*time.Second, func(time.Time) tea.Msg {
					// Reload builds after a short delay to let the new build appear
					builds, err := m.client.GetBuilds(m.pipeline, m.job, 50)
					if err != nil {
						return BuildsLoadedMsg{Error: err, Job: m.job, Pipeline: m.pipeline}
					}
					return BuildsLoadedMsg{Builds: builds, Job: m.job, Pipeline: m.pipeline}
				}),
			)
		} else {
			m.state = buildsStateList
			m.rerunMessage = fmt.Sprintf("✗ Failed to rerun build %s/%s #%d: %s", m.pipeline, m.job, msg.Build, msg.Output)
		}
		// Clear the message after 5 seconds
		return m, tea.Tick(5*time.Second, func(time.Time) tea.Msg {
			return ClearRerunMessageMsg{}
		})
	case BuildRerunTickMsg:
		if m.state == buildsStateRerunning {
			// Continue ticking animation
			return m, tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
				return BuildRerunTickMsg{}
			})
		}
	case ClearRerunMessageMsg:
		m.rerunMessage = ""
	}
	
	return m, nil
}

// LoadBuilds loads builds for a specific job
func (m *BuildsViewModel) LoadBuilds(pipeline, job string) tea.Cmd {
	m.state = buildsStateLoading
	m.err = nil
	m.job = job
	m.pipeline = pipeline
	m.cursor = 0
	
	return func() tea.Msg {
		builds, err := m.client.GetBuilds(pipeline, job, 50) // Get last 50 builds
		if err != nil {
			return BuildsLoadedMsg{Error: err, Job: job, Pipeline: pipeline}
		}
		return BuildsLoadedMsg{Builds: builds, Job: job, Pipeline: pipeline}
	}
}

// HandleBuildsLoaded handles the builds loaded message
func (m *BuildsViewModel) HandleBuildsLoaded(msg BuildsLoadedMsg) {
	m.builds = msg.Builds
	m.err = msg.Error
	m.job = msg.Job
	m.pipeline = msg.Pipeline
	m.state = buildsStateList
	m.cursor = 0
}

// formatTimeAgo returns a human-readable relative time string
func formatBuildTimeAgo(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}
	
	duration := time.Since(t)
	
	if duration < time.Minute {
		return "just now"
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1min ago"
		}
		return fmt.Sprintf("%dmin ago", minutes)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1hr ago"
		}
		return fmt.Sprintf("%dhr ago", hours)
	} else if duration < 7*24*time.Hour {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1day ago"
		}
		return fmt.Sprintf("%dd ago", days)
	} else {
		return t.Format("Jan 2")
	}
}

// View renders the builds view
func (m BuildsViewModel) View() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		MarginBottom(1)

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		PaddingLeft(1).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.Color("205"))

	itemStyle := lipgloss.NewStyle().
		PaddingLeft(2)

	var content strings.Builder
	title := "Builds"
	if m.job != "" {
		title = fmt.Sprintf("Builds - %s/%s", m.pipeline, m.job)
	}
	content.WriteString(titleStyle.Render(title))
	content.WriteString("\n\n")

	switch m.state {
	case buildsStateLoading:
		content.WriteString("Loading builds...\n")
	case buildsStateList, buildsStateRerunning:
		if m.err != nil {
			errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
			content.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
			content.WriteString("\n")
		} else if len(m.builds) == 0 {
			content.WriteString("No builds found.\n")
		} else {
			// Show builds list
			for i, build := range m.builds {
				status := strings.ToUpper(build.Status)
				statusColor := "240" // default gray
				
				switch status {
				case "SUCCEEDED":
					statusColor = "46" // green
				case "FAILED":
					statusColor = "196" // red
				case "STARTED", "PENDING":
					statusColor = "226" // yellow
				}
				
				statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor)).Bold(true)
				
				startTime := formatBuildTimeAgo(build.GetStartTime())
				duration := "unknown"
				
				if !build.GetStartTime().IsZero() && !build.GetEndTime().IsZero() {
					dur := build.GetEndTime().Sub(build.GetStartTime())
					if dur < time.Minute {
						duration = fmt.Sprintf("%ds", int(dur.Seconds()))
					} else if dur < time.Hour {
						duration = fmt.Sprintf("%dm%ds", int(dur.Minutes()), int(dur.Seconds())%60)
					} else {
						duration = fmt.Sprintf("%dh%dm", int(dur.Hours()), int(dur.Minutes())%60)
					}
				}
				
				line := fmt.Sprintf("#%s %s %s (%s)", build.Name, statusStyle.Render(fmt.Sprintf("[%s]", status)), startTime, duration)
				
				if i == m.cursor {
					content.WriteString(selectedStyle.Render("> " + line))
				} else {
					content.WriteString(itemStyle.Render("  " + line))
				}
				content.WriteString("\n")
			}

			// Show selected build info
			content.WriteString("\n")
			infoStyle := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("240")).
				Padding(1).
				MarginTop(1)

			build := m.builds[m.cursor]
			info := fmt.Sprintf("Build: #%s\nJob: %s/%s\nStatus: %s\nTeam: %s", 
				build.Name, build.PipelineName, build.JobName, strings.ToUpper(build.Status), build.TeamName)
			
			if !build.GetStartTime().IsZero() {
				info += fmt.Sprintf("\nStarted: %s", build.GetStartTime().Format("2006-01-02 15:04:05"))
			}
			
			if !build.GetEndTime().IsZero() {
				info += fmt.Sprintf("\nEnded: %s", build.GetEndTime().Format("2006-01-02 15:04:05"))
			}

			content.WriteString(infoStyle.Render(info))
		}
		
		// Show rerun status/message
		if m.state == buildsStateRerunning {
			content.WriteString("\n\n")
			loadingStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("226")).
				Bold(true).
				MarginTop(1)
			content.WriteString(loadingStyle.Render("🔄 " + m.rerunMessage))
		} else if m.rerunMessage != "" {
			content.WriteString("\n\n")
			if strings.Contains(m.rerunMessage, "✓") {
				successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true)
				content.WriteString(successStyle.Render(m.rerunMessage))
			} else if strings.Contains(m.rerunMessage, "✗") || strings.Contains(m.rerunMessage, "Error") {
				errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
				content.WriteString(errorStyle.Render(m.rerunMessage))
			} else {
				content.WriteString(m.rerunMessage)
			}
		}
	}

	// Add instructions
	content.WriteString("\n\n")
	instructionsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true)
	
	switch m.state {
	case buildsStateLoading:
		content.WriteString(instructionsStyle.Render("Press 'q' or 'esc' to go back"))
	case buildsStateList:
		content.WriteString(instructionsStyle.Render("↑/↓: Navigate • Enter: Rerun build • q/esc: Back to jobs"))
	case buildsStateRerunning:
		content.WriteString(instructionsStyle.Render("Rerunning build... • q/esc: Back to jobs"))
	}

	return content.String()
}