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
	"github.com/mainflux/mproxy/pkg/coap"
	hproxy "github.com/mainflux/mproxy/pkg/http"
	"github.com/mainflux/mproxy/pkg/mqtt"
	"github.com/mainflux/mproxy/pkg/mqtt/websocket"
	"github.com/mainflux/mproxy/pkg/session"
	mptls "github.com/mainflux/mproxy/pkg/tls"
	"github.com/mainflux/mproxy/pkg/websocket"
	"github.com/pion/dtls/v2"
)

const (
	// WS - MQTT
	defMQTTWSHost         = "0.0.0.0"
	defMQTTWSPath         = "/mqtt"
	defMQTTWSPort         = "8080"
	defMQTTWSSPath        = "/mqtt"
	defMQTTWSSPort        = "8081"
	defMQTTWSTargetScheme = "ws"
	defMQTTWSTargetHost   = "localhost"
	defMQTTWSTargetPort   = "8888"
	defMQTTWSTargetPath   = "/mqtt"

	envMQTTWSHost         = "MPROXY_MQTT_WS_HOST"
	envMQTTWSPort         = "MPROXY_MQTT_WS_PORT"
	envMQTTWSPath         = "MPROXY_MQTT_WS_PATH"
	envMQTTWSSPort        = "MPROXY_MQTT_WSS_PORT"
	envMQTTWSSPath        = "MPROXY_MQTT_WSS_PATH"
	envMQTTWSTargetScheme = "MPROXY_MQTT_WS_TARGET_SCHEME"
	envMQTTWSTargetHost   = "MPROXY_MQTT_WS_TARGET_HOST"
	envMQTTWSTargetPort   = "MPROXY_MQTT_WS_TARGET_PORT"
	envMQTTWSTargetPath   = "MPROXY_MQTT_WS_TARGET_PATH"

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

	// WS
	defWSHost       = "0.0.0.0"
	defWSPort       = "8081"
	defWSTargetHost = "ws://localhost"
	defWSTargetPort = "8889"

	envWSHost       = "MPROXY_WS_HOST"
	envWSPort       = "MPROXY_WS_PORT"
	envWSTargetHost = "MPROXY_MQTT_WS_TARGET_HOST"
	envWSTargetPort = "MPROXY_MQTT_WS_TARGET_PORT"

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
	defHTTPTargetHost = "http://localhost"
	defHTTPTargetPort = "8081"
	defHTTPServerCert = ""
	defHTTPServerKey  = ""
)

type config struct {
	clientTLS  bool
	caCerts    string
	serverCert string
	serverKey  string

	httpConfig   HTTPConfig
	mqttConfig   MQTTConfig
	wsMQTTConfig WSMQTTConfig
	wsConfig     WSConfig

	coapHost       string
	coapPort       string
	coapTLS        bool
	coapDTLS       bool
	coapTargetHost string
	coapTargetPort string

	logLevel string
}

type WSConfig struct {
	host       string
	port       string
	targetHost string
	targetPort string
}

type WSMQTTConfig struct {
	host         string
	port         string
	path         string
	wssPort      string
	wssPath      string
	targetScheme string
	targetHost   string
	targetPort   string
	targetPath   string
}

type MQTTConfig struct {
	host       string
	port       string
	mqttsPort  string
	targetHost string
	targetPort string
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

		// WSS - MQTT
		logger.Info(fmt.Sprintf("Starting encrypted WebSocket proxy on port %s ", cfg.wsMQTTConfig.wssPort))
		go proxyMQTTWSS(cfg, tlsCfg, logger, h, errs)
		// MQTTS
		logger.Info(fmt.Sprintf("Starting MQTTS proxy on port %s ", cfg.mqttConfig.mqttsPort))
		go proxyMQTTS(ctx, cfg.mqttConfig, tlsCfg, logger, h, errs)
		// WSS
		logger.Info(fmt.Sprintf("Starting WSS proxy on port %s ", cfg.wsConfig.port))
		go proxyWSS(ctx, cfg, logger, h, errs)
		// HTTPS
		logger.Info(fmt.Sprintf("Starting HTTPS proxy on port %s ", cfg.httpConfig.port))
		go proxyHTTPS(ctx, cfg.httpConfig, logger, h, errs)
	} else {
		// WS - MQTT
		logger.Info(fmt.Sprintf("Starting WebSocket proxy on port %s ", cfg.wsMQTTConfig.port))
		go proxyMQTTWS(cfg.wsMQTTConfig, logger, h, errs)

		// MQTT
		logger.Info(fmt.Sprintf("Starting MQTT proxy on port %s ", cfg.mqttConfig.port))
		go proxyMQTT(ctx, cfg.mqttConfig, logger, h, errs)
		// WS
		logger.Info(fmt.Sprintf("Starting WS proxy on port %s ", cfg.wsConfig.port))
		go proxyWS(ctx, cfg.wsConfig, logger, h, errs)
		// HTTP
		logger.Info(fmt.Sprintf("Starting HTTP proxy on port %s ", cfg.httpConfig.port))
		go proxyHTTP(ctx, cfg.httpConfig, logger, h, errs)
	}

	switch {
	case cfg.coapDTLS:
		tlsCfg, err := mptls.LoadTLSCfg(cfg.caCerts, cfg.serverCert, cfg.serverKey)
		if err != nil {
			errs <- err
		}
		dtlsCfg := &dtls.Config{
			Certificates: tlsCfg.Certificates,
			ClientCAs:    tlsCfg.ClientCAs,
		}
		go proxyCoapDTLS(cfg, dtlsCfg, logger, errs)
	case cfg.coapTLS:
		tlsCfg, err := mptls.LoadTLSCfg(cfg.caCerts, cfg.serverCert, cfg.serverKey)
		if err != nil {
			errs <- err
		}
		go proxyCoapTLS(cfg, tlsCfg, logger, errs)
	default:
		go proxyCoap(cfg, logger, errs)
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
		wsMQTTConfig: WSMQTTConfig{
			host:         env(envMQTTWSHost, defMQTTWSHost),
			port:         env(envMQTTWSPort, defMQTTWSPort),
			path:         env(envMQTTWSPath, defMQTTWSPath),
			wssPort:      env(envMQTTWSSPort, defMQTTWSSPort),
			wssPath:      env(envMQTTWSSPath, defMQTTWSSPath),
			targetScheme: env(envMQTTWSTargetScheme, defMQTTWSTargetScheme),
			targetHost:   env(envMQTTWSTargetHost, defMQTTWSTargetHost),
			targetPort:   env(envMQTTWSTargetPort, defMQTTWSTargetPort),
			targetPath:   env(envMQTTWSTargetPath, defMQTTWSTargetPath),
		},

		// MQTT
		mqttConfig: MQTTConfig{
			host:       env(envMQTTHost, defMQTTHost),
			port:       env(envMQTTPort, defMQTTPort),
			mqttsPort:  env(envMQTTSPort, defMQTTSPort),
			targetHost: env(envMQTTTargetHost, defMQTTTargetHost),
			targetPort: env(envMQTTTargetPort, defMQTTTargetPort),
		},
		clientTLS:  tls,
		caCerts:    env(envCACerts, defCACerts),
		serverCert: env(envServerCert, defServerCert),
		serverKey:  env(envServerKey, defServerKey),

		// HTTP
		httpConfig: HTTPConfig{
			port:       env(envHTTPPort, defHTTPPort),
			host:       env(envHTTPHost, defHTTPHost),
			targetHost: env(envHTTPargetHost, defHTTPTargetHost),
			targetPort: env(envHTTPTargetPort, defHTTPTargetPort),
			serverCert: env(envHTTPServerCert, defHTTPServerCert),
			serverKey:  env(envHTTPServerKey, defHTTPServerKey),
		},

		// WS
		wsConfig: WSConfig{
			host:       env(envWSHost, defWSHost),
			port:       env(envWSPort, defWSPort),
			targetHost: env(envWSTargetHost, defWSTargetHost),
			targetPort: env(envWSTargetPort, defWSTargetPort),
		},

		// Log
		logLevel: env(envLogLevel, defLogLevel),
	}
}

func proxyMQTTWS(cfg WSMQTTConfig, logger mflog.Logger, handler session.Handler, errs chan error) {
	target := fmt.Sprintf("%s:%s", cfg.targetHost, cfg.targetPort)
	wp := websocket.New(target, cfg.targetPath, cfg.targetScheme, handler, logger)
	http.Handle(cfg.path, wp.Handler())

	errs <- wp.Listen(cfg.port)
}

func proxyMQTTWSS(cfg config, tlsCfg *tls.Config, logger mflog.Logger, handler session.Handler, errs chan error) {
	target := fmt.Sprintf("%s:%s", cfg.wsMQTTConfig.targetHost, cfg.wsMQTTConfig.targetPort)
	wp := websocket.New(target, cfg.wsMQTTConfig.targetPath, cfg.wsMQTTConfig.targetScheme, handler, logger)
	http.Handle(cfg.wsMQTTConfig.wssPath, wp.Handler())
	errs <- wp.ListenTLS(tlsCfg, cfg.serverCert, cfg.serverKey, cfg.wsMQTTConfig.wssPort)
}

func proxyMQTT(ctx context.Context, cfg MQTTConfig, logger mflog.Logger, handler session.Handler, errs chan error) {
	address := fmt.Sprintf("%s:%s", cfg.host, cfg.port)
	target := fmt.Sprintf("%s:%s", cfg.targetHost, cfg.targetPort)
	mp := mqtt.New(address, target, handler, logger)

	errs <- mp.Listen(ctx)
}

func proxyMQTTS(ctx context.Context, cfg MQTTConfig, tlsCfg *tls.Config, logger mflog.Logger, handler session.Handler, errs chan error) {
	address := fmt.Sprintf("%s:%s", cfg.host, cfg.mqttsPort)
	target := fmt.Sprintf("%s:%s", cfg.targetHost, cfg.targetPort)
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

func proxyHTTPS(ctx context.Context, cfg HTTPConfig, logger mflog.Logger, handler session.Handler, errs chan error) {
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

func proxyWS(ctx context.Context, cfg WSConfig, logger mflog.Logger, handler session.Handler, errs chan error) {
	address := fmt.Sprintf("%s:%s", cfg.host, cfg.port)
	target := fmt.Sprintf("%s:%s", cfg.targetHost, cfg.targetPort)
	wp, err := websockets.NewProxy(address, target, logger, handler)
	if err != nil {
		errs <- err
	}
	errs <- wp.Listen()
}

func proxyWSS(ctx context.Context, cfg config, logger mflog.Logger, handler session.Handler, errs chan error) {
	address := fmt.Sprintf("%s:%s", cfg.wsConfig.host, cfg.wsConfig.port)
	target := fmt.Sprintf("%s:%s", cfg.wsConfig.targetHost, cfg.wsConfig.targetPort)
	wp, err := websockets.NewProxy(address, target, logger, handler)
	if err != nil {
		errs <- err
	}
	errs <- wp.ListenTLS(cfg.serverCert, cfg.serverKey)
func proxyCoapTLS(cfg config, tlsCfg *tls.Config, logger mflog.Logger, errs chan error) {
	address := fmt.Sprintf("%s:%s", cfg.coapHost, cfg.coapPort)
	target := fmt.Sprintf("%s:%s", cfg.coapTargetHost, cfg.coapTargetPort)
	cp, err := coap.NewProxy(address, target, logger)
	if err != nil {
		errs <- err
	}

	errs <- cp.ListenTLS(tlsCfg)
}

func proxyCoapDTLS(cfg config, dtlsCfg *dtls.Config, logger mflog.Logger, errs chan error) {
	address := fmt.Sprintf("%s:%s", cfg.coapHost, cfg.coapPort)
	target := fmt.Sprintf("%s:%s", cfg.coapTargetHost, cfg.coapTargetPort)
	cp, err := coap.NewProxy(address, target, logger)
	if err != nil {
		errs <- err
	}

	errs <- cp.ListenDLS(dtlsCfg)
}

func proxyCoap(cfg config, logger mflog.Logger, errs chan error) {
	address := fmt.Sprintf("%s:%s", cfg.coapHost, cfg.coapPort)
	target := fmt.Sprintf("%s:%s", cfg.coapTargetHost, cfg.coapTargetPort)
	cp, err := coap.NewProxy(address, target, logger)
	if err != nil {
		errs <- err
	}

	errs <- cp.Listen()
}
