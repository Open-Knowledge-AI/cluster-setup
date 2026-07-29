package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	// These are local Docker subnets, NOT the external IP subnets.
	// They are passed to ufw-docker as a space-separated list.
	defaultDockerSubnets = "172.16.0.0/12 192.168.0.0/16"

	defaultProtocol = "tcp"
)

type field int

const (
	externalSubnetsField field = iota
	dockerSubnetsField
	actionField
	protocolField
	portField
)

type model struct {
	externalSubnets textinput.Model
	dockerSubnets   textinput.Model
	action          textinput.Model
	protocol        textinput.Model
	port            textinput.Model

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
	t.CharLimit = 512
	t.Width = 60

	return t
}

func initialModel() model {
	m := model{
		// Intentionally empty: these are the external/source IPs.
		externalSubnets: newInput(
			"203.0.113.10/32;198.51.100.0/24",
			"",
		),

		// These are local Docker networks and are passed to ufw-docker.
		dockerSubnets: newInput(
			"172.16.0.0/12 192.168.0.0/16",
			defaultDockerSubnets,
		),

		action:   newInput("add or delete", "add"),
		protocol: newInput("tcp, udp, or all", defaultProtocol),
		port:     newInput("container port", ""),
	}

	m.externalSubnets.Focus()

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
	case externalSubnetsField:
		m.externalSubnets, cmd = m.externalSubnets.Update(msg)

	case dockerSubnetsField:
		m.dockerSubnets, cmd = m.dockerSubnets.Update(msg)

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
	case externalSubnetsField:
		m.focused = dockerSubnetsField

	case dockerSubnetsField:
		m.focused = actionField

	case actionField:
		m.focused = protocolField

	case protocolField:
		m.focused = portField

	case portField:
		m.focused = externalSubnetsField
	}

	m.focusCurrent()
}

func (m *model) previousField() {
	m.blurAll()

	switch m.focused {
	case externalSubnetsField:
		m.focused = portField

	case dockerSubnetsField:
		m.focused = externalSubnetsField

	case actionField:
		m.focused = dockerSubnetsField

	case protocolField:
		m.focused = actionField

	case portField:
		m.focused = protocolField
	}

	m.focusCurrent()
}

func (m *model) blurAll() {
	m.externalSubnets.Blur()
	m.dockerSubnets.Blur()
	m.action.Blur()
	m.protocol.Blur()
	m.port.Blur()
}

func (m *model) focusCurrent() {
	switch m.focused {
	case externalSubnetsField:
		m.externalSubnets.Focus()

	case dockerSubnetsField:
		m.dockerSubnets.Focus()

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
	b.WriteString("\n\n")

	b.WriteString(m.renderField(
		"External IP subnets",
		m.externalSubnets,
		externalSubnetsField,
	))
	b.WriteString("\n")

	b.WriteString(m.renderField(
		"Docker subnets",
		m.dockerSubnets,
		dockerSubnetsField,
	))
	b.WriteString("\n")

	b.WriteString(m.renderField(
		"Action",
		m.action,
		actionField,
	))
	b.WriteString("\n")

	b.WriteString(m.renderField(
		"Protocol",
		m.protocol,
		protocolField,
	))
	b.WriteString("\n")

	b.WriteString(m.renderField(
		"Port",
		m.port,
		portField,
	))
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render(
		"External subnets: semicolon-separated • Docker subnets: space-separated\n" +
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
	externalInput := strings.TrimSpace(m.externalSubnets.Value())
	dockerSubnets := strings.TrimSpace(m.dockerSubnets.Value())
	action := strings.ToLower(strings.TrimSpace(m.action.Value()))
	protocol := strings.ToLower(strings.TrimSpace(m.protocol.Value()))
	port := strings.TrimSpace(m.port.Value())

	if externalInput == "" {
		m.err = fmt.Errorf("at least one external IP subnet is required")
		m.done = true
		return m, nil
	}

	if dockerSubnets == "" {
		m.err = fmt.Errorf("at least one Docker subnet is required")
		m.done = true
		return m, nil
	}

	if action != "add" && action != "delete" {
		m.err = fmt.Errorf("action must be 'add' or 'delete'")
		m.done = true
		return m, nil
	}

	if protocol != "tcp" && protocol != "udp" && protocol != "all" {
		m.err = fmt.Errorf("protocol must be 'tcp', 'udp', or 'all'")
		m.done = true
		return m, nil
	}

	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		m.err = fmt.Errorf("port must be a number between 1 and 65535")
		m.done = true
		return m, nil
	}

	externalSubnets, err := parseExternalSubnets(externalInput)
	if err != nil {
		m.err = err
		m.done = true
		return m, nil
	}

	var commands []string

	// Apply the UFW rule separately for every external/source subnet.
	for _, subnet := range externalSubnets {
		args := []string{
			"ufw",
			"route",
		}

		if action == "delete" {
			args = append(args, "delete")
		}

		args = append(args, "allow")

		// "all" means no protocol restriction, so do not include
		// "proto all" in the UFW command.
		if protocol != "all" {
			args = append(args,
				"proto", protocol,
			)
		}

		args = append(args,
			"from", subnet,
			"to", "any",
			"port", port,
		)

		if err := runCommand(args...); err != nil {
			m.err = fmt.Errorf(
				"ufw command failed for external subnet %s: %w",
				subnet,
				err,
			)
			m.done = true
			return m, nil
		}

		commands = append(
			commands,
			"sudo "+strings.Join(args, " "),
		)
	}

	// Docker subnets are deliberately independent from the external
	// IP subnets above. They are passed to ufw-docker as ONE argument
	// containing a space-separated list.
	dockerArgs := []string{
		"ufw-docker",
		"install",
		"--docker-subnets",
		dockerSubnets,
	}

	if err := runCommand(dockerArgs...); err != nil {
		m.err = fmt.Errorf(
			"ufw-docker command failed: %w",
			err,
		)
		m.done = true
		return m, nil
	}

	commands = append(
		commands,
		"sudo "+strings.Join(dockerArgs, " "),
	)

	// Finally reload UFW.
	reloadArgs := []string{
		"ufw",
		"reload",
	}

	if err := runCommand(reloadArgs...); err != nil {
		m.err = fmt.Errorf(
			"ufw reload failed: %w",
			err,
		)
		m.done = true
		return m, nil
	}

	commands = append(
		commands,
		"sudo "+strings.Join(reloadArgs, " "),
	)

	m.output = strings.Join(commands, "\n")
	m.done = true

	return m, nil
}

func parseExternalSubnets(input string) ([]string, error) {
	raw := strings.Split(input, ";")
	result := make([]string, 0, len(raw))

	for _, value := range raw {
		subnet := strings.TrimSpace(value)

		if subnet == "" {
			continue
		}

		ip, network, err := net.ParseCIDR(subnet)
		if err != nil {
			return nil, fmt.Errorf(
				"invalid external subnet %q: %w",
				subnet,
				err,
			)
		}

		if ip.To4() == nil {
			return nil, fmt.Errorf(
				"external subnet %q is not IPv4",
				subnet,
			)
		}

		ones, bits := network.Mask.Size()
		if bits != 32 {
			return nil, fmt.Errorf(
				"external subnet %q is not IPv4",
				subnet,
			)
		}

		// Normalise the network address.
		result = append(
			result,
			fmt.Sprintf("%s/%d", network.IP.String(), ones),
		)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf(
			"at least one external IPv4 subnet is required",
		)
	}

	return result, nil
}

func runCommand(args ...string) error {
	cmd := exec.Command(args[0], args[1:]...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		if len(output) > 0 {
			return fmt.Errorf(
				"%w: %s",
				err,
				strings.TrimSpace(string(output)),
			)
		}

		return err
	}

	return nil
}

func main() {
	// The program must be started with sudo or directly as root.
	// Exit immediately before starting Bubble Tea otherwise.
	if os.Geteuid() != 0 {
		fmt.Fprintln(
			os.Stderr,
			"This program must be run as root. Try: sudo ufw-docker-expose",
		)
		os.Exit(1)
	}

	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
