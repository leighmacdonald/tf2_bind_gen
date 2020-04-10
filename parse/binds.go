package parse

import (
	"bind_generator/consts"
	"bind_generator/google"
	"bind_generator/model"
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"math/rand"
	"net/url"
	"regexp"
	"strings"
	"text/template"
)

const bindFmt = "echo \"Loaded bind_generator.cfg\"; alias bind_gen \"say %s\"\n"

type bindTemplate struct {
	Tmpl       *template.Template
	queryTypes []consts.QueryType
}

var bindsMap map[string][]bindTemplate

// exec bind_gen/scratch.cfg
// bind_gen
// con_logfile bind_gen/scratch_cfg
// echo alias bindgen ""
// con_logfile console.log
//
func WriteBindFile(path string, msg string) error {
	bind := fmt.Sprintf(bindFmt, msg)
	log.Debugf("Writing alias data: %s", bind)
	return ioutil.WriteFile(path, []byte(bind), 0644)
}

var rxBindTemplate = regexp.MustCompile(`^\[(.+?)]\s+?(.+?)$`)
var rxBindGoogleResult = regexp.MustCompile(`\$google_result\s+`)

func getQueryTypes(b string) []consts.QueryType {
	bl := strings.ToLower(b)
	var qt []consts.QueryType
	if rxBindGoogleResult.MatchString(bl) {
		qt = append(qt, consts.GoogleResult)
	}
	return qt
}

func ReadBinds(r io.Reader) error {
	binds := make(map[string][]bindTemplate)
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	bindId := 0
	for _, line := range strings.Split(string(b), "\n") {
		if m := rxBindTemplate.FindStringSubmatch(line); m != nil {
			if len(m) != 3 {
				log.Warnf("Failed to fully match bind: %s", line)
				continue
			}
			bindStr := m[2]
			killType := strings.ToLower(m[1])
			t, err := template.New(fmt.Sprintf("t_%d", bindId)).Parse(bindStr)
			if err != nil {
				log.Warnf("Failed to parse template: %s", err.Error())
				continue
			}
			binds[killType] = append(binds[killType], bindTemplate{
				Tmpl:       t,
				queryTypes: getQueryTypes(bindStr),
			})
			bindId++
		} else {
			log.Warnf("Failed to parse bind: %s", line)
		}
	}
	log.Infof("Loaded %d binds", bindId)
	bindsMap = binds
	return nil
}

func getTemplate(event *model.LogEvent) (bindTemplate, error) {
	killType := string(event.Weapon)
	if event.IsCritical {
		killType += ".crit"
	}
	var bindTmpl bindTemplate
	f, found := bindsMap[killType]
	// Look for a matching specific event key
	// default to "generic" if no templates are found for the event
	if found {
		bindTmpl = f[rand.Intn(len(f))]
	} else {
		g, found := bindsMap["generic"]
		if !found {
			return bindTmpl, consts.ErrNoTemplate
		}
		bindTmpl = g[rand.Intn(len(g))]
	}
	return bindTmpl, nil
}

func simplifyUrl(inURL string) string {
	u, err := url.Parse(inURL)
	if err != nil {
		log.Warnf("Failed to parse url for simplification: %s", inURL)
		return inURL
	}
	u.RawQuery = ""
	return strings.Replace(strings.Replace(u.String(), "https://", "", 1), "http://", "", 1)
}

func GenerateMessage(event *model.LogEvent) (string, error) {
	tryCount := 0
	// Loop until we successfully get a bind
	for tryCount < consts.MaxBindGenAttempts {
		tryCount++
		bindTmpl, e := getTemplate(event)
		if e != nil {
			continue
		}
		buf := new(bytes.Buffer)
		if err := bindTmpl.Tmpl.Execute(buf, event); err != nil {
			return "", err
		}
		msg := buf.String()
		if msg == "" {
			continue
		}
		for _, qt := range bindTmpl.queryTypes {
			switch qt {
			case consts.GoogleResult:
				res, err := google.Search(event.Victim)
				if err != nil || len(res) == 0 {
					continue
				}
				idx := 0
				if len(res) > 10 {
					idx = rand.Intn(10)
				}
				msg = strings.Replace(msg, string(consts.GoogleResultVar), simplifyUrl(res[idx].ResultURL), 1)
			}
		}
		return msg, nil
	}
	return "", consts.ErrEmptyResult
}

func init() {
	bindsMap = make(map[string][]bindTemplate)
}
