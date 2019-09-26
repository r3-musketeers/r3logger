package main

import (
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

// Logger struct represents the Logger process state. Member of the Raft cluster as a
// non-Voter participant and thus, just recording proposed commands to the FSM
type Logger struct {
	workers []Scribe

	incoming chan pb.Command
}

// NewLogger ...
func NewLogger() *Logger {
	return nil
}

// Listen ...
func (l *Logger) Listen() {}

// TODO: Receive a single stream of 'LogEntry' messages and distributed each to their
// corresponding Scribe, identified by an unique ID.
func (l *Logger) fanOutPattern() {}
