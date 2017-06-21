package main

import (
	"errors"
	"io"

	"github.com/fatih/color"
	"github.com/jroimartin/gocui"
)

var state = &State{
	ActionIndex: -1,
}

// State is the central function managing the information of claws.
type State struct {
	LastActions   []string
	Mode          int
	ActionIndex   int
	Writer        io.Writer
	Conn          *WebSocket
	FirstDrawDone bool
	ExecuteFunc   func(func(*gocui.Gui) error)
}

// PushAction adds an action to LastActions
func (s *State) PushAction(act string) {
	s.LastActions = append([]string{act}, s.LastActions...)
	if len(s.LastActions) > 100 {
		s.LastActions = s.LastActions[:100]
	}
}

// BrowseActions changes the ActionIndex and returns the value at the specified index.
// move is the number of elements to move (negatives go into more recent history,
// 0 returns the current element, positives go into older history)
func (s *State) BrowseActions(move int) string {
	s.ActionIndex += move
	if s.ActionIndex >= len(s.LastActions) {
		s.ActionIndex = len(s.LastActions) - 1
	} else if s.ActionIndex < -1 {
		s.ActionIndex = -1
	}

	// -1 always indicates the "next" element, thus empty
	if s.ActionIndex == -1 {
		return ""
	}
	return s.LastActions[s.ActionIndex]
}

// StartConnection begins a WebSocket connection to url.
func (s *State) StartConnection(url string) error {
	if s.Conn != nil {
		return errors.New("state: conn is not nil")
	}
	ws, err := CreateWebSocket(url)
	if err != nil {
		return err
	}
	s.Conn = ws
	go s.wsReader()
	return nil
}

func (s *State) wsReader() {
	ch := s.Conn.ReadChannel()
	for msg := range ch {
		s.Server(msg)
	}
}

var (
	printDebug  = color.New(color.FgCyan).Fprint
	printError  = color.New(color.FgRed).Fprint
	printUser   = color.New(color.FgGreen).Fprint
	printServer = color.New(color.FgWhite).Fprint
)

// Debug prints debug information to the Writer, using light blue.
func (s *State) Debug(x string) {
	s.printToOut(printDebug, x)
}

// Error prints an error to the Writer, using red.
func (s *State) Error(x string) {
	s.printToOut(printError, x)
}

// User prints user-provided messages to the Writer, using green.
func (s *State) User(x string) {
	s.printToOut(printUser, x)
}

// Server prints server-returned messages to the Writer, using white.
func (s *State) Server(x string) {
	s.printToOut(printServer, x)
}

func (s *State) printToOut(f func(io.Writer, ...interface{}) (int, error), str string) {
	if str != "" && str[len(str)-1] != '\n' {
		str += "\n"
	}
	s.ExecuteFunc(func(*gocui.Gui) error {
		_, err := f(s.Writer, str)
		return err
	})
}