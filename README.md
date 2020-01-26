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
Event handlers should implement the following interface defined in [pkg/events/events.go](pkg/events/events.go):

```go
// Event is an interface for mProxy hooks
type Event interface {
	// Athorization on client `CONNECT`
	// Each of the params are passed by reference, so that it can be changed
	AuthRegister(clientID, username *string, password *[]byte) error

	// Athorization on client `PUBLISH`
	// Topic is passed by reference, so that it can be modified
	AuthPublish(clientID string, topic *string, payload *[]byte) error

	// Athorization on client `SUBSCRIBE`
	// Topics are passed by reference, so that they can be modified
	AuthSubscribe(clientID string, topics *[]string) error

	// After client sucesfully connected
	Register(clientID string)

	// After client sucesfully published
	Publish(clientID, topic string, payload []byte)

	// After client sucesfully subscribed
	Subscribe(clientID string, topics []string)

	// After client unsubscribed
	Unubscribe(clientID string, topics []string)
}
```

An example of implementation is given [here](examples/simple/simple.go), alongside with it's [`main()` function](cmd/main.go).

## Deployment
mProxy does not do load balancing - just pure and simple proxying. This is why it should be deployed
right in front of it's corresponding MQTT broker instance: one mProxy for each MQTT broker instance in the MQTT cluster.

Usually this is done by deploying mProxy as a side-car in the same Kubernetes pod alongside with MQTT broker instance (MQTT cluster node).

<p align="center"><img src="docs/img/mproxy-cluster.png"></p>

TLS termination and LB tasks can be offloaded to a standard ingress proxy - for example NginX.

## License
[Apache-2.0](LICENSE)
