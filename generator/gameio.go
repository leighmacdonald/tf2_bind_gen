package generator

import (
	"bind_generator/consts"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
)

const bindFmt = `echo "Loaded bind_generator.cfg"; alias bind_gen "say %s";\n%s`
const clearFmt = `alias bind_gen ""`

type BindWriter struct {
	configPath string
	writeCount int
	writeEvery int
}

// exec bind_gen/scratch.cfg
// bind_gen
// con_logfile bind_gen/scratch_cfg
// echo alias bindgen ""
// con_logfile console.log
//
func (bw *BindWriter) Write(p []byte) (n int, err error) {
	clearMsg := ""
	if bw.writeCount%bw.writeEvery == 0 {
		clearMsg = consts.ClearConfigMsg
	}
	bind := fmt.Sprintf(bindFmt, string(p), clearMsg)
	log.Debugf("Writing alias data: %s", bind)
	if err := ioutil.WriteFile(bw.configPath, []byte(bind), 0644); err != nil {
		return 0, err
	}
	return len(bind), nil
}

func (bw *BindWriter) Reset() error {
	log.Debugf("Cleared bind config")
	err := ioutil.WriteFile(bw.configPath, []byte(clearFmt), 0644)
	if err != nil {
		return err
	}
	return nil
}

type ConsoleIOQueue struct {
	output chan string
	clear  chan interface{}
	stop   chan interface{}
	writer BindWriter
	// If queue size = 0, we can still be awaiting our output in-game to get triggered.
	// This will only be false if we have no enqueued message and have output our message in-game already if any.
	awaitingClear bool
}

func NewConsoleIOQueue(configPath string, stopChan chan interface{}) ConsoleIOQueue {
	return ConsoleIOQueue{
		clear:  make(chan interface{}),
		output: make(chan string),
		stop:   stopChan,
		writer: BindWriter{configPath, 0, 3},
	}
}

// Add will enqueue messages for the user if we are waiting on another message to execute
func (mq *ConsoleIOQueue) Add(msg string) {
	if mq.awaitingClear || len(mq.output) > 0 {
		mq.output <- msg
	} else {
		if _, err := mq.writer.Write([]byte(msg)); err != nil {
			log.Error("Failed to write queue config file")
		}
	}
}

// Wait for clear msg
// if queue size of output > 0 write new cfg
func (mq *ConsoleIOQueue) Start() {
	for {
		select {
		case <-mq.stop:
			return
		case <-mq.clear:
			if err := mq.writer.Reset(); err != nil {
				log.Error("Failed to clear config file")
				continue
			}
			if len(mq.output) > 0 {
				msg := <-mq.output
				if _, err := mq.writer.Write([]byte(msg)); err != nil {
					log.Error("Failed to write queue config file")
				}
			}
		}
	}
}
