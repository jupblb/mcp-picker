package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/adrg/xdg"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

type ServerConfig struct {
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	URL     string            `json:"url,omitempty"`
}

type AgentType string

const (
	AgentAmp    AgentType = "amp"
	AgentClaude AgentType = "claude"
)

func parseAgentType(s string) (AgentType, error) {
	switch s {
	case "amp", "":
		return AgentAmp, nil
	case "claude":
		return AgentClaude, nil
	default:
		return "", fmt.Errorf("unknown agent: %s (valid: amp, claude)", s)
	}
}

var (
	gruvboxOrange = lipgloss.AdaptiveColor{Dark: "#fe8019", Light: "#af3a03"}
	gruvboxGreen  = lipgloss.AdaptiveColor{Dark: "#b8bb26", Light: "#79740e"}
	gruvboxPurple = lipgloss.AdaptiveColor{Dark: "#d3869b", Light: "#8f3f71"}
	gruvboxGray   = lipgloss.AdaptiveColor{Dark: "#928374", Light: "#7c6f64"}

	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(gruvboxOrange)
	selectedStyle = lipgloss.NewStyle().Foreground(gruvboxGreen)
	cursorStyle   = lipgloss.NewStyle().Foreground(gruvboxPurple)
	normalStyle   = lipgloss.NewStyle()
)

type serverItem string

func (i serverItem) FilterValue() string { return string(i) }
func (i serverItem) Title() string       { return string(i) }
func (i serverItem) Description() string { return "" }

type serverItemDelegate struct {
	selected map[string]bool
}

func (d serverItemDelegate) Height() int  { return 1 }
func (d serverItemDelegate) Spacing() int { return 0 }

func (d serverItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok || keyMsg.String() != " " {
		return nil
	}

	item, ok := m.SelectedItem().(serverItem)
	if !ok {
		return nil
	}

	d.selected[string(item)] = !d.selected[string(item)]
	return nil
}

func (d serverItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(serverItem)
	if !ok {
		return
	}

	cursor := "  "
	if index == m.Index() {
		cursor = cursorStyle.Render("> ")
	}

	checkbox := "[ ]"
	style := normalStyle
	if d.selected[string(item)] {
		checkbox = "[x]"
		style = selectedStyle
	}

	fmt.Fprint(w, cursor+style.Render(checkbox+" "+string(item)))
}

type pickerModel struct {
	list      list.Model
	selected  map[string]bool
	confirmed bool
}

func (m pickerModel) Init() tea.Cmd { return nil }

func (m pickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			if m.list.FilterState() == list.Unfiltered && m.list.FilterValue() == "" {
				return m, tea.Quit
			}
		case "enter":
			if m.list.FilterState() == list.Filtering {
				break
			}
			m.confirmed = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m pickerModel) View() string {
	return m.list.View()
}

func loadServerConfigs(path string) (map[string]ServerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	configs := make(map[string]ServerConfig)
	if err := json.Unmarshal(data, &configs); err != nil {
		return nil, err
	}
	return configs, nil
}

func writeSelectedServerConfigs(
	selected map[string]bool, configs map[string]ServerConfig, agent AgentType,
) (string, error) {
	result := make(map[string]ServerConfig, len(selected))
	for name, isSelected := range selected {
		if !isSelected {
			continue
		}
		cfg, ok := configs[name]
		if !ok {
			continue
		}
		result[name] = cfg
	}

	var output any
	switch agent {
	case AgentClaude:
		output = map[string]any{"mcpServers": result}
	default:
		output = result
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", err
	}

	tmpFile, err := os.CreateTemp("", "mcp-config-*.json")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	if _, err := tmpFile.Write(data); err != nil {
		return "", err
	}
	return tmpFile.Name(), nil
}

func main() {
	agentStr := flag.String("agent", "amp", "agent type: amp, claude")
	flag.StringVar(agentStr, "a", "amp", "agent type: amp, claude (short)")
	flag.Parse()

	agent, err := parseAgentType(*agentStr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	lipgloss.SetColorProfile(termenv.NewOutput(os.Stderr).Profile)

	configs, err := loadServerConfigs(
		filepath.Join(xdg.ConfigHome, "mcp-picker", "servers.json"),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error loading servers.json:", err)
		os.Exit(1)
	}

	servers := make([]string, 0, len(configs))
	for name := range configs {
		servers = append(servers, name)
	}
	sort.Strings(servers)

	items := make([]list.Item, len(servers))
	for i, name := range servers {
		items[i] = serverItem(name)
	}

	selected := make(map[string]bool)
	delegate := serverItemDelegate{selected: selected}

	l := list.New(items, delegate, 40, 12)
	l.Title = "Select MCP Servers"
	l.SetShowStatusBar(false)
	l.SetShowHelp(true)
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("space"), key.WithHelp("space", "toggle")),
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
		}
	}
	l.Styles.Title = titleStyle
	l.Styles.HelpStyle = lipgloss.NewStyle().Foreground(gruvboxGray)
	l.SetFilteringEnabled(true)

	result, err := tea.NewProgram(pickerModel{
		list:     l,
		selected: selected,
	}, tea.WithOutput(os.Stderr)).Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error running TUI:", err)
		os.Exit(1)
	}

	finalModel := result.(pickerModel)
	if !finalModel.confirmed {
		os.Exit(1)
	}
	tmpPath, err := writeSelectedServerConfigs(
		finalModel.selected, configs, agent,
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error writing config:", err)
		os.Exit(1)
	}
	fmt.Println(tmpPath)
}
