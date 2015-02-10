package constants


const (
	PING_INTERVAL = 10		// multicast every PING_INTERVAL seconds if something has changed
	LIST = 0
	LIST_REPLY = 1

	FILE_REQUEST = 2
	FILE_RESPONSE = 3

	MAX_FILE_REQUESTS = 20	// maximum outstanding requests

	REQUEST_TTL = 5
)
