package generator

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"testing"
)

type ConsoleGenerator struct {
	lines      []string
	outputFile *os.File
	stopped    bool
}

func NewConsoleGenerator(exampleLog io.Reader, outputPath string) ConsoleGenerator {
	data, err := ioutil.ReadAll(exampleLog)
	if err != nil {
		log.Panic("Failed to read console data")
	}
	l := strings.Split(string(data), "\n")
	if len(l) == 0 {
		log.Panic("Empty console data")
	}

	of, err := os.OpenFile(outputPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Panic("Failed to open outputPath")
	}
	return ConsoleGenerator{lines: l, outputFile: of, stopped: true}
}

func (c *ConsoleGenerator) Stop() {
	c.stopped = true
}

func (c *ConsoleGenerator) Start() {
	c.stopped = false
	defer func() { _ = c.outputFile.Close() }()
	for _, s := range c.lines {
		if c.stopped {
			return
		}
		if s == "" {
			continue
		}
		_, err := c.outputFile.WriteString(fmt.Sprintf("%s\n", s))
		if err != nil {
			log.Panic("failed to write output string")
		}
	}
}

func exists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func exampleData(name string) (io.Reader, error) {
	p := path.Join("testdata", name)
	if !exists(p) {
		p = path.Join("..", p)
	}
	if !exists(p) {
		log.Panic("Invalid example data name")
	}
	b, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}

func TestCommandParser(t *testing.T) {
	r, _ := exampleData("console_db_1round.log")
	f, err := ioutil.TempFile("", "sample")
	if err != nil {
		log.Panic("Failed to create temp file")
	}
	defer func() { _ = os.Remove(f.Name()) }()
	cg := NewConsoleGenerator(r, f.Name())
	go cg.Start()
	cg.Stop()
}
