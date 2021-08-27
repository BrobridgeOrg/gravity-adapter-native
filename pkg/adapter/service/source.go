package adapter

import (
	"fmt"
	"sync"
	"time"
	"unsafe"

	"github.com/BrobridgeOrg/gravity-sdk/core"
	gravity_subscriber "github.com/BrobridgeOrg/gravity-sdk/subscriber"
	gravity_state_store "github.com/BrobridgeOrg/gravity-sdk/subscriber/state_store"

	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
)

var counter uint64

type Packet struct {
	EventName string
	Payload   []byte
}

type Source struct {
	adapter    *Adapter
	stateStore *gravity_state_store.StateStore
	subscriber *gravity_subscriber.Subscriber

	name string
	info *SourceInfo
}

var packetPool = sync.Pool{
	New: func() interface{} {
		return &Packet{}
	},
}

func StrToBytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

func NewSource(adapter *Adapter, name string, sourceInfo *SourceInfo) *Source {

	return &Source{
		adapter: adapter,
		name:    name,
		info:    sourceInfo,
	}
}

func (source *Source) processData(msg *gravity_subscriber.Message) error {
	/*
		id := atomic.AddUint64((*uint64)(&counter), 1)

		if id%100 == 0 {
			log.Info(id)
		}
	*/
	event := msg.Payload.(*gravity_subscriber.DataEvent)
	pj := event.Payload

	// Convert projection to JSON
	data, err := pj.ToJSON()
	if err != nil {
		return err
	}

	eventName := ""
	//Mapping event Name from source.info.Events
	switch pj.Method {
	case "insert":
		eventName = source.info.Events.Create
	case "update":
		eventName = source.info.Events.Update
	case "delete":
		eventName = source.info.Events.Delete

	}
	payload := jsoniter.Get(data, "payload").ToString()

	connector := source.adapter.app.GetAdapterConnector()
	for {
		err := connector.Publish(eventName, StrToBytes(payload), nil)
		if err != nil {
			log.Error(err)
			time.Sleep(time.Second)
			continue
		}

		break
	}

	msg.Ack()

	return nil
}

func (source *Source) InitStateStore() error {

	// Initializing state store
	options := gravity_state_store.NewOptions()
	options.Name = source.name
	stateStore := gravity_state_store.NewStateStoreWithStore(source.adapter.store, options)
	err := stateStore.Initialize()
	if err != nil {
		return err
	}

	source.stateStore = stateStore

	return nil
}

func (source *Source) Init() error {

	address := fmt.Sprintf("%s:%d", source.info.Host, source.info.Port)

	log.WithFields(log.Fields{
		"address": address,
	}).Info("Initializing source connector")

	err := source.InitStateStore()
	if err != nil {
		return err
	}

	// Initializing gravity subscriber and connecting to server
	options := gravity_subscriber.NewOptions()
	options.StateStore = source.stateStore
	options.WorkerCount = source.info.WorkerCount
	options.ChunkSize = source.info.ChunkSize
	options.InitialLoad.Enabled = source.info.InitialLoad
	options.InitialLoad.OmittedCount = source.info.OmittedCount
	options.Verbose = source.info.Verbose

	source.subscriber = gravity_subscriber.NewSubscriber(options)
	opts := core.NewOptions()
	opts.PingInterval = time.Second * time.Duration(source.info.PingInterval)
	opts.MaxPingsOutstanding = source.info.MaxPingsOutstanding
	opts.MaxReconnects = source.info.MaxReconnects
	err = source.subscriber.Connect(address, opts)
	if err != nil {
		return err
	}

	// Setup data handler
	source.subscriber.SetEventHandler(source.eventHandler)
	source.subscriber.SetSnapshotHandler(source.snapshotHandler)

	// Register subscriber
	log.WithFields(log.Fields{
		"subscriberID":   source.info.SubscriberID,
		"subscriberName": source.info.SubscriberName,
	}).Info("Registering subscriber")
	err = source.subscriber.Register(gravity_subscriber.SubscriberType_Transmitter, "native", source.info.SubscriberID, source.info.SubscriberName)
	if err != nil {
		return err
	}

	// Subscribe to collections
	subscriptions := make(map[string][]string)
	for _, collection := range source.info.Collections {
		log.Info("Subscribing to collection: " + collection)
		subscriptions[collection] = make([]string, 0)
	}
	err = source.subscriber.SubscribeToCollections(subscriptions)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{}).Info("Subscribing to gravity pipelines...")
	err = source.subscriber.AddAllPipelines()
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func (source *Source) eventHandler(msg *gravity_subscriber.Message) {

	err := source.processData(msg)
	if err != nil {
		log.Error(err)
		return
	}
}

func (source *Source) snapshotHandler(msg *gravity_subscriber.Message) {

	event := msg.Payload.(*gravity_subscriber.SnapshotEvent)
	snapshotRecord := event.Payload

	eventName := source.info.Events.Snapshot
	payload, err := snapshotRecord.Payload.MarshalJSON()
	if err != nil {
		log.Error(err)
		return
	}

	connector := source.adapter.app.GetAdapterConnector()
	for {
		err := connector.Publish(eventName, payload, nil)
		if err != nil {
			log.Error(err)
			time.Sleep(time.Second)
			continue
		}
		break
	}

	msg.Ack()

}

func (source *Source) Run() error {

	log.WithFields(log.Fields{}).Info("Starting to fetch data from gravity...")

	source.subscriber.Start()

	return nil
}
