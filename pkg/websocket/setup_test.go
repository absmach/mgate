// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package websocket

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
)

var testURL string

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	container, err := pool.Run("ksdn117/web-socket-test", "latest", nil)
	if err != nil {
		log.Fatalf("Could not start container: %s", err)
	}

	testURL = fmt.Sprintf("ws://localhost:%s", container.GetPort("8010/tcp"))

	if err = pool.Retry(func() error {
		_, err := net.Dial("tcp", strings.TrimPrefix(testURL, "ws://"))

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
