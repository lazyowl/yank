#Yank
Yank is a peer to peer file sharing system for LANs written in Go.

######News Flash
There's been some rather chaotic initial development as I grapple with various designs and golang best practices. I am slowly converging upon a stable design to build on though.

####Mostly Complete:

* Basic networking capability - each machine runs a small server subscribed to a multicast network.
* Basic custom configuration
* Inotify watcher in place watching the directory specified in the config file
* Basic yank prompt allows users to see available public files and users on the LAN (as well as some stats), though its not very pretty
* File transfer - extremely simplistic, fixed size chunks

####To do:

* Show download stats when file download is complete and allow subsequent files to be downloaded
* Add a file index. Right now yank performs a readdir to get to the file in question based on one of its attributes

####Run:
`go run loop.go`
