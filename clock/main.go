package clock

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
}

func (m Model) Init() tea.Cmd {
	return tick()
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg.(type) {
	case TickMsg:
		return m, tick()
	}
	return m, nil
}

func (m Model) View() string {

	s := time.Now().Format("15:04:05")
	return s
}

func New() Model {
	return Model{}
}

type TickMsg time.Time

func tick() tea.Cmd {
	return tea.Every(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}
