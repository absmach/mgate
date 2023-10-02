// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mainflux/mproxy/pkg/logger"
	"github.com/ory/dockertest/v3"
)

var testURL string

var errSessionMissing = errors.New("session is missing")

var _ Handler = (*customHandler)(nil)

// Handler implements mqtt.Handler interface
type customHandler struct {
	logger logger.Logger
}

// New creates new Event entity
func newHandler(logger logger.Logger) *customHandler {
	return &customHandler{
		logger: logger,
	}
}

// AuthConnect is called on device connection,
// prior forwarding to the MQTT broker
func (h *customHandler) AuthConnect(ctx context.Context) error {
	s, ok := FromContext(ctx)
	if !ok {
		h.logger.Error(errSessionMissing.Error())
		return errSessionMissing
	}
	h.logger.Info(fmt.Sprintf("AuthConnect() - sessionID: %s, username: %s, password: %s, client_CN: %s", s.ID, s.Username, string(s.Password), s.Cert.Subject.CommonName))
	return nil
}

// AuthPublish is called on device publish,
// prior forwarding to the MQTT broker
func (h *customHandler) AuthPublish(ctx context.Context, topic *string, payload *[]byte) error {
	s, ok := FromContext(ctx)
	if !ok {
		h.logger.Error(errSessionMissing.Error())
		return errSessionMissing
	}
	h.logger.Info(fmt.Sprintf("AuthPublish() - sessionID: %s, topic: %s, payload: %s", s.ID, *topic, string(*payload)))

	return nil
}

// AuthSubscribe is called on device publish,
// prior forwarding to the MQTT broker
func (h *customHandler) AuthSubscribe(ctx context.Context, topics *[]string) error {
	s, ok := FromContext(ctx)
	if !ok {
		h.logger.Error(errSessionMissing.Error())
		return errSessionMissing
	}
	h.logger.Info(fmt.Sprintf("AuthSubscribe() - sessionID: %s, topics: %s", s.ID, strings.Join(*topics, ",")))
	return nil
}

// Connect - after client successfully connected
func (h *customHandler) Connect(ctx context.Context) error {
	s, ok := FromContext(ctx)
	if !ok {
		h.logger.Error(errSessionMissing.Error())
		return errSessionMissing
	}
	h.logger.Info(fmt.Sprintf("Connect() - username: %s, sessionID: %s", s.Username, s.ID))
	return nil
}

// Publish - after client successfully published
func (h *customHandler) Publish(ctx context.Context, topic *string, payload *[]byte) error {
	s, ok := FromContext(ctx)
	if !ok {
		h.logger.Error(errSessionMissing.Error())
		return errSessionMissing
	}
	h.logger.Info(fmt.Sprintf("Publish() - username: %s, sessionID: %s, topic: %s, payload: %s", s.Username, s.ID, *topic, string(*payload)))
	return nil
}

// Subscribe - after client successfully subscribed
func (h *customHandler) Subscribe(ctx context.Context, topics *[]string) error {
	s, ok := FromContext(ctx)
	if !ok {
		h.logger.Error(errSessionMissing.Error())
		return errSessionMissing
	}
	h.logger.Info(fmt.Sprintf("Subscribe() - username: %s, sessionID: %s, topics: %s", s.Username, s.ID, strings.Join(*topics, ",")))
	return nil
}

// Unsubscribe - after client unsubscribed
func (h *customHandler) Unsubscribe(ctx context.Context, topics *[]string) error {
	s, ok := FromContext(ctx)
	if !ok {
		h.logger.Error(errSessionMissing.Error())
		return errSessionMissing
	}
	h.logger.Info(fmt.Sprintf("Unsubscribe() - username: %s, sessionID: %s, topics: %s", s.Username, s.ID, strings.Join(*topics, ",")))
	return nil
}

// Disconnect on connection lost
func (h *customHandler) Disconnect(ctx context.Context) error {
	s, ok := FromContext(ctx)
	if !ok {
		h.logger.Error(errSessionMissing.Error())
		return errSessionMissing
	}
	h.logger.Info(fmt.Sprintf("Disconnect() - client with username: %s and ID: %s disconnected", s.Username, s.ID))
	return nil
}

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	container, err := pool.Run("ksdn117/tcp-udp-test", "latest", nil)
	if err != nil {
		log.Fatalf("Could not start container: %s", err)
	}

	testURL = fmt.Sprintf("http://localhost:%s", container.GetPort("1234/tcp"))

	if err = pool.Retry(func() error {
		_, err := net.Dial("tcp", strings.TrimPrefix(testURL, "http://"))

		time.Sleep(1 * time.Second)

		return err
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	code := m.Run()

	// Defers will not be run when using os.Exit
	if err = pool.Purge(container); err != nil {
		log.Fatalf("Could not purge container: %s", err)
	}

	os.Exit(code)
}
