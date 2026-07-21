// Package form provides a reusable multi-field text-input form for the TUI.
// It orchestrates an ordered set of fields, handles focus navigation and
// validation, and emits a Submitted or Canceled message. It knows nothing
// about the domain: callers map the submitted values back to their types.
package form

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lvlcn-t/azctx/tui/keys"
	"github.com/lvlcn-t/azctx/tui/styles"
)

// Field describes a single form input.
type Field struct {
	Validate    func(value string) error
	Key         string
	Label       string
	Placeholder string
	Value       string
	Required    bool
}

// Submitted is emitted when the form passes validation and the user submits.
type Submitted struct {
	// Values maps each Field.Key to its final trimmed value.
	Values map[string]string
}

// Canceled is emitted when the user aborts the form.
type Canceled struct{}

// Model is a focusable multi-field form.
type Model struct {
	title  string
	err    string
	keys   formKeys
	fields []Field
	inputs []textinput.Model
	focus  int
	width  int
}

type formKeys struct {
	Next   key.Binding
	Prev   key.Binding
	Submit key.Binding
	Cancel key.Binding
}

func newFormKeys() formKeys {
	return formKeys{
		Next:   keys.New(keys.Tab).WithHelp("next").WithAliases(keys.ArrowDown).Bind(),
		Prev:   keys.New(keys.ShiftTab).WithHelp("prev").WithAliases(keys.ArrowUp).Bind(),
		Submit: keys.New(keys.Enter).WithHelp("submit").Bind(),
		Cancel: keys.New(keys.Escape).WithHelp("cancel").Bind(),
	}
}

// New builds a form from the given title and fields. The first field is
// focused.
func New(title string, fields []Field) Model {
	inputs := make([]textinput.Model, len(fields))
	for i, f := range fields {
		in := textinput.New()
		in.Placeholder = f.Placeholder
		in.SetValue(f.Value)
		if i == 0 {
			in.Focus()
		}
		inputs[i] = in
	}

	return Model{
		title:  title,
		fields: fields,
		inputs: inputs,
		keys:   newFormKeys(),
	}
}

// SetWidth records the available width for layout.
func (m *Model) SetWidth(width int) {
	m.width = width
}

// Values returns the current trimmed value of every field, keyed by Field.Key.
func (m *Model) Values() map[string]string {
	values := make(map[string]string, len(m.fields))
	for i, f := range m.fields {
		values[f.Key] = strings.TrimSpace(m.inputs[i].Value())
	}
	return values
}

// Update handles navigation, editing, submission, and cancellation. It returns
// a Submitted or Canceled tea.Cmd when the form terminates.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) { //nolint:gocritic // Bubble Tea value-receiver idiom; refactor tracked separately
	if _, ok := msg.(tea.KeyMsg); ok {
		switch {
		case keys.Matches(msg, m.keys.Cancel):
			return m, cancel

		case keys.Matches(msg, m.keys.Submit):
			return m.trySubmit()

		case keys.Matches(msg, m.keys.Next):
			m.focusDelta(1)
			return m, nil

		case keys.Matches(msg, m.keys.Prev):
			m.focusDelta(-1)
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.inputs[m.focus], cmd = m.inputs[m.focus].Update(msg)
	return m, cmd
}

// trySubmit validates every field and, on success, emits Submitted.
func (m Model) trySubmit() (Model, tea.Cmd) { //nolint:gocritic // Bubble Tea value-receiver idiom; refactor tracked separately
	values := make(map[string]string, len(m.fields))
	for i, f := range m.fields {
		value := strings.TrimSpace(m.inputs[i].Value())
		if f.Required && value == "" {
			m.err = f.Label + " is required"
			m.focusOn(i)
			return m, nil
		}

		if f.Validate != nil {
			if err := f.Validate(value); err != nil {
				m.err = err.Error()
				m.focusOn(i)
				return m, nil
			}
		}

		values[f.Key] = value
	}

	return m, submit(values)
}

func (m *Model) focusDelta(delta int) {
	m.err = ""
	m.focusOn((m.focus + delta + len(m.inputs)) % len(m.inputs))
}

// focusOn moves focus to index without touching the error message.
func (m *Model) focusOn(index int) {
	m.inputs[m.focus].Blur()
	m.focus = index
	m.inputs[m.focus].Focus()
}

// View renders the form as a bordered panel.
func (m Model) View() string { //nolint:gocritic // Bubble Tea value-receiver idiom; refactor tracked separately
	labelWidth := 0
	for _, f := range m.fields {
		if len(f.Label) > labelWidth {
			labelWidth = len(f.Label)
		}
	}

	labelStyle := lipgloss.NewStyle().Width(labelWidth + 1).Foreground(styles.ColorPrimary)

	var rows []string
	rows = append(rows, styles.TitleStyle.Render(m.title), "")
	for i, f := range m.fields {
		row := lipgloss.JoinHorizontal(
			lipgloss.Top,
			labelStyle.Render(f.Label),
			m.inputs[i].View(),
		)
		rows = append(rows, row)
	}

	if m.err != "" {
		rows = append(rows, "", styles.ErrorStyle.Render(m.err))
	}

	rows = append(rows, "", styles.HelpStyle.Render("tab/↑↓ move · enter submit · esc cancel"))

	return styles.ViewerStyle.Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

func submit(values map[string]string) tea.Cmd {
	return func() tea.Msg { return Submitted{Values: values} }
}

func cancel() tea.Msg {
	return Canceled{}
}
