package define

//default values define
const (
	DefaultMinConnects = 128
	DefaultMaxConnects = 1024
	DefaultTcpVersion = "tcp"
	DefaultTcpReadBuffSize = 1024
	DefaultLazySendChanSize = 1024
	DefaultManagerTicker = 60 //xx seconds
	DefaultUnActiveSeconds = 60 //xx seconds
)


//general
const (
	ConnectWriteChanSize = 1024
	HandlerQueueSizeDefault = 5
	HandlerQueueSizeMax = 128
	HandlerQueueChanSize = 1024
	PacketMaxSize = 4096 //4KB
)