package google

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestScrape(t *testing.T) {
	res, err := Search("team fortress 2")
	assert.NoError(t, err, "Search query failure")
	assert.True(t, len(res) >= 100)
}

/**

bind w +mfwd
bind s +mback
bind a +mleft
bind d +mright

alias +mfwd "exec bind_generator; bind_gen;-back;+forward;alias checkfwd +forward;"
alias +mback "exec bind_generator; bind_gen;-forward;+back;alias checkback +back"
alias +mleft "exec bind_generator; bind_gen;-moveright;+moveleft;alias checkleft +moveleft"
alias +mright "exec bind_generator; bind_gen;-moveleft;+moveright;alias checkright +moveright"
alias -mfwd "-forward;checkback;alias checkfwd none"
alias -mback "-back;checkfwd;alias checkback none"
alias -mleft "-moveleft;checkright;alias checkleft none"
alias -mright "-moveright;checkleft;alias checkright none"
alias checkfwd none
alias checkback none
alias checkleft none
alias checkright none
alias none ""
*/
