# Irc

Irc is an Altid service used to connect to an IRC network

![Go](https://github.com/altid/irc/workflows/Go/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/altid/ircfs)](https://goreportcard.com/report/github.com/altid/ircfs) [![License](http://img.shields.io/:license-mit-blue.svg)](http://doge.mit-license.org)

`go install github.com/altid/ircfs/cmd/irc@latest`

## Usage

`irc [-s <servicename>] [-d] [-a <address to bind to>] [-m]`
 - -d enables debug output
 - if no service name is given, `irc` is used

[Wiki!](https://github.com/altid/irc/wiki)

## Configuration

```
# altid/config - place this in your operating systems' default config directory

service=irc address=libera.chat port=6697 auth=pass=hunter2 ssl=none
	nick=guest user=guest name=guest
	channels=#altid
	log=/home/guest/logs/irc/
	filter=all

service=irc2 address=supersecure.ircserver.net port=28888 auth=factotum
	ssl=cert cert=/path/to/some/cert.pem key=/path/to/some/key.pem
	nick=fsociety user=elliot
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
 - filter
   - all filters all JOIN/PART/QUIT messages
   - smart filters JOIN/PART/QUIT messages for people who haven't written to the channel recently
   - none does not filter any messages
