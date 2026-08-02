package ui

import (
	"fmt"
	"strings"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ProgressMsg struct {
	ID       string
	Status   string
	Progress float64
}

type DoneMsg struct{
	Image string
}

type layer struct {
	status   string
	progress float64
}

type Model struct {
	events <-chan tea.Msg
	order  []string
	layers map[string]*layer
	done   bool
}

func New(events <-chan tea.Msg) Model {
	return Model{
		events: events,
		layers: make(map[string]*layer),
	}
}

func waitForMsg(events <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-events
		if !ok {
			return DoneMsg{}
		}
		return msg
	}
}

func (m Model) Init() tea.Cmd {
	return waitForMsg(m.events)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ProgressMsg:
		l, ok := m.layers[msg.ID]
		if !ok {
			l = &layer{}
			m.layers[msg.ID] = l
			m.order = append(m.order, msg.ID)
		}
		l.status = msg.Status
		l.progress = msg.Progress
		return m, waitForMsg(m.events)

	case DoneMsg:
		m.done = true
		return m, tea.Quit

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

const barWidth = 30

var (
	barStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	doneStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
)

func renderBar(progress float64) string {
	filled := int(progress * barWidth)
	if filled > barWidth {
		filled = barWidth
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
	return barStyle.Render(bar)
}

func (m Model) View() string {
	var b strings.Builder
	for _, id := range m.order {
		l := m.layers[id]
		if l.status == "Pull complete" {
			b.WriteString(doneStyle.Render(fmt.Sprintf("%s Pull complete", id)) + "\n")
			continue
		}
		b.WriteString(fmt.Sprintf("%s %-15s %s %3.0f%%\n",
			id, l.status, renderBar(l.progress), l.progress*100))
	}
	if m.done {
		b.WriteString("\n" + doneStyle.Render("Status: Downloaded newer image") + "\n")
	}
	return b.String()
}