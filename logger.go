package main

import (
	"fmt"
	"net"
	"r3logger/pb"
)

const (
	// Used in catastrophic fault models, where crash faults must be recoverable even if
	// all nodes presented in the consensus cluster are down. Always set to false in any
	// other cases, because this strong assumption greatly degradates performance.
	catastrophicFaults = false

	// Each second writes current throughput to stdout.
	monitoringThroughtput = false
)

type loggerInfo struct {
	id       string
	raft     string
	recov    string
	joinHost string
}

// Logger struct represents the Logger process state. Member of the Raft cluster as a
// non-Voter participant and thus, just recording proposed commands to the FSM
type Logger struct {
	workers  []Scribe
	incoming chan pb.Command
}

// NewLogger ...
func NewLogger() *Logger {
	return &Logger{
		workers:  make([]Scribe, 0),
		incoming: make(chan pb.Command, 0),
	}
}

// GetLogger ...
func GetLogger() *Logger {
	singleton.Do(func() {
		instanceLog = NewLogger()
	})
	return instanceLog
}

// DeployScribe ...
func (l *Logger) DeployScribe(li loggerInfo) error {

	s := NewScribe(li.id, li.recov)
	err := s.StartRaft(li.id, li.raft)
	if err != nil {
		return err
	}

	err = sendJoinRequest(li.id, li.raft, li.joinHost)
	if err != nil {
		return err
	}

	l.workers = append(l.workers, *s)
	return nil
}

// TODO: Receive a single stream of 'LogEntry' messages and distributed each to their
// corresponding Scribe, identified by an unique ID.
func (l *Logger) fanOutPattern() {}

// Shutdown ...
func (l *Logger) Shutdown() {
	for _, w := range l.workers {
		w.Stop()
	}
}

func sendJoinRequest(logID, raftAddr, joinAddr string) error {
	conn, err := net.Dial("tcp", joinAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = fmt.Fprint(conn, logID+"-"+raftAddr+"-"+"false"+"\n")
	if err != nil {
		return err
	}
	return nil
}
