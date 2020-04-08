package model

import (
	"bind_generator/consts"
	"bind_generator/steam"
)

type PlayerSet struct {
	Player  string
	NameMap map[string]steam.SID64
	SIDMap  map[steam.SID64]string
}

func NewPlayerSet() PlayerSet {
	return PlayerSet{
		Player:  "",
		NameMap: make(map[string]steam.SID64),
		SIDMap:  make(map[steam.SID64]string),
	}
}

// Catchall struct for log events. Not all fields will be required
// for all event types
type LogEvent struct {
	Type        consts.EventType
	Player      string
	PlayerSID   steam.SID64
	Victim      string
	VictimSID   steam.SID64
	Weapon      consts.Weapon
	IsCritical  bool
	TimesKilled int
	Message     string
}

func NewLogEvent(t consts.EventType) *LogEvent {
	return &LogEvent{
		Type:        t,
		Victim:      "",
		VictimSID:   0,
		Weapon:      "",
		IsCritical:  false,
		TimesKilled: 0,
	}
}
