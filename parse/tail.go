package parse

import (
	"github.com/hpcloud/tail"
	"strings"
)

func FileReader(path string, c chan string, s chan interface{}) error {
	t, err := tail.TailFile(path, tail.Config{Follow: true})
	if err != nil {
		return err
	}
	for line := range t.Lines {
		if line.Text != "" {
			c <- strings.ReplaceAll(line.Text, "\r", "")
		}
	}
	return err
}
