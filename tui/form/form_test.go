package form

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func keyMsg(k string) tea.KeyMsg {
	switch k {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "shift+tab":
		return tea.KeyMsg{Type: tea.KeyShiftTab}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
	}
}

// typeText feeds each rune of s to the focused input.
func typeText(m Model, s string) Model { //nolint:gocritic // mirrors the Bubble Tea value-receiver idiom under test
	for _, r := range s {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	return m
}

func twoFieldForm() Model {
	return New("Test", []Field{
		{Key: "name", Label: "Name", Required: true},
		{Key: "id", Label: "ID"},
	})
}

func TestModel_Submit(t *testing.T) {
	m := typeText(twoFieldForm(), "corp")
	m, _ = m.Update(keyMsg("tab"))
	m = typeText(m, "tenant-1")

	_, cmd := m.Update(keyMsg("enter"))
	require.NotNil(t, cmd)

	msg := cmd()
	submitted, ok := msg.(Submitted)
	require.True(t, ok)
	assert.Equal(t, "corp", submitted.Values["name"])
	assert.Equal(t, "tenant-1", submitted.Values["id"])
}

func TestModel_Submit_TrimsWhitespace(t *testing.T) {
	m := typeText(twoFieldForm(), "  corp  ")

	_, cmd := m.Update(keyMsg("enter"))
	require.NotNil(t, cmd)

	submitted, ok := cmd().(Submitted)
	require.True(t, ok)
	assert.Equal(t, "corp", submitted.Values["name"])
}

func TestModel_Cancel(t *testing.T) {
	m := twoFieldForm()

	_, cmd := m.Update(keyMsg("esc"))
	require.NotNil(t, cmd)

	_, ok := cmd().(Canceled)
	assert.True(t, ok)
}

func TestModel_RequiredFieldBlocksSubmit(t *testing.T) {
	m := twoFieldForm() // name left empty

	next, cmd := m.Update(keyMsg("enter"))
	assert.Nil(t, cmd, "submit must be blocked")
	assert.NotEmpty(t, next.err, "an inline error must be shown")
	assert.Equal(t, 0, next.focus, "focus returns to the invalid field")
}

func TestModel_ValidatorBlocksSubmit(t *testing.T) {
	sentinel := errors.New("bad value")
	m := New("Test", []Field{
		{Key: "id", Label: "ID", Validate: func(string) error { return sentinel }},
	})
	m = typeText(m, "anything")

	next, cmd := m.Update(keyMsg("enter"))
	assert.Nil(t, cmd)
	assert.Equal(t, sentinel.Error(), next.err)
}

func TestModel_Navigation(t *testing.T) {
	m := twoFieldForm()
	require.Equal(t, 0, m.focus)

	m, _ = m.Update(keyMsg("tab"))
	assert.Equal(t, 1, m.focus)

	m, _ = m.Update(keyMsg("tab")) // wraps
	assert.Equal(t, 0, m.focus)

	m, _ = m.Update(keyMsg("shift+tab")) // wraps back
	assert.Equal(t, 1, m.focus)
}

func TestModel_NavigationClearsError(t *testing.T) {
	m := twoFieldForm()

	m, _ = m.Update(keyMsg("enter")) // triggers required error
	require.NotEmpty(t, m.err)

	m, _ = m.Update(keyMsg("tab"))
	assert.Empty(t, m.err, "moving focus clears the error")
}

func TestModel_ReadOnlyFieldSkipped(t *testing.T) {
	m := New("Edit", []Field{
		{Key: "name", Label: "Name", Value: "corp", ReadOnly: true},
		{Key: "id", Label: "ID", Required: true},
	})

	// The editable id field is focused initially, not the read-only name.
	assert.Equal(t, 1, m.focus)

	// Typing edits id; the read-only name keeps its value.
	m = typeText(m, "tenant-9")
	m, _ = m.Update(keyMsg("tab")) // wraps, still lands on id (only editable)
	assert.Equal(t, 1, m.focus)

	_, cmd := m.Update(keyMsg("enter"))
	require.NotNil(t, cmd)
	submitted, ok := cmd().(Submitted)
	require.True(t, ok)
	assert.Equal(t, "corp", submitted.Values["name"], "read-only value is preserved")
	assert.Equal(t, "tenant-9", submitted.Values["id"])
}
