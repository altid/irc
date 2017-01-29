package main

// Nicks - returns a list of nicknames in current channel according to format string
func (c *client) Nicks() string {
	//TODO: Implement this in the future
	return "fake\nsidebar\njohn\nbob\npeter\npaul\n"
}

// Status - return status bar for current channel according to format string
func (c *client) Status() string {
	return "this is not a real status\n"
}

// Title - returns topic for channel, etc
func (c *client) Title() string {
	return "Channel for the discussion of meth and frogs\n"
}

// Ctl - returns list of commands available for clients to issue to server
func (c *client) Ctl() string {
	return "foo\nbar\nbaz\nbalogna\nchicken\n"
}

// Tabs - returns formatted string of buffers with activity
func (c *client) Tabs() string {
	return "this will eventually be a list of buffers with updates\n"
}
