package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AddTargetViewModel represents the add target form
type AddTargetViewModel struct {
	fields     []string
	values     []string
	focused    int
	submitted  bool
	err        error
	saving     bool
	flyCommand string
	saveResult string
}

// TargetCreateMsg represents the result of target creation
type TargetCreateMsg struct {
	Success bool
	Output  string
	Error   error
	Command string
}

// ExitAndRunCommandMsg represents a request to exit TUI and run a command
type ExitAndRunCommandMsg struct {
	Command string
	Message string
}

// NewAddTargetViewModel creates a new add target view model
func NewAddTargetViewModel() AddTargetViewModel {
	return AddTargetViewModel{
		fields: []string{"Name", "URL", "Team"},
		values: []string{"", "", ""},
		focused: 0,
	}
}

// Init initializes the add target view model
func (m AddTargetViewModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the add target view
func (m AddTargetViewModel) Update(msg tea.Msg) (AddTargetViewModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			if !m.saving && m.saveResult == "" {
				m.focused = (m.focused + 1) % len(m.fields)
			}
		case "shift+tab":
			if !m.saving && m.saveResult == "" {
				m.focused = (m.focused - 1 + len(m.fields)) % len(m.fields)
			}
		case "enter":
			if m.canSubmit() && !m.saving && m.saveResult == "" {
				return m.startSave()
			} else if m.saveResult != "" {
				// If showing results, go back to targets view
				return m, func() tea.Msg {
					return SwitchViewMsg{View: ViewTargets}
				}
			}
		case "r":
			// Retry checking target authentication
			if !m.saving && m.saveResult != "" && strings.Contains(m.saveResult, "Interactive authentication required") {
				// Clear results and check if target is now authenticated
				m.saveResult = ""
				m.err = nil
				
				name := strings.TrimSpace(m.values[0])
				if name != "" {
					m.saving = true
					return m, func() tea.Msg {
						checkCmd := exec.Command("fly", "-t", name, "status")
						checkOutput, checkErr := checkCmd.CombinedOutput()
						
						if checkErr == nil && strings.Contains(string(checkOutput), "logged in successfully") {
							return TargetCreateMsg{
								Success: true,
								Output:  fmt.Sprintf("✅ Target '%s' is now authenticated and ready to use!", name),
								Error:   nil,
								Command: fmt.Sprintf("fly -t %s status", name),
							}
						}
						
						// Still not authenticated, check what the error is
						outputStr := strings.TrimSpace(string(checkOutput))
						if strings.Contains(outputStr, "not found") || strings.Contains(outputStr, "no such") {
							return TargetCreateMsg{
								Success: false,
								Output:  fmt.Sprintf("❌ Target '%s' not found. Please run the fly login command in a separate terminal:\n\nfly -t %s login -c %s -n %s\n\nThen press 'r' again to retry.", 
									name, name, strings.TrimSpace(m.values[1]), strings.TrimSpace(m.values[2])),
								Error:   nil,
								Command: "",
							}
						}
						
						return TargetCreateMsg{
							Success: false,
							Output:  fmt.Sprintf("⏳ Target '%s' exists but authentication is still pending.\n\nIf you're still completing browser authentication, wait and press 'r' again.\n\nIf authentication failed, run this command in a separate terminal:\nfly -t %s login -c %s -n %s", 
								name, name, strings.TrimSpace(m.values[1]), strings.TrimSpace(m.values[2])),
							Error:   nil,
							Command: "",
						}
					}
				}
			} else {
				// If we're in input mode and not showing auth error, treat 'r' as regular text input
				if m.focused < len(m.values) && !m.saving && m.saveResult == "" {
					m.values[m.focused] += "r"
				}
			}
			return m, nil
		case "c":
			// Copy command to clipboard (when showing interactive auth message)
			if !m.saving && m.saveResult != "" && strings.Contains(m.saveResult, "Interactive authentication required") {
				name := strings.TrimSpace(m.values[0])
				url := strings.TrimSpace(m.values[1])  
				team := strings.TrimSpace(m.values[2])
				if team == "" {
					team = "main"
				}
				
				command := fmt.Sprintf("fly -t %s login -c %s -n %s", name, url, team)
				
				// Try to copy to clipboard using pbcopy on macOS
				copyCmd := exec.Command("pbcopy")
				copyCmd.Stdin = strings.NewReader(command)
				err := copyCmd.Run()
				
				if err == nil {
					// Update the result to show command was copied
					m.saveResult = fmt.Sprintf("Interactive authentication required.\n\n✅ Command copied to clipboard!\n\nTo complete target creation:\n\n1. Open a new terminal window\n2. Paste and run the command (Cmd+V)\n3. Complete browser authentication  \n4. Press 'r' here to retry checking the target\n\nCommand: fly -t %s login -c %s -n %s", name, url, team)
				}
			} else {
				// If we're in input mode and not showing auth error, treat 'c' as regular text input
				if m.focused < len(m.values) && !m.saving && m.saveResult == "" {
					m.values[m.focused] += "c"
				}
			}
			return m, nil
		case "esc":
			return m, func() tea.Msg {
				return SwitchViewMsg{View: ViewTargets}
			}
		case "ctrl+c":
			// Allow copying - handled by terminal
			return m, nil
		case "ctrl+v":
			// Paste is handled by terminal and comes as regular text input
			return m, nil
		default:
			// Handle text input for the focused field
			if m.focused < len(m.values) && !m.saving && m.saveResult == "" {
				switch msg.String() {
				case "backspace":
					if len(m.values[m.focused]) > 0 {
						m.values[m.focused] = m.values[m.focused][:len(m.values[m.focused])-1]
					}
				case "ctrl+a":
					// Select all - not implemented but don't add as text
				case "ctrl+u":
					// Clear line
					m.values[m.focused] = ""
				default:
					// Handle multi-character input (paste) and regular typing
					if msg.String() != "" {
						// Handle bracketed paste - remove the brackets if present
						text := msg.String()
						
						// Check for bracketed paste sequences
						if strings.HasPrefix(text, "[") && strings.HasSuffix(text, "]") && len(text) > 2 {
							// Remove brackets from pasted content
							text = text[1 : len(text)-1]
						}
						
						// Filter out non-printable characters except spaces and common URL/name chars
						cleaned := ""
						for _, r := range text {
							if (r >= 32 && r <= 126) || r == '\t' {
								if r == '\t' {
									// Convert tab to nothing in input
									continue
								}
								cleaned += string(r)
							}
						}
						if cleaned != "" {
							m.values[m.focused] += cleaned
						}
					}
				}
			}
		}
	case TargetCreateMsg:
		m.saving = false
		if msg.Error != nil {
			m.err = msg.Error
			m.saveResult = ""
		} else if msg.Success {
			m.saveResult = fmt.Sprintf("✓ Target created successfully: %s", msg.Output)
			m.err = nil
			// After successful creation, go back to targets view after a short delay
			return m, tea.Tick(2*time.Second, func(time.Time) tea.Msg {
				return SwitchViewMsg{View: ViewTargets}
			})
		} else {
			m.err = fmt.Errorf("Failed to create target: %s", msg.Output)
			m.saveResult = ""
		}
	}
	
	return m, nil
}

// canSubmit checks if the form can be submitted
func (m AddTargetViewModel) canSubmit() bool {
	for _, value := range m.values {
		if strings.TrimSpace(value) == "" {
			return false
		}
	}
	return true
}

// startSave starts the target creation process
func (m AddTargetViewModel) startSave() (AddTargetViewModel, tea.Cmd) {
	if !m.canSubmit() {
		return m, nil
	}
	
	// Prepare values
	name := strings.TrimSpace(m.values[0])
	url := strings.TrimSpace(m.values[1])
	team := strings.TrimSpace(m.values[2])
	
	// Default team to "main" if empty
	if team == "" {
		team = "main"
	}
	
	// Generate fly command
	m.flyCommand = fmt.Sprintf("fly -t %s login -c %s -n %s", name, url, team)
	m.saving = true
	m.err = nil
	m.saveResult = ""
	
	// Execute the fly command
	return m, func() tea.Msg {
		// First, check if the target already exists and is authenticated
		checkCmd := exec.Command("fly", "-t", name, "status")
		checkOutput, checkErr := checkCmd.CombinedOutput()
		
		if checkErr == nil && strings.Contains(string(checkOutput), "logged in successfully") {
			// Target already exists and is logged in
			return TargetCreateMsg{
				Success: true,
				Output:  fmt.Sprintf("Target '%s' already exists and is authenticated", name),
				Error:   nil,
				Command: fmt.Sprintf("fly -t %s status", name),
			}
		}
		
		// Perform interactive login using the same approach as the auth view
		args := []string{"login", "-c", url}
		if team != "" {
			args = append(args, "-n", team)
		}
		args = append([]string{"-t", name}, args...)
		
		// Execute interactively (this will open browser) - same as LoginInteractive in client.go
		cmd := exec.Command("fly", args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		
		err := cmd.Run()
		if err != nil {
			return TargetCreateMsg{
				Success: false,
				Output:  fmt.Sprintf("Failed to create target: %s", err.Error()),
				Error:   err,
				Command: fmt.Sprintf("fly -t %s login -c %s -n %s", name, url, team),
			}
		}
		
		// Login succeeded
		return TargetCreateMsg{
			Success: true,
			Output:  fmt.Sprintf("Target '%s' created successfully!", name),
			Error:   nil,
			Command: fmt.Sprintf("fly -t %s login -c %s -n %s", name, url, team),
		}
	}
}

// submit submits the form (old method - kept for compatibility)
func (m AddTargetViewModel) submit() tea.Cmd {
	return func() tea.Msg {
		// This is now handled by startSave
		return SwitchViewMsg{View: ViewTargets}
	}
}

// Reset resets the form
func (m *AddTargetViewModel) Reset() {
	m.values = []string{"", "", ""}
	m.focused = 0
	m.submitted = false
	m.err = nil
	m.saving = false
	m.flyCommand = ""
	m.saveResult = ""
}

// View renders the add target view
func (m AddTargetViewModel) View(width, height int) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		MarginBottom(2)
	
	labelStyle := lipgloss.NewStyle().
		Bold(true).
		MarginRight(2)
		
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Width(40)
		
	focusedInputStyle := inputStyle.Copy().
		BorderForeground(lipgloss.Color("205"))
	
	var content strings.Builder
	content.WriteString(titleStyle.Render("Add New Target"))
	content.WriteString("\n")
	
	for i, field := range m.fields {
		content.WriteString(labelStyle.Render(field + ":"))
		
		var inputBox string
		value := m.values[i]
		placeholder := ""
		
		// Add placeholders
		switch field {
		case "Name":
			if value == "" {
				placeholder = "e.g., production"
			}
		case "URL":
			if value == "" {
				placeholder = "e.g., https://ci.example.com"
			}
		case "Team":
			if value == "" {
				placeholder = "e.g., main (default: main)"
			}
		}
		
		displayValue := value
		if displayValue == "" && placeholder != "" {
			displayValue = placeholder
		}
		
		if i == m.focused && !m.saving {
			// Show cursor
			if value == "" && placeholder != "" {
				inputBox = focusedInputStyle.Render(placeholder + "█")
			} else {
				inputBox = focusedInputStyle.Render(value + "█")
			}
		} else {
			if value == "" && placeholder != "" {
				placeholderStyle := inputStyle.Copy().Foreground(lipgloss.Color("240"))
				inputBox = placeholderStyle.Render(placeholder)
			} else {
				inputBox = inputStyle.Render(value)
			}
		}
		
		content.WriteString(inputBox)
		content.WriteString("\n\n")
	}
	
	// Show fly command if saving or saved
	if m.saving || m.flyCommand != "" {
		commandStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("33")).
			Padding(1).
			MarginBottom(1)
		
		if m.saving {
			content.WriteString(commandStyle.Render("🔄 Executing: " + m.flyCommand))
		} else if m.flyCommand != "" {
			content.WriteString(commandStyle.Render("📝 Command executed: " + m.flyCommand))
		}
		content.WriteString("\n")
	}
	
	// Show save result
	if m.saveResult != "" {
		// Check if this is an interactive authentication message
		if strings.Contains(m.saveResult, "Interactive authentication required") {
			// Show interactive auth message
			authStyle := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("220")).
				Padding(1).
				MarginBottom(1).
				Foreground(lipgloss.Color("220"))
			
			content.WriteString(authStyle.Render("🔐 " + m.saveResult))
			content.WriteString("\n")
			
			helpStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Italic(true)
			content.WriteString(helpStyle.Render("Press Esc to return to targets view"))
		} else if strings.Contains(m.saveResult, "already exists") {
			// Show success message for existing target
			successStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("46")).
				Bold(true).
				MarginBottom(1)
			content.WriteString(successStyle.Render("✅ " + m.saveResult))
			content.WriteString("\n")
			content.WriteString("Returning to targets view...\n")
		} else {
			// Show regular result
			resultStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("46")).
				Bold(true).
				MarginBottom(1)
			content.WriteString(resultStyle.Render(m.saveResult))
			content.WriteString("\n")
			content.WriteString("Returning to targets view...\n")
		}
	}
	
	// Show help text
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true).
		MarginTop(1)
	
	var help string
	if m.saving {
		help = "Creating target... Please wait"
	} else if m.saveResult != "" && strings.Contains(m.saveResult, "Interactive authentication required") {
		help = "Press 'r' to retry checking authentication • 'c' to copy command • Esc: Return to targets"
	} else if m.saveResult != "" {
		help = "Enter: Return to targets • Esc: Return to targets"
	} else {
		help = "Tab/Shift+Tab: Navigate • Enter: Create Target • Ctrl+U: Clear field • Esc: Cancel"
	}
	content.WriteString(helpStyle.Render(help))
	
	if m.err != nil {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			MarginTop(1)
		content.WriteString("\n")
		content.WriteString(errorStyle.Render("Error: " + m.err.Error()))
	}
	
	return content.String()
}