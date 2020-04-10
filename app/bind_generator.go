package app

import (
	"bind_generator/consts"
	"bind_generator/model"
	"bind_generator/parse"
	"bufio"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
)

type BindGenerator struct {
	logMsgChan chan string
	stopChan   chan interface{}
	evtChan    chan *model.LogEvent
	sigChan    chan os.Signal
	players    model.PlayerSet
	parser     *parse.LogParser
	logFile    string
	bindFile   string
	cfgPath    string
}

func New(logFile, bindFile, cfgPath string) BindGenerator {
	g := BindGenerator{
		logMsgChan: make(chan string),
		stopChan:   make(chan interface{}, 1),
		evtChan:    make(chan *model.LogEvent),
		sigChan:    make(chan os.Signal, 1),
		players:    model.NewPlayerSet(),
		logFile:    logFile,
		cfgPath:    cfgPath,
		bindFile:   bindFile,
	}
	signal.Notify(g.sigChan, os.Interrupt)
	g.parser = parse.NewLogParser(g.logMsgChan, g.evtChan)
	log.SetFormatter(&log.TextFormatter{
		ForceColors: true,
	})
	return g
}

func (bg *BindGenerator) Start() {
	// Read in our bind templates to memory
	f, err := os.Open(bg.bindFile)
	if err != nil {
		log.Errorf("Failed to open %s")
		return
	}
	if err := parse.ReadBinds(bufio.NewReader(f)); err != nil {
		log.Errorf("Failed to initialize bind data")
		os.Exit(1)
	}

	// Log file reader "tail -f"
	go func() {
		if err := parse.FileReader(bg.logFile, bg.logMsgChan, bg.stopChan); err != nil {
			log.Errorf("Error reading console.log: %s", err.Error())
		}
	}()

	go bg.parser.Start(bg.stopChan)
	go bg.StartHandler(bg.stopChan)

	select {
	case <-bg.sigChan:
		log.Infof("Shutting down")
		bg.stopChan <- true
	}
}
func (bg *BindGenerator) handleEvtKill(evt *model.LogEvent) {
	log.WithFields(log.Fields{
		"player": evt.Player,
		"victim": evt.Victim,
		"weapon": evt.Weapon,
		"class":  parse.GetPlayerClass(evt.Weapon),
		"crit":   evt.IsCritical,
	}).Infof("Kill event")
	if evt.Player == bg.players.Player {
		msg, err := parse.GenerateMessage(evt)
		if err != nil {
			return
		}
		log.Debugf(msg)
		if err := parse.WriteBindFile(bg.cfgPath, msg); err != nil {
			log.Errorf("Failed to write binds to file: %s", err.Error())
		}
	}
}
func (bg *BindGenerator) handleEvtMsg(evt *model.LogEvent) {
	log.WithFields(log.Fields{"name": evt.Player}).Printf("%s", evt.Message)
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
				case consts.EvtConnect:
					// On connect player only set us as the player on first connection.
					// subsequent connections are always other players
					if bg.players.Player == "" {
						log.WithFields(log.Fields{"name": evt.Player}).Infof("Detected local player name")
						bg.players.Player = evt.Player
					}
				case consts.EvtDisconnect:
					log.WithFields(log.Fields{"player": evt.Player}).Infof("Disconnected from server.")
					//bg.players = model.NewPlayerSet()
				case consts.EvtStatusId:
				}
			}
		case <-stopChan:
			return
		}
	}
}
