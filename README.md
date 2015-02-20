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
* Kind of shows download stats when file download is complete; also new files can be downloaded too

####To do:

* Add a file index. Right now yank performs a readdir to get to the file in question based on one of its attributes

####Run:
`go run loop.go`

To view the command line options, `go run loop.go -h`

####Prompt syntax
This may be just a temporary instruction set for now:
```
ls - list public files on the network
lls - list local files
lu - list users on the network
get <hash> <dest> - get file with <hash> and save it as <dest> (if <dest> not provided, <hash> itself is used as the filename)
q - quit
```
