package model

import (
	"bind_generator/consts"
	"bind_generator/steam"
)

// Current known player state
type Player struct {
	Class consts.PlayerClass
	SID   steam.SID64
	Team  consts.Team
	// Times we have killed this player
	PersonalKillCount int
	KillCount         int
	DeathCount        int
}

type PlayerState struct {
	player  string
	players []*Player
}

func (s *PlayerState) Clear() {
	s.players = []*Player{}
}

func (s *PlayerState) SetPlayer(player string) bool {
	return s.player == player
}

func (s *PlayerState) IsSelf(player string) bool {
	return s.player == player
}

func NewPlayerState() PlayerState {
	return PlayerState{
		player: "",
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
	Team        consts.Team
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
