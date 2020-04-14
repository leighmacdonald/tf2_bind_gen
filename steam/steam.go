package steam

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
)

const (
	urlVanity = "http://api.steampowered.com/ISteamUser/ResolveVanityURL/v0001/?"
)

var (
	reStatusID    *regexp.Regexp
	reGroupIDTags *regexp.Regexp
	reSID         *regexp.Regexp
	reSID64       *regexp.Regexp
	reSID3        *regexp.Regexp
	reGroupURL    *regexp.Regexp
)

type AppID int64
type SID string   // STEAM_0:0:86173181
type SID64 uint64 // 76561198132612090
type SID32 uint32 // 172346362
type SID3 string  // [U:1:172346362]
type GID uint64   // 103582791453729676

type GIDs []GID
type SID64s []SID64
type AppIDs []AppID

func (t GID) String() string {
	return strconv.FormatInt(int64(t), 10)
}

func (t SID64) String() string {
	return strconv.FormatInt(int64(t), 10)
}

func (t SID64) Int64() int64 {
	return int64(t)
}

func (t SID64) Valid() bool {
	return t > 0
}

var idGen uint64 = 76561198206310000

// GSID generated a unique random (numerically) valid steamid for testing
func GSID() SID64 {
	return SID64(atomic.AddUint64(&idGen, 1))
}

func SID64FromString(steamId string) SID64 {
	if steamId == "" {
		return SID64(0)
	}
	i, err := strconv.ParseInt(steamId, 10, 64)
	if err != nil {
		return SID64(0)
	}
	return SID64(i)
}

func GIDFromString(steamId string) GID {
	if steamId == "" {
		return GID(0)
	}
	i, err := strconv.ParseInt(steamId, 10, 64)
	if err != nil {
		return GID(0)
	}
	return GID(i)
}

func (t GID) Valid() bool {
	return t > 0
}

func (t GID) Int64() int64 {
	return int64(t)
}

func SID64In(sid SID64, list SID64s) bool {
	for _, b := range list {
		if b == sid {
			return true
		}
	}
	return false
}

// SIDToSID64 converts a given SteamID to a SID64.
// eg. STEAM_0:0:86173181 -> 76561198132612090
//
// 0 is returned if the process was unsuccessful.
func SIDToSID64(steamID SID) SID64 {
	idParts := strings.Split(string(steamID), ":")
	magic, _ := new(big.Int).SetString("76561197960265728", 10)
	steam64, _ := new(big.Int).SetString(idParts[2], 10)
	steam64 = steam64.Mul(steam64, big.NewInt(2))
	steam64 = steam64.Add(steam64, magic)
	auth, _ := new(big.Int).SetString(idParts[1], 10)
	return SID64(steam64.Add(steam64, auth).Int64())
}

// SIDToSID32 converts a given SteamID to a SID32.
// eg. STEAM_0:0:86173181 -> 172346362
//
// 0 is returned if the process was unsuccessful.
func SIDToSID32(steamID SID) SID32 {
	return SID64ToSID32(SIDToSID64(steamID))
}

// SIDToSID3 converts a given SteamID to a SID3.
// eg. STEAM_0:0:86173181 -> [U:1:172346362]
//
// An empty SID3 (string) is returned if the process was unsuccessful.
func SIDToSID3(steamID SID) SID3 {
	steamIDParts := strings.Split(string(steamID), ":")
	steamLastPart, err := strconv.ParseUint(string(steamIDParts[len(steamIDParts)-1]), 10, 64)
	if err != nil {
		return SID3("")
	}

	return SID3("[U:1:" + strconv.FormatUint(steamLastPart*2, 10) + "]")
}

// SID64ToSteamID converts a given SID64 to a SteamID.
// eg. 76561198132612090 -> STEAM_0:0:86173181
//
// An empty SteamID (string) is returned if the process was unsuccessful.
func SID64ToSteamID(steam64 SID64) SID {
	steamID := new(big.Int).SetInt64(int64(steam64))
	magic, _ := new(big.Int).SetString("76561197960265728", 10)
	steamID = steamID.Sub(steamID, magic)
	isServer := new(big.Int).And(steamID, big.NewInt(1))
	steamID = steamID.Sub(steamID, isServer)
	steamID = steamID.Div(steamID, big.NewInt(2))
	return SID("STEAM_0:" + isServer.String() + ":" + steamID.String())
}

// SID64ToSID32 converts a given SID64 to a SID32.
// eg. 76561198132612090 -> 172346362
//
// 0 is returned if the process was unsuccessful.
func SID64ToSID32(steam64 SID64) SID32 {
	steam64Str := strconv.FormatUint(uint64(steam64), 10)
	if len(steam64Str) < 3 {
		return 0
	}
	steam32, err := strconv.ParseInt(steam64Str[3:], 10, 64)
	if err != nil {
		return 0
	}
	return SID32(steam32 - 61197960265728)
}

// SID64ToSID3 converts a given SID64 to a SID3.
// eg. 76561198132612090 -> [U:1:172346362]
//
// An empty SID3 (string) is returned if the process was unsuccessful.
func SID64ToSID3(steam64 SID64) SID3 {
	steamID := SID64ToSteamID(steam64)
	if steamID == SID(0) {
		return SID3("")
	}

	return SIDToSID3(steamID)
}

// SID32ToSteamID converts a given SID32 to a SteamID.
// eg. 172346362 -> STEAM_0:0:86173181
//
// An empty SteamID (string) is returned if the process was unsuccessful.
func SID32ToSteamID(steam32 SID32) SID {
	return SID64ToSteamID(SID32ToSID64(steam32))
}

// SID32ToSID64 converts a given SID32 to a SID64.
// eg. 172346362 -> 76561198132612090
//
// 0 is returned if the process was unsuccessful.
func SID32ToSID64(steam32 SID32) SID64 {
	steam64, err := strconv.ParseInt("765"+strconv.FormatInt(int64(steam32)+61197960265728, 10), 10, 64)
	if err != nil {
		return 0
	}
	return SID64(steam64)
}

// SID32ToSID3 converts a given SID32 to a SID3.
// eg. 172346362 -> [U:1:172346362]
//
// An empty SID3 (string) is returned if the process was unsuccessful.
func SID32ToSID3(steam32 SID32) SID3 {
	steamID := SID32ToSteamID(steam32)
	if steamID == SID(0) {
		return ""
	}
	return SIDToSID3(steamID)
}

// SID3ToSID converts a given SID3 to a SteamID.
// eg. [U:1:172346362] -> STEAM_0:0:86173181
//
// An empty SteamID (string) is returned if the process was unsuccessful.
func SID3ToSID(steam3 SID3) SID {
	IDparts := strings.Split(string(steam3), ":")

	id32 := IDparts[len(IDparts)-1]

	if len(id32) <= 0 {
		return ""
	}

	if id32[len(id32)-1:] == "]" {
		id32 = id32[:len(id32)-1]
	}

	steam32, err := strconv.ParseUint(id32, 10, 64)
	if err != nil {
		return ""
	}

	return SID32ToSteamID(SID32(steam32))
}

// SID3ToSID64 converts a given SID3 to a SID64.
// eg. [U:1:172346362] -> 76561198132612090
//
// 0 is returned if the process was unsuccessful.
func SID3ToSID64(steam3 SID3) SID64 {
	idParts := strings.Split(string(steam3), ":")

	id32 := idParts[len(idParts)-1]

	if len(id32) <= 0 {
		return SID64(0)
	}

	if id32[len(id32)-1:] == "]" {
		id32 = id32[:len(id32)-1]
	}

	steam32, err := strconv.ParseUint(id32, 10, 64)
	if err != nil {
		return SID64(0)
	}

	return SID32ToSID64(SID32(steam32))
}

// SID3ToSID64 converts a given SID3 to a SID64.
// eg. [U:1:172346362] -> 172346362
//
// 0 is returned if the process was unsuccessful.
func SID3ToSID32(steam3 SID3) SID32 {
	IDparts := strings.Split(string(steam3), ":")

	id32 := IDparts[len(IDparts)-1]

	if len(id32) <= 0 {
		return SID32(0)
	}

	if id32[len(id32)-1:] == "]" {
		id32 = id32[:len(id32)-1]
	}

	steam32, err := strconv.ParseUint(id32, 10, 64)
	if err != nil {
		return SID32(0)
	}

	return SID32(steam32)
}

func SIDSFromStatus(text string) SID64s {
	var ids SID64s
	found := reStatusID.FindAllString(text, -1)
	if found == nil {
		return nil
	}
	for _, strId := range found {
		ids = append(ids, SID3ToSID64(SID3(strId)))
	}
	return ids
}

// ResolveGroupID tried to resolve the GroupID from a group custom url.
func ResolveGID(groupVanityURL string) GID {
	m := reGroupURL.FindStringSubmatch(groupVanityURL)
	if len(m) > 0 {
		groupVanityURL = m[1]
	}
	resp, err := http.Get("https://steamcommunity.com/groups/" + groupVanityURL + "/memberslistxml?xml=1")
	if err != nil {
		log.Println("Failed to fetch GID", err)
		return GID(0)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.WithError(err).Errorln("Failed to close body")
		}
	}()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Failed to read GID body", err)
		return GID(0)
	}
	groupIDTags := reGroupIDTags.FindSubmatch(content)
	if len(groupIDTags) >= 2 {
		groupid, err := strconv.ParseUint(string(groupIDTags[1]), 10, 64)
		if err != nil {
			log.Println("Failed to parse GID", err)
			return GID(0)
		}
		return GID(groupid)
	}
	return GID(0)
}

// SearchForID tries to retrieve a SteamID64 using a query (search).
//
// If an error occurs or the SteamID was unable to be resolved from the query then a 0 is returned.
func ResolveSID64(query string) SID64 {
	query = strings.Replace(query, " ", "", -1)

	if strings.Index(query, "steamcommunity.com/profiles/") != -1 {
		if string(query[len(query)-1]) == "/" {
			query = query[0 : len(query)-1]
		}

		output, err := strconv.ParseInt(query[strings.Index(query, "steamcommunity.com/profiles/")+len("steamcommunity.com/profiles/"):], 10, 64)
		if err != nil {
			return SID64(0)
		}

		query = strings.Replace(query, "/", "", -1)

		if len(strconv.FormatInt(output, 10)) != 17 {
			return SID64(0)
		}

		return SID64(output)
	} else if strings.Index(query, "steamcommunity.com/id/") != -1 {
		if string(query[len(query)-1]) == "/" {
			query = query[0 : len(query)-1]
		}

		query = query[strings.Index(query, "steamcommunity.com/id/")+len("steamcommunity.com/id/"):]
		apiKey := viper.GetString("steam_api_key")
		log.Panic("Unimplemented api key resolver")
		resp, err := http.Get(urlVanity + url.Values{
			"key":       {apiKey},
			"vanityurl": {query},
		}.Encode())
		if err != nil {
			return SID64(0)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				log.WithError(err).Errorln("Failed to close body")
			}
		}()

		content, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return SID64(0)
		}

		var vanityUrlResponse struct {
			Response struct {
				Steamid string
				Success int
			}
		}

		if err := json.Unmarshal(content, &vanityUrlResponse); err != nil {
			return SID64(0)
		}

		if vanityUrlResponse.Response.Success != 1 {
			return SID64(0)
		}

		if len(vanityUrlResponse.Response.Steamid) != 17 {
			return SID64(0)
		}

		output, err := strconv.ParseInt(vanityUrlResponse.Response.Steamid, 10, 64)
		if err != nil {
			return SID64(0)
		}

		return SID64(output)
	} else if reSID.MatchString(query) {
		steam64 := SIDToSID64(SID(query))

		if len(strconv.FormatUint(uint64(steam64), 10)) != 17 {
			return SID64(0)
		}

		return SID64(steam64)
	} else if reSID64.MatchString(query) && len(query) == 17 {
		output, err := strconv.ParseInt(query, 10, 64)
		if err != nil {
			return SID64(0)
		}

		if len(strconv.FormatInt(output, 10)) != 17 {
			return SID64(0)
		}

		return SID64(output)
	} else if reSID3.MatchString(strings.ToUpper(query)) {
		return SID3ToSID64(SID3(query))
	}
	apiKey := viper.GetString("steam_api_key")
	log.Panic("Unimplemented api key resolver")
	resp, err := http.Get("http://api.steampowered.com/ISteamUser/ResolveVanityURL/v0001/?" + url.Values{
		"key":       {apiKey},
		"vanityurl": {query},
	}.Encode())
	if err != nil {
		return SID64(0)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.WithError(err).Errorln("Failed to close body")
		}
	}()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return SID64(0)
	}

	var vanityUrlResponse struct {
		Response struct {
			Steamid string
			Success int
		}
	}

	if err := json.Unmarshal(content, &vanityUrlResponse); err != nil {
		return SID64(0)
	}

	if vanityUrlResponse.Response.Success != 1 {
		return SID64(0)
	}

	if len(vanityUrlResponse.Response.Steamid) != 17 {
		return SID64(0)
	}

	output, err := strconv.ParseInt(vanityUrlResponse.Response.Steamid, 10, 64)
	if err != nil {
		return SID64(0)
	}

	return SID64(output)
}

func init() {
	reSID = regexp.MustCompile(`^STEAM_0:([01]):[0-9]{1}[0-9]{0,8}$`)
	reSID64 = regexp.MustCompile(`^\d+$`)
	reSID3 = regexp.MustCompile(`(\[)?U:1:\d+(\])?`)
	reGroupIDTags = regexp.MustCompile(`<groupID64>(\w+)</groupID64>`)
	reStatusID = regexp.MustCompile(`"(.+?)"\s+(\[U:\d+:\d+\]|STEAM_\d:\d:\d+)`)
	reGroupURL = regexp.MustCompile(`steamcommunity.com/groups/(\S+)/?`)
}
