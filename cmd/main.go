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
	"github.com/absmach/mproxy/pkg/mqtt"
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

	mqttConfig := mproxy.Config{}
	mqttConfigEnv := map[string]string{
		"ADDRESS": ":1884",
		"TARGET":  "localhost:1883",
	}
	if err := mqttConfig.EnvParse(env.Options{Environment: mqttConfigEnv}); err != nil {
		panic(err)
	}

	mqttProxy := mqtt.New(mqttConfig, handler, interceptor, logger)
	g.Go(func() error {
		return mqttProxy.Listen(ctx)
	})

	mqttTlsConfig := mproxy.Config{}
	mqttTlsConfigEnv := map[string]string{
		"ADDRESS":                        ":8883",
		"TARGET":                         "localhost:1883",
		"CERT_FILE":                      "ssl/certs/server.crt",
		"KEY_FILE":                       "ssl/certs/server.key",
		"SERVER_CA_FILE":                 "ssl/certs/ca.crt",
		"CLIENT_CA_FILE":                 "ssl/certs/ca.crt",
		"PREFIX_PATH":                    "",
		"CLIENT_CERT_VALIDATION_METHODS": "OCSP",
	}
	if err := mqttTlsConfig.EnvParse(env.Options{Environment: mqttTlsConfigEnv}); err != nil {
		panic(err)
	}

	mqttTlsProxy := mqtt.New(mqttTlsConfig, handler, interceptor, logger)
	g.Go(func() error {
		return mqttTlsProxy.Listen(ctx)
	})

	g.Go(func() error {
		return StopSignalHandler(ctx, cancel, logger)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("mqtt proxy service terminated with error: %s", err))
	} else {
		logger.Info("mqtt proxy service stopped")
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
