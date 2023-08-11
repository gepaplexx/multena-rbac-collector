package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	ginprometheus "github.com/zsais/go-gin-prometheus"
	"net/http"
	"os"
	"os/signal"
	"rbac-collector/collectors"
	"rbac-collector/config"
	"syscall"
	"time"
)

func main() {
	conf := config.ReadConfig()
	configureLogger(conf)
	log.Infof("Starting RBAC collector with config: %+v", conf)

	setupGracefulShutdown()

	go runMonitoringEndpoints(conf)
	runRbacCollector(conf)
}

func runMonitoringEndpoints(conf config.Configuration) {
	engine := setupGinEngine()
	setEngineRoutes(engine)
	if err := engine.Run(fmt.Sprintf(":%d", conf.Port)); err != nil {
		log.Fatalf("error running application: %v", err)
	}
}

func runRbacCollector(conf config.Configuration) {
	for {
		collectClusterPermissions(conf)
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
	setLogLevel(conf.LogLevel)
}

func collectClusterPermissions(conf config.Configuration) {
	apiCollector, err := collectors.NewKubeApiCollector(conf)
	if err != nil {
		log.Fatal(err)
	}
	apiCollector.Collect()
}

func setupGracefulShutdown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Info("Received SIGTERM, exiting gracefully...")
		os.Exit(0)
	}()
}

func setupGinEngine() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery(), gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/healthz", "/readyz", "/metrics", "/favicon.ico"},
	}))
	ginprometheus.NewPrometheus("gin").Use(engine)
	return engine
}

func setEngineRoutes(engine *gin.Engine) {
	engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, "OK")
	})
	engine.GET("/readyz", func(c *gin.Context) {
		c.JSON(http.StatusOK, "OK")
	})
}

func setLogLevel(logLevel string) {
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		level = log.InfoLevel
		log.Info("Log level not set, defaulting to info")
	}
	log.SetLevel(level)
}
