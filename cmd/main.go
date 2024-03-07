// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/absmach/mproxy"
	"github.com/absmach/mproxy/examples/simple"
	"github.com/absmach/mproxy/pkg/http"
	"github.com/absmach/mproxy/pkg/mqtt"
	"github.com/absmach/mproxy/pkg/mqtt/websocket"
	"github.com/absmach/mproxy/pkg/session"
	"github.com/caarlos0/env/v10"
	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(logHandler)

	handler := simple.New(logger)

	var interceptor session.Interceptor

	// MQTT Proxy Configuration without TLS
	mqttConfig := mproxy.Config{}
	mqttConfigEnv := map[string]string{
		"ADDRESS": ":1884",
		"TARGET":  "localhost:1883",
	}
	if err := mqttConfig.EnvParse(env.Options{Environment: mqttConfigEnv}); err != nil {
		panic(err)
	}

	// MQTT Proxy without TLS
	mqttProxy := mqtt.New(mqttConfig, handler, interceptor, logger)
	g.Go(func() error {
		return mqttProxy.Listen(ctx)
	})

	// MQTT Proxy Configuration with TLS
	mqttTLSConfig := mproxy.Config{}
	mqttTLSConfigEnv := map[string]string{
		"ADDRESS":                        ":8883",
		"TARGET":                         "localhost:1883",
		"CERT_FILE":                      "ssl/certs/server.crt",
		"KEY_FILE":                       "ssl/certs/server.key",
		"SERVER_CA_FILE":                 "ssl/certs/ca.crt",
		"PREFIX_PATH":                    "",
		"CLIENT_CERT_VALIDATION_METHODS": "",
		"OCSP_RESPONDER_URL":             "",
	}
	if err := mqttTLSConfig.EnvParse(env.Options{Environment: mqttTLSConfigEnv}); err != nil {
		panic(err)
	}

	// MQTT Proxy with TLS
	mqttTLSProxy := mqtt.New(mqttTLSConfig, handler, interceptor, logger)
	g.Go(func() error {
		return mqttTLSProxy.Listen(ctx)
	})

	// MQTT Proxy Configuration with mTLS
	mqttMTLSConfig := mproxy.Config{}
	mqttMTLSConfigEnv := map[string]string{
		"ADDRESS":                        ":8884",
		"TARGET":                         "localhost:1883",
		"CERT_FILE":                      "ssl/certs/server.crt",
		"KEY_FILE":                       "ssl/certs/server.key",
		"SERVER_CA_FILE":                 "ssl/certs/ca.crt",
		"CLIENT_CA_FILE":                 "ssl/certs/ca.crt",
		"PREFIX_PATH":                    "",
		"CLIENT_CERT_VALIDATION_METHODS": "OCSP",
		"OCSP_RESPONDER_URL":             "http://localhost:8080/ocsp",
	}
	if err := mqttMTLSConfig.EnvParse(env.Options{Environment: mqttMTLSConfigEnv}); err != nil {
		panic(err)
	}

	// MQTT Proxy with mTLS
	mqttMTlsProxy := mqtt.New(mqttMTLSConfig, handler, interceptor, logger)
	g.Go(func() error {
		return mqttMTlsProxy.Listen(ctx)
	})

	// Websocket MQTT Configuration without TLS
	wsConfig := mproxy.Config{}
	wsConfigEnv := map[string]string{
		"ADDRESS": ":8083",
		"TARGET":  "ws://localhost:8000/",
	}
	if err := wsConfig.EnvParse(env.Options{Environment: wsConfigEnv}); err != nil {
		panic(err)
	}

	// Websocket MQTT Proxy without TLS
	wsProxy := websocket.New(wsConfig, handler, interceptor, logger)
	g.Go(func() error {
		return wsProxy.Listen(ctx)
	})

	g.Go(func() error {
		return StopSignalHandler(ctx, cancel, logger)
	})

	// Websocket MQTT Proxy Configuration with TLS
	wsTLSConfig := mproxy.Config{}
	wsTLSConfigEnv := map[string]string{
		"ADDRESS":                        ":8084",
		"TARGET":                         "ws://localhost:8000/",
		"CERT_FILE":                      "ssl/certs/server.crt",
		"KEY_FILE":                       "ssl/certs/server.key",
		"SERVER_CA_FILE":                 "ssl/certs/ca.crt",
		"PREFIX_PATH":                    "",
		"CLIENT_CERT_VALIDATION_METHODS": "",
		"OCSP_RESPONDER_URL":             "",
	}
	if err := wsTLSConfig.EnvParse(env.Options{Environment: wsTLSConfigEnv}); err != nil {
		panic(err)
	}

	// Websocket MQTT Proxy with TLS
	wsTLSProxy := websocket.New(wsTLSConfig, handler, interceptor, logger)
	g.Go(func() error {
		return wsTLSProxy.Listen(ctx)
	})

	// Websocket MQTT Proxy Configuration with mTLS
	wsMTLSConfig := mproxy.Config{}
	wsMTLSConfigEnv := map[string]string{
		"ADDRESS":                        ":8085",
		"TARGET":                         "ws://localhost:8000/",
		"CERT_FILE":                      "ssl/certs/server.crt",
		"KEY_FILE":                       "ssl/certs/server.key",
		"SERVER_CA_FILE":                 "ssl/certs/ca.crt",
		"CLIENT_CA_FILE":                 "ssl/certs/ca.crt",
		"PREFIX_PATH":                    "/mqtt",
		"CLIENT_CERT_VALIDATION_METHODS": "OCSP",
		"OCSP_RESPONDER_URL":             "http://localhost:8080/ocsp",
	}
	if err := wsMTLSConfig.EnvParse(env.Options{Environment: wsMTLSConfigEnv}); err != nil {
		panic(err)
	}

	// HTTP Proxy with mTLS
	wsMTLSProxy := websocket.New(wsMTLSConfig, handler, interceptor, logger)
	g.Go(func() error {
		return wsMTLSProxy.Listen(ctx)
	})

	// HTTP Configuration without TLS
	httpConfig := mproxy.Config{}
	httpConfigEnv := map[string]string{
		"ADDRESS":     ":8086",
		"TARGET":      "http://localhost:8888/",
		"PREFIX_PATH": "/messages",
	}
	if err := httpConfig.EnvParse(env.Options{Environment: httpConfigEnv}); err != nil {
		panic(err)
	}

	// HTTP Proxy without TLS
	httpProxy, err := http.NewProxy(httpConfig, handler, logger)
	if err != nil {
		panic(err)
	}
	g.Go(func() error {
		return httpProxy.Listen(ctx)
	})

	// Websocket MQTT Proxy Configuration with TLS
	httpTLSConfig := mproxy.Config{}
	httpTLSConfigEnv := map[string]string{
		"ADDRESS":                        ":8087",
		"TARGET":                         "http://localhost:8888/",
		"CERT_FILE":                      "ssl/certs/server.crt",
		"KEY_FILE":                       "ssl/certs/server.key",
		"SERVER_CA_FILE":                 "ssl/certs/ca.crt",
		"PREFIX_PATH":                    "/messages",
		"CLIENT_CERT_VALIDATION_METHODS": "",
		"OCSP_RESPONDER_URL":             "",
	}
	if err := httpTLSConfig.EnvParse(env.Options{Environment: httpTLSConfigEnv}); err != nil {
		panic(err)
	}

	// HTTP Proxy with TLS
	httpTLSProxy, err := http.NewProxy(httpTLSConfig, handler, logger)
	if err != nil {
		panic(err)
	}
	g.Go(func() error {
		return httpTLSProxy.Listen(ctx)
	})

	// HTTP Proxy Configuration with mTLS
	httpMTLSConfig := mproxy.Config{}
	httpMTLSConfigEnv := map[string]string{
		"ADDRESS":                        ":8088",
		"TARGET":                         "http://localhost:8888/",
		"CERT_FILE":                      "ssl/certs/server.crt",
		"KEY_FILE":                       "ssl/certs/server.key",
		"SERVER_CA_FILE":                 "ssl/certs/ca.crt",
		"CLIENT_CA_FILE":                 "ssl/certs/ca.crt",
		"PREFIX_PATH":                    "/messages",
		"CLIENT_CERT_VALIDATION_METHODS": "OCSP",
		"OCSP_RESPONDER_URL":             "http://localhost:8080/ocsp",
	}
	if err := httpMTLSConfig.EnvParse(env.Options{Environment: httpMTLSConfigEnv}); err != nil {
		panic(err)
	}

	// HTTP Proxy with mTLS
	httpMTLSProxy, err := http.NewProxy(httpMTLSConfig, handler, logger)
	if err != nil {
		panic(err)
	}
	g.Go(func() error {
		return httpMTLSProxy.Listen(ctx)
	})

	g.Go(func() error {
		return StopSignalHandler(ctx, cancel, logger)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("mProxy service terminated with error: %s", err))
	} else {
		logger.Info("mProxy service stopped")
	}
}

func StopSignalHandler(ctx context.Context, cancel context.CancelFunc, logger *slog.Logger) error {
	c := make(chan os.Signal, 2)
	signal.Notify(c, syscall.SIGINT, syscall.SIGABRT)
	select {
	case <-c:
		cancel()
		return nil
	case <-ctx.Done():
		return nil
	}
}
