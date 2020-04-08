package app

import (
	"bind_generator/consts"
	"bind_generator/model"
	"bind_generator/parse"
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
}

func New() BindGenerator {
	g := BindGenerator{
		logMsgChan: make(chan string),
		stopChan:   make(chan interface{}, 1),
		evtChan:    make(chan *model.LogEvent),
		sigChan:    make(chan os.Signal, 1),
		players:    model.NewPlayerSet(),
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
	if err := parse.ReadBinds("binds_custom.txt"); err != nil {
		log.Errorf("Failed to initialize bind data")
		os.Exit(1)
	}
	logPath := `S:\Steam\steamapps\common\Team Fortress 2\tf\console_test.log`
	// Log file reader "tail -f"
	go func() {
		if err := parse.FileReader(logPath, bg.logMsgChan, bg.stopChan); err != nil {
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

func (bg *BindGenerator) StartHandler(stopChan chan interface{}) {
	for {
		select {
		case evt := <-bg.evtChan:
			{
				switch evt.Type {
				case consts.EvtKill:
					log.WithFields(log.Fields{
						"player": evt.Player,
						"victim": evt.Victim,
						"weapon": evt.Weapon,
						"class":  parse.GetPlayerClass(evt.Weapon),
						"crit":   evt.IsCritical,
					}).Infof("Kill event")

				// Write bind to file
				case consts.EvtMsg:
					log.WithFields(log.Fields{"name": evt.Player}).Printf("%s", evt.Message)
				case consts.EvtConnect:
					// On connect player only set us as the player on first connection.
					// subsequent connections are always other players
					if bg.players.Player == "" {
						log.WithFields(log.Fields{"name": evt.Player}).Infof("Detected local player name")
						bg.players.Player = evt.Player
					}
				case consts.EvtDisconnect:
					log.Infof("Disconnected from server.")
					bg.players = model.NewPlayerSet()
				case consts.EvtStatusId:
				}
			}
		case <-stopChan:
			return
		}

	}
}
