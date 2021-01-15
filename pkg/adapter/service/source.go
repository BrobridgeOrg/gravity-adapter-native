package adapter

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
	"unsafe"

	jsoniter "github.com/json-iterator/go"

	grpc_server "github.com/BrobridgeOrg/gravity-adapter-native/pkg/grpc_server"
	controller "github.com/BrobridgeOrg/gravity-api/service/controller"
	dsa "github.com/BrobridgeOrg/gravity-api/service/dsa"
	parallel_chunked_flow "github.com/cfsghost/parallel-chunked-flow"
	log "github.com/sirupsen/logrus"
)

//var counter uint64

type Source struct {
	adapter    *Adapter
	incoming   chan []byte
	name       string
	host       string
	port       int
	grpcServer grpc_server.Server
	parser     *parallel_chunked_flow.ParallelChunkedFlow
}

var requestPool = sync.Pool{
	New: func() interface{} {
		return &dsa.PublishRequest{}
	},
}

func StrToBytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

func NewSource(adapter *Adapter, name string, sourceInfo *SourceInfo) *Source {

	// Initialize parapllel chunked flow
	pcfOpts := parallel_chunked_flow.Options{
		BufferSize: 204800,
		ChunkSize:  512,
		ChunkCount: 512,
		Handler: func(data interface{}, output chan interface{}) {
			/*
				id := atomic.AddUint64((*uint64)(&counter), 1)
				if id%1000 == 0 {
					log.Info(id)
				}
			*/

			eventName := jsoniter.Get(data.([]byte), "event").ToString()
			payload := jsoniter.Get(data.([]byte), "payload").ToString()

			// Preparing request
			request := requestPool.Get().(*dsa.PublishRequest)
			request.EventName = eventName
			request.Payload = StrToBytes(payload)

			output <- request
		},
	}

	return &Source{
		adapter:  adapter,
		incoming: make(chan []byte, 204800),
		name:     name,
		host:     sourceInfo.Host,
		port:     sourceInfo.Port,
		parser:   parallel_chunked_flow.NewParallelChunkedFlow(&pcfOpts),
	}
}

func (source *Source) InitSubscription(controllerHost string) error {

	log.WithFields(log.Fields{
		"source": source.name,
	}).Info("Register to controller and routing data from synchronizer to dsa.")

	// handle grpc incoming data
	source.grpcServer = source.adapter.app.GetGRPCServer()
	eventIncoming := source.grpcServer.GetEventChan()
	go func() {
		for {
			select {
			case msg := <-eventIncoming:
				source.incoming <- msg
			}
		}
	}()

	// Register client
	// Initializing gRPC pool
	err := source.adapter.app.InitGRPCPool(controllerHost)
	if err != nil {
		log.Error("Failed to init grpc pool: ", err)
		return err
	}
	conn, err := source.adapter.app.GetGRPCPool().Get()
	if err != nil {
		log.Error("Failed to get connection: ", err)
		return err
	}

	client := controller.NewControllerClient(conn)

	// Preparing context
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	registerAdapterRequest := &controller.RegisterAdapterRequest{
		ClientID: source.adapter.clientName + "-" + source.name,
		Url:      fmt.Sprintf("%s:%d", source.host, source.port),
		Offset:   0,
	}
	resp, err := client.RegisterAdapter(ctx, registerAdapterRequest)
	if err != nil {
		log.Error("Registation failed: ", err)
		return err
	}
	if !resp.Success {
		log.Error("Registation failed: ", resp.Success)
		return errors.New(resp.Reason)
	}

	return nil
}

func (source *Source) Init() error {

	address := fmt.Sprintf("%s:%d", source.host, source.port)

	log.WithFields(log.Fields{
		"source":      source.name,
		"address":     address,
		"client_name": source.adapter.clientName + "-" + source.name,
	}).Info("Initializing source connector")

	// Connect to data source
	go source.eventReceiver()
	go source.requestHandler()

	return source.InitSubscription(address)
}

func (source *Source) Uninit() error {

	// Unregister client
	// Initializing gRPC pool
	log.Info("Connection to controller grpc ...")
	conn, err := source.adapter.app.GetGRPCPool().Get()
	if err != nil {
		log.Error("Failed to get connection: ", err)
		return err
	}

	client := controller.NewControllerClient(conn)

	// Preparing context
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	log.Info("Unregister client ...")
	registerAdapterRequest := &controller.UnregisterAdapterRequest{
		ClientID: source.adapter.clientName + "-" + source.name,
	}
	resp, err := client.UnregisterAdapter(ctx, registerAdapterRequest)
	if err != nil {
		log.Error("Unregistation failed: ", err)
		return err
	}
	if !resp.Success {
		log.Error("Unregistation failed: ", resp.Success)
		return errors.New(resp.Reason)
	}

	return nil
}

func (source *Source) eventReceiver() {

	log.WithFields(log.Fields{
		"source":      source.name,
		"client_name": source.adapter.clientName + "-" + source.name,
	}).Info("Initializing workers ...")

	for {
		select {
		case msg := <-source.incoming:
			source.parser.Push(msg)
		}
	}
}

func (source *Source) requestHandler() {

	for {
		select {
		case req := <-source.parser.Output():
			source.HandleRequest(req.(*dsa.PublishRequest))
			requestPool.Put(req)
		}
	}
}

func (source *Source) HandleRequest(request *dsa.PublishRequest) {

	for {
		connector := source.adapter.app.GetAdapterConnector()
		err := connector.Publish(request.EventName, request.Payload, nil)
		if err != nil {
			log.Error(err)
			time.Sleep(time.Second)
			continue
		}

		break
	}
}
