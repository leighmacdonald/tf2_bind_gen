package parse

import (
	"bind_generator/consts"
	"bind_generator/model"
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"math/rand"
	"regexp"
	"strings"
	"text/template"
)

var bindsMap map[string][]*template.Template

func WriteBindFile(path string, msg string) error {
	bind := fmt.Sprintf(`alias bind_gen "say %s"\n`, msg)
	log.Debugf("Writing alias data: %s", bind)
	return ioutil.WriteFile(path, []byte(bind), 0644)
}

var rxBindTemplate = regexp.MustCompile(`^\[(.+?)]\s+?(.+?)$`)

func ReadBinds(path string) error {
	binds := make(map[string][]*template.Template)
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	bindId := 0
	for _, line := range strings.Split(string(b), "\n") {
		if m := rxBindTemplate.FindStringSubmatch(line); m != nil {
			log.Println(m)
			if len(m) != 3 {
				log.Warnf("Failed to fully match bind: %s", line)
				continue
			}
			killType := strings.ToLower(m[1])
			t, err := template.New(fmt.Sprintf("t_%d", bindId)).Parse(m[2])
			if err != nil {
				log.Errorf("Failed to parse template: %s", err.Error())
				continue
			}
			binds[killType] = append(binds[killType], t)
			bindId++
		} else {
			log.Warnf("Failed to parse bind: %s", line)
		}
	}
	log.Debugf("Parsed %d binds", bindId)
	bindsMap = binds
	return nil
}

func generateMessage(event model.LogEvent) (string, error) {
	killType := string(event.Weapon)
	if event.IsCritical {
		killType += ".crit"
	}
	var tmpl *template.Template
	f, found := bindsMap[killType]
	// Look for a match9ng specific event key
	// default to "generic" if no templates are found for the event
	if found {
		tmpl = f[rand.Intn(len(f))]
	} else {
		g, found := bindsMap["generic"]
		if !found {
			return "", consts.ErrNoTemplate
		}
		tmpl = g[rand.Intn(len(g))]
	}
	buf := new(bytes.Buffer)
	if err := tmpl.Execute(buf, event); err != nil {
		return "", err
	}
	msg := buf.String()
	if msg == "" {
		return msg, consts.ErrEmptyResult
	} else {
		return msg, nil
	}

}

func init() {
	bindsMap = make(map[string][]*template.Template)
}
