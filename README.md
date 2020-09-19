# mProxy
mProxy is an MQTT proxy.

It is deployed in front of MQTT broker and can be used for authorization, packet inspection and modification,
logging and debugging and various other purposes.

## Usage
```bash
go get github.com/mainflux/mproxy
cd $(GOPATH)/github.com/mainflux/mproxy
make
./mproxy
```

## Architecture
mProxy starts TCP and WS servers, offering connections to devices. Upon the connection, it establishes a session with a remote MQTT broker.
It then pipes packets from devices to MQTT broker, inspecting or modifying them as they flow through proxy.

Here is the flow in more details:
- Device connects to mProxy's TCP server
- mProxy accepts the inbound (IN) connection and estabishes a new session with remote MQTT broker
(i.e. it dials out to MQTT broker only once it accepted new connection from a device.
This way one device-mProxy connection corresponds to one mProxy-MQTT broker connection.)
- mProxy then spawn 2 goroutines: one that will read incoming packets from device-mProxy socket (INBOUND or UPLINK),
inspect them (calling event handlers) and write them to mProxy-broker socket (forwarding them towards the broker)
and other that will be reading MQTT broker responses from mProxy-broker socket and writing them towards device,
in device-mProxy socket (OUTBOUND or DOWNLINK).

<p align="center"><img src="docs/img/mproxy.png"></p>

mProxy can parse and understand MQTT packages, and upon their detection it actually calls external event handlers.
Event handlers should implement the following interface defined in [pkg/mqtt/events.go](pkg/mqtt/events.go):

```go
// Event is an interface for mProxy hooks
type Event interface {
	// Authorization on client `CONNECT`
	// Each of the params are passed by reference, so that it can be changed
	AuthConnect(client *Client) error

	// Authorization on client `PUBLISH`
	// Topic is passed by reference, so that it can be modified
	AuthPublish(client *Client, topic *string, payload *[]byte) error

	// Authorization on client `SUBSCRIBE`
	// Topics are passed by reference, so that they can be modified
	AuthSubscribe(client *Client, topics *[]string) error

	// After client successfully connected
	Connect(client *Client)

	// After client successfully published
	Publish(client *Client, topic *string, payload *[]byte)

	// After client successfully subscribed
	Subscribe(client *Client, topics *[]string)

	// After client unsubscribed
	Unsubscribe(client *Client, topics *[]string)

	// Disconnect on connection with client lost
	Disconnect(client *Client)
}
```

An example of implementation is given [here](examples/simple/simple.go), alongside with it's [`main()` function](cmd/main.go).

## Deployment
mProxy does not do load balancing - just pure and simple proxying. This is why it should be deployed
right in front of it's corresponding MQTT broker instance: one mProxy for each MQTT broker instance in the MQTT cluster.

Usually this is done by deploying mProxy as a side-car in the same Kubernetes pod alongside with MQTT broker instance (MQTT cluster node).

<p align="center"><img src="docs/img/mproxy-cluster.png"></p>

TLS termination and LB tasks can be offloaded to a standard ingress proxy - for example NginX.

## Configuration

The service is configured using the environment variables presented in the following table. Note that any unset variables will be replaced with their default values.

| Variable                | Description                                    | Default   |
|-------------------------|------------------------------------------------|-----------|
| MPROXY_WS_HOST          | WebSocket inbound (IN) connection host         | 0.0.0.0   |
| MPROXY_WS_PORT          | WebSocket inbound (IN) connection port         | 8080      |
| MPROXY_WS_PATH          | WebSocket inbound (IN) connection path         | /mqtt     |
| MPROXY_WSS_PORT         | WebSocket Secure inbound (IN) connection port  | 8080      |
| MPROXY_WSS_PATH         | WebSocket Secure inbound (IN) connection path  | /mqtt     |
| MPROXY_WS_TARGET_SCHEME | WebSocket Target schema                        | ws        |
| MPROXY_WS_TARGET_HOST   | WebSocket Target host                          | localhost |
| MPROXY_WS_TARGET_PORT   | WebSocket Target port                          | 8888      |
| MPROXY_WS_TARGET_PATH   | WebSocket Target path                          | /mqtt     |
| MPROXY_MQTT_HOST        | MQTT inbound connection host                   | 0.0.0.0   |
| MPROXY_MQTT_PORT        | MQTT inbound connection port                   | 1883      |
| MPROXY_MQTTS_PORT       | MQTTS inbound connection port                  | 8883      |
| MPROXY_MQTT_TARGET_HOST | MQTT broker host                               | 0.0.0.0   |
| MPROXY_MQTT_TARGET_PORT | MQTT broker port                               | 1884      |
| MPROXY_CLIENT_TLS       | Flag that indicates if TLS should be turned on | false     |
| MPROXY_CA_CERTS         | Path to trusted CAs in PEM format              |           |
| MPROXY_SERVER_CERT      | Path to server certificate in pem format       |           |
| MPROXY_SERVER_KEY       | Path to server key in pem format               |           |
| MPROXY_LOG_LEVEL        | Log level                                      | debug     |

## License
[Apache-2.0](LICENSE)
