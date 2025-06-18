// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"net/http"
)

// Interceptor is an interface for mGate intercept hook.
type Interceptor interface {
	// Intercept is called on every request flowing through the Proxy.
	// Requests can be modified before being sent to the broker or the client.
	// If the interceptor returns a non-nil request, the modified request is sent.
	// The error indicates unsuccessful interception and mGate is cancelling the request.
	Intercept(ctx context.Context, r *http.Request) (*http.Request, error)
}
