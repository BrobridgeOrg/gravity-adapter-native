package grpc_server

type Server interface {
	Init(string) error
	Serve() error
	GetEventChan() chan []byte
}
