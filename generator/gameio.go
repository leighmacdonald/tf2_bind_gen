package generator

import (
	"bind_generator/consts"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
)

const bindFmt = `echo "Loaded bind_generator.cfg"; alias bind_gen "say %s";\n%s`
const clearFmt = `alias bind_gen ""`

type MessageQueue struct {
	configPath string
	output     chan string
	clear      chan interface{}
	stop       chan interface{}
	// If queue size = 0, we can still be awaiting our output in-game to get triggered.
	// This will only be false if we have no enqueued message and have output our message in-game already if any.
	awaitingClear bool
}

func NewMessageQueue(configPath string, stopChan chan interface{}) MessageQueue {
	return MessageQueue{
		configPath: configPath,
		clear:      make(chan interface{}),
		output:     make(chan string),
		stop:       stopChan,
	}
}

// Add will enqueue messages for the user if we are waiting on another message to execute
func (mq *MessageQueue) Add(msg string) {
	if mq.awaitingClear || len(mq.output) > 0 {
		mq.output <- msg
	} else {
		if err := mq.WriteCfg(msg, false); err != nil {
			log.Error("Failed to write queue config file: %s", mq.configPath)
		}
	}
}

// Wait for clear msg
// if queue size of output > 0 write new cfg
func (mq *MessageQueue) Start() {
	for {
		select {
		case <-mq.stop:
			return
		case <-mq.clear:
			if err := mq.ClearCfg(); err != nil {
				log.Error("Failed to clear config file: %s", mq.configPath)
				continue
			}
			if len(mq.output) > 0 {
				msg := <-mq.output
				if err := mq.WriteCfg(msg, false); err != nil {
					log.Error("Failed to write queue config file: %s", mq.configPath)
				}
			}
		}
	}
}

// exec bind_gen/scratch.cfg
// bind_gen
// con_logfile bind_gen/scratch_cfg
// echo alias bindgen ""
// con_logfile console.log
//
func (mq *MessageQueue) WriteCfg(msg string, status bool) error {

	clearMsg := ""
	if status {
		clearMsg = consts.ClearConfigMsg
	}
	bind := fmt.Sprintf(bindFmt, msg, clearMsg)
	log.Debugf("Writing alias data: %s", bind)
	err := ioutil.WriteFile(mq.configPath, []byte(bind), 0644)
	if err != nil {
		return err
	}
	mq.awaitingClear = true
	return nil
}

func (mq *MessageQueue) ClearCfg() error {
	log.Debugf("Cleared bind config")
	err := ioutil.WriteFile(mq.configPath, []byte(clearFmt), 0644)
	if err != nil {
		return err
	}
	mq.awaitingClear = false
	return nil
}
