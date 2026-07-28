// filename: create-user.go
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Group definitions
var availableGroups = []struct {
	name        string
	description string
	required    bool
}{
	{"docker", "Run Docker containers", false},
	{"podman", "Run Podman containers", false},
	{"kubectl", "Access Kubernetes", false},
	{"devtools", "Access development tools (conda, mamba, uv, node, go)", false},
	{"ml-users", "Access ML models and datasets", false},
	{"shared-users", "Access shared directories", true},
	{"sudo", "Administrative privileges", false},
}

// ---------------------------------------------------------------------------
// Custom multi-select "checkbox list" component.
// bubbles has no checkbox package, so this replaces that dependency.
// ---------------------------------------------------------------------------

type groupOption struct {
	name        string
	description string
	checked     bool
	required    bool
}

type groupSelector struct {
	options []groupOption
	cursor  int
}

func newGroupSelector() groupSelector {
	opts := make([]groupOption, len(availableGroups))
	for i, g := range availableGroups {
		opts[i] = groupOption{
			name:        g.name,
			description: g.description,
			checked:     g.required,
			required:    g.required,
		}
	}
	return groupSelector{options: opts}
}

func (g groupSelector) Update(msg tea.Msg) (groupSelector, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "up", "k":
			if g.cursor > 0 {
				g.cursor--
			}
		case "down", "j":
			if g.cursor < len(g.options)-1 {
				g.cursor++
			}
		case " ":
			if !g.options[g.cursor].required {
				g.options[g.cursor].checked = !g.options[g.cursor].checked
			}
		}
	}
	return g, nil
}

func (g groupSelector) View() string {
	var b strings.Builder
	for i, opt := range g.options {
		cursor := "  "
		if i == g.cursor {
			cursor = "> "
		}
		check := "[ ]"
		if opt.checked {
			check = "[x]"
		}
		lock := ""
		if opt.required {
			lock = " (required)"
		}
		fmt.Fprintf(&b, "%s%s %s — %s%s\n", cursor, check, opt.name, opt.description, lock)
	}
	b.WriteString("\n(↑/↓ or j/k to move, space to toggle, enter to continue)")
	return b.String()
}

func (g groupSelector) SelectedNames() []string {
	var names []string
	for _, opt := range g.options {
		if opt.checked {
			names = append(names, opt.name)
		}
	}
	return names
}

// ---------------------------------------------------------------------------
// Model represents the application state
// ---------------------------------------------------------------------------

type Model struct {
	step          int
	username      textinput.Model
	fullname      textinput.Model
	password      textinput.Model
	confirmPass   textinput.Model
	groups        groupSelector
	quotaSize     textinput.Model
	creating      bool
	done          bool
	err           error
	successMsg    string
	showPasswords bool
}

type statusMsg string
type errMsg error

// Initial model setup
func initialModel() Model {
	username := textinput.New()
	username.Placeholder = "username"
	username.Focus()
	username.CharLimit = 32
	username.Width = 30

	fullname := textinput.New()
	fullname.Placeholder = "Full Name"
	fullname.Width = 30

	password := textinput.New()
	password.Placeholder = "••••••••"
	password.EchoMode = textinput.EchoPassword
	password.EchoCharacter = '•'
	password.Width = 30

	confirmPass := textinput.New()
	confirmPass.Placeholder = "••••••••"
	confirmPass.EchoMode = textinput.EchoPassword
	confirmPass.EchoCharacter = '•'
	confirmPass.Width = 30

	quotaSize := textinput.New()
	quotaSize.Placeholder = "15"
	quotaSize.SetValue("15")
	quotaSize.Width = 10

	return Model{
		step:          0,
		username:      username,
		fullname:      fullname,
		password:      password,
		confirmPass:   confirmPass,
		groups:        newGroupSelector(),
		quotaSize:     quotaSize,
		creating:      false,
		done:          false,
		showPasswords: false,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles all events
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if m.creating {
		// Wait for creation to complete
		switch msg := msg.(type) {
		case statusMsg:
			m.successMsg = string(msg)
			m.creating = false
			m.done = true
			return m, tea.Quit
		case errMsg:
			m.err = msg
			m.creating = false
			return m, nil
		}
		return m, nil
	}

	if m.done {
		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			if m.step == 0 {
				if m.username.Value() != "" {
					m.step = 1
					m.fullname.Focus()
					return m, textinput.Blink
				}
				return m, nil
			}
			if m.step == 1 {
				m.step = 2
				m.password.Focus()
				return m, textinput.Blink
			}
			if m.step == 2 {
				m.step = 3
				m.confirmPass.Focus()
				return m, textinput.Blink
			}
			if m.step == 3 {
				if m.password.Value() != m.confirmPass.Value() {
					m.err = fmt.Errorf("passwords do not match")
					return m, nil
				}
				m.step = 4
				return m, nil
			}
			if m.step == 4 {
				m.step = 5
				m.quotaSize.Focus()
				return m, textinput.Blink
			}
			if m.step == 5 {
				// Validate quota
				quota, err := strconv.Atoi(m.quotaSize.Value())
				if err != nil || quota < 1 {
					m.err = fmt.Errorf("invalid quota size: must be a positive number")
					return m, nil
				}
				m.step = 6
				return m, nil
			}
			if m.step == 6 {
				// Create user
				m.creating = true
				return m, m.createUser()
			}
		}
	}

	// Handle updates for different steps
	switch m.step {
	case 0:
		var cmd tea.Cmd
		m.username, cmd = m.username.Update(msg)
		cmds = append(cmds, cmd)
	case 1:
		var cmd tea.Cmd
		m.fullname, cmd = m.fullname.Update(msg)
		cmds = append(cmds, cmd)
	case 2:
		var cmd tea.Cmd
		m.password, cmd = m.password.Update(msg)
		cmds = append(cmds, cmd)
	case 3:
		var cmd tea.Cmd
		m.confirmPass, cmd = m.confirmPass.Update(msg)
		cmds = append(cmds, cmd)
	case 4:
		var cmd tea.Cmd
		m.groups, cmd = m.groups.Update(msg)
		cmds = append(cmds, cmd)
		// Handle enter key in groups
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "enter" {
			m.step = 5
			m.quotaSize.Focus()
			return m, textinput.Blink
		}
	case 5:
		var cmd tea.Cmd
		m.quotaSize, cmd = m.quotaSize.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Handle errors
	if em, ok := msg.(errMsg); ok {
		m.err = em
	}

	return m, tea.Batch(cmds...)
}


// Create user asynchronously
func (m Model) createUser() tea.Cmd {
	return func() tea.Msg {
		username := m.username.Value()
		fullname := m.fullname.Value()
		password := m.password.Value()
		quotaSize := m.quotaSize.Value()

		selectedGroups := m.groups.SelectedNames()

		// Create user
		if err := createSystemUser(username, fullname, password, selectedGroups, quotaSize); err != nil {
			return errMsg(fmt.Errorf("failed to create user: %v", err))
		}

		// Create README
		if err := createReadme(username, selectedGroups); err != nil {
			return errMsg(fmt.Errorf("user created but failed to create README: %v", err))
		}

		return statusMsg(fmt.Sprintf("✓ User '%s' created successfully!", username))
	}
}

// System functions
func createSystemUser(username, fullname, password string, groups []string, quotaSize string) error {
	// Check if user exists
	if _, err := exec.Command("id", username).Output(); err == nil {
		return fmt.Errorf("user '%s' already exists", username)
	}

	// Create user with home directory
	cmd := exec.Command("useradd", "-m", "-s", "/bin/bash", "-c", fullname, username)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}

	// Set password
	cmd = exec.Command("chpasswd")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to set password: %v", err)
	}
	go func() {
		defer stdin.Close()
		fmt.Fprintf(stdin, "%s:%s", username, password)
	}()
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set password: %v", err)
	}

	// Add user to groups
	for _, group := range groups {
		if group == "sudo" {
			// Special handling for sudo group
			if err := exec.Command("usermod", "-aG", "sudo", username).Run(); err != nil {
				return fmt.Errorf("failed to add user to sudo group: %v", err)
			}
		} else if group == "docker" {
			// Ensure docker group exists
			if err := exec.Command("groupadd", "-f", "docker").Run(); err != nil {
				return fmt.Errorf("failed to ensure docker group exists: %v", err)
			}
			if err := exec.Command("usermod", "-aG", "docker", username).Run(); err != nil {
				return fmt.Errorf("failed to add user to docker group: %v", err)
			}
		} else {
			// Create group if it doesn't exist
			if err := exec.Command("groupadd", "-f", group).Run(); err != nil {
				return fmt.Errorf("failed to ensure group %s exists: %v", group, err)
			}
			if err := exec.Command("usermod", "-aG", group, username).Run(); err != nil {
				return fmt.Errorf("failed to add user to group %s: %v", group, err)
			}
		}
	}

	// Set quota
	quota, err := strconv.Atoi(quotaSize)
	if err == nil && quota > 0 {
		if err := setQuota(username, quota); err != nil {
			// Don't fail, just warn
			fmt.Printf("Warning: Failed to set quota: %v\n", err)
		}
	}

	// Set up HuggingFace cache directory
	homeDir := fmt.Sprintf("/home/%s", username)
	hfCacheDir := filepath.Join(homeDir, ".cache", "huggingface")
	if err := os.MkdirAll(hfCacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create HuggingFace cache directory: %v", err)
	}

	// Copy HuggingFace environment file
	if err := copyFile("/etc/profile.d/huggingface.sh", filepath.Join(hfCacheDir, ".env")); err != nil {
		// Don't fail if file doesn't exist
		if !os.IsNotExist(err) {
			fmt.Printf("Warning: Failed to copy HuggingFace config: %v\n", err)
		}
	}

	// Set ownership
	if err := exec.Command("chown", "-R", fmt.Sprintf("%s:%s", username, username), homeDir).Run(); err != nil {
		return fmt.Errorf("failed to set ownership: %v", err)
	}

	return nil
}

func setQuota(username string, quotaGB int) error {
	// Convert GB to blocks (1 block = 1KB)
	blocks := quotaGB * 1024 * 1024

	// Create temporary quota file
	tempFile, err := os.CreateTemp("/tmp", "quota-")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Write quota settings
	_, err = tempFile.WriteString(fmt.Sprintf("%s soft=%d hard=%d\n", username, blocks, blocks))
	if err != nil {
		return fmt.Errorf("failed to write quota settings: %v", err)
	}
	tempFile.Close()

	// Apply quota
	cmd := exec.Command("edquota", "-p", "root", username)
	if err := cmd.Run(); err != nil {
		// Try setting quota directly
		cmd = exec.Command("setquota", username, fmt.Sprintf("%d", blocks), fmt.Sprintf("%d", blocks), "0", "0", "/home")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to set quota: %v", err)
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, 0644)
}

func createReadme(username string, groups []string) error {
	homeDir := fmt.Sprintf("/home/%s", username)
	readmePath := filepath.Join(homeDir, "README.md")

	// Determine if user has sudo
	hasSudo := false
	for _, g := range groups {
		if g == "sudo" {
			hasSudo = true
			break
		}
	}

	content := fmt.Sprintf(`# Welcome to the Cluster, %s!

## System Overview
This is a shared computing cluster with centralized tools and data storage. Here's everything you need to know:

### Shared Directories
- **/shared/projects** - Project workspace for collaborative development
- **/shared/models** - Shared ML models and weights
- **/shared/huggingface** - HuggingFace cache directory
- **/shared/datasets** - Shared datasets
- **/shared/tools** - Shared development tools

### Available Tools
All tools are available system-wide:

#### Container Runtimes
- **Docker** - Container management
- **Podman** - Rootless container runtime
- **kubectl** - Kubernetes cluster management

#### Development Tools
- **Python** - Miniconda3 installation at /shared/tools/miniconda3
- **Mamba** - Fast conda package manager
- **UV** - Python package installer
- **Node.js** - JavaScript runtime
- **Go** - Go programming language

### HuggingFace Configuration
The HuggingFace cache is automatically configured to use the shared directory:
- HF_HOME=/shared/huggingface
- HF_DATASETS_CACHE=/shared/huggingface/datasets
- TRANSFORMERS_CACHE=/shared/huggingface/cache
- HUGGINGFACE_HUB_CACHE=/shared/huggingface/cache

### Your Groups
You have been added to the following groups:
%s

### Storage Quotas
- **Home Directory**: %sGB maximum
- **Shared Directories**: No quota (but please be considerate)

### Best Practices
1. Store large files in /shared/ directories
2. Keep your home directory clean (under 15GB)
3. Use the HuggingFace cache for model files
4. When using containers, mount shared directories as volumes
5. Be mindful of resource usage

### Quick Commands
` + "```bash" + `
# Check your quota
quota -s

# View shared directories
ls -la /shared/

# Use mamba
mamba create -n myenv python=3.11
mamba activate myenv

# Check Docker status
docker info

# Check Podman status
podman info
` + "```" + `

### Getting Help
- For tool issues: Check the tool's documentation
- For cluster issues: Contact the cluster administrator
- For quota issues: Use the shared directories for large files

### Important Notes
- %s
- Log out and log back in for group changes to take effect
- Use shared directories for collaborative work
- Report any issues to the administrator

Welcome aboard and happy computing!
`,
		username,
		formatGroups(groups),
		getQuotaStr(username),
		getSudoWarning(hasSudo))

	return os.WriteFile(readmePath, []byte(content), 0644)
}

func formatGroups(groups []string) string {
	if len(groups) == 0 {
		return "  - No special groups (basic user only)"
	}
	var result strings.Builder
	for _, g := range groups {
		result.WriteString(fmt.Sprintf("  - %s\n", g))
	}
	return result.String()
}

func getQuotaStr(username string) string {
	// Try to get actual quota
	cmd := exec.Command("quota", "-s", username)
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "home") || strings.Contains(line, "/home") {
				fields := strings.Fields(line)
				if len(fields) >= 3 {
					return fields[2]
				}
			}
		}
	}
	return "15" // Default
}

func getSudoWarning(hasSudo bool) string {
	if hasSudo {
		return "⚠️  You have sudo access - use it responsibly"
	}
	return "You do not have sudo access - contact admin if you need it"
}

// View renders the UI
func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress any key to exit...", m.err)
	}

	if m.creating {
		return "\n Creating user... Please wait.\n\n ⏳ Processing..."
	}

	if m.done {
		return fmt.Sprintf("\n %s\n\n Press any key to exit...", m.successMsg)
	}

	var content string
	style := lipgloss.NewStyle().Padding(1).Width(80)

	switch m.step {
	case 0:
		content = style.Render(fmt.Sprintf(
			"Create New User (Step 1/6)\n\n"+
				"Enter username:\n%s\n\n"+
				"Press Enter to continue, Ctrl+C to quit",
			m.username.View(),
		))
	case 1:
		content = style.Render(fmt.Sprintf(
			"Create New User (Step 2/6)\n\n"+
				"Enter full name:\n%s\n\n"+
				"Press Enter to continue, Ctrl+C to quit",
			m.fullname.View(),
		))
	case 2:
		content = style.Render(fmt.Sprintf(
			"Create New User (Step 3/6)\n\n"+
				"Enter password:\n%s\n\n"+
				"Press Enter to continue, Ctrl+C to quit",
			m.password.View(),
		))
	case 3:
		content = style.Render(fmt.Sprintf(
			"Create New User (Step 4/6)\n\n"+
				"Confirm password:\n%s\n\n"+
				"Press Enter to continue, Ctrl+C to quit",
			m.confirmPass.View(),
		))
	case 4:
		content = style.Render(fmt.Sprintf(
			"Create New User (Step 5/6)\n\n"+
				"Select groups:\n%s\n\n"+
				"Press Enter to continue, Ctrl+C to quit",
			m.groups.View(),
		))
	case 5:
		content = style.Render(fmt.Sprintf(
			"Create New User (Step 6/6)\n\n"+
				"Set quota size (GB):\n%s\n\n"+
				"Press Enter to create user, Ctrl+C to quit",
			m.quotaSize.View(),
		))
	case 6:
		// Final confirmation before creation
		selected := m.groups.SelectedNames()
		content = style.Render(fmt.Sprintf(
			"Create New User - Confirmation\n\n"+
				"Username: %s\n"+
				"Full Name: %s\n"+
				"Groups: %v\n"+
				"Quota: %s GB\n\n"+
				"Press Enter to confirm and create, Ctrl+C to quit",
			m.username.Value(),
			m.fullname.Value(),
			selected,
			m.quotaSize.Value(),
		))
	}

	return "\n" + content + "\n"
}

func main() {
	// Check if running as root
	if os.Geteuid() != 0 {
		fmt.Println("Error: This program must be run as root (sudo)")
		os.Exit(1)
	}

	// Create the program
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
