package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	app "github.com/BrobridgeOrg/gravity-adapter-native/pkg/app/instance"
)

func init() {

	debugLevel := log.InfoLevel
	switch os.Getenv("GRAVITY_DEBUG") {
	case log.TraceLevel.String():
		debugLevel = log.TraceLevel
	case log.DebugLevel.String():
		debugLevel = log.DebugLevel
	case log.ErrorLevel.String():
		debugLevel = log.ErrorLevel
	}

	log.SetLevel(debugLevel)

	fmt.Printf("Debug level is set to \"%s\"\n", debugLevel.String())

	// From the environment
	viper.SetEnvPrefix("GRAVITY_ADAPTER_NATIVE")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// From config file
	viper.SetConfigName("config")
	viper.AddConfigPath("./")
	viper.AddConfigPath("./configs")

	if err := viper.ReadInConfig(); err != nil {
		log.Warn("No configuration file was loaded")
	}

	runtime.GOMAXPROCS(8)
	/*
		return
		go func() {

			f, err := os.Create("cpu-profile.prof")
			if err != nil {
				log.Fatal(err)
			}

			pprof.StartCPUProfile(f)

			sig := make(chan os.Signal, 1)
			signal.Notify(sig, os.Interrupt, os.Kill)
			<-sig
			pprof.StopCPUProfile()

			os.Exit(0)
		}()
	*/
}

func main() {

	// Initializing application
	a := app.NewAppInstance()

	err := a.Init()
	if err != nil {
		log.Fatal(err)
		return
	}

	// uninit
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, os.Kill)
		<-sig

		// Call controller to  unregister Client
		err := a.Uninit()
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}

		log.Info("Bye!")
		os.Exit(0)
	}()
	// Starting application
	err = a.Run()
	if err != nil {
		log.Fatal(err)
		return
	}
}
