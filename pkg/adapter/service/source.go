package adapter

import (
	"fmt"
	"sync"
	"time"
	"unsafe"

	"github.com/BrobridgeOrg/gravity-sdk/core"
	gravity_subscriber "github.com/BrobridgeOrg/gravity-sdk/subscriber"
	gravity_state_store "github.com/BrobridgeOrg/gravity-sdk/subscriber/state_store"
	gravity_sdk_types_projection "github.com/BrobridgeOrg/gravity-sdk/types/projection"
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

var projectionPool = sync.Pool{
	New: func() interface{} {
		return &gravity_sdk_types_projection.Projection{}
	},
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
	pj := projectionPool.Get().(*gravity_sdk_types_projection.Projection)
	defer projectionPool.Put(pj)

	// Parsing data
	err := gravity_sdk_types_projection.Unmarshal(msg.Event.Data, pj)
	if err != nil {
		return err
	}

	// Convert projection to JSON
	data, err := pj.ToJSON()
	if err != nil {
		return err
	}

	eventName := jsoniter.Get(data, "event").ToString()
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

func (source *Source) InitSubscription() error {

	// Subscribe to collections
	subscriptions := make(map[string][]string)
	for _, collection := range source.info.Collections {
		log.Info("Subscribing to collection: " + collection)
		subscriptions[collection] = make([]string, 0)
	}
	err := source.subscriber.SubscribeToCollections(subscriptions)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{}).Info("Subscribing to gravity pipelines...")
	err = source.subscriber.AddAllPipelines()
	if err != nil {
		return err
	}

	// Start receiving
	log.WithFields(log.Fields{}).Info("Starting to fetch data from gravity...")
	_, err = source.subscriber.Subscribe(func(msg *gravity_subscriber.Message) {

		err := source.processData(msg)
		if err != nil {
			log.Error(err)
			return
		}

		msg.Ack()
	})
	if err != nil {
		log.Error(err)
		return err
	}

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
	options.Verbose = false
	options.StateStore = source.stateStore
	options.WorkerCount = source.info.WorkerCount

	subscriber := gravity_subscriber.NewSubscriber(options)
	opts := core.NewOptions()
	opts.PingInterval = time.Second * time.Duration(source.info.PingInterval)
	opts.MaxPingsOutstanding = source.info.MaxPingsOutstanding
	opts.MaxReconnects = source.info.MaxReconnects
	err = subscriber.Connect(address, opts)
	if err != nil {
		return err
	}

	// Register subscriber
	log.WithFields(log.Fields{
		"subscriberID":   source.info.SubscriberID,
		"subscriberName": source.info.SubscriberName,
	}).Info("Registering subscriber")
	err = subscriber.Register(gravity_subscriber.SubscriberType_Exporter, "native", source.info.SubscriberID, source.info.SubscriberName)
	if err != nil {
		return err
	}

	source.subscriber = subscriber

	return source.InitSubscription()
}
