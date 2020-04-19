package app

import (
	"bind_generator/consts"
	"bind_generator/generator"
	"bind_generator/model"
	"bind_generator/store"
	"bufio"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
)

type BindGenerator struct {
	logMsgChan    chan string
	stopChan      chan interface{}
	evtChan       chan *model.LogEvent
	sigChan       chan os.Signal
	state         model.PlayerState
	parser        *generator.LogParser
	commandParser generator.CommandParser
	msgQueue      generator.ConsoleIOQueue
	store         store.DataStoreI
	logFile       string
	bindFile      string
	cfgPath       string
	eventCount    int
}

func New(logFile, bindFile, cfgPath string) BindGenerator {
	g := BindGenerator{
		logMsgChan:    make(chan string),
		stopChan:      make(chan interface{}, 1),
		evtChan:       make(chan *model.LogEvent),
		sigChan:       make(chan os.Signal, 1),
		state:         model.NewPlayerState(),
		store:         &store.SQLiteDataStore{},
		commandParser: generator.NewCommandParser("!"),
		logFile:       logFile,
		cfgPath:       cfgPath,
		bindFile:      bindFile,
	}
	signal.Notify(g.sigChan, os.Interrupt)
	g.parser = generator.NewLogParser(g.logMsgChan, g.evtChan)
	g.msgQueue = generator.NewConsoleIOQueue(cfgPath, g.stopChan)
	return g
}

func (bg *BindGenerator) Start() {
	// Read in our bind templates to memory
	f, err := os.Open(bg.bindFile)
	if err != nil {
		log.Errorf("Failed to open %s")
		return
	}
	if err := generator.ReadBinds(bufio.NewReader(f)); err != nil {
		log.Errorf("Failed to initialize bind data")
		os.Exit(1)
	}
	// Log file reader "tail -f"
	go func() {
		if err := generator.FileReader(bg.logFile, bg.logMsgChan, bg.stopChan); err != nil {
			log.Errorf("Error reading console.log: %s", err.Error())
		}
	}()

	go bg.parser.Start(bg.stopChan)
	go bg.StartHandler(bg.stopChan)
	go bg.msgQueue.Start()

	select {
	case <-bg.sigChan:
		log.Infof("Shutting down")
		bg.stopChan <- true
	}
	_ = bg.store.Close()
}
func (bg *BindGenerator) handleEvtKill(evt *model.LogEvent) {
	log.WithFields(log.Fields{
		"player": evt.Player,
		"victim": evt.Victim,
		"weapon": evt.Weapon,
		"class":  generator.GetPlayerClass(evt.Weapon),
		"crit":   evt.IsCritical,
	}).Infof("Kill event")
	if bg.state.IsSelf(evt.Player) {
		msg, err := generator.GenerateMessage(evt)
		if err != nil {
			return
		}
		log.Debugf(msg)
		bg.eventCount++
		bg.msgQueue.Add(msg)
	}
}

func (bg *BindGenerator) handleEvtMsg(evt *model.LogEvent) {
	log.WithFields(log.Fields{"name": evt.Player}).Printf("%s", evt.Message)
}

func (bg *BindGenerator) handleEvtClearCfg(evt *model.LogEvent) {
	if err := bg.msgQueue.ClearCfg(); err != nil {
		log.Errorf("Error trying to clear bind config: %s", err.Error())
	}
}

func (bg *BindGenerator) handleEvtStatusId(evt *model.LogEvent) {

}

func (bg *BindGenerator) StartHandler(stopChan chan interface{}) {
	for {
		select {
		case evt := <-bg.evtChan:
			{
				switch evt.Type {
				case consts.EvtKill:
					bg.handleEvtKill(evt)
				case consts.EvtMsg:
					bg.handleEvtMsg(evt)
					c, err := bg.commandParser.ParseMsg(evt.Message)
					if err != nil {
						return
					}

				case consts.EvtConnect:
					// On connect player only set us as the player on first connection.
					// subsequent connections are always other state
					if bg.state.IsSelf("") {
						log.WithFields(log.Fields{"name": evt.Player}).Infof("Detected local player name")
						bg.state.SetPlayer(evt.Player)
					}
				case consts.EvtDisconnect:
					log.WithFields(log.Fields{"player": evt.Player}).Infof("Disconnected from server.")
					//bg.state = model.NewPlayerState()
				case consts.EvtStatusId:
					bg.handleEvtStatusId(evt)
				case consts.EvtClearCfg:
					bg.handleEvtClearCfg(evt)
				}
			}
		case <-stopChan:
			return
		}
	}
}
