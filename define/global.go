package define

//default values define
const (
	DefaultPort = 5300
	DefaultTcpVersion = "tcp"
	DefaultTcpReadBuffSize = 1024

	DefaultTcpDialTimeOut = 10 //xx seconds
	DefaultManagerTicker = 60 //xx seconds
	DefaultUnActiveSeconds = 60 //xx seconds
	DefaultGCRate = 300 //xx seconds

	DefaultBuckets = 3
	DefaultBucketReadRate = 0.2 //xx seconds
)

//general
const (
	ConnectWriteChanSize = 1024
	HandlerQueueSizeDefault = 5
	HandlerQueueSizeMax = 128
	HandlerQueueChanSize = 1024
	PacketMaxSize = 2048 //2KB
)