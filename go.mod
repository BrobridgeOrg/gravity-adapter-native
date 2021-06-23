module github.com/BrobridgeOrg/gravity-adapter-native

go 1.15

require (
	github.com/BrobridgeOrg/gravity-sdk v0.0.17
	github.com/cfsghost/parallel-chunked-flow v0.0.6
	github.com/jinzhu/copier v0.3.0
	github.com/json-iterator/go v1.1.10
	github.com/nats-io/nats.go v1.10.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/viper v1.7.1
	golang.org/x/net v0.0.0-20200625001655-4c5254603344
)

//replace github.com/BrobridgeOrg/gravity-sdk => ../gravity-sdk
