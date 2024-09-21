package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"example.com/clock"
	"example.com/input"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	boldText     = lipgloss.NewStyle().Bold(true)

	focusedButton = focusedStyle.Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
	cols          = []table.Column{
		{Title: "Time", Width: 8},
		{Title: "Path", Width: 20},
		{Title: "Done", Width: 4},
	}
)

type model struct {
	triggers   *Triggers
	keys       keyMap
	help       help.Model
	quitting   bool
	focusIndex int
	inputs     []input.Model
	table      table.Model

	cursorMode cursor.Mode
	clock      clock.Model
	width      int

	host             string
	port             int
	displaySaved     int
	configChanged    bool
	showCloseConfirm bool
}

type TickMsg time.Time

func tick() tea.Cmd {
	return tea.Every(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.clock.Init(), tick())
}

func (m *model) save() {
	config := Config{
		Host:   m.host,
		Port:   m.port,
		Points: m.triggers.GetTriggers(),
	}
	err := SaveConfig(config)
	if err != nil {
		LOG(err.Error())
	} else {
		LOG("success")
		m.configChanged = false
	}
}
func initialModel(config Config) model {
	m := model{
		triggers:   NewTriggers(config.Host, config.Port),
		keys:       keys,
		help:       help.New(),
		cursorMode: cursor.CursorBlink,
		focusIndex: -1,
		inputs: []input.Model{
			input.New("Host", focusedStyle, blurredStyle),
			input.New("Port", focusedStyle, blurredStyle),
			input.New("Path", focusedStyle, blurredStyle),
			input.New("Time", focusedStyle, blurredStyle),
		},
		table: table.New(table.WithColumns(cols), table.WithHeight(7)),
		clock: clock.New(),
		host:  config.Host,
		port:  config.Port,
	}
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	m.table.SetStyles(s)
	m.inputs[0].SetValue(config.Host)
	m.inputs[1].SetValue(strconv.Itoa(config.Port))
	m.inputs[1].SetAllowedChars("0123456789")
	m.inputs[3].SetAllowedChars("0123456789:")
	for _, point := range config.Points {
		m.triggers.AddPoint(TriggerFromTime(point.Path, point.Time))
	}
	m.table.SetRows(m.triggers.ToRows())
	return m
}

func (m model) focusInput(index int) (model, tea.Cmd) {
	m.focusIndex = index
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		if i == index {
			m.inputs[i], cmds[i] = m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
	return m, tea.Batch(cmds...)
}
func (m model) updateInputs(msg tea.Msg) (model, tea.Cmd) {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return m, tea.Batch(cmds...)
}
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	m.clock, cmd = m.clock.Update(msg)
	cmds = append(cmds, cmd)
	m, cmd = m.updateInputs(msg)
	cmds = append(cmds, cmd)
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)
	switch msg := msg.(type) {
	case TickMsg:
		if m.displaySaved > 0 {
			m.displaySaved -= 1
		}
		m.table.SetRows(m.triggers.ToRows())
		return m, tick()
	case tea.WindowSizeMsg:
		// If we set a width on the help menu it can gracefully truncate
		// its view as needed.
		m.help.Width = msg.Width
		m.width = msg.Width
		cols[1].Width = msg.Width - 18
		m.table.SetHeight(msg.Height - 16)
		m.table, cmd = m.table.Update(msg)
		cmds = append(cmds, cmd)

	case tea.KeyMsg:
		if m.focusIndex >= 0 {
			switch {
			case key.Matches(msg, m.keys.Enter):
				switch m.focusIndex {
				case 0:
					m.host = m.inputs[0].Value()
					m.triggers.Host = m.host
					m.configChanged = true
					m, cmd = m.focusInput(-1)
					cmds = append(cmds, cmd)
				case 1:
					p, err := strconv.Atoi(m.inputs[1].Value())
					if err == nil {
						m.port = p
						m.triggers.Port = m.port
						m.configChanged = true
						m, cmd = m.focusInput(-1)
						cmds = append(cmds, cmd)
					}
				case 2:
					m, cmd = m.focusInput(3)
					cmds = append(cmds, cmd)
				case 3:
					tp := TriggerFromString(m.inputs[2].Value(), m.inputs[3].Value())
					m.triggers.AddPoint(tp)
					m.configChanged = true
					m, cmd = m.focusInput(-1)
					cmds = append(cmds, cmd)
					m.table.SetRows(m.triggers.ToRows())
				}
			case key.Matches(msg, m.keys.Esc):
				switch m.focusIndex {
				case 0:
					m.inputs[0].SetValue(m.host)
				case 1:
					m.inputs[1].SetValue(strconv.Itoa(m.port))
				}
				m, cmd = m.focusInput(-1)
				cmds = append(cmds, cmd)
			}
			return m, tea.Batch(cmds...)
		}
		switch {
		case key.Matches(msg, m.keys.Enter):
			if m.showCloseConfirm {
				return m, tea.Quit
			}
		case key.Matches(msg, m.keys.Esc):
			if m.showCloseConfirm {
				m.showCloseConfirm = false
				return m, tea.Batch(cmds...)
			}
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, m.keys.Quit):
			if m.configChanged {
				m.showCloseConfirm = true
				return m, nil
			}
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, m.keys.Host):
			m, cmd = m.focusInput(0)
			cmds = append(cmds, cmd)
		case key.Matches(msg, m.keys.Port):
			m, cmd = m.focusInput(1)
			cmds = append(cmds, cmd)
		case key.Matches(msg, m.keys.Trigger):
			m, cmd = m.focusInput(2)
			cmds = append(cmds, cmd)
		case key.Matches(msg, m.keys.Down):
			m.table.MoveDown(1)
		case key.Matches(msg, m.keys.Up):
			m.table.MoveUp(1)
		case key.Matches(msg, m.keys.Save):
			m.displaySaved = 2
			m.save()
		case key.Matches(msg, m.keys.Delete):
			if len(m.triggers.GetTriggers()) > 0 {
				m.triggers.RemovePoint(m.table.Cursor())
				m.table.SetRows(m.triggers.ToRows())
			}
			m.configChanged = true
		}
	}
	return m, tea.Batch(cmds...)
}

// t, err := time.Parse(time.TimeOnly, ftime)
// now := time.Now()
// td := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), now.Location())
func (m model) View() string {
	if m.quitting {
		return "powered by it creates media\n"
	}
	var b strings.Builder
	header := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(m.width).
		Render("OSC Timed triggerer by henne <3")
	b.WriteString(header)
	b.WriteRune('\n')
	clock := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(m.width).
		Render(m.clock.View())
	b.WriteString(clock)
	b.WriteRune('\n')
	if m.showCloseConfirm {
		saved := lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("205")).Bold(true).Align(lipgloss.Center).Width(m.width).Render("Config has been changed but not saved, are you sure to quit? [ Enter ] to Quit [ ESC ] to Cancel")
		b.WriteString(saved)
		return b.String()
	}
	b.WriteRune('\n')
	b.WriteString("Target Host: ")
	b.WriteString(m.inputs[0].View())
	b.WriteRune('\n')
	b.WriteString("Target Port: ")
	b.WriteString(m.inputs[1].View())
	b.WriteRune('\n')
	b.WriteString(m.table.View())
	b.WriteRune('\n')
	b.WriteString(boldText.Render("Add Trigger:"))
	b.WriteRune('\n')
	b.WriteString("Path: ")
	b.WriteString(m.inputs[2].View())
	b.WriteRune('\n')
	b.WriteString("Time: ")
	b.WriteString(m.inputs[3].View())
	b.WriteString(boldText.Render("  (HH:MM:SS)"))
	b.WriteRune('\n')
	if m.focusIndex > 2 {
		b.WriteString(focusedButton)
	} else {
		b.WriteString(blurredButton)
	}
	b.WriteRune('\n')
	if !m.help.ShowAll {
		b.WriteString("\n\n\n\n")
	}
	if m.displaySaved != 0 {
		saved := lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("205")).Bold(true).Align(lipgloss.Center).Width(m.width).Render("[ CONFIG SAVED ]")
		b.WriteString(saved)
	}
	b.WriteRune('\n')
	b.WriteString(m.help.View(m.keys) + "\n")
	footer := lipgloss.NewStyle().
		Align(lipgloss.Right).
		Width(m.width).
		Render("powered by it creates media")
	b.WriteString(footer)
	return b.String()
}
