# ircfs
--------

## Overview

`ircfs -c <myconf> -d <dir>`

This will set up a directory with `<dir>` as the base

For each server defined in your config file, it will create a directory:

```
<dir>/
	irc.freenode.net/
	irc.oftc.net/
	...
```

Upon connection, it will attempt to join any channels listed in the INI
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
`printf '%s\n' "This is a message that I wish to write to the normal channel" > <dir>/irc.freenode.net/\#ubqt/input`

### Direct Message To User

To send a message to "halfwit" on Freenode, you'd do the following:
`printf '%s\n' "msg halfwit I'm just a poor boy no body loves me" > <dir>/irc.freenode.net/ctl`

### Controlling A Session

Several examples:
```
printf '%s\n' "join #ubqt" > <dir>/irc.freenode.net/ctl
printf '%s\n' "part #ubqt" > <dir>/irc.freenode.net/ctl
printf '%s\n' "reconnect" > <dir>/irc/freenode.net/ctl
# Reload the configuration. This will affect all connected servers
printf '%s\n' "reload" > <dir>/irc.freenode.net/ctl 
# This will quit the application outright
printf '%s\n' "quit" > <dir>/irc.freenode.net/ctl
``` 

TODO: This will have some changes in the future, as things formalise with the rest of the ecosystem.
A large change that is currently planned is that ircfs will only store the feed of a particular channel, with the rest only persisting the lifetime of the program itself - this will facilitate most of the server living on faster storage, such as a ramdisk.
