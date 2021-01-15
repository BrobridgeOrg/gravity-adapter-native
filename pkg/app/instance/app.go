package instance

import (
	"runtime"

	grpc_connection_pool "github.com/cfsghost/grpc-connection-pool"

	adapter_service "github.com/BrobridgeOrg/gravity-adapter-native/pkg/adapter/service"
	grpc_server "github.com/BrobridgeOrg/gravity-adapter-native/pkg/grpc_server/server"
	mux_manager "github.com/BrobridgeOrg/gravity-adapter-native/pkg/mux_manager/manager"
	gravity_adapter "github.com/BrobridgeOrg/gravity-sdk/adapter"
	log "github.com/sirupsen/logrus"
)

type AppInstance struct {
	done             chan bool
	adapterConnector *gravity_adapter.AdapterConnector
	adapter          *adapter_service.Adapter
	grpcServer       *grpc_server.Server
	grpcPool         *grpc_connection_pool.GRPCPool
	muxManager       *mux_manager.MuxManager
}

func NewAppInstance() *AppInstance {

	a := &AppInstance{
		done: make(chan bool),
	}

	a.adapter = adapter_service.NewAdapter(a)

	return a
}

func (a *AppInstance) Init() error {

	log.WithFields(log.Fields{
		"max_procs": runtime.GOMAXPROCS(0),
	}).Info("Starting application")

	a.muxManager = mux_manager.NewMuxManager(a)
	a.grpcServer = grpc_server.NewServer(a)

	a.initMuxManager()
	//Initializing grpc Server
	err := a.initGRPCServer()
	if err != nil {
		return err
	}

	/*
		// Initializing gRPC pool
		err = a.initGRPCPool()
		if err != nil {
			return err
		}
	*/

	// Initializing adapter connector
	err = a.initAdapterConnector()
	if err != nil {
		return err
	}

	err = a.adapter.Init()
	if err != nil {
		return err
	}

	return nil
}

func (a *AppInstance) Uninit() error {
	//Unregister to controller
	return a.adapter.Uninit()
}

func (a *AppInstance) Run() error {

	//GRPC
	go func() {
		err := a.runGRPCServer()
		if err != nil {
			log.Fatal(err)
		}
	}()

	err := a.runMuxManager()
	if err != nil {
		return err
	}

	<-a.done

	return nil
}
