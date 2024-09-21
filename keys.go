package main

import "github.com/charmbracelet/bubbles/key"

// keyMap defines a set of keybindings. To work for help it must satisfy
// key.Map. It could also very easily be a map[string]key.Binding.
type keyMap struct {
	Up      key.Binding
	Down    key.Binding
	Help    key.Binding
	Quit    key.Binding
	Delete  key.Binding
	Trigger key.Binding
	Host    key.Binding
	Port    key.Binding
	Esc     key.Binding
	Enter   key.Binding
	Save    key.Binding
	Edit    key.Binding
	Copy    key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Save: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "save config"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "enter to confirm"),
	),
	Esc: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "esc to cancel"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete trigger from list"),
	),
	Copy: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "copy list item"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit trigger"),
	),
	Trigger: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "add new trigger"),
	),
	Host: key.NewBinding(
		key.WithKeys("h"),
		key.WithHelp("h", "change target host / ip"),
	),
	Port: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "change target port"),
	),
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Enter, k.Esc, k.Up, k.Down},
		{k.Host, k.Port, k.Trigger, k.Copy}, // first column
		{k.Save, k.Help, k.Quit},            // second column
	}
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Enter, k.Esc, k.Help, k.Quit}
}
