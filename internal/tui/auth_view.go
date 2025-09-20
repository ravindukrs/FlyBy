package tui

import (
	"fmt"
	"strings"

	"flyby/internal/concourse"
	"flyby/internal/config"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AuthViewModel represents the authentication view
type AuthViewModel struct {
	target        config.Target
	client        *concourse.Client
	authenticating bool
	error         error
	success       bool
}

// AuthenticationMsg represents authentication result
type AuthenticationMsg struct {
	Success bool
	Error   error
	Target  string
}

// NewAuthViewModel creates a new authentication view model
func NewAuthViewModel() AuthViewModel {
	return AuthViewModel{
		authenticating: false,
		success:        false,
	}
}

// SetTarget sets the target to authenticate with
func (m *AuthViewModel) SetTarget(target config.Target, client *concourse.Client) {
	m.target = target
	m.client = client
	m.authenticating = false
	m.error = nil
	m.success = false
}

// StartAuthentication begins the authentication process
func (m *AuthViewModel) StartAuthentication() tea.Cmd {
	m.authenticating = true
	m.error = nil
	
	client := m.client
	target := m.target
	
	return func() tea.Msg {
		// Perform interactive login
		err := client.LoginInteractive(target.GetURL(), target.Team)
		return AuthenticationMsg{
			Success: err == nil,
			Error:   err,
			Target:  target.Name,
		}
	}
}

// Update handles messages for the authentication view
func (m AuthViewModel) Update(msg tea.KeyMsg) (AuthViewModel, tea.Cmd) {
	if m.authenticating {
		// Don't handle keys during authentication
		return m, nil
	}
	
	switch msg.String() {
	case "enter", "y":
		cmd := m.StartAuthentication()
		return m, cmd
	case "n":
		// Go back to targets
		return m, func() tea.Msg {
			return SwitchViewMsg{View: ViewTargets}
		}
	case "esc":
		return m, func() tea.Msg {
			return SwitchViewMsg{View: ViewTargets}
		}
	}
	
	return m, nil
}

// HandleAuthResult handles authentication result message
func (m AuthViewModel) HandleAuthResult(msg AuthenticationMsg) (AuthViewModel, tea.Cmd) {
	m.authenticating = false
	m.success = msg.Success
	m.error = msg.Error
	
	if m.success {
		// Authentication successful, go to pipelines
		return m, func() tea.Msg {
			return SwitchViewMsg{View: ViewPipelines, Target: msg.Target}
		}
	}
	
	return m, nil
}

// View renders the authentication view
func (m AuthViewModel) View(width, height int) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		MarginBottom(2)
	
	contentStyle := lipgloss.NewStyle().
		Padding(1).
		MarginBottom(1)
	
	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true)
	
	successStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("46")).
		Bold(true)
	
	promptStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true)
	
	var content strings.Builder
	
	if m.authenticating {
		content.WriteString(titleStyle.Render("Authenticating..."))
		content.WriteString("\n\n")
		content.WriteString(contentStyle.Render("Opening browser for authentication..."))
		content.WriteString("\n")
		content.WriteString(contentStyle.Render("Please complete the login process in your browser."))
		content.WriteString("\n\n")
		content.WriteString(promptStyle.Render("Waiting for authentication to complete..."))
		
	} else if m.success {
		content.WriteString(titleStyle.Render("Authentication Successful!"))
		content.WriteString("\n\n")
		content.WriteString(successStyle.Render("✓ Successfully logged in to " + m.target.Name))
		content.WriteString("\n")
		content.WriteString(contentStyle.Render("Redirecting to pipelines..."))
		
	} else if m.error != nil {
		content.WriteString(titleStyle.Render("Authentication Failed"))
		content.WriteString("\n\n")
		content.WriteString(errorStyle.Render("✗ " + m.error.Error()))
		content.WriteString("\n\n")
		content.WriteString(contentStyle.Render("Would you like to try again?"))
		content.WriteString("\n\n")
		content.WriteString(promptStyle.Render("Press Enter/y to retry, n to go back, or Esc to cancel"))
		
	} else {
		content.WriteString(titleStyle.Render("Authentication Required"))
		content.WriteString("\n\n")
		content.WriteString(contentStyle.Render(fmt.Sprintf("Target: %s", m.target.Name)))
		content.WriteString("\n")
		content.WriteString(contentStyle.Render(fmt.Sprintf("Team: %s", m.target.Team)))
		content.WriteString("\n")
		content.WriteString(contentStyle.Render(fmt.Sprintf("URL: %s", m.target.GetURL())))
		content.WriteString("\n\n")
		content.WriteString(contentStyle.Render("You need to log in to access this Concourse instance."))
		content.WriteString("\n")
		content.WriteString(contentStyle.Render("This will open your browser for authentication."))
		content.WriteString("\n\n")
		content.WriteString(promptStyle.Render("Press Enter/y to login, n to go back, or Esc to cancel"))
	}
	
	return content.String()
}