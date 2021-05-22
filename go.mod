module github.com/BrobridgeOrg/gravity-adapter-native

go 1.15

require (
	github.com/BrobridgeOrg/gravity-adapter-nats v0.0.0-20210103201642-95f4806b99ef
	github.com/BrobridgeOrg/gravity-api v0.2.14
	github.com/BrobridgeOrg/gravity-exporter-nats v0.0.0-20210521224753-688f250cd6c7
	github.com/BrobridgeOrg/gravity-sdk v0.0.9
	github.com/cfsghost/grpc-connection-pool v0.6.0
	github.com/cfsghost/parallel-chunked-flow v0.0.3
	github.com/json-iterator/go v1.1.10
	github.com/nats-io/nats.go v1.10.0
	github.com/sirupsen/logrus v1.7.0
	github.com/soheilhy/cmux v0.1.4
	github.com/spf13/viper v1.7.1
	golang.org/x/net v0.0.0-20200625001655-4c5254603344
	google.golang.org/grpc v1.32.0
)

//replace github.com/BrobridgeOrg/gravity-sdk => ../gravity-sdk
