# Ircfs

Ircfs is a file service used to connect to an IRC network

![Go](https://github.com/altid/ircfs/workflows/Go/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/altid/ircfs)](https://goreportcard.com/report/github.com/altid/ircfs) [![License](http://img.shields.io/:license-mit-blue.svg)](http://doge.mit-license.org)

`go install github.com/altid/ircfs`

## Usage

`ircfs [-p <servicename>]`

 - if no service name is given, `irc` is used

[Wiki!](https://github.com/altid/ircfs/wiki)

## Configuration

```
# altid/config - place this in your operating systems' default config directory

service=irc address=irc.freenode.net port=6697 auth=pass=hunter2 ssl=none
	nick=guest user=guest name=guest
	channels=#altid,#hwwm
	log=/home/guest/logs/irc/
	filter=all
	# listen_address=192.168.1.144:12345

service=irc2 address=supersecure.ircserver.net port=28888 auth=factotum
	ssl=cert cert=/path/to/some/cert.pem key=/path/to/some/key.pem
	nick=fsociety user=elliot
	log=none
	filter=smart
``` 

 - service matches the given servicename (default "irc")
 - address is the address for the remote IRC server
 - port is the port on the remote IRC server
 - auth is the authentication method
   - pass will send the string following pass= as your user password to the remote IRC server
   - factotum uses a local factotum (Plan9, plan9port) to find your password
   - none bypasses password authentication at connection time (default)
 - ssl
   - none will use a plain TCP connection with the remote IRC server
   - simple will use a simple, generally insecure connection to the remote IRC server. This is for testing purposes only and should generally not be usd
   - cert uses the values provided for cert= and key= to use certificate based authenticationover a TLS connection to the remote server
 - nick, user, and name are your respective nickname, username that you registered to the IRC server, and your real name (optional)
 - log is the directory that stores channel logs. A special value of `none` can be used to bypass logging
 - filter
   - all filters all JOIN/PART/QUIT messages
   - smart filters JOIN/PART/QUIT messages for people who haven't written to the channel recently
   - none does not filter any messages
 - listen_address is a more advanced topic, explained here: [Using listen_address](https://altid.github.io/using-listen-address.html)

## Plan9 Caveat

You must run this in the same namespace which you plan on running your servers from (9p-server, and the to-be-implemented html5 server for example) since the bind mounts don't translate across arbitrary namespaces. A simple script such as, 


```

ircfs &
discordfs &
docfs &
9p-server &
html5-server

```

will suffice to ensure everything runs in the same namespace.
