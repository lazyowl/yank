#Yank
Yank is a peer to peer file sharing system for LANs written in Go.

####Mostly Complete:

* Basic networking capability - each machine runs a small server subscribed to a multicast network.
* Basic custom configuration
* Inotify watcher in place watching the directory specified in the config file
* Basic yank prompt allows users to see available public files and users on the LAN (as well as some stats)

####To do:

* The actual file transfer module - thinking of constant size chunks and fetch the ones the user doesn't have. May possibly look into something like Rabin-Karp hashing
* Add a file index. Right now yank performs a readdir to get to the file in question based on one of its attributes

####Run:
`go run loop.go`
