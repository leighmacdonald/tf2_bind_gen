package backpacktf

import (
	"bind_generator/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInventoryValues(t *testing.T) {
	config.InitConfig()
	if Authenticate() != nil {
		t.Skip("No token available")
		return
	}
	v, e := InventoryValues(76561197961279983)
	assert.NoError(t, e)
	assert.True(t, v.Value > 0)
	assert.True(t, v.MarketValue > 0)
}
