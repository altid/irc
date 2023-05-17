package commands

import (
	"github.com/altid/libs/service/commander"
)

var Commands = []*commander.Command{
	{
		Name:        "action",
		Alias:       []string{"me", "act"},
		Description: "Send an emote to the server",
		Heading:     commander.ActionGroup,
	},
	{
		Name:        "nick",
		Description: "Set new nickname",
		Args:        []string{"<name>"},
		Heading:     commander.DefaultGroup,
	},
	{
		Name:        "msg",
		Heading:     commander.DefaultGroup,
		Description: "Send a message to user",
		Args:        []string{"<name>", "<msg>"},
		Alias:       []string{"query", "m", "q"},
	},
}
