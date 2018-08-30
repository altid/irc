package main

/* Watch each input file in the tree for activity, firing off input event
Possibilities: 

Listen on a channel for new/closing on an event channel and multitail
  - Input would care about our client joining/parting/quitting/closing and related, (also being kicked)
  - closing and deleting the file would be resultant from an event, not dissimilar opening initially. Allows us to compartmentalize the logic here to any way we want

Inotifywatch for all files named input
  - this would suffer from inotifywatch limitations, as well as require a way to query the status of each channel, whether it was allowed to write or not and wether the server itself was instantiated
  - We would need to query whether or not to close this file and delete it

Ghetto inotify 
  - this is roughly similar to the normal inotifywatch solution, it wouldn't hit the limitations on watches, but it would need querying of the server state before action.

Listen on event channel for new/closing, then open an input loop in a goroutine that also carries a close channel.  (Map of all open channels, closing send kill and input loop cleans itself up, including deleting the file)

*/ 

func (st *State) InLoop() {
	<- st.done
}
