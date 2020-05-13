package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/mainflux/mainflux"
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
	defKeepAlive       = "true"
	defKeepAlivePeriod = "60" // In seconds
	defMQTTHost        = "0.0.0.0"
	defMQTTPort        = "1883"
	defMQTTTargetHost  = "0.0.0.0"
	defMQTTTargetPort  = "1884"

	envKeepAlive       = "MPROXY_KEEP_ALIVE"
	envKeepAlivePeriod = "MPROXY_KEEP_ALIVE_PERIOD"
	envMQTTHost        = "MPROXY_MQTT_HOST"
	envMQTTPort        = "MPROXY_MQTT_PORT"
	envMQTTTargetHost  = "MPROXY_MQTT_TARGET_HOST"
	envMQTTTargetPort  = "MPROXY_MQTT_TARGET_PORT"

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

	keepAlive       bool
	keepAlivePariod time.Duration

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
	ka, err := strconv.ParseBool(mainflux.Env(envKeepAlive, defKeepAlive))
	if err != nil {
		log.Fatalf("Invalid value passed for %s\n", envKeepAlive)
	}
	kaPeriod, err := strconv.ParseInt(mainflux.Env(envKeepAlivePeriod, defKeepAlivePeriod), 10, 64)
	if err != nil {
		log.Fatalf("Invalid %s value: %s", envKeepAlivePeriod, err.Error())
	}

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

		// mProxy config
		keepAlive:       ka,
		keepAlivePariod: time.Duration(kaPeriod) * time.Second,
		logLevel:        env(envLogLevel, defLogLevel),
	}
}

func proxyHTTP(cfg config, logger mflog.Logger, evt session.Handler, errs chan error) {
	target := fmt.Sprintf("%s:%s", cfg.httpTargetHost, cfg.httpTargetPort)
	wp := websocket.New(target, cfg.httpTargetPath, cfg.httpScheme, evt, logger)
	http.Handle("/mqtt", wp.Handler())

	p := fmt.Sprintf(":%s", cfg.httpPort)
	errs <- http.ListenAndServe(p, nil)
}

func proxyMQTT(cfg config, logger mflog.Logger, handler session.Handler, errs chan error) {
	address := fmt.Sprintf("%s:%s", cfg.mqttHost, cfg.mqttPort)
	target := fmt.Sprintf("%s:%s", cfg.mqttTargetHost, cfg.mqttTargetPort)
	mp := mqtt.New(address, target, cfg.keepAlive, cfg.keepAlivePariod, handler, logger)

	errs <- mp.Proxy()
}
