# How To Contribute To This Project

## Bug Fixes

Pull Requests are the preferred method.
Please also attach the bug being addressed in your PR message, using one of the keywords
Example: 
 - closes #2
 - fix #4/fixes #4
 - fixed #43
 - resolve/resolves/resolved #5923

Please also run go fmt over any golang code, and _always_ follow the general coding style of the source you're working on.

## Bug Reports

Open an issue which describes the bug, and give as much information as possible.
Examples of useful information are:
 - Which services you were using
 - Which server
 - Which client
 - What is happening, what should be happening (what do you expect to happen)

## New Features To Existing Projects

Pull Requests, as described in Bug Fixes are the preferred method.
Code style for new source should follow any language idioms where they're well known, such as Golang, but any style for languages such as C is fine - so long as it's coherent, and reasonably easy to parse. 

### Tabs, Spaces, Alignment

You do you. Enough blood has been shed.

## New Servers

A server cannot modify the underlying servers' structured directly, outside of deleting `notification`. This indicates that a client has seen a notification, and has actioned it.
9p-server is considered authoritative on behavior in this regard, and any deviance from that behavior can be vetted through a Pull Requests' comment thread.

## New Services

Testfs can be considered as a simplest-case, complete service for Altid. A few notable caveats exist for services, and are as follows:
 - Events must be a normal file
   - FIFO are tempting to use, but suffer only allowing one listener at a time. This limits an Altid network to a single server type, therefor is undesirable.
 - Tabs must list all buffers still open since startup. A client close must remove a tab from the file.

## New Clients

Arguably the most simple thing to implement in Altid is a client. They simply have to attach to a server, and issue ctrl/(input when applicable) messages. Optionally, they can handle and delete notifications for the buffer they are viewing.
