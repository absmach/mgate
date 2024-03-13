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
	"github.com/joho/godotenv"
	"golang.org/x/sync/errgroup"
)

const (
	mqttWithoutTLS = "MPROXY_MQTT_WITHOUT_TLS_"
	mqttWithTLS    = "MPROXY_MQTT_WITH_TLS_"
	mqttWithmTLS   = "MPROXY_MQTT_WITH_MTLS_"

	mqttWSWithoutTLS = "MPROXY_MQTT_WS_WITHOUT_TLS_"
	mqttWSWithTLS    = "MPROXY_MQTT_WS_WITH_TLS_"
	mqttWSWithmTLS   = "MPROXY_MQTT_WS_WITH_MTLS_"

	httpWithoutTLS = "MPROXY_HTTP_WITHOUT_TLS_"
	httpWithTLS    = "MPROXY_HTTP_WITH_TLS_"
	httpWithmTLS   = "MPROXY_HTTP_WITH_MTLS_"
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

	// Loading .env file to environment
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	// MQTT Proxy Configuration without TLS
	mqttConfig := mproxy.Config{}
	if err := mqttConfig.EnvParse(env.Options{Prefix: mqttWithoutTLS}); err != nil {
		panic(err)
	}

	// MQTT Proxy without TLS
	mqttProxy := mqtt.New(mqttConfig, handler, interceptor, logger)
	g.Go(func() error {
		return mqttProxy.Listen(ctx)
	})

	// MQTT Proxy Configuration with TLS
	mqttTLSConfig := mproxy.Config{}
	if err := mqttTLSConfig.EnvParse(env.Options{Prefix: mqttWithTLS}); err != nil {
		panic(err)
	}

	// MQTT Proxy with TLS
	mqttTLSProxy := mqtt.New(mqttTLSConfig, handler, interceptor, logger)
	g.Go(func() error {
		return mqttTLSProxy.Listen(ctx)
	})

	// MQTT Proxy Configuration with mTLS
	mqttMTLSConfig := mproxy.Config{}
	if err := mqttMTLSConfig.EnvParse(env.Options{Prefix: mqttWithmTLS}); err != nil {
		panic(err)
	}

	// MQTT Proxy with mTLS
	mqttMTlsProxy := mqtt.New(mqttMTLSConfig, handler, interceptor, logger)
	g.Go(func() error {
		return mqttMTlsProxy.Listen(ctx)
	})

	// Websocket MQTT Configuration without TLS
	wsConfig := mproxy.Config{}
	if err := wsConfig.EnvParse(env.Options{Prefix: mqttWSWithoutTLS}); err != nil {
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
	if err := wsTLSConfig.EnvParse(env.Options{Prefix: mqttWSWithTLS}); err != nil {
		panic(err)
	}

	// Websocket MQTT Proxy with TLS
	wsTLSProxy := websocket.New(wsTLSConfig, handler, interceptor, logger)
	g.Go(func() error {
		return wsTLSProxy.Listen(ctx)
	})

	// Websocket MQTT Proxy Configuration with mTLS
	wsMTLSConfig := mproxy.Config{}
	if err := wsMTLSConfig.EnvParse(env.Options{Prefix: mqttWSWithmTLS}); err != nil {
		panic(err)
	}

	// HTTP Proxy with mTLS
	wsMTLSProxy := websocket.New(wsMTLSConfig, handler, interceptor, logger)
	g.Go(func() error {
		return wsMTLSProxy.Listen(ctx)
	})

	// HTTP Configuration without TLS
	httpConfig := mproxy.Config{}
	if err := httpConfig.EnvParse(env.Options{Prefix: httpWithoutTLS}); err != nil {
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
	if err := httpTLSConfig.EnvParse(env.Options{Prefix: httpWithTLS}); err != nil {
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
	if err := httpMTLSConfig.EnvParse(env.Options{Prefix: httpWithmTLS}); err != nil {
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
