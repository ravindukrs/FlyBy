package tui

import (
	"fmt"
	"strings"

	"flyby/internal/config"
	"flyby/internal/concourse"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ViewType represents the current view
type ViewType int

const (
	ViewMain ViewType = iota
	ViewTargets
	ViewPipelines
	ViewJobs
	ViewResources
	ViewBuilds
	ViewAddTarget
	ViewAuth
)

// Model represents the main TUI model
type Model struct {
	currentView   ViewType
	width, height int
	
	// Components
	mainView      MainViewModel
	targetsView   TargetsViewModel  
	pipelinesView PipelinesViewModel
	jobsView      JobsViewModel
	resourcesView ResourcesViewModel
	buildsView    BuildsViewModel
	addTargetView AddTargetViewModel
	authView      AuthViewModel
	
	// Dependencies
	configManager *config.ConfigManager
	client        *concourse.Client
	
	// State
	currentTarget string
	err           error
}

// App represents the TUI application
type App struct {
	model *Model
}

// NewApp creates a new TUI application
func NewApp() *App {
	return &App{}
}

// Run starts the TUI application
func (a *App) Run() error {
	configManager, err := config.NewConfigManager()
	if err != nil {
		return fmt.Errorf("failed to initialize config manager: %w", err)
	}
	
	model := &Model{
		currentView:   ViewMain,
		configManager: configManager,
	}
	
	// Initialize sub-models
	model.mainView = NewMainViewModel()
	model.targetsView = NewTargetsViewModel(configManager)
	model.pipelinesView = NewPipelinesViewModel()
	model.jobsView = NewJobsViewModel()
	model.resourcesView = NewResourcesViewModel()
	model.buildsView = NewBuildsViewModel(nil) // Client will be set when switching views
	model.addTargetView = NewAddTargetViewModel()
	model.authView = NewAuthViewModel()
	
	a.model = model
	
	program := tea.NewProgram(model, tea.WithAltScreen())
	_, err = program.Run()
	return err
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
		
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			// Handle hierarchical navigation
			switch m.currentView {
			case ViewBuilds:
				m.currentView = ViewJobs
				return m, nil
			case ViewResources:
				m.currentView = ViewPipelines
				return m, nil
			case ViewJobs:
				m.currentView = ViewPipelines
				return m, nil
			case ViewPipelines:
				m.currentView = ViewTargets
				return m, nil
			case ViewAddTarget:
				m.currentView = ViewTargets
				return m, nil
			case ViewAuth:
				m.currentView = ViewTargets
				return m, nil
			default:
				// From main menu or targets, do nothing (stay where we are)
			}
		}
		
		// Route key messages to current view
		return m.handleViewUpdate(msg)
		
	case SwitchViewMsg:
		m.currentView = msg.View
		m.currentTarget = msg.Target
		if msg.Target != "" {
			m.client = concourse.NewClient(msg.Target)
		}
		
		// Handle builds view switching with specific job/pipeline
		if msg.View == ViewBuilds && msg.Job != "" && msg.Pipeline != "" {
			if m.client != nil {
				// Set the client for the builds view
				m.buildsView.client = m.client
				return m, m.buildsView.LoadBuilds(msg.Pipeline, msg.Job)
			}
		}
		
		return m, m.handleViewSwitch()
		
	case PipelinesLoadedMsg:
		// Check if this is an authentication error
		if concourse.IsAuthError(msg.Error) && m.currentTarget != "" {
			// Get the target config and switch to auth view
			if target, exists := m.configManager.GetTarget(m.currentTarget); exists {
				m.authView.SetTarget(target, m.client)
				m.currentView = ViewAuth
				return m, nil
			}
		}
		m.pipelinesView = m.pipelinesView.HandlePipelinesLoaded(msg)
		return m, nil
		
	case JobsLoadedMsg:
		m.jobsView = m.jobsView.HandleJobsLoaded(msg)
		return m, nil
		
	case ResourcesLoadedMsg:
		m.resourcesView = m.resourcesView.HandleResourcesLoaded(msg)
		return m, nil
		
	case BuildsLoadedMsg:
		m.buildsView.HandleBuildsLoaded(msg)
		return m, nil
		
	case BuildRerunResultMsg:
		// Handle build rerun result messages - let the builds view handle it
		var cmd tea.Cmd
		var newModel tea.Model
		newModel, cmd = m.buildsView.Update(msg)
		m.buildsView = newModel.(BuildsViewModel)
		return m, cmd
		
	case BuildRerunTickMsg:
		// Handle build rerun tick messages - let the builds view handle it
		var cmd tea.Cmd  
		var newModel tea.Model
		newModel, cmd = m.buildsView.Update(msg)
		m.buildsView = newModel.(BuildsViewModel)
		return m, cmd
		
	case ClearRerunMessageMsg:
		// Handle clear rerun message - let the builds view handle it
		var cmd tea.Cmd
		var newModel tea.Model  
		newModel, cmd = m.buildsView.Update(msg)
		m.buildsView = newModel.(BuildsViewModel)
		return m, cmd
		
	case ResourceCheckMsg:
		var cmd tea.Cmd
		m.resourcesView, cmd = m.resourcesView.HandleResourceCheck(msg)
		return m, cmd
		
	case ReloadResourcesMsg:
		if m.client != nil {
			return m, m.resourcesView.ReloadResources(m.client)
		}
		return m, nil
		
	case TriggerJobMsg:
		m.jobsView = m.jobsView.HandleTriggerJob(msg)
		return m, nil
		
	case TriggerJobRequestMsg:
		if m.client != nil {
			jobName := fmt.Sprintf("%s/%s", msg.Pipeline, msg.Job)
			m.jobsView = m.jobsView.StartJobTrigger(jobName)
			return m, func() tea.Msg {
				success, output, err := m.client.TriggerJobWithOutput(msg.Pipeline, msg.Job)
				return TriggerJobMsg{
					Job:     jobName,
					Output:  output,
					Error:   err,
					Success: success,
				}
			}
		}
		return m, nil
		
	case CheckResourceRequestMsg:
		if m.client != nil {
			resourceName := fmt.Sprintf("%s/%s", msg.Pipeline, msg.Resource)
			m.resourcesView = m.resourcesView.StartResourceCheck(resourceName)
			return m, func() tea.Msg {
				success, output, err := m.client.CheckResourceWithOutput(msg.Pipeline, msg.Resource)
				return ResourceCheckMsg{
					Resource: resourceName,
					Output:   output,
					Error:    err,
					Success:  success,
				}
			}
		}
		return m, nil
		
	case AuthenticationMsg:
		var cmd tea.Cmd
		m.authView, cmd = m.authView.HandleAuthResult(msg)
		return m, cmd
		
	case TargetCreateMsg:
		// Handle target creation result - let the add target view handle it
		var cmd tea.Cmd
		newModel, cmd := m.addTargetView.Update(msg)
		m.addTargetView = newModel
		
		// If creation was successful, refresh targets when we switch back
		if msg.Success {
			// Reload targets configuration
			m.targetsView = NewTargetsViewModel(m.configManager)
		}
		
		return m, cmd
	}
	
	return m, nil
}

// handleViewUpdate routes updates to the current view
func (m *Model) handleViewUpdate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	
	switch m.currentView {
	case ViewMain:
		m.mainView, cmd = m.mainView.Update(msg)
	case ViewTargets:
		m.targetsView, cmd = m.targetsView.Update(msg)
	case ViewPipelines:
		m.pipelinesView, cmd = m.pipelinesView.Update(msg)
	case ViewJobs:
		m.jobsView, cmd = m.jobsView.Update(msg)
	case ViewResources:
		m.resourcesView, cmd = m.resourcesView.Update(msg)
	case ViewBuilds:
		var newModel tea.Model
		newModel, cmd = m.buildsView.Update(msg)
		m.buildsView = newModel.(BuildsViewModel)
	case ViewAddTarget:
		newModel, cmd := m.addTargetView.Update(msg)
		m.addTargetView = newModel
		return m, cmd
	case ViewAuth:
		m.authView, cmd = m.authView.Update(msg)
	}
	
	return m, cmd
}

// handleViewSwitch handles switching between views
func (m *Model) handleViewSwitch() tea.Cmd {
	switch m.currentView {
	case ViewPipelines:
		if m.client != nil {
			return m.pipelinesView.LoadPipelines(m.client)
		}
	case ViewJobs:
		if m.client != nil && m.pipelinesView.GetSelectedPipeline() != "" {
			// Set client for jobs view so it can refresh
			m.jobsView.client = m.client
			return m.jobsView.LoadJobs(m.client, m.pipelinesView.GetSelectedPipeline())
		}
	case ViewResources:
		if m.client != nil && m.pipelinesView.GetSelectedPipeline() != "" {
			// Set client for resources view so it can refresh
			m.resourcesView.client = m.client
			return m.resourcesView.LoadResources(m.client, m.pipelinesView.GetSelectedPipeline())
		}
	}
	return nil
}

// View renders the current view
func (m *Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}
	
	// Header
	header := m.renderHeader()
	
	// Content
	var content string
	switch m.currentView {
	case ViewMain:
		content = m.mainView.View(m.width, m.height-3)
	case ViewTargets:
		content = m.targetsView.View(m.width, m.height-3)
	case ViewPipelines:
		content = m.pipelinesView.View(m.width, m.height-3)
	case ViewJobs:
		content = m.jobsView.View(m.width, m.height-3, m.client.GetTarget())
	case ViewResources:
		content = m.resourcesView.View(m.width, m.height-3, m.client.GetTarget())
	case ViewBuilds:
		content = m.buildsView.View()
	case ViewAddTarget:
		content = m.addTargetView.View(m.width, m.height-3)
	case ViewAuth:
		content = m.authView.View(m.width, m.height-3)
	}
	
	// Footer
	footer := m.renderFooter()
	
	return lipgloss.JoinVertical(lipgloss.Left, header, content, footer)
}

// renderHeader renders the application header
func (m *Model) renderHeader() string {
	style := lipgloss.NewStyle().
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230")).
		Bold(true).
		Padding(0, 1).
		Width(m.width)
	
	title := "FlyBy - Concourse CI Terminal UI"
	if m.currentTarget != "" {
		title += fmt.Sprintf(" | Target: %s", m.currentTarget)
	}
	
	return style.Render(title)
}

// renderFooter renders the application footer
func (m *Model) renderFooter() string {
	style := lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("252")).
		Padding(0, 1).
		Width(m.width)
	
	var keyHelp []string
	
	switch m.currentView {
	case ViewMain:
		keyHelp = []string{"↑/↓: navigate", "enter: select", "q: quit"}
	case ViewTargets:
		keyHelp = []string{"↑/↓: navigate", "enter: select", "a: add target", "d: delete", "esc: back", "q: quit"}
	case ViewPipelines:
		keyHelp = []string{"↑/↓: navigate", "j: jobs", "r: resources", "t: trigger", "p: pause/unpause", "F5: refresh", "esc: back", "q: quit"}
	case ViewJobs:
		keyHelp = []string{"↑/↓: navigate", "enter: trigger", "b: builds", "F5: refresh", "esc: back", "q: quit"}
	case ViewResources:
		keyHelp = []string{"↑/↓: navigate", "enter: check", "F5: refresh", "esc: back", "q: quit"}
	case ViewBuilds:
		keyHelp = []string{"↑/↓: navigate", "enter: rerun build", "F5: refresh", "esc: back", "q: quit"}
	case ViewAddTarget:
		keyHelp = []string{"tab: next field", "enter: save", "esc: cancel", "q: quit"}
	case ViewAuth:
		keyHelp = []string{"enter/y: login", "n: cancel", "esc: back", "q: quit"}
	}
	
	return style.Render(strings.Join(keyHelp, " • "))
}

// SwitchViewMsg is a message for switching views
type SwitchViewMsg struct {
	View     ViewType
	Target   string
	Job      string
	Pipeline string
	Data     interface{}
}