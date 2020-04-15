package generator

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewCommandParser(t *testing.T) {
	type testCases struct {
		Msg      string
		Expected UserCommand
	}
	tests := []testCases{
		{"!help", UserCommand{command: "help"}},
		{"!nextfact", UserCommand{command: "nextfact"}},
		{"!price", UserCommand{command: "price"}},
		{"!bp", UserCommand{command: "bp"}},
		{"!bp 76561197961279983", UserCommand{command: "bp", args: []string{"76561197961279983"}}},
		{"!bp [U:1:1014255]", UserCommand{command: "bp", args: []string{"76561197961279983"}}},
		{"!bp STEAM_0:1:507127", UserCommand{command: "bp", args: []string{"76561197961279983"}}},
		{"!help addfact", UserCommand{command: "help", args: []string{"addfact"}}},
	}
	cp := NewCommandParser("!")
	for _, tc := range tests {
		uc, err := cp.ParseMsg(tc.Msg)
		assert.NoError(t, err)
		assert.Equal(t, uc.command, tc.Expected.command)
		assert.EqualValues(t, uc.args, tc.Expected.args)
	}
}
