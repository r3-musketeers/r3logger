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
	"r3logger/pb"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

// Scribe ...
type Scribe struct {
	LogFile  io.Writer
	req      uint64
	incoming chan pb.Command
	cancel   context.CancelFunc
}

// NewScribe ...
func NewScribe(uniqueID string) *Scribe {

	l, err := createStateLog(uniqueID)
	if err != nil {
		log.Fatalln("could not init log file: ", err.Error())
	}

	ctx, c := context.WithCancel(context.Background())
	s := &Scribe{
		LogFile:  l,
		req:      0,
		incoming: make(chan pb.Command, 0),
		cancel:   c,
	}

	if !httpAPI {
		go s.ListenStateTransfer(ctx, recovAddr)
	}

	if monitoringThroughtput {
		go s.monitor(ctx)
	}
	return s
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

			err = conn.Close()
			if err != nil {
				log.Fatalf("Error encountered on connection close: %s", err.Error())
			}
		}
	}
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
