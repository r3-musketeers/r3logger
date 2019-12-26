package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/hashicorp/raft"
)

// Custom configuration over default for testing
func configRaft() *raft.Config {
	config := raft.DefaultConfig()
	config.SnapshotInterval = 24 * time.Hour
	config.SnapshotThreshold = 2 << 62
	config.LogLevel = "WARN"
	return config
}

// Scribe ...
type Scribe struct {
	LogFile io.Writer
	raft    *raft.Raft

	req    uint64
	cancel context.CancelFunc
}

// NewScribe ...
func NewScribe(id, recov string) *Scribe {

	l, err := createStateLog(id)
	if err != nil {
		log.Fatalln("could not init log file: ", err.Error())
	}

	ctx, c := context.WithCancel(context.Background())
	s := &Scribe{
		LogFile: l,
		req:     0,
		cancel:  c,
	}

	if !httpAPI {
		go s.ListenStateTransfer(ctx, recov)
	}

	if monitoringThroughtput {
		go s.monitor(ctx)
	}
	return s
}

// StartRaft ...
func (s *Scribe) StartRaft(localID, raftAddr string) error {

	// Setup Raft configuration.
	config := configRaft()
	config.LocalID = raft.ServerID(localID)

	// Setup Raft communication.
	addr, err := net.ResolveTCPAddr("tcp", raftAddr)
	if err != nil {
		return err
	}
	transport, err := raft.NewTCPTransport(raftAddr, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return err
	}

	logStore := raft.NewInmemStore()
	stableStore := raft.NewInmemStore()

	// Create a fake snapshot store
	dir := "checkpoints/" + localID
	snapshots, err := raft.NewFileSnapshotStore(dir, 2, os.Stderr)
	if err != nil {
		return err
	}

	// Instantiate the Raft systems.
	ra, err := raft.NewRaft(config, (*fsm)(s), logStore, stableStore, snapshots, transport)
	if err != nil {
		return fmt.Errorf("new raft: %s", err)
	}
	s.raft = ra
	return nil
}

// ListenStateTransfer ...
func (s *Scribe) ListenStateTransfer(c context.Context, addr string) {

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to bind connection at %s: %s", addr, err.Error())
	}

	for {
		select {
		case <-c.Done():
			return

		default:
			conn, err := listener.Accept()
			if err != nil {
				log.Fatalf("accept failed: %s", err.Error())
			}
			defer conn.Close()

			request, _ := bufio.NewReader(conn).ReadString('\n')

			data := strings.Split(request, "-")
			if len(data) != 2 {
				log.Fatalf("incorrect state request, got: %s", data)
			}

			data[1] = strings.TrimSuffix(data[1], "\n")
			requestedLogIndex, _ := strconv.Atoi(data[1])

			err = s.UnsafeStateRecover(uint64(requestedLogIndex), conn)
			if err != nil {
				log.Fatalf("failed to transfer log to node located at %s: %s", data[0], err.Error())
			}
		}
	}
}

func (s *Scribe) monitor(c context.Context) {
	for {
		select {
		case <-c.Done():
			return

		default:
			time.Sleep(1 * time.Second)
			cont := atomic.SwapUint64(&s.req, 0)
			fmt.Println(cont)
		}
	}
}

// UnsafeStateRecover ...
func (s *Scribe) UnsafeStateRecover(logIndex uint64, activePipe net.Conn) error {

	// Create a read-only file descriptor
	logFileName := *logfolder + "log-file-" + logID + ".txt"
	fd, _ := os.OpenFile(logFileName, os.O_RDONLY, 0644)
	defer fd.Close()

	logFileContent, err := readAll(fd)
	if err != nil {
		return err
	}

	signalError := make(chan error, 0)
	go func(dataToSend []byte, pipe net.Conn, signal chan<- error) {

		_, err := pipe.Write(dataToSend)
		signal <- err

	}(logFileContent, activePipe, signalError)
	return <-signalError
}

// Stop ...
func (s *Scribe) Stop() {
	s.cancel()
	s.raft.Shutdown()
}

// readAll is a slightly derivation of 'ioutil.ReadFile()'. It skips the file descriptor creation
// and is declared to avoid unecessary dependency from the whole ioutil package.
// 'A little copying is better than a little dependency.'
func readAll(fileDescriptor *os.File) ([]byte, error) {
	// It's a good but not certain bet that FileInfo will tell us exactly how much to
	// read, so let's try it but be prepared for the answer to be wrong.
	var n int64 = bytes.MinRead

	if fi, err := fileDescriptor.Stat(); err == nil {
		// As initial capacity for readAll, use Size + a little extra in case Size
		// is zero, and to avoid another allocation after Read has filled the
		// buffer. The readAll call will read into its allocated internal buffer
		// cheaply. If the size was wrong, we'll either waste some space off the end
		// or reallocate as needed, but in the overwhelmingly common case we'll get
		// it just right.
		if size := fi.Size() + bytes.MinRead; size > n {
			n = size
		}
	}
	return func(r io.Reader, capacity int64) (b []byte, err error) {
		// readAll reads from r until an error or EOF and returns the data it read
		// from the internal buffer allocated with a specified capacity.
		var buf bytes.Buffer
		// If the buffer overflows, we will get bytes.ErrTooLarge.
		// Return that as an error. Any other panic remains.
		defer func() {
			e := recover()
			if e == nil {
				return
			}
			if panicErr, ok := e.(error); ok && panicErr == bytes.ErrTooLarge {
				err = panicErr
			} else {
				panic(e)
			}
		}()
		if int64(int(capacity)) == capacity {
			buf.Grow(int(capacity))
		}
		_, err = buf.ReadFrom(r)
		return buf.Bytes(), err
	}(fileDescriptor, n)
}

func createStateLog(id string) (io.Writer, error) {
	var flags int
	var w io.Writer

	logFileName := *logfolder + "log-file-" + id + ".txt"
	if catastrophicFaults {
		flags = os.O_SYNC | os.O_WRONLY
	} else {
		flags = os.O_WRONLY
	}

	if _, exists := os.Stat(logFileName); exists == nil {
		w, _ = os.OpenFile(logFileName, flags, 0644)
	} else if os.IsNotExist(exists) {
		w, _ = os.OpenFile(logFileName, os.O_CREATE|flags, 0644)
	} else {
		return nil, exists
	}
	return w, nil
}
