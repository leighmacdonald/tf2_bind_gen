package steam

import (
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseStatus(t *testing.T) {
	s := `version : 4620606/24 4620606 secure
hostname: Valve Matchmaking Server (Washington mwh-1/srcds135 #48)
udp/ip  : 192.69.97.58:27062  (public ip: 192.69.97.58)
steamid : [A:1:729372672:10372] (90116540677576704)
map     : koth_suijin at: 0 x, 0 y, 0 z
account : not logged in  (No account specified)
tags    : cp,hidden,increased_maxplayers,valve
players : 24 humans, 0 bots (32 max)
edicts  : 731 used of 2048 max
# userid name                uniqueid            connected ping loss state
#      2 "WolfXine"          [U:1:166779318]     15:22       85    0 active
#      3 "mdaniels5746"      [U:1:361821288]     15:22       87    0 active
#     28 "KRGonzales"        [U:1:875620767]     00:29       76   10 active
#      4 "juan.martinez2009" [U:1:79002518]      15:22       72    0 active
#      9 "Luuá¸°e"           [U:1:123675776]     15:18      109    0 active
#      5 "[LBJ] â™› King James â™›" [U:1:87772789] 15:22     76    0 active
#     10 "elirobot"          [U:1:167562538]     15:18       91    0 active
#      6 "guy (actual human feces)" [U:1:416855641] 15:22    83    0 active
#     26 "=/TFA\= CosmicTat" [U:1:163325254]     00:38       94    0 active
#      7 "alterego312"       [U:1:242237960]     15:22      128    0 active
#     12 "KcTheCray"         [U:1:332143895]     15:17       90    0 active
#      8 "Fag Bag McGee | Trade.tf" [U:1:861259628] 15:22   127    0 active
#     13 "Prototype x1-5150" [U:1:339990071]     15:17       77    0 active
#     14 "VAVI"              [U:1:122890196]     15:09      132    0 active
#     15 "Mecha Engineer Alfredo" [U:1:196165302] 15:06     132    0 active
#     16 "Ceebee324"         [U:1:132135410]     14:45      102    0 active
#     19 "Lil Dave"          [U:1:123147588]     14:39       87    0 active
#     22 "Stede Bonnet the pirate" [U:1:206922652] 10:37    165    0 active
#     20 "hard aim pootis serbia" [U:1:49974197] 14:13       84    0 active
#     18 "Enderz"            [U:1:202535707]     14:41       83    0 active
#     23 "WAFFLEDUDE"        [U:1:878783526]     10:33      128    0 active
#     24 "smokehousesteve"   [U:1:130361378]     09:54      128    0 active
#     29 "à¸¸"               [U:1:123868297]     00:24       59    0 active
#     27 "Cyndaquil"         [U:1:198198697]     00:31      131    0 active
`
	ids := SIDSFromStatus(s)
	assert.NotNil(t, ids)
	assert.Equal(t, len(ids), 24)
}

func TestConversions(t *testing.T) {
	assert.Equal(t, SID3ToSID32("[U:1:172346362]"), SID32(172346362))
	assert.Equal(t, SID3ToSID64("[U:1:172346362]"), SID64(76561198132612090))
	assert.Equal(t, SID3ToSID("[U:1:172346362]"), SID("STEAM_0:0:86173181"))
	assert.Equal(t, SID32ToSID3(172346362), SID3("[U:1:172346362]"))
	assert.Equal(t, SID32ToSID64(172346362), SID64(76561198132612090))
	assert.Equal(t, SID32ToSteamID(172346362), SID("STEAM_0:0:86173181"))
	assert.Equal(t, SID64ToSID3(76561198132612090), SID3("[U:1:172346362]"))
	assert.Equal(t, SID64ToSID32(76561198132612090), SID32(172346362))
	assert.Equal(t, SID64ToSteamID(76561198132612090), SID("STEAM_0:0:86173181"))
	assert.Equal(t, SIDToSID3("STEAM_0:0:86173181"), SID3("[U:1:172346362]"))
	assert.Equal(t, SIDToSID32("STEAM_0:0:86173181"), SID32(172346362))
	assert.Equal(t, SIDToSID64("STEAM_0:0:86173181"), SID64(76561198132612090))
}

func TestResolveGID(t *testing.T) {
	gid1 := ResolveGID("SQTreeHouse")
	assert.True(t, gid1.Valid())
	assert.Equal(t, gid1, GID(103582791441572968))
	gid2 := ResolveGID("SQTreeHouseHJHJHSDAF")
	assert.False(t, gid2.Valid())
}

func TestResolveSID(t *testing.T) {
	apiKey := viper.GetString("steam_api_key")
	if apiKey == "" {
		t.Skip("steam_api_key unset")
	}
	sid1 := ResolveSID64("https://steamcommunity.com/id/SQUIRRELLY")
	assert.Equal(t, sid1, SID64(76561197961279983))

	sid2 := ResolveSID64("https://steamcommunity.com/id/FAKEXXXXXXXXXX123123")
	assert.False(t, sid2.Valid())

	sid3 := ResolveSID64("http://steamcommunity.com/profiles/76561197961279983")
	assert.Equal(t, sid3, SID64(76561197961279983))

	sid4 := ResolveSID64("[U:1:1014255]")
	assert.Equal(t, sid4, SID64(76561197961279983))

	sid5 := ResolveSID64("STEAM_0:1:507127")
	assert.Equal(t, sid5, SID64(76561197961279983))

	sid6 := ResolveSID64("")
	assert.False(t, sid6.Valid())

}
