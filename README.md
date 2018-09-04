# ircfs
--------

## Usage

### Startup
`ircfs -c <myconf.ini> -d <dir>`

This will set up a directory with <dir> as the base.
For each server defined in your configuration INI file, it will create a directory:

```
ircfs/
	irc.freenode.net/
	irc.oftc.net/
	...
```

Upon connection, it will attempt to join any channels listed in the INI
It will result in a structure similar to this.

```
ircfs/
	irc.freenode.net/
		ctl          // Used for control messages to server. Use in append-only to have a ctl history.
		feed         // The server log
		#ubqt/       // An IRC "channel"
			status   // File containing the username/mode, and other information about the session
			title    // File containing the topic of the given channel
			feed     // The buffer log
			sidebar  // List of nicknames and their modes in a channel
			input    // Used to send messages to the channel. Use in append-only to have an input history.
		#foo/
			...
	irc.oftc.net/
		ctl
		feed
```

### Runtime Usage


