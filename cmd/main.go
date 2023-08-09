package main

import (
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"rbac-collector/collectors"
	"rbac-collector/config"
	"sync"
	"syscall"
	"time"
)

func main() {
	conf := config.ReadConfig()
	configureLogger(conf)
	log.Infof("Starting RBAC collector with config: %+v", conf)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Info("Received SIGTERM, exiting gracefully...")
		os.Exit(0)
	}()

	for {
		wg := &sync.WaitGroup{}
		wg.Add(1)
		go collectClusterPermissions(conf, wg)
		wg.Wait()

		log.Infof("Sleeping for %d minutes", conf.ReconcileLoopInterval)
		time.Sleep(time.Duration(conf.ReconcileLoopInterval) * time.Minute)
		log.Infof("Waking up...")
	}

}

func configureLogger(conf config.Configuration) {
	log.SetFormatter(&log.JSONFormatter{
		TimestampFormat:   "2023-12-01 - 15:04:05",
		DisableTimestamp:  false,
		DisableHTMLEscape: false,
		DataKey:           "",
		CallerPrettyfier:  nil,
		PrettyPrint:       false,
	})
	log.SetOutput(os.Stdout)
	lvl, err := log.ParseLevel(conf.LogLevel)
	if err != nil {
		lvl = log.InfoLevel
		log.Info("Log level not set, defaulting to info")
	}
	log.SetLevel(lvl)
}

func collectClusterPermissions(conf config.Configuration, wg *sync.WaitGroup) {
	defer wg.Done()
	apiCollector, err := collectors.NewKubeApiCollector(conf)
	if err != nil {
		log.Fatal(err)
	}
	apiCollector.Collect()
}
