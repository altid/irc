# ircfs
--------

## Overview

`ircfs -p <dir>`

This will set up a directory with `<dir>` as the base

For each server defined in your config file, it will create a directory:

```
<dir>/
	irc.freenode.net/
	irc.oftc.net/
	...
```

Upon connection, it will attempt to join any channels listed in config
It will result in a structure similar to this.

```
<dir>/
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

## Usage

### Writing To Channel

To write a message to the channel #ubqt on Freenode, you'd do the following:
`printf '%s\n' "This is a message that I wish to write to the normal channel" >> <dir>/irc.freenode.net/\#ubqt/input`

### Direct Message To User

To send a message to "halfwit" on Freenode, you'd do the following:
`printf '%s\n' "msg halfwit I'm just a poor boy no body loves me" >> <dir>/irc.freenode.net/ctl`

### Controlling A Session

Several examples:
```
printf '%s\n' "join #ubqt" >> <dir>/irc.freenode.net/ctl
printf '%s\n' "part #ubqt" >> <dir>/irc.freenode.net/ctl
printf '%s\n' "reconnect" >> <dir>/irc/freenode.net/ctl
# Reload the configuration. This will affect all connected servers
printf '%s\n' "reload" >> <dir>/irc.freenode.net/ctl 
# This will quit the application outright
printf '%s\n' "quit" >> <dir>/irc.freenode.net/ctl
``` 
## Current Status

The ctrl file needs to be implemented still, it's at this moment a placeholder. As it stands, the program will block on stdin, after anything is typed it will exit.

## Caveat

Firstly, on plan9, you must run this in the same namespace which you plan on running your servers from (9p-server, and the to-be-implemented html5 server for example) since the bind mounts don't translate across arbitrary namespaces. A simple script such as, 


```
ircfs &
discordfs &
docfs &
9p-server &
html5-server &
```

You would then simply Kill(1) the named processes when you're finished with them.

Secondly, this fs uses polling for all input in its current form, which is expensive to do on a networked plan9 device that isn't the file server itself. Consider running on a machine with a very fast connection to the file server, if not the file server itself (future versions will create a 9p file tree on plan9 outright, which will drastically improve this case)
