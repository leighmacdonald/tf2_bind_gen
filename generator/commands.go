package generator

import (
	"bind_generator/consts"
	"bind_generator/steam"
	"fmt"
	"strings"
)

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

type UserCommand struct {
	command userCommandType
	args    []string
}

type CommandParser struct {
	prefix     string
	commandMap map[userCommandType]commandHandler
}

func NewCommandParser(prefix string) CommandParser {
	cp := CommandParser{prefix: prefix}
	cp.commandMap = map[userCommandType]commandHandler{
		"help":     {cp.help, 0, 1, ""},
		"addfact":  {cp.addFact, 0, 1, "Add a new fact"},
		"nextfact": {cp.nextFact, 0, 1, "Get the next random fact"},
		"price":    {cp.price, 0, 1, "Get the price of a item"},
		"bp":       {cp.backpack, 0, 1, "Get the price a users backpack"},
	}
	return cp
}

func (c *CommandParser) ParseMsg(commandStr string) (UserCommand, error) {
	var uc UserCommand
	if !strings.HasPrefix(commandStr, c.prefix) {
		return uc, consts.ErrNoMatch
	}
	args := strings.SplitN(commandStr, " ", 2)
	uct := userCommandType(strings.Replace(args[0], c.prefix, "", 1))
	_, found := c.commandMap[uct]
	if !found {
		uc.command = "help"
		return uc, consts.ErrInvalidCommand
	} else {
		uc.command = uct
		if len(args) == 2 {
			values := strings.Split(args[1], " ")
			if uct == "bp" {
				// Convert the first value to a ID
				sid := steam.StringToSID64(values[0])
				if sid.Valid() {
					values[0] = sid.String()
				}
			}
			uc.args = values
		}
	}
	return uc, nil
}

func (c *CommandParser) help(args []string) (CommandResult, error) {
	var r CommandResult
	if len(args) == 0 {
		var m []string
		for k := range c.commandMap {
			m = append(m, fmt.Sprintf("%s%s", c.prefix, k))
		}
		r.Message = strings.Join(m, " ")
	} else {
		cs := userCommandType(strings.ToLower(args[0]))
		cmd, found := c.commandMap[cs]
		if !found {
			r.Message = fmt.Sprintf("Unknown command: %s", cs)
		} else {
			r.Message = fmt.Sprintf("%s%s - %s", c.prefix, cs, cmd.helpMsg)
		}
	}
	return r, nil
}

func (c *CommandParser) addFact(args []string) (CommandResult, error) {
	return CommandResult{}, nil
}

func (c *CommandParser) nextFact(args []string) (CommandResult, error) {
	return CommandResult{}, nil
}

func (c *CommandParser) price(args []string) (CommandResult, error) {
	return CommandResult{}, nil
}

func (c *CommandParser) backpack(args []string) (CommandResult, error) {
	return CommandResult{}, nil
}
