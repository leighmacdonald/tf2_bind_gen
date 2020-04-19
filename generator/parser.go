package generator

import (
	"bind_generator/consts"
	"bind_generator/model"
	"fmt"
	"github.com/leighmacdonald/steamid"
	log "github.com/sirupsen/logrus"
	"regexp"
	"strings"
)

type LogParser struct {
	evtChan       chan *model.LogEvent
	ReadChannel   chan string
	rxKill        *regexp.Regexp
	rxMsg         *regexp.Regexp
	rxConnected   *regexp.Regexp
	rxDisconnect  *regexp.Regexp
	rxStatusID    *regexp.Regexp
	rxLobbyPlayer *regexp.Regexp
	rxClearCfg    *regexp.Regexp
	rx            []*regexp.Regexp
}

func (l *LogParser) ParseEvent(msg string) (*model.LogEvent, error) {
	// the i index must to the EventType const values
	for i, rx := range l.rx {
		if m := rx.FindStringSubmatch(msg); m != nil {
			t := consts.EventType(i)
			le := model.NewLogEvent(t)
			switch t {
			case consts.EvtLobbyPlayerTeam:
				le.PlayerSID = steamid.SID3ToSID64(steamid.SID3(m[1]))
				if m[2] == "DEFENDER" {
					le.Team = consts.RED
				} else {
					le.Team = consts.BLU
				}
			case consts.EvtConnect:
				le.Player = m[1]
			case consts.EvtDisconnect:
				le.Player = m[1]
			case consts.EvtMsg:
				le.Player = m[1]
				le.Message = m[2]
			case consts.EvtStatusId:
			case consts.EvtKill:
				le.Player = m[1]
				le.Victim = m[2]
				le.Weapon = consts.Weapon(m[3])
				le.IsCritical = strings.HasSuffix(msg, "(crit)")
			case consts.EvtClearCfg:

			}
			return le, nil
		}
	}
	return nil, consts.ErrEmptyResult
}

func (l *LogParser) Start(stopChan chan interface{}) {
	for {
		select {
		case msg := <-l.ReadChannel:
			e, err := l.ParseEvent(msg)
			if err != nil {
				continue
			}
			l.evtChan <- e
		case <-stopChan:
			log.Debug("Stopped LogParser")
			return
		}
	}
}

func NewLogParser(readChannel chan string, evtChan chan *model.LogEvent) *LogParser {
	lp := &LogParser{
		evtChan:     evtChan,
		ReadChannel: readChannel,
		rxMsg:       regexp.MustCompile(`^(.+?)\s:\s\s(.+?)$`),
		rxKill:      regexp.MustCompile(`^(.+?)\skilled\s(.+?)\swith\s(.+)(\.|\. \(crit\))$`),
		rxConnected: regexp.MustCompile(`(?:.+?\.)?(\S+)\sconnected$`),
		//rxConnectedAlt: regexp.MustCompile(`(.+?\.)?(\S+)\sconnected$`),
		rxDisconnect:  regexp.MustCompile(`(^Disconnecting from abandoned match server$|\(Server shutting down\)$)`),
		rxStatusID:    regexp.MustCompile(`^#\s+(?P<id>\d+)\s"(?P<name>.+?)"\s+(?P<sid>\[U:\d:\d+])\s+(?P<time>\d+:\d+)\s+(?P<ping>\d+)\s+(?P<loss>\d+)`),
		rxLobbyPlayer: regexp.MustCompile(`\s+(Member|Pending)\[\d+]\s+(?P<sid>\[.+?]).+?TF_GC_TEAM_(?P<team>(DEFENDERS|INVADERS))`),
		rxClearCfg:    regexp.MustCompile(fmt.Sprintf(`^%s`, consts.ClearConfigMsg)),
	}
	lp.rx = []*regexp.Regexp{lp.rxKill, lp.rxMsg, lp.rxConnected, lp.rxDisconnect, lp.rxStatusID}
	return lp
}

func GetPlayerClass(weapon consts.Weapon) consts.PlayerClass {
	for pClass, weapons := range consts.Weapons {
		for _, w := range weapons {
			if weapon == w {
				return pClass
			}
		}
	}
	return consts.Multi
}
