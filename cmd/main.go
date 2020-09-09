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
	// WS
	defWSHost       = "0.0.0.0"
	defWSPort       = "8080"
	defWSScheme     = "ws"
	defWSTargetHost = "localhost"
	defWSTargetPort = "8888"
	defWSTargetPath = "/mqtt"

	envWSHost       = "MPROXY_WS_HOST"
	envWSPort       = "MPROXY_WS_PORT"
	envWSScheme     = "MPROXY_WS_SCHEMA"
	envWSTargetHost = "MPROXY_WS_TARGET_HOST"
	envWSTargetPort = "MPROXY_WS_TARGET_PORT"
	envWSTargetPath = "MPROXY_WS_TARGET_PATH"

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
	wsHost       string
	wsPort       string
	wsScheme     string
	wsTargetHost string
	wsTargetPort string
	wsTargetPath string

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

	// WS
	logger.Info(fmt.Sprintf("Starting WebSocket proxy on port %s ", cfg.wsPort))
	go proxyWS(cfg, logger, h, errs)

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
		// WS
		wsHost:       env(envWSHost, defWSHost),
		wsPort:       env(envWSPort, defWSPort),
		wsScheme:     env(envWSScheme, defWSScheme),
		wsTargetHost: env(envWSTargetHost, defWSTargetHost),
		wsTargetPort: env(envWSTargetPort, defWSTargetPort),
		wsTargetPath: env(envWSTargetPath, defWSTargetPath),

		// MQTT
		mqttHost:       env(envMQTTHost, defMQTTHost),
		mqttPort:       env(envMQTTPort, defMQTTPort),
		mqttTargetHost: env(envMQTTTargetHost, defMQTTTargetHost),
		mqttTargetPort: env(envMQTTTargetPort, defMQTTTargetPort),

		// Log
		logLevel: env(envLogLevel, defLogLevel),
	}
}

func proxyWS(cfg config, logger mflog.Logger, handler session.Handler, errs chan error) {
	target := fmt.Sprintf("%s:%s", cfg.wsTargetHost, cfg.wsTargetPort)
	wp := websocket.New(target, cfg.wsTargetPath, cfg.wsScheme, handler, logger)
	http.Handle("/mqtt", wp.Handler())

	p := fmt.Sprintf(":%s", cfg.wsPort)
	errs <- http.ListenAndServe(p, nil)
}

func proxyMQTT(cfg config, logger mflog.Logger, handler session.Handler, errs chan error) {
	address := fmt.Sprintf("%s:%s", cfg.mqttHost, cfg.mqttPort)
	target := fmt.Sprintf("%s:%s", cfg.mqttTargetHost, cfg.mqttTargetPort)
	mp := mqtt.New(address, target, handler, logger)

	errs <- mp.Proxy()
}
