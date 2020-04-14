package parse

import (
	"bind_generator/consts"
	"fmt"
	"strings"
)

var commandMap map[userCommandType]commandHandler

type CommandResult struct {
	Message string
}

type userCommandType string

type commandHandler struct {
	handler func(args []string) (CommandResult, error)
	minArgs int
	maxArgs int
	helpMsg string
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

func help(args []string) (CommandResult, error) {
	var r CommandResult
	if len(args) == 0 {
		var m []string
		for k := range commandMap {
			m = append(m, fmt.Sprintf("%s%s", "!", k))
		}
		r.Message = strings.Join(m, " ")
	} else {
		cs := userCommandType(strings.ToLower(args[0]))
		c, found := commandMap[cs]
		if !found {
			r.Message = fmt.Sprintf("Unknown command: %s", cs)
		} else {
			r.Message = fmt.Sprintf("!%s - %s", cs, c.helpMsg)
		}
	}
	return r, nil
}

func addFact(args []string) (CommandResult, error) {
	return CommandResult{}, nil
}

func nextFact(args []string) (CommandResult, error) {
	return CommandResult{}, nil
}

func price(args []string) (CommandResult, error) {
	return CommandResult{}, nil
}

func backpack(args []string) (CommandResult, error) {
	return CommandResult{}, nil
}

func init() {
	commandMap = map[userCommandType]commandHandler{
		"help":  {help, 0, 1, ""},
		"add":   {addFact, 0, 1, "Add a new fact"},
		"next":  {nextFact, 0, 1, "Get the next random fact"},
		"price": {price, 0, 1, "Get the price of a item"},
		"bp":    {backpack, 0, 0, "Get the price a users backpack"},
	}
}
