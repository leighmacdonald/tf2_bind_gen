package parse

import (
	"bind_generator/consts"
	"strings"
)

var commandMap map[userCommandType]commandHandler

type userCommandType string

type commandHandler struct {
	handler func(args []string) error
	minArgs int
	maxArgs int
}

type userCommand struct {
	command userCommandType
	args    []string
}

type CommandParser struct {
	prefix string
}

func NewCommandParser(prefix string) CommandParser {
	return CommandParser{prefix: prefix}
}

func (c *CommandParser) parse(commandStr string) (userCommand, error) {
	var uc userCommand
	if !strings.HasPrefix(commandStr, c.prefix) {
		return uc, consts.ErrNoMatch
	}
	args := strings.SplitN(commandStr, " ", 1)
	uct := userCommandType(args[0])
	_, found := commandMap[uct]
	if !found {
		uc.command = "help"
		return uc, consts.ErrInvalidCommand
	} else {
		uc.command = uct
	}
	return uc, nil
}

func help(args []string) error {
	return nil
}

func addFact(args []string) error {
	return nil
}

func nextFact(args []string) error {
	return nil
}

func price(args []string) error {
	return nil
}

func backpack(args []string) error {
	return nil
}

func init() {
	commandMap = map[userCommandType]commandHandler{
		"help":  {help, 0, 1},
		"add":   {addFact, 0, 1},
		"next":  {nextFact, 0, 1},
		"price": {price, 0, 1},
		"bp":    {backpack, 0, 0},
	}
}
