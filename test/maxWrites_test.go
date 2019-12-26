package main

import (
	"encoding/binary"
	"log"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"r3logger/pb"

	"github.com/golang/protobuf/proto"
)

const (
	numKeys  = 1000000
	execTime = 60
)

var testLogFilename = "/tmp/log-max-test.txt"

func TestMaxWritesByClientTime(b *testing.T) {

	storeValue := strings.Repeat("@", 1024)

	signal := make(chan bool)
	requests := make(chan *pb.Command, 512)
	logFile, err := os.OpenFile(testLogFilename, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalln(err)
	}

	go generateProtobufRequests(requests, signal, numKeys, storeValue)
	go killWorkers(execTime, signal)

	var numWrites uint64
	for {
		msg, ok := <-requests
		if !ok {
			break
		}

		serializedMessage, err := proto.Marshal(msg)
		if err != nil {
			log.Fatal(err)
		}

		command := &pb.Command{}
		err = proto.Unmarshal(serializedMessage, command)
		if err != nil {
			log.Fatal(err)
		}

		command.Id = 10
		serializedCmd, _ := proto.Marshal(command)
		binary.Write(logFile, binary.BigEndian, int32(len(serializedCmd)))
		_, err = logFile.Write(serializedCmd)
		if err != nil {
			log.Fatal(err)
		}
		numWrites++
	}
	b.Logf("Executed %d in %d seconds of experiment", numWrites, execTime)
}

func generateProtobufRequests(reqs chan<- *pb.Command, signal <-chan bool, numKey int, storeValue string) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	for {
		var msg *pb.Command
		op := rand.Intn(2)

		switch op {
		case 0:
			msg = &pb.Command{
				Op:    pb.Command_SET,
				Key:   strconv.Itoa(rand.Intn(numKey)),
				Value: storeValue,
			}
			break
		case 1:
			msg = &pb.Command{
				Op:  pb.Command_GET,
				Key: strconv.Itoa(rand.Intn(numKey)),
			}
			// case 2:
			// 	msg = &pb.Command{
			// 		Op:  pb.Command_DELETE,
			// 		Key: strconv.Itoa(rand.Intn(numKey)),
			// 	}
		}

		select {
		case <-signal:
			close(reqs)
			return

		default:
			reqs <- msg
		}
	}
}

func killWorkers(timeSeconds int64, signal chan<- bool) {
	t := time.NewTimer(time.Duration(timeSeconds) * time.Second)
	<-t.C
	signal <- true
}
