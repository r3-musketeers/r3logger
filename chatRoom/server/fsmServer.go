package main

import (
	"io"
	"strconv"
	"strings"

	"github.com/hashicorp/raft"
)

// Must implement the Raft FSM interface, even if the chatRoom application
// won't do any procedure after successfully applys by consensus protocol
type fsm Server

// Apply proposes a new value to the consensus cluster
func (s *fsm) Apply(l *raft.Log) interface{} {

	// Using a string slice to concatenate strings in O(n) complexity
	auxBuffer := []string{strconv.FormatUint(l.Index, 10), string(l.Data)}

	// Broadcast a message to every other client on the room
	for _, client := range s.clients {
		client.outgoing <- strings.Join(auxBuffer, "-")
	}

	return nil
}

// Restore stores the key-value store to a previous state.
func (s *fsm) Restore(rc io.ReadCloser) error {
	return nil
}

// Snapshot returns a snapshot of the key-value store.
func (s *fsm) Snapshot() (raft.FSMSnapshot, error) {
	return nil, nil
}
