// Copyright (c) 2021 Romano
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.

package engineManager

import (
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	cm "github.com/charmbracelet/charm/ui/common"
	"github.com/muesli/reflow/indent"
	"github.com/muesli/termenv"
	"time"
)

// General states
const (
	StatusRunning status = iota
	StatusError
	StatusSuccess
	StatusDone
	StatusQuitting
)

const indentAmount = 2

type status int

var color = termenv.ColorProfile().Color

type FailedMsg struct{ Err error }

func (f FailedMsg) Error() string { return f.Err.Error() }

type SuccessMsg struct{ Msg string }

func (s SuccessMsg) String() string { return s.String() }

// DoneMsg is sent when the keygen has completely finished running.
type DoneMsg struct{}

// Model is the Bubble Tea model which stores the state of the keygen.
type Model struct {
	Status        status
	err           error
	standalone    bool
	fancy         bool
	spinner       spinner.Model
	terminalWidth int
}

func LoadModel() Model {
	s := spinner.NewModel()
	s.Spinner = spinner.Dot
	return Model{spinner: s}
}

func (m Model) Loading() string {
	if m.err != nil {
		return m.err.Error()
	}
	s := termenv.String(m.spinner.View()).Foreground(color("205")).String()
	str := fmt.Sprintf("\n\n   %s Loading forever...press q to quit\n\n", s)
	return str
}

// View renders the view from the keygen model.
func (m Model) View() string {
	var s string

	switch m.Status {
	case StatusRunning:
		if m.standalone {
			s += m.spinner.View()
		}
		s += " Spanning new engine..."
	case StatusSuccess:
		s += termenv.String("✔").Foreground(cm.Green.Color()).String()
		s += "  Engine started "
	case StatusError:
		switch m.err.(type) {
		//case keygen.SSHKeysAlreadyExistErr:
		//	s += "You already have SSH keys :)"
		default:
			s += fmt.Sprintf("Uh oh, there's been an error: %v", m.err)
		}
	case StatusQuitting:
		s += "Exiting..."
	}

	if m.standalone && m.fancy {
		//return termenv.String(m.spinner.View()).Foreground(color("205")).String()
		return indent.String(fmt.Sprintf("\n%s\n\n", s), indentAmount)
	}

	return s
}

// NewModel returns a new keygen model in its initial state.
func NewModel() Model {
	return Model{
		Status: StatusRunning,
	}
}

// Init is the Bubble Tea initialization function for the keygen.
func (m Model) Init() tea.Cmd {
	//s := engineManager.StartMainEngine()
	return tea.Batch(StartMainEngine, spinner.Tick)
}

func (m Model) StopMainEngine() tea.Cmd {
	return tea.Batch(StopMainEngine(), spinner.Tick)
}

//func (m Model) Load() Model {
//	s := spinner.NewModel()
//	s.Spinner = spinner.Dot
//	return Model{spinner: s}
//}

//func (m Model) ViewStop() string {
//	if m.err != nil {
//		return m.err.Error()
//	}
//	s := termenv.String(m.spinner.View()).Foreground(color("205")).String()
//	str := fmt.Sprintf("\n\n   %s Loading forever...press q to quit\n\n", s)
//	return str
//}

// Update is the Bubble Tea update loop for the keygen.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			m.Status = StatusQuitting
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.terminalWidth = msg.Width
		return m, nil
	case FailedMsg:
		m.err = msg.Err
		m.Status = StatusError
		return m, tea.Quit
	case SuccessMsg:
		m.Status = StatusSuccess
		return m, pause
	case DoneMsg:
		if m.standalone {
			return m, tea.Quit
		}
		m.Status = StatusDone
		return m, nil
	case spinner.TickMsg:
		if m.Status == StatusRunning {
			newSpinnerModel, cmd := m.spinner.Update(msg)
			m.spinner = newSpinnerModel
			return m, cmd
		}
	}

	return m, nil
}

// pause runs the final pause before we wrap things up.
func pause() tea.Msg {
	time.Sleep(time.Millisecond * 600)
	return DoneMsg{}
}

//func NewProgram(fancy bool) *tea.Program {
//	m := NewModel()
//	m.standalone = true
//	m.fancy = fancy
//	m.spinner = spinner.NewModel()
//	m.spinner.Spinner = cm.Spinner
//	m.spinner.ForegroundColor = cm.SpinnerColor.String()
//	return tea.NewProgram(m)
//}
