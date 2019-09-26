package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
)

var logID string
var raftAddr string
var joinAddr string
var logfolder *string

func init() {
	flag.StringVar(&logID, "id", "", "Set the logger unique ID")
	flag.StringVar(&raftAddr, "raft", ":12000", "Set RAFT consensus bind address")
	flag.StringVar(&joinAddr, "join", ":13000", "Set join address to an already configured raft node")

	logfolder = flag.String("logfolder", "", "log received commands to a file at specified destination folder using Journey")
}

func main() {

	flag.Parse()
	if logID == "" {
		log.Fatalln("must set a logger ID, run with: ./logger -id 'logID'")
	}

	listOfLogIds := strings.Split(logID, ",")
	numDiffIds := countDiffStrInSlice(listOfLogIds)

	listOfRaftAddrs := strings.Split(raftAddr, ",")
	numDiffRaft := countDiffStrInSlice(listOfRaftAddrs)

	listOfJoinAddrs := strings.Split(joinAddr, ",")
	numDiffServices := countDiffStrInSlice(listOfJoinAddrs)

	if numDiffServices != numDiffIds || numDiffIds != numDiffRaft || numDiffRaft != numDiffServices {
		log.Fatalln("must run with the same number of unique IDs, raft and join addrs: ./logger -id 'X,Y' -raft 'A,B' -join 'W,Z'")
	}

	for i := 0; i < numDiffServices; i++ {
		go func(j int) {
			// TODO: ...
		}(i)
	}

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)
	<-terminate

	// TODO: ...

}

func countDiffStrInSlice(elements []string) int {

	foundMarker := make(map[string]bool, len(elements))
	numDiff := 0

	for _, str := range elements {
		if !foundMarker[str] {
			foundMarker[str] = true
			numDiff++
		}
	}
	return numDiff
}
