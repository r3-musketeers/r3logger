package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"

	"github.com/gorilla/handlers"
)

const (
	defaultPort = ":8181"
	httpAPI     = true

	logInFile   = true
	logFilename = "events.log"
)

var (
	instanceLog *Logger
	singleton   sync.Once

	logID     string
	raftAddr  string
	joinAddr  string
	recovAddr string
	logfolder *string
)

func init() {
	flag.StringVar(&logID, "id", "", "Set the logger unique ID")
	flag.StringVar(&raftAddr, "raft", ":12000", "Set RAFT consensus bind address")
	flag.StringVar(&joinAddr, "join", ":13000", "Set join address to an already configured raft node")
	flag.StringVar(&recovAddr, "hrecov", "", "Set port id to receive state transfer requests from the application log")
	logfolder = flag.String("logfolder", "", "log received commands to a file at specified destination folder using Journey")
}

func main() {

	if httpAPI {
		r := createRouter()
		w := createEventLog()
		logR := handlers.LoggingHandler(w, r)

		go func() {
			err := http.ListenAndServe(defaultPort, logR)
			if err != nil {
				log.Fatal(err)
			}
		}()
	} else {
		initStaticConfig()
	}

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)
	<-terminate
	GetLogger().Shutdown()
}

func initStaticConfig() {
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
			// TODO: Launch logger routines.
		}(i)
	}
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

func createEventLog() io.Writer {
	w := os.Stderr
	flags := os.O_APPEND | os.O_WRONLY

	if logInFile {
		if _, exists := os.Stat(logFilename); exists == nil {
			w, _ = os.OpenFile(logFilename, flags, 0644)
		} else if os.IsNotExist(exists) {
			w, _ = os.OpenFile(logFilename, os.O_CREATE|flags, 0644)
		} else {
			log.Fatalln("could not create log file:", exists.Error())
		}
	}
	return w
}
