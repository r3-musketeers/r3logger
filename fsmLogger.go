package main

import (
	"io"
	"sync/atomic"

	"r3logger/pb"

	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/raft"
)

// Must implement the Raft FSM interface, even if the chatRoom application
// won't do any procedure after successfully applys by consensus protocol
type fsm Scribe

// Apply proposes a new value to the consensus cluster
func (s *fsm) Apply(l *raft.Log) interface{} {

	command := &pb.Command{}
	err := proto.Unmarshal(l.Data, command)
	if err != nil {
		return err
	}
	command.Id = l.Index
	serializedCmd, _ := proto.Marshal(command)
	_, err = s.LogFile.Write(serializedCmd)

	if s.monit {
		atomic.AddUint64(&s.req, 1)
	}
	return err
}

// Restore stores the key-value store to a previous state.
func (s *fsm) Restore(rc io.ReadCloser) error {
	return nil
}

// Snapshot returns a snapshot of the key-value store.
func (s *fsm) Snapshot() (raft.FSMSnapshot, error) {
	return nil, nil
}
