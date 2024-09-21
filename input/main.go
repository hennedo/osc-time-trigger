package input

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	input        textinput.Model
	focusStyle   lipgloss.Style
	blurStyle    lipgloss.Style
	allowedChars string
}

func (m Model) View() string {
	return m.input.View()
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if len(m.allowedChars) > 0 && msg.String() != "backspace" && msg.String() != "left" && msg.String() != "right" && !strings.Contains(m.allowedChars, msg.String()) {
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Focus() (Model, tea.Cmd) {
	m.input.PromptStyle = m.focusStyle
	m.input.TextStyle = m.focusStyle
	return m, m.input.Focus()
}

func (m *Model) Blur() {
	m.input.PromptStyle = m.blurStyle
	m.input.TextStyle = m.blurStyle
	m.input.Blur()
}

func (m Model) Value() string {
	return m.input.Value()
}

func (m *Model) SetValue(value string) {
	m.input.SetValue(value)
}

func (m *Model) SetAllowedChars(chars string) {
	m.allowedChars = chars
}
func New(placeholder string, focusStyle, blurStyle lipgloss.Style) Model {
	m := Model{
		focusStyle:   focusStyle,
		blurStyle:    blurStyle,
		allowedChars: "",
	}

	m.input = textinput.New()
	m.input.Cursor.Style = focusStyle
	m.input.Placeholder = placeholder
	m.input.PromptStyle = blurStyle
	m.input.TextStyle = blurStyle
	return m
}
