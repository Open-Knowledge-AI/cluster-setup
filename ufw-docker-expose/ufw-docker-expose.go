package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

const (
	// Replace these with the subnets you want to use when the user
	// leaves the subnet field empty.
	defaultSubnets = "172.16.0.0/12;192.168.0.0/16"

	defaultProtocol = "tcp"
)

type field int

const (
	subnetField field = iota
	actionField
	protocolField
	portField
)

type model struct {
	subnets  textinput.Model
	action   textinput.Model
	protocol textinput.Model
	port     textinput.Model

	focused field
	done    bool
	err     error
	output  string
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	focusedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42"))
)

func newInput(placeholder, value string) textinput.Model {
	t := textinput.New()
	t.Placeholder = placeholder
	t.SetValue(value)
	t.CharLimit = 256
	t.Width = 50
	return t
}

func initialModel() model {
	m := model{
		subnets:  newInput(defaultSubnets, defaultSubnets),
		action:   newInput("add or delete", "add"),
		protocol: newInput("tcp or udp", defaultProtocol),
		port:     newInput("container port", ""),
	}

	m.subnets.Focus()
	return m
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.done {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "q", "esc", "ctrl+c", "enter":
				return m, tea.Quit
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "tab", "down":
			m.nextField()

		case "shift+tab", "up":
			m.previousField()

		case "enter":
			if m.focused == portField {
				return m.execute()
			}
			m.nextField()
		}
	}

	var cmd tea.Cmd

	switch m.focused {
	case subnetField:
		m.subnets, cmd = m.subnets.Update(msg)
	case actionField:
		m.action, cmd = m.action.Update(msg)
	case protocolField:
		m.protocol, cmd = m.protocol.Update(msg)
	case portField:
		m.port, cmd = m.port.Update(msg)
	}

	return m, cmd
}

func (m *model) nextField() {
	m.blurAll()

	switch m.focused {
	case subnetField:
		m.focused = actionField
	case actionField:
		m.focused = protocolField
	case protocolField:
		m.focused = portField
	case portField:
		m.focused = subnetField
	}

	m.focusCurrent()
}

func (m *model) previousField() {
	m.blurAll()

	switch m.focused {
	case subnetField:
		m.focused = portField
	case actionField:
		m.focused = subnetField
	case protocolField:
		m.focused = actionField
	case portField:
		m.focused = protocolField
	}

	m.focusCurrent()
}

func (m *model) blurAll() {
	m.subnets.Blur()
	m.action.Blur()
	m.protocol.Blur()
	m.port.Blur()
}

func (m *model) focusCurrent() {
	switch m.focused {
	case subnetField:
		m.subnets.Focus()
	case actionField:
		m.action.Focus()
	case protocolField:
		m.protocol.Focus()
	case portField:
		m.port.Focus()
	}
}

func (m model) View() string {
	if m.done {
		if m.err != nil {
			return titleStyle.Render("UFW Docker Expose") +
				"\n\n" +
				errorStyle.Render("Command failed:") +
				"\n\n" +
				m.err.Error() +
				"\n\n" +
				helpStyle.Render("Press Enter or q to exit.")
		}

		return titleStyle.Render("UFW Docker Expose") +
			"\n\n" +
			successStyle.Render("✓ Rules updated successfully.") +
			"\n\n" +
			m.output +
			"\n\n" +
			helpStyle.Render("Press Enter or q to exit.")
	}

	var b strings.Builder

	b.WriteString(titleStyle.Render("UFW Docker Expose"))
	b.WriteString("\n")

	b.WriteString(m.renderField("Subnets", m.subnets, subnetField))
	b.WriteString("\n")

	b.WriteString(m.renderField("Action", m.action, actionField))
	b.WriteString("\n")

	b.WriteString(m.renderField("Protocol", m.protocol, protocolField))
	b.WriteString("\n")

	b.WriteString(m.renderField("Port", m.port, portField))
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render(
		"Tab/↑↓: navigate • Enter: next/execute • Esc: quit",
	))

	return b.String()
}

func (m model) renderField(
	label string,
	input textinput.Model,
	f field,
) string {
	style := labelStyle

	if m.focused == f {
		style = focusedStyle
	}

	return style.Render(label+": ") + input.View()
}

func (m model) execute() (tea.Model, tea.Cmd) {
	subnets := strings.TrimSpace(m.subnets.Value())
	action := strings.ToLower(strings.TrimSpace(m.action.Value()))
	protocol := strings.ToLower(strings.TrimSpace(m.protocol.Value()))
	port := strings.TrimSpace(m.port.Value())

	if subnets == "" {
		subnets = defaultSubnets
	}

	if action != "add" && action != "delete" {
		m.err = fmt.Errorf("action must be 'add' or 'delete'")
		m.done = true
		return m, nil
	}

	if protocol != "tcp" && protocol != "udp" {
		m.err = fmt.Errorf("protocol must be 'tcp' or 'udp'")
		m.done = true
		return m, nil
	}

	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		m.err = fmt.Errorf("port must be a number between 1 and 65535")
		m.done = true
		return m, nil
	}

	parsedSubnets, err := parseSubnets(subnets)
	if err != nil {
		m.err = err
		m.done = true
		return m, nil
	}

	commands := make([]string, 0, len(parsedSubnets)*2+1)

	for _, subnet := range parsedSubnets {
		ufwAction := "allow"
		if action == "delete" {
			ufwAction = "delete allow"
		}

		args := []string{
			"ufw",
			"route",
		}

		args = append(args, strings.Fields(ufwAction)...)

		args = append(args,
			"proto", protocol,
			"from", subnet,
			"to", "any",
			"port", port,
		)

		if err := runCommand(args...); err != nil {
			m.err = fmt.Errorf("ufw command failed for %s: %w", subnet, err)
			m.done = true
			return m, nil
		}

		commands = append(commands, "sudo "+strings.Join(args, " "))
	}

	// Update ufw-docker's Docker subnet configuration.
	//
	// The value is passed as a single --docker-subnets argument.
	// Adjust this if your installed ufw-docker version expects a
	// different separator or syntax.
	dockerSubnets := strings.Join(parsedSubnets, ",")

	dockerArgs := []string{
		"ufw-docker",
		"install",
		"--docker-subnets",
		dockerSubnets,
	}

	if err := runCommand(dockerArgs...); err != nil {
		m.err = fmt.Errorf("ufw-docker command failed: %w", err)
		m.done = true
		return m, nil
	}

	commands = append(commands, "sudo "+strings.Join(dockerArgs, " "))

	reloadArgs := []string{
		"ufw",
		"reload",
	}

	if err := runCommand(reloadArgs...); err != nil {
		m.err = fmt.Errorf("ufw reload failed: %w", err)
		m.done = true
		return m, nil
	}

	commands = append(commands, "sudo "+strings.Join(reloadArgs, " "))

	m.output = strings.Join(commands, "\n")
	m.done = true

	return m, nil
}

func parseSubnets(input string) ([]string, error) {
	raw := strings.Split(input, ";")
	result := make([]string, 0, len(raw))

	for _, value := range raw {
		subnet := strings.TrimSpace(value)

		if subnet == "" {
			continue
		}

		ip, network, err := net.ParseCIDR(subnet)
		if err != nil {
			return nil, fmt.Errorf("invalid subnet %q: %w", subnet, err)
		}

		if ip.To4() == nil {
			return nil, fmt.Errorf("subnet %q is not IPv4", subnet)
		}

		ones, bits := network.Mask.Size()
		if bits != 32 {
			return nil, fmt.Errorf("subnet %q is not IPv4", subnet)
		}

		// Normalise the address so e.g. 192.168.1.10/24 becomes
		// 192.168.1.0/24.
		result = append(result, fmt.Sprintf("%s/%d", network.IP.String(), ones))
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("at least one IPv4 subnet is required")
	}

	return result, nil
}

func runCommand(args ...string) error {
	cmd := exec.Command("sudo", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		if len(output) > 0 {
			return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
		}
		return err
	}

	return nil
}

func main() {
	// The utility itself can be run as a normal user because every
	// system command is explicitly executed through sudo.
	//
	// If you want to require the entire application to be started
	// with sudo, uncomment the following check.
	/*
		if os.Geteuid() != 0 {
			fmt.Fprintln(os.Stderr, "Please run with sudo.")
			os.Exit(1)
		}
	*/

	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
