package server

import (
	"net"

	app "github.com/BrobridgeOrg/gravity-adapter-native/pkg/app"
	adapter_service "github.com/BrobridgeOrg/gravity-adapter-native/pkg/grpc_server/services/adapter"
	adapter "github.com/BrobridgeOrg/gravity-api/service/adapter"
	"google.golang.org/grpc"

	log "github.com/sirupsen/logrus"
	"github.com/soheilhy/cmux"
)

type Server struct {
	app            app.App
	instance       *grpc.Server
	listener       net.Listener
	adapterService *adapter_service.Service
	host           string
}

func NewServer(a app.App) *Server {
	return &Server{
		app:      a,
		instance: &grpc.Server{},
	}
}

func (server *Server) Init(host string) error {

	// Put it to mux
	mux, err := server.app.GetMuxManager().AssertMux("grpc", host)
	if err != nil {
		return err
	}

	// Preparing listener
	lis := mux.MatchWithWriters(
		cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"),
	)

	server.host = host
	server.listener = lis
	server.instance = grpc.NewServer()

	// Services
	server.adapterService = adapter_service.NewService(server.app)
	adapter.RegisterAdapterServer(server.instance, server.adapterService)

	return nil
}

func (server *Server) Serve() error {

	log.WithFields(log.Fields{
		"host": server.host,
	}).Info("Starting GRPC server")

	// Starting server
	if err := server.instance.Serve(server.listener); err != cmux.ErrListenerClosed {
		log.Error(err)
		return err
	}

	return nil
}

func (server *Server) GetApp() app.App {
	return server.app
}

func (server *Server) GetEventChan() chan []byte {
	return server.adapterService.Incoming

}
