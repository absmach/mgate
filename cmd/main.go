package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"

	"os"
	"os/signal"
	"strconv"
	"syscall"

	mflog "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mproxy/examples/simple"
	hproxy "github.com/mainflux/mproxy/pkg/http"
	"github.com/mainflux/mproxy/pkg/mqtt"
	"github.com/mainflux/mproxy/pkg/session"
	mptls "github.com/mainflux/mproxy/pkg/tls"
	"github.com/mainflux/mproxy/pkg/websocket"
)

const (
	// WS
	defWSHost         = "0.0.0.0"
	defWSPath         = "/mqtt"
	defWSPort         = "8080"
	defWSSPath        = "/mqtt"
	defWSSPort        = "8081"
	defWSTargetScheme = "ws"
	defWSTargetHost   = "localhost"
	defWSTargetPort   = "8888"
	defWSTargetPath   = "/mqtt"

	envWSHost         = "MPROXY_WS_HOST"
	envWSPort         = "MPROXY_WS_PORT"
	envWSPath         = "MPROXY_WS_PATH"
	envWSSPort        = "MPROXY_WSS_PORT"
	envWSSPath        = "MPROXY_WSS_PATH"
	envWSTargetScheme = "MPROXY_WS_TARGET_SCHEME"
	envWSTargetHost   = "MPROXY_WS_TARGET_HOST"
	envWSTargetPort   = "MPROXY_WS_TARGET_PORT"
	envWSTargetPath   = "MPROXY_WS_TARGET_PATH"

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

	envHTTPHost       = "MPROXY_HTTP_HOST"
	envHTTPPort       = "MPROXY_HTTP_PORT"
	envHTTPargetHost  = "MPROXY_HTTP_TARGET_HOST"
	envHTTPTargetPort = "MPROXY_HTTP_TARGET_PORT"
	envHTTPServerCert = "MPROXY_HTTP_SERVER_CERT"
	envHTTPServerKey  = "MPROXY_HTTP_SERVER_KEY"

	defHTTPHost       = "0.0.0.0"
	defHTTPPort       = "8888"
	defHTTPTargetHost = "0.0.0.0"
	defHTTPTargetPort = "8081"
	defHTTPServerCert = ""
	defHTTPServerKey  = ""
)

type config struct {
	wsHost         string
	wsPort         string
	wsPath         string
	wssPort        string
	wssPath        string
	wsTargetScheme string
	wsTargetHost   string
	wsTargetPort   string
	wsTargetPath   string

	mqttHost       string
	mqttPort       string
	mqttsPort      string
	mqttTargetHost string
	mqttTargetPort string
	clientTLS      bool
	caCerts        string
	serverCert     string
	serverKey      string

	httpConfig HTTPConfig

	logLevel string
}

type HTTPConfig struct {
	host       string
	port       string
	targetHost string
	targetPort string
	serverCert string
	serverKey  string
}

func main() {
	cfg := loadConfig()

	logger, err := mflog.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	h := simple.New(logger)

	errs := make(chan error, 3)

	ctx := context.Background()

	if cfg.clientTLS {
		tlsCfg, err := mptls.LoadTLSCfg(cfg.caCerts, cfg.serverCert, cfg.serverKey)
		if err != nil {
			errs <- err
		}

		// WSS
		logger.Info(fmt.Sprintf("Starting encrypted WebSocket proxy on port %s ", cfg.wssPort))
		go proxyWSS(cfg, tlsCfg, logger, h, errs)
		// MQTTS
		logger.Info(fmt.Sprintf("Starting MQTTS proxy on port %s ", cfg.mqttsPort))
		go proxyMQTTS(ctx, cfg, tlsCfg, logger, h, errs)
		go proxyHTTPs(ctx, cfg.httpConfig, logger, h, errs)
	} else {
		// WS
		logger.Info(fmt.Sprintf("Starting WebSocket proxy on port %s ", cfg.wsPort))
		go proxyWS(cfg, logger, h, errs)

		// MQTT
		logger.Info(fmt.Sprintf("Starting MQTT proxy on port %s ", cfg.mqttPort))
		go proxyMQTT(ctx, cfg, logger, h, errs)

		//HTTP
		logger.Info(fmt.Sprintf("Starting HTTP proxy on port %s ", cfg.httpConfig.port))
		go proxyHTTP(ctx, cfg.httpConfig, logger, h, errs)
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
		wsHost:         env(envWSHost, defWSHost),
		wsPort:         env(envWSPort, defWSPort),
		wsPath:         env(envWSPath, defWSPath),
		wssPort:        env(envWSSPort, defWSSPort),
		wssPath:        env(envWSSPath, defWSSPath),
		wsTargetScheme: env(envWSTargetScheme, defWSTargetScheme),
		wsTargetHost:   env(envWSTargetHost, defWSTargetHost),
		wsTargetPort:   env(envWSTargetPort, defWSTargetPort),
		wsTargetPath:   env(envWSTargetPath, defWSTargetPath),

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

		// HTTP
		httpConfig: HTTPConfig{
			port:       env(envHTTPPort, defHTTPPort),
			host:       env(envHTTPHost, defHTTPHost),
			targetHost: env(envHTTPargetHost, defHTTPTargetHost),
			targetPort: env(envHTTPTargetPort, defHTTPTargetPort),
			serverCert: env(envHTTPServerCert, defHTTPServerCert),
			serverKey:  env(envHTTPServerKey, defHTTPServerKey),
		},

		// Log
		logLevel: env(envLogLevel, defLogLevel),
	}
}

func proxyWS(cfg config, logger mflog.Logger, handler session.Handler, errs chan error) {
	target := fmt.Sprintf("%s:%s", cfg.wsTargetHost, cfg.wsTargetPort)
	wp := websocket.New(target, cfg.wsTargetPath, cfg.wsTargetScheme, handler, logger)
	http.Handle(cfg.wsPath, wp.Handler())

	errs <- wp.Listen(cfg.wsPort)
}

func proxyWSS(cfg config, tlsCfg *tls.Config, logger mflog.Logger, handler session.Handler, errs chan error) {
	target := fmt.Sprintf("%s:%s", cfg.wsTargetHost, cfg.wsTargetPort)
	wp := websocket.New(target, cfg.wsTargetPath, cfg.wsTargetScheme, handler, logger)
	http.Handle(cfg.wssPath, wp.Handler())
	errs <- wp.ListenTLS(tlsCfg, cfg.serverCert, cfg.serverKey, cfg.wssPort)
}

func proxyMQTT(ctx context.Context, cfg config, logger mflog.Logger, handler session.Handler, errs chan error) {
	address := fmt.Sprintf("%s:%s", cfg.mqttHost, cfg.mqttPort)
	target := fmt.Sprintf("%s:%s", cfg.mqttTargetHost, cfg.mqttTargetPort)
	mp := mqtt.New(address, target, handler, logger)

	errs <- mp.Listen(ctx)
}

func proxyMQTTS(ctx context.Context, cfg config, tlsCfg *tls.Config, logger mflog.Logger, handler session.Handler, errs chan error) {
	address := fmt.Sprintf("%s:%s", cfg.mqttHost, cfg.mqttsPort)
	target := fmt.Sprintf("%s:%s", cfg.mqttTargetHost, cfg.mqttTargetPort)
	mp := mqtt.New(address, target, handler, logger)

	errs <- mp.ListenTLS(ctx, tlsCfg)
}

func proxyHTTP(ctx context.Context, cfg HTTPConfig, logger mflog.Logger, handler session.Handler, errs chan error) {
	address := fmt.Sprintf("%s:%s", cfg.host, cfg.port)
	target := fmt.Sprintf("%s:%s", cfg.targetHost, cfg.targetPort)
	hp, err := hproxy.NewProxy(address, target, handler, logger)
	if err != nil {
		errs <- err
		return
	}
	http.HandleFunc("/", hp.Handler)
	errs <- hp.Listen()
}

func proxyHTTPs(ctx context.Context, cfg HTTPConfig, logger mflog.Logger, handler session.Handler, errs chan error) {
	address := fmt.Sprintf("%s:%s", cfg.host, cfg.port)
	target := fmt.Sprintf("%s:%s", cfg.targetHost, cfg.targetPort)
	hp, err := hproxy.NewProxy(address, target, handler, logger)
	if err != nil {
		errs <- err
		return
	}
	http.HandleFunc("/", hp.Handler)
	errs <- hp.ListenTLS(cfg.serverCert, cfg.serverKey)
}
