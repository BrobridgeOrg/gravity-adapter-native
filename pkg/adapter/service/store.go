package adapter

import (
	gravity_store "github.com/BrobridgeOrg/gravity-sdk/core/store"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func (adapter *Adapter) initializeStore() error {

	viper.SetDefault("adapter.state_store", "./datastore")
	storePath := viper.GetString("adapter.state_store")

	log.WithFields(log.Fields{
		"path": storePath,
	}).Info("Initializing store")

	options := gravity_store.NewOptions()
	options.StoreOptions.DatabasePath = storePath
	store, err := gravity_store.NewStore(options)
	if err != nil {
		return err
	}

	adapter.store = store

	return nil
}
