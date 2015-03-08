#Yank
Yank is a peer to peer file sharing system for LANs written in Go.

####Mostly Complete:

* Basic networking capability - each machine runs a small server subscribed to a multicast network.
* Basic custom configuration
* Inotify watcher in place watching the directory specified in the config file
* Basic yank prompt allows users to see available public files and users on the LAN (as well as some stats), though its not very pretty
* File transfer - extremely simplistic, fixed size chunks
* Kind of shows download stats when file download is complete; also new files can be downloaded too

####Current (Known) Limitations/Issues
* **Important** The hash of the file should be updated when the file is modified. However, by doing so, other peers who were fetching data related to the file will now have their requests invalidated since the hash is different. This needs to be handled somehow. One possibility is to store chunk hashes for each file. When a file is to be fetched, fetch all of its chunk hashes first and then query those (rather than chunk positions). Does this even need to be handled? Maybe we should just abort the fetch if the file changes and start the whole thing all over again. May not be terribly efficient though.
* The current implementation uses a 64 bit integer as a bit vector to store file chunk presence/absence. Therefore, file sizes are limited to 64 * CHUNK_SIZE.
* It assumes that each user will have a unique name. Currently, it does not check for this.

####Config File Format
The config file should be in JSON format. By default, Yank will look for a file named `config.txt` in the base directory of the repository.
```
{
	"Name": "lazyowl",
	"PublicDir": "public_dir",
	"MetaDir": "meta_dir",
	"PingInterval": 8
}
```


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


####Architecture
This is a basic high-level overview of the system. Plenty of room for improvement.

#####File Representation
Each file in the public folder is represented by a `MyFile` struct storing the name, hash, hash bit vector, size and some internal fields. The hash uniquely identifies the file. The hash bit vector indicates which chunks are present and which aren't.

#####Caches
Each peer maintains two caches:

1. *host cache* - Maps name to IP address. An `lu` command always queries this cache.
2. *file cache* - Stores the files present with each peer on the network. An `ls` command always queries this cache.

#####Ping
At regular  intervals, each peer sends out a ping which includes a list of public files (`MyFile` structs). Each peer, on receiving such a ping, updates its caches.

#####File Fetch
When the client issues a `get <hash name> <dest name>` command, Yank checks its file cache to see if the file is present with at least one peer. If not, it aborts. Otherwise, it builds a map of each peer to its corresponding `MyFile` object.
It also checks whether a possibly incomplete copy of the file is present locally. It then starts the sending round.

In the sending round, Yank cycles through each user and identifies all the chunks that it can request from that user (accounting for chunks requested for previous users as well as chunks already present), marks them as `REQUEST_SENT` and sends the request. It starts a timer for that user as well. It also stores the chunks requested for that user.

If the timer times out without a response, the chunks requested are unmarked and set to `UNREQUESTED` again.

Whenever a response is received, Yank checks that the received response is current and that its timer has not expired. It then writes the chunks locally (creating the file if not present), setting bits in the bit vector accordingly.

Whenever a chunk request is received, the peer creates a list of chunks and sends them back to the requester. If the peer does not have the file (maybe it has been deleted since the last ping), the peer ignores the request. This will cause its timer on the sender's side to time out and the sender will act accordingly.
