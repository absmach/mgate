// Package proxy implements both TCP server and WS server which holds connections
package proxy

import (
	"fmt"

	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mproxy/pkg/mqtt"
)

// Proxy is main MQTT proxy struct
type mqttProxy struct {
	host   string
	port   string
	target string
	event  mqtt.Event
	logger logger.Logger
}

// Proxy - struct that holds HTTP proxy info
type wsProxy struct {
	host   string
	port   string
	path   string
	scheme string
	event  mqtt.Event
	logger logger.Logger
}

// New will setup a new Proxy struct after parsing the options
func NewMQTT(host, port, targetHost, targetPort string, event mqtt.Event, logger logger.Logger) *mqttProxy {
	return &mqttProxy{
		host:   host,
		port:   port,
		target: fmt.Sprintf("%s:%s", targetHost, targetPort),
		event:  event,
		logger: logger,
	}
}

// New - creates new HTTP proxy
func NewWS(host, port, path, scheme string, event mqtt.Event, logger logger.Logger) *wsProxy {
	return &wsProxy{
		host:   host,
		port:   port,
		path:   path,
		scheme: scheme,
		event:  event,
		logger: logger,
	}
}
