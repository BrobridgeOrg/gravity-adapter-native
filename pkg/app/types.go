package app

import (
	grpc_server "github.com/BrobridgeOrg/gravity-adapter-native/pkg/grpc_server"
	mux_manager "github.com/BrobridgeOrg/gravity-adapter-native/pkg/mux_manager"
	gravity_adapter "github.com/BrobridgeOrg/gravity-sdk/adapter"
	grpc_connection_pool "github.com/cfsghost/grpc-connection-pool"
)

type App interface {
	GetAdapterConnector() *gravity_adapter.AdapterConnector
	GetMuxManager() mux_manager.Manager
	GetGRPCServer() grpc_server.Server
	InitGRPCPool(string) error
	GetGRPCPool() *grpc_connection_pool.GRPCPool
}
