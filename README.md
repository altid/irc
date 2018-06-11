# ircfs
--------

## Usage

### Startup
`ircfs -c <configuration> -d <dir>`

This will set up a directory with <dir> as the base.
For each server defined in your <configuration> file, it will create a directory:

```
ircfs/
	irc.freenode.net/
	irc.oftc.net/
	...
```

Additionally, each server can auto connect to channels, based on the configuration settings, leading to the following structure.

```
ircfs/
	ctl
	irc.freenode.net/
		input
		feed
		#ubqt/
			status
			title
			feed
			sidebar
			input
		#foo/
			...
	irc.oftc.net/
		input
		feed
```

### Runtime Usage

By writing to the base level ctl file, ircfs/ctl, one can connect, disconnect, and reload servers.
`printf '%s\n' "/reconnect irc.oftc.net" > ircfs/ctl
Joining channels, leaving channels, and related is achieved by writing to the input file on your current buffer or server.
`printf '%s\n' "/join #ubqt" > ircfs/irc.freenode.net/input`

(The choice to break out to input and ctl is arbitrary in reference to the ircfs itself, but lends itself well to the design of a ubqt system on a whole.)

Simply writing to the input of an arbitrary buffer will send to that channel. Messages prefaced with a `/` are commands, such as `/join /part /quit /ignore`.
