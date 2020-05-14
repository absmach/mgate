package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	mflog "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mproxy/examples/simple"
	"github.com/mainflux/mproxy/pkg/mqtt"
	"github.com/mainflux/mproxy/pkg/session"
	"github.com/mainflux/mproxy/pkg/websocket"
)

const (
	// HTTP
	defHTTPHost       = "0.0.0.0"
	defHTTPPort       = "8080"
	defHTTPScheme     = "ws"
	defHTTPTargetHost = "localhost"
	defHTTPTargetPort = "8888"
	defHTTPTargetPath = "/mqtt"

	envHTTPHost       = "MPROXY_HTTP_HOST"
	envHTTPPort       = "MPROXY_HTTP_PORT"
	envHTTPScheme     = "MPROXY_HTTP_SCHEMA"
	envHTTPTargetHost = "MPROXY_HTTP_TARGET_HOST"
	envHTTPTargetPort = "MPROXY_HTTP_TARGET_PORT"
	envHTTPTargetPath = "MPROXY_HTTP_TARGET_PATH"

	// MQTT
	defMQTTHost       = "0.0.0.0"
	defMQTTPort       = "1883"
	defMQTTTargetHost = "0.0.0.0"
	defMQTTTargetPort = "1884"

	envMQTTHost       = "MPROXY_MQTT_HOST"
	envMQTTPort       = "MPROXY_MQTT_PORT"
	envMQTTTargetHost = "MPROXY_MQTT_TARGET_HOST"
	envMQTTTargetPort = "MPROXY_MQTT_TARGET_PORT"

	defLogLevel = "debug"
	envLogLevel = "MPROXY_LOG_LEVEL"
)

type config struct {
	httpHost       string
	httpPort       string
	httpScheme     string
	httpTargetHost string
	httpTargetPort string
	httpTargetPath string

	mqttHost       string
	mqttPort       string
	mqttTargetHost string
	mqttTargetPort string

	logLevel string
}

func main() {
	cfg := loadConfig()

	logger, err := mflog.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	h := simple.New(logger)

	errs := make(chan error, 3)

	// HTTP
	logger.Info(fmt.Sprintf("Starting HTTP proxy on port %s ", cfg.httpPort))
	go proxyHTTP(cfg, logger, h, errs)

	// MQTT
	logger.Info(fmt.Sprintf("Starting MQTT proxy on port %s ", cfg.mqttPort))
	go proxyMQTT(cfg, logger, h, errs)

	go func() {
		c := make(chan os.Signal, 2)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	err = <-errs
	logger.Error(fmt.Sprintf("mProxy terminated: %s", err))
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}

func loadConfig() config {
	return config{
		// HTTP
		httpHost:       env(envHTTPHost, defHTTPHost),
		httpPort:       env(envHTTPPort, defHTTPPort),
		httpScheme:     env(envHTTPScheme, defHTTPScheme),
		httpTargetHost: env(envHTTPTargetHost, defHTTPTargetHost),
		httpTargetPort: env(envHTTPTargetPort, defHTTPTargetPort),
		httpTargetPath: env(envHTTPTargetPath, defHTTPTargetPath),

		// MQTT
		mqttHost:       env(envMQTTHost, defMQTTHost),
		mqttPort:       env(envMQTTPort, defMQTTPort),
		mqttTargetHost: env(envMQTTTargetHost, defMQTTTargetHost),
		mqttTargetPort: env(envMQTTTargetPort, defMQTTTargetPort),

		// Log
		logLevel: env(envLogLevel, defLogLevel),
	}
}

func proxyHTTP(cfg config, logger mflog.Logger, handler session.Handler, errs chan error) {
	target := fmt.Sprintf("%s:%s", cfg.httpTargetHost, cfg.httpTargetPort)
	wp := websocket.New(target, cfg.httpTargetPath, cfg.httpScheme, handler, logger)
	http.Handle("/mqtt", wp.Handler())

	p := fmt.Sprintf(":%s", cfg.httpPort)
	errs <- http.ListenAndServe(p, nil)
}

func proxyMQTT(cfg config, logger mflog.Logger, handler session.Handler, errs chan error) {
	address := fmt.Sprintf("%s:%s", cfg.mqttHost, cfg.mqttPort)
	target := fmt.Sprintf("%s:%s", cfg.mqttTargetHost, cfg.mqttTargetPort)
	mp := mqtt.New(address, target, handler, logger)

	errs <- mp.Proxy()
}
