package control

// Event represents a user action or intent derived from key input.
type Event int

// Event constants representing various user actions or intents.
const (
	EventNone Event = iota

	EventNext
	EventPrev

	EventSelect
	EventView
	EventClose
	EventQuit
)
