package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
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
	defMQTTSPort      = "8883"
	defMQTTTargetHost = "0.0.0.0"
	defMQTTTargetPort = "1884"
	defCACerts        = ""
	defServerCert     = ""
	defServerKey      = ""

	envMQTTHost       = "MPROXY_MQTT_HOST"
	envMQTTPort       = "MPROXY_MQTT_PORT"
	envMQTTSPort      = "MPROXY_MQTTS_PORT"
	envMQTTTargetHost = "MPROXY_MQTT_TARGET_HOST"
	envMQTTTargetPort = "MPROXY_MQTT_TARGET_PORT"
	envCACerts        = "MPROXY_CA_CERTS"
	envServerCert     = "MPROXY_SERVER_CERT"
	envServerKey      = "MPROXY_SERVER_KEY"

	defClientTLS = "false"
	envClientTLS = "MPROXY_CLIENT_TLS"
	defLogLevel  = "debug"
	envLogLevel  = "MPROXY_LOG_LEVEL"
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
	mqttsPort      string
	mqttTargetHost string
	mqttTargetPort string
	clientTLS      bool
	caCerts        string
	serverCert     string
	serverKey      string

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

	if cfg.clientTLS {
		// MQTTS
		logger.Info(fmt.Sprintf("Starting MQTTS proxy on port %s ", cfg.mqttsPort))
		go proxyMQTTS(cfg, logger, h, errs)
	} else {
		// HTTP
		logger.Info(fmt.Sprintf("Starting HTTP proxy on port %s ", cfg.httpPort))
		go proxyHTTP(cfg, logger, h, errs)

		// MQTT
		logger.Info(fmt.Sprintf("Starting MQTT proxy on port %s ", cfg.mqttPort))
		go proxyMQTT(cfg, logger, h, errs)
	}

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
	tls, err := strconv.ParseBool(env(envClientTLS, defClientTLS))
	if err != nil {
		log.Fatalf("Invalid value passed for %s\n", envClientTLS)
	}

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
		mqttsPort:      env(envMQTTSPort, defMQTTSPort),
		mqttTargetHost: env(envMQTTTargetHost, defMQTTTargetHost),
		mqttTargetPort: env(envMQTTTargetPort, defMQTTTargetPort),
		clientTLS:      tls,
		caCerts:        env(envCACerts, defCACerts),
		serverCert:     env(envServerCert, defServerCert),
		serverKey:      env(envServerKey, defServerKey),

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
	mp := mqtt.New(address, target, handler, logger, cfg.caCerts, cfg.serverCert, cfg.serverKey)

	errs <- mp.Listen()
}

func proxyMQTTS(cfg config, logger mflog.Logger, handler session.Handler, errs chan error) {
	address := fmt.Sprintf("%s:%s", cfg.mqttHost, cfg.mqttsPort)
	target := fmt.Sprintf("%s:%s", cfg.mqttTargetHost, cfg.mqttTargetPort)
	mp := mqtt.New(address, target, handler, logger, cfg.caCerts, cfg.serverCert, cfg.serverKey)

	errs <- mp.ListenTLS()
}
