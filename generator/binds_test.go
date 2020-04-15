package generator

import (
	"bind_generator/consts"
	"bind_generator/model"
	"github.com/stretchr/testify/assert"
	"log"
	"strings"
	"testing"
)

func TestGenerateMessage(t *testing.T) {
	binds := "[generic] $google_result 4 more hilarious deaths by {{ .Victim }}!"
	if err := ReadBinds(strings.NewReader(binds)); err != nil {
		t.Errorf(err.Error())
		return
	}
	e1 := model.NewLogEvent(consts.EvtKill)
	e1.Player = "player_1"
	e1.Victim = "victim_1"
	e1.Weapon = consts.ProjectileRocket
	m, e := GenerateMessage(e1)
	assert.NotNil(t, e, "Error generating message")
	log.Print(m)

}
