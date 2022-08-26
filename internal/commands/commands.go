package commands

import "github.com/altid/libs/service"

var Commands = []*service.Command{
	{
		Name:        "action",
		Alias:       []string{"me", "act"},
		Description: "Send an emote to the server",
		Heading:     service.ActionGroup,
	},
	{
		Name:        "nick",
		Description: "Set new nickname",
		Args:        []string{"<name>"},
		Heading:     service.DefaultGroup,
	},
	{
		Name:    	 "msg",
		Heading: 	 service.DefaultGroup,
		Description: "Send a message to user",
		Args:    	 []string{"<name>", "<msg>"},
		Alias:   	 []string{"query", "m", "q"},
	},
}