package tui

import (
	"fmt"
	"strings"

	"flyby/internal/concourse"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// JobsViewModel represents the jobs view
type JobsViewModel struct {
	client         *concourse.Client
	jobs           []concourse.Job
	filteredJobs   []concourse.Job
	selected       int
	loading        bool
	err            error
	pipeline       string
	triggeringJob  string
	triggerResult  string
	triggerError   error
	searchQuery    string
	searchMode     bool
}

// NewJobsViewModel creates a new jobs view model
func NewJobsViewModel() JobsViewModel {
	return JobsViewModel{
		selected:     0,
		loading:      false,
		searchQuery:  "",
		searchMode:   false,
	}
}

// JobsLoadedMsg represents loaded jobs
type JobsLoadedMsg struct {
	Jobs     []concourse.Job
	Error    error
	Pipeline string
}

// TriggerJobMsg represents a job trigger result
type TriggerJobMsg struct {
	Job     string
	Output  string
	Error   error
	Success bool
}

// TriggerJobRequestMsg represents a request to trigger a job
type TriggerJobRequestMsg struct {
	Pipeline string
	Job      string
}

// LoadJobs loads jobs from Concourse
func (m JobsViewModel) LoadJobs(client *concourse.Client, pipeline string) tea.Cmd {
	return func() tea.Msg {
		jobs, err := client.GetJobs(pipeline)
		return JobsLoadedMsg{Jobs: jobs, Error: err, Pipeline: pipeline}
	}
}

// filterJobs filters jobs based on the current search query
func (m *JobsViewModel) filterJobs() {
	if m.searchQuery == "" {
		m.filteredJobs = make([]concourse.Job, len(m.jobs))
		copy(m.filteredJobs, m.jobs)
	} else {
		m.filteredJobs = nil
		query := strings.ToLower(m.searchQuery)
		for _, job := range m.jobs {
			if strings.Contains(strings.ToLower(job.Name), query) ||
			   strings.Contains(strings.ToLower(job.PipelineName), query) ||
			   strings.Contains(strings.ToLower(job.TeamName), query) {
				m.filteredJobs = append(m.filteredJobs, job)
			}
		}
	}
	
	// Reset selection and scroll if it's out of bounds
	if m.selected >= len(m.filteredJobs) {
		m.selected = 0
	}
	if m.selected < 0 && len(m.filteredJobs) > 0 {
		m.selected = 0
	}
}

// Update handles messages for the jobs view
func (m JobsViewModel) Update(msg tea.KeyMsg) (JobsViewModel, tea.Cmd) {
	// Handle search mode
	if m.searchMode {
		switch msg.String() {
		case "enter":
			m.searchMode = false
		case "esc":
			m.searchMode = false
			m.searchQuery = ""
			m.filterJobs()
		case "backspace":
			if len(m.searchQuery) > 0 {
				m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
				m.filterJobs()
			}
		case "ctrl+u":
			m.searchQuery = ""
			m.filterJobs()
		default:
			if len(msg.String()) == 1 {
				m.searchQuery += msg.String()
				m.filterJobs()
			}
		}
		return m, nil
	}
	
	// Handle normal navigation mode
	switch msg.String() {
	case "f5":
		// Refresh jobs
		if m.client != nil && m.pipeline != "" {
			m.loading = true
			return m, m.LoadJobs(m.client, m.pipeline)
		}
	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}
		// Clear trigger results when navigating
		m.triggerResult = ""
		m.triggerError = nil
	case "down", "j":
		if m.selected < len(m.filteredJobs)-1 {
			m.selected++
		}
		// Clear trigger results when navigating
		m.triggerResult = ""
		m.triggerError = nil
	case "enter", "t":
		if len(m.filteredJobs) > 0 {
			return m, m.triggerJob()
		}
	case "x", "clear":
		// Clear trigger results
		m.triggerResult = ""
		m.triggerError = nil
		m.triggeringJob = ""
	case "b":
		if len(m.filteredJobs) > 0 {
			job := m.filteredJobs[m.selected]
			return m, func() tea.Msg {
				return SwitchViewMsg{View: ViewBuilds, Job: job.Name, Pipeline: job.PipelineName}
			}
		}
	case "/", "s":
		m.searchMode = true
	}
	
	return m, nil
}

// triggerJob triggers the selected job
func (m JobsViewModel) triggerJob() tea.Cmd {
	if len(m.filteredJobs) == 0 {
		return nil
	}
	
	job := m.filteredJobs[m.selected]
	return func() tea.Msg {
		return TriggerJobRequestMsg{
			Pipeline: job.PipelineName,
			Job:      job.Name,
		}
	}
}

// HandleJobsLoaded handles the jobs loaded message
func (m JobsViewModel) HandleJobsLoaded(msg JobsLoadedMsg) JobsViewModel {
	m.jobs = msg.Jobs
	m.err = msg.Error
	m.pipeline = msg.Pipeline
	m.loading = false
	m.selected = 0
	m.filterJobs() // Filter the loaded jobs
	return m
}

// HandleTriggerJob handles the job trigger result message
func (m JobsViewModel) HandleTriggerJob(msg TriggerJobMsg) JobsViewModel {
	m.triggeringJob = ""
	
	if msg.Error != nil {
		// Actual command execution error
		m.triggerError = msg.Error
		m.triggerResult = ""
	} else if msg.Success {
		// Job trigger succeeded
		m.triggerResult = msg.Output
		m.triggerError = nil
	} else {
		// Job trigger failed (but fly command ran)
		m.triggerResult = ""
		m.triggerError = fmt.Errorf("Job trigger failed: %s", msg.Output)
	}
	
	return m
}

// StartJobTrigger starts triggering a job
func (m JobsViewModel) StartJobTrigger(jobName string) JobsViewModel {
	m.triggeringJob = jobName
	m.triggerResult = ""
	m.triggerError = nil
	return m
}

// View renders the jobs view
func (m JobsViewModel) View(width, height int, target string) string {
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
	title := "Jobs"
	if m.pipeline != "" {
		title = fmt.Sprintf("Jobs - %s", m.pipeline)
	}
	content.WriteString(titleStyle.Render(title))
	content.WriteString("\n\n")
	
	if m.loading {
		content.WriteString("Loading jobs...\n")
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
	
	if len(m.filteredJobs) == 0 {
		if m.searchQuery != "" {
			content.WriteString("No jobs match search query.\n")
		} else {
			content.WriteString("No jobs found.\n")
		}
		return content.String()
	}
	
	// Show jobs list
	for i, job := range m.filteredJobs {
		status := ""
		if job.FinishedBuild.Status != "" {
			status = fmt.Sprintf(" [%s]", strings.ToUpper(job.FinishedBuild.Status))
		}
		
		line := fmt.Sprintf("%s%s", job.Name, status)
		
		if i == m.selected {
			content.WriteString(selectedStyle.Render("> " + line))
		} else {
			content.WriteString(itemStyle.Render("  " + line))
		}
		content.WriteString("\n")
	}
	
	// Show selected job info
	if len(m.filteredJobs) > 0 {
		content.WriteString("\n")
		infoStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1).
			MarginTop(1)
		
		job := m.filteredJobs[m.selected]
		info := fmt.Sprintf("Job: %s\nPipeline: %s\nTeam: %s", 
			job.Name, job.PipelineName, job.TeamName)
		
		if job.FinishedBuild.Status != "" {
			info += fmt.Sprintf("\nLast Build: #%d (%s)", job.FinishedBuild.ID, job.FinishedBuild.Status)
		}
		
		if job.NextBuild.ID != 0 {
			info += fmt.Sprintf("\nNext Build: #%d", job.NextBuild.ID)
		}
		
		content.WriteString(infoStyle.Render(info))
	}
	
	// Show triggering status
	if m.triggeringJob != "" {
		content.WriteString("\n")
		statusStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")).
			Bold(true).
			MarginTop(1)
		content.WriteString(statusStyle.Render(fmt.Sprintf("🔄 Triggering job: %s", m.triggeringJob)))
		content.WriteString("\n")
		content.WriteString(fmt.Sprintf("Command: fly -t %s trigger-job -j %s", target, m.triggeringJob))
	} else if m.triggerResult != "" || m.triggerError != nil {
		content.WriteString("\n")
		
		if m.triggerError != nil {
			errorStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")).
				Bold(true).
				MarginTop(1)
			content.WriteString(errorStyle.Render("❌ Job trigger failed:"))
			content.WriteString("\n")
			
			errorDetailStyle := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("196")).
				Padding(1).
				MarginTop(1)
			content.WriteString(errorDetailStyle.Render("Error:\n" + m.triggerError.Error()))
		}
		
		if m.triggerResult != "" {
			successStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("46")).
				Bold(true).
				MarginTop(1)
			content.WriteString(successStyle.Render("✅ Job triggered successfully:"))
			content.WriteString("\n")
			
			resultStyle := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("46")).
				Padding(1).
				MarginTop(1)
			content.WriteString(resultStyle.Render("Output:\n" + m.triggerResult))
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
		help = "↑/↓: navigate • Enter/t: trigger • b: builds • /,s: search • x: clear • F5: refresh • Esc: back"
	}
	content.WriteString(helpStyle.Render(help))

	return content.String()
}