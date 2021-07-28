module github.com/BrobridgeOrg/gravity-adapter-native

go 1.15

require (
	github.com/BrobridgeOrg/gravity-sdk v0.0.25
	github.com/cfsghost/parallel-chunked-flow v0.0.6
	github.com/jinzhu/copier v0.3.0
	github.com/json-iterator/go v1.1.10
	github.com/nats-io/nats.go v1.10.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/viper v1.7.1
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97 // indirect
	golang.org/x/net v0.0.0-20210226172049-e18ecbb05110
)

//replace github.com/BrobridgeOrg/gravity-sdk => ../gravity-sdk
