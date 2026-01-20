<div align="center">

# mGate

### Lightweight multi-protocol IoT proxy

### Pluggable Auth - Observability - Packet Manipulation

[![Go Report Card](https://goreportcard.com/badge/github.com/absmach/mgate)](https://goreportcard.com/report/github.com/absmach/mgate)
[![Release](https://img.shields.io/github/v/release/absmach/mgate?display_name=tag&sort=semver)](https://github.com/absmach/mgate/releases)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

Made with love by [Abstract Machines](https://www.absmach.eu)

</div>

mGate is a lightweight, scalable, and customizable IoT API gateway designed to support seamless communication across multiple protocols. It enables real-time packet manipulation, features pluggable authentication mechanisms, and offers observability for monitoring and troubleshooting. Built for flexibility, mGate can be deployed as a sidecar or standalone service and can also function as a library for easy integration into applications.

The extensible nature of mGate allows developers to customize it to fit various IoT ecosystems, ensuring optimal performance and security.

## Key Features

Some of the key features of mGate include multi-protocol support, real-time packet manipulation, pluggable authentication, observability, and scalability, all while being lightweight, customizable, and easily deployable as a sidecar or standalone service.

<p align="center"><img src="docs/img/mgate-features.png"></p>

#### Multi-Protocol Support

mGate is built to interface with a wide range of IoT protocols, including:

- MQTT
- CoAP
- HTTP
- WebSocket
- Easily extendable to support additional protocols.

### On-the-Fly Packet Manipulation

Allows real-time packet transformation and processing.
Custom logic or packet interceptors can be injected for modifying incoming and outgoing messages.

### Authentication and Authorization

Pluggable authentication system supporting different providers like OAuth, JWT, API Keys, and more.
Access Control for fine-grained resource authorization.
Easily replaceable auth modules for integration with custom or enterprise identity systems.

### Observability

Provides real-time metrics for monitoring system health and performance.
Offers logging and tracing to facilitate troubleshooting and optimization and options to easily integrate with Prometheus, Grafana, and OpenTelemetry for detailed tracing and visualization.

### Scalable Architecture

mGate is designed to scale horizontally, ensuring it can handle high-throughput environments.

### Pluggable and Extensible

Core components are modular, making it easy to plug in custom modules or replace existing ones.
Extendable to add new IoT protocols, middleware, and features as needed.

### Customizable

Highly configurable, allowing adjustment of protocol-specific behaviors, observability, and performance optimizations.
Minimal configuration is required for default deployment but supports deep customization.

### Lightweight

Built with Go programming language, it is optimized for low resource usage, making it suitable for both high-performance data centers and resource-constrained IoT edge devices.

### Deployment Flexibility

Can be deployed as a sidecar to enhance existing microservices or as a standalone service for direct IoT device interaction.
Available as a library for integration into existing applications.

## Quickstart

- Build and run the sample proxy:

```bash
make
./build/mgate
```

- Alternatively, run directly for development:

```bash
go run cmd/main.go
```

## Try It Now

Spin up local dependencies and run protocol proxies:

```bash
# Start MQTT broker (Mosquitto) with WS support
examples/server/mosquitto/server.sh

# Start HTTP echo server
go run examples/server/http-echo/main.go &

# Start OCSP/CRL mock responder
go run examples/ocsp-crl-responder/main.go &

# Start mGate example servers
go run cmd/main.go
```

Client examples:

- MQTT (no TLS): `examples/client/mqtt/without_tls.sh`
- MQTT (TLS): `examples/client/mqtt/with_tls.sh`
- MQTT (mTLS): `examples/client/mqtt/with_mtls.sh`
- MQTT over WebSocket (no TLS): `go run examples/client/websocket/without_tls/main.go`
- MQTT over WebSocket (TLS): `go run examples/client/websocket/with_tls/main.go`
- MQTT over WebSocket (mTLS): `go run examples/client/websocket/with_mtls/main.go`
- HTTP (no TLS): `examples/client/http/without_tls.sh`
- HTTP (TLS): `examples/client/http/with_tls.sh`
- HTTP (mTLS): `examples/client/http/with_mtls.sh`
- CoAP (no DTLS): `examples/client/coap/without_dtls.sh`
- CoAP (DTLS): `examples/client/coap/with_dtls.sh`

## Protocol Matrix & Examples

| Protocol | Mode        | Port   | Path                 |
|:---------|:------------|-------:|:---------------------|
| `MQTT`   | no TLS      | `1884` | -                    |
| `MQTT`   | TLS         | `8883` | -                    |
| `MQTT`   | mTLS        | `8884` | -                    |
| `MQTT/WS`| no TLS      | `8083` | /mgate-ws            |
| `MQTT/WS`| TLS         | `8084` | /mgate-ws            |
| `MQTT/WS`| mTLS        | `8085` | /mgate-ws            |
| `HTTP`   | no TLS      | `8086` | /mgate-http/messages |
| `HTTP`   | TLS         | `8087` | /mgate-http/messages |
| `HTTP`   | mTLS        | `8088` | /mgate-http/messages |
| `CoAP`   | no DTLS     | `5682` | -                    |
| `CoAP`   | DTLS        | `5684` | -                    |

Note: HTTP uses `PATH_PREFIX + TARGET_PATH` (default `/mgate-http` + `/messages`), while MQTT over WebSocket uses `PATH_PREFIX` (default `/mgate-ws`).

Examples:

- Servers: `examples/server/mosquitto`, `examples/server/http-echo`, `examples/ocsp-crl-responder`
- Clients: `examples/client/mqtt`, `examples/client/websocket`, `examples/client/http`, `examples/client/coap`

## Usage

```bash
git clone https://github.com/absmach/mgate.git
cd mgate
make
./build/mgate
```

## Architecture

mGate starts protocol servers, offering connections to devices. Upon the connection, it establishes a session with a remote protocol server. It then pipes packets from devices to the protocol server, inspecting or modifying them as they flow through the proxy.

Here is the flow in more detail:

- The device connects to mGate's server
- mGate accepts the inbound (IN) connection and establishes a new session with the remote server (e.g. it dials out to the MQTT broker only once it accepts a new connection from a device. This way one device-mGate connection corresponds to one mGate-MQTT broker connection.)
- mGate then spawns 2 goroutines: one that will read incoming packets from the device-mGate socket (INBOUND or UPLINK), inspect them (calling event handlers) and write them to mGate-server socket (forwarding them towards the server) and other that will be reading server responses from mGate-server socket and writing them towards device, in device-mGate socket (OUTBOUND or DOWNLINK).

<p align="center"><img src="docs/img/mgate.png"></p>

mGate can parse and understand protocol packets, and upon their detection, it calls external event handlers. Event handlers should implement the following interface defined in [pkg/mqtt/events.go](pkg/mqtt/events.go):

```go
// Handler is an interface for mGate hooks
type Handler interface {
    // Authorization on client `CONNECT`
    // Each of the params are passed by reference, so that it can be changed
    AuthConnect(ctx context.Context) error

    // Authorization on client `PUBLISH`
    // Topic is passed by reference, so that it can be modified
    AuthPublish(ctx context.Context, topic *string, payload *[]byte) error

    // Authorization on client `SUBSCRIBE`
    // Topics are passed by reference, so that they can be modified
    AuthSubscribe(ctx context.Context, topics *[]string) error

    // After client successfully connected
    Connect(ctx context.Context)

    // After client successfully published
    Publish(ctx context.Context, topic *string, payload *[]byte)

    // After client successfully subscribed
    Subscribe(ctx context.Context, topics *[]string)

    // After client unsubscribed
    Unsubscribe(ctx context.Context, topics *[]string)

    // Disconnect on connection with client lost
    Disconnect(ctx context.Context)
}
```

The Handler interface is inspired by MQTT protocol control packets; if the underlying protocol does not support some of these actions, the implementation can simply omit them. An example of implementation is given in [examples/simple/simple.go](examples/simple/simple.go), alongside with its [`main()` function](cmd/main.go).

## Deployment

To explain the deployment process, an MQTT broker will be used as an example, given that MQTT is one of the most widely used and feature-rich protocols. mGate does not do load balancing - just pure and simple proxying with TLS termination. This is why it should be deployed right in front of its corresponding MQTT broker instance: one mGate for each MQTT broker instance in the MQTT cluster.

Usually, this is done by deploying mGate as a side-car in the same Kubernetes pod alongside with MQTT broker instance (MQTT cluster node).

<p align="center"><img src="docs/img/mgate-cluster.png"></p>

LB tasks can be offloaded to a standard ingress proxy - for example, NginX.

## Example Setup & Testing of mGate

### Requirements

- Golang
- Mosquitto MQTT Server
- Mosquitto Publisher and Subscriber Client
- coap-client or Magistrala coap-cli

### Example Setup of mGate

mGate is used to proxy requests to a backend server. For the example setup, we will use Mosquitto server as the backend for MQTT, and MQTT over WebSocket and an HTTP echo server for HTTP.

1. Start the Mosquitto MQTT Server with the following command. This bash script will initiate the Mosquitto MQTT server with WebSocket support. The Mosquitto Server will listen for MQTT connections on port 1883 and MQTT over WebSocket connections on port 8000.

   ```bash
   examples/server/mosquitto/server.sh
   ```

2. Start the HTTP Echo Server:

   ```bash
   go run examples/server/http-echo/main.go
   ```

3. Start the OCSP/CRL Mock responder:

   ```bash
   go run examples/ocsp-crl-responder/main.go
   ```

4. Start the example mGate servers for various protocols:

   ```bash
   go run cmd/main.go
   ```

   The `cmd/main.go` Go program initializes mGate servers for the following protocols:

   - mGate server for `MQTT` protocol `without TLS` on port `1884`
   - mGate server for `MQTT` protocol `with TLS` on port `8883`
   - mGate server for `MQTT` protocol `with mTLS` on port `8884`
   - mGate server for `MQTT over WebSocket without TLS` on port `8083` with prefix path `/mgate-ws`
   - mGate server for `MQTT over WebSocket with TLS` on port `8084` with prefix path `/mgate-ws`
   - mGate server for `MQTT over WebSocket with mTLS` on port `8085` with prefix path `/mgate-ws`
   - mGate server for `HTTP protocol without TLS` on port `8086` with path `/mgate-http/messages`
   - mGate server for `HTTP protocol with TLS` on port `8087` with path `/mgate-http/messages`
   - mGate server for `HTTP protocol with mTLS` on port `8088` with path `/mgate-http/messages`
   - mGate server for `CoAP protocol without DTLS` on port `5682`
   - mGate server for `CoAP protocol with DTLS` on port `5684`

### Example testing of mGate

#### Test mGate server for MQTT protocols

Bash scripts available in `examples/client/mqtt` directory help to test the mGate servers running for MQTT protocols.

- Script to test mGate server running at port 1884 for MQTT without TLS

  ```bash
  examples/client/mqtt/without_tls.sh
  ```

- Script to test mGate server running at port 8883 for MQTT with TLS

  ```bash
  examples/client/mqtt/with_tls.sh
  ```

- Script to test mGate server running at port 8884 for MQTT with mTLS

  ```bash
  examples/client/mqtt/with_mtls.sh
  ```

#### Test mGate server for MQTT over WebSocket protocols

Go programs available in `examples/client/websocket/*/main.go` directory help to test the mGate servers running for MQTT over WebSocket protocols.

- Go program to test mGate server running at port 8083 for MQTT over WebSocket without TLS

  ```bash
  go run examples/client/websocket/without_tls/main.go
  ```

- Go program to test mGate server running at port 8084 for MQTT over WebSocket with TLS

  ```bash
  go run examples/client/websocket/with_tls/main.go
  ```

- Go program to test mGate server running at port 8085 for MQTT over WebSocket with mTLS

  ```bash
  go run examples/client/websocket/with_mtls/main.go
  ```

#### Test mGate server for HTTP protocols

Bash scripts available in `examples/client/http` directory help to test the mGate servers running for HTTP protocols.

- Script to test mGate server running at port 8086 for HTTP without TLS

  ```bash
  examples/client/http/without_tls.sh
  ```

- Script to test mGate server running at port 8087 for HTTP with TLS

  ```bash
  examples/client/http/with_tls.sh
  ```

- Script to test mGate server running at port 8088 for HTTP with mTLS

  ```bash
  examples/client/http/with_mtls.sh
  ```

### Test mGate server for CoAP protocols

Bash scripts available in `examples/client/coap` directory help to test the mGate servers running for CoAP protocols. You will require to have either the [coap-client](https://libcoap.net/doc/reference/4.3.1/man_coap-client.html) or the [Magistrala coap-cli](https://github.com/absmach/coap-cli).
The script can be used alongside the simple go-coap server provided at `examples/server/coap`.

- Script to test mGate server running at 5682 for CoAP without DTLS

  ```bash
  examples/client/coap/without_dtls.sh
  ```

- Script to test mGate server running at 5684 for CoAP with DTLS

  ```bash
  examples/client/coap/with_dtls.sh
  ```

## Configuration

mGate is configured via environment variables with per-protocol prefixes. Required keys are `PORT`, `TARGET_PROTOCOL`, `TARGET_HOST`, and `TARGET_PORT`. `HOST`, `PATH_PREFIX`, and `TARGET_PATH` are optional. The table below mirrors the local `.env` values (empty means unset).

| Variable                                          | Description                                                     | Example (.env)               |
| ------------------------------------------------- | --------------------------------------------------------------- | -----------------------------|
| MGATE_MQTT_WITHOUT_TLS_HOST                       | MQTT without TLS inbound bind host                              | localhost                    |
| MGATE_MQTT_WITHOUT_TLS_PORT                       | MQTT without TLS inbound port                                   | 1884                         |
| MGATE_MQTT_WITHOUT_TLS_TARGET_PROTOCOL            | MQTT without TLS outbound protocol                              | mqtt                         |
| MGATE_MQTT_WITHOUT_TLS_TARGET_HOST                | MQTT without TLS outbound host                                  | localhost                    |
| MGATE_MQTT_WITHOUT_TLS_TARGET_PORT                | MQTT without TLS outbound port                                  | 1883                         |
| MGATE_MQTT_WITH_TLS_HOST                          | MQTT with TLS inbound bind host                                 | localhost                    |
| MGATE_MQTT_WITH_TLS_PORT                          | MQTT with TLS inbound port                                      | 8883                         |
| MGATE_MQTT_WITH_TLS_TARGET_PROTOCOL               | MQTT with TLS outbound protocol                                 | mqtt                         |
| MGATE_MQTT_WITH_TLS_TARGET_HOST                   | MQTT with TLS outbound host                                     | localhost                    |
| MGATE_MQTT_WITH_TLS_TARGET_PORT                   | MQTT with TLS outbound port                                     | 1883                         |
| MGATE_MQTT_WITH_TLS_CERT_FILE                     | MQTT with TLS certificate file path                             | ssl/certs/server.crt         |
| MGATE_MQTT_WITH_TLS_KEY_FILE                      | MQTT with TLS key file path                                     | ssl/certs/server.key         |
| MGATE_MQTT_WITH_TLS_SERVER_CA_FILE                | MQTT with TLS server CA file path                               | ssl/certs/ca.crt             |
| MGATE_MQTT_WITH_MTLS_HOST                         | MQTT with mTLS inbound bind host                                | localhost                    |
| MGATE_MQTT_WITH_MTLS_PORT                         | MQTT with mTLS inbound port                                     | 8884                         |
| MGATE_MQTT_WITH_MTLS_TARGET_PROTOCOL              | MQTT with mTLS outbound protocol                                | mqtt                         |
| MGATE_MQTT_WITH_MTLS_TARGET_HOST                  | MQTT with mTLS outbound host                                    | localhost                    |
| MGATE_MQTT_WITH_MTLS_TARGET_PORT                  | MQTT with mTLS outbound port                                    | 1883                         |
| MGATE_MQTT_WITH_MTLS_CERT_FILE                    | MQTT with mTLS certificate file path                            | ssl/certs/server.crt         |
| MGATE_MQTT_WITH_MTLS_KEY_FILE                     | MQTT with mTLS key file path                                    | ssl/certs/server.key         |
| MGATE_MQTT_WITH_MTLS_SERVER_CA_FILE               | MQTT with mTLS server CA file path                              | ssl/certs/ca.crt             |
| MGATE_MQTT_WITH_MTLS_CLIENT_CA_FILE               | MQTT with mTLS client CA file path                              | ssl/certs/ca.crt             |
| MGATE_MQTT_WITH_MTLS_CERT_VERIFICATION_METHODS    | MQTT with mTLS certificate verification methods                 | ocsp                         |
| MGATE_MQTT_WITH_MTLS_OCSP_RESPONDER_URL           | MQTT with mTLS OCSP responder URL                               | <http://localhost:8080/ocsp> |
| MGATE_MQTT_WS_WITHOUT_TLS_HOST                    | MQTT over WebSocket without TLS inbound bind host               | localhost                    |
| MGATE_MQTT_WS_WITHOUT_TLS_PORT                    | MQTT over WebSocket without TLS inbound port                    | 8083                         |
| MGATE_MQTT_WS_WITHOUT_TLS_PATH_PREFIX             | MQTT over WebSocket without TLS inbound path prefix             | /mgate-ws                    |
| MGATE_MQTT_WS_WITHOUT_TLS_TARGET_PROTOCOL         | MQTT over WebSocket without TLS outbound protocol               | ws                           |
| MGATE_MQTT_WS_WITHOUT_TLS_TARGET_HOST             | MQTT over WebSocket without TLS outbound host                   | localhost                    |
| MGATE_MQTT_WS_WITHOUT_TLS_TARGET_PORT             | MQTT over WebSocket without TLS outbound port                   | 8000                         |
| MGATE_MQTT_WS_WITHOUT_TLS_TARGET_PATH             | MQTT over WebSocket without TLS outbound path                   | (empty)                      |
| MGATE_MQTT_WS_WITH_TLS_HOST                       | MQTT over WebSocket with TLS inbound bind host                  | localhost                    |
| MGATE_MQTT_WS_WITH_TLS_PORT                       | MQTT over WebSocket with TLS inbound port                       | 8084                         |
| MGATE_MQTT_WS_WITH_TLS_PATH_PREFIX                | MQTT over WebSocket with TLS inbound path prefix                | /mgate-ws                    |
| MGATE_MQTT_WS_WITH_TLS_TARGET_PROTOCOL            | MQTT over WebSocket with TLS outbound protocol                  | ws                           |
| MGATE_MQTT_WS_WITH_TLS_TARGET_HOST                | MQTT over WebSocket with TLS outbound host                      | localhost                    |
| MGATE_MQTT_WS_WITH_TLS_TARGET_PORT                | MQTT over WebSocket with TLS outbound port                      | 8000                         |
| MGATE_MQTT_WS_WITH_TLS_TARGET_PATH                | MQTT over WebSocket with TLS outbound path                      | (empty)                      |
| MGATE_MQTT_WS_WITH_TLS_CERT_FILE                  | MQTT over WebSocket with TLS certificate file path              | ssl/certs/server.crt         |
| MGATE_MQTT_WS_WITH_TLS_KEY_FILE                   | MQTT over WebSocket with TLS key file path                      | ssl/certs/server.key         |
| MGATE_MQTT_WS_WITH_TLS_SERVER_CA_FILE             | MQTT over WebSocket with TLS server CA file path                | ssl/certs/ca.crt             |
| MGATE_MQTT_WS_WITH_MTLS_HOST                      | MQTT over WebSocket with mTLS inbound bind host                 | localhost                    |
| MGATE_MQTT_WS_WITH_MTLS_PORT                      | MQTT over WebSocket with mTLS inbound port                      | 8085                         |
| MGATE_MQTT_WS_WITH_MTLS_PATH_PREFIX               | MQTT over WebSocket with mTLS inbound path prefix               | /mgate-ws                    |
| MGATE_MQTT_WS_WITH_MTLS_TARGET_PROTOCOL           | MQTT over WebSocket with mTLS outbound protocol                 | ws                           |
| MGATE_MQTT_WS_WITH_MTLS_TARGET_HOST               | MQTT over WebSocket with mTLS outbound host                     | localhost                    |
| MGATE_MQTT_WS_WITH_MTLS_TARGET_PORT               | MQTT over WebSocket with mTLS outbound port                     | 8000                         |
| MGATE_MQTT_WS_WITH_MTLS_TARGET_PATH               | MQTT over WebSocket with mTLS outbound path                     | (empty)                      |
| MGATE_MQTT_WS_WITH_MTLS_CERT_FILE                 | MQTT over WebSocket with mTLS certificate file path             | ssl/certs/server.crt         |
| MGATE_MQTT_WS_WITH_MTLS_KEY_FILE                  | MQTT over WebSocket with mTLS key file path                     | ssl/certs/server.key         |
| MGATE_MQTT_WS_WITH_MTLS_SERVER_CA_FILE            | MQTT over WebSocket with mTLS server CA file path               | ssl/certs/ca.crt             |
| MGATE_MQTT_WS_WITH_MTLS_CLIENT_CA_FILE            | MQTT over WebSocket with mTLS client CA file path               | ssl/certs/ca.crt             |
| MGATE_MQTT_WS_WITH_MTLS_CERT_VERIFICATION_METHODS | MQTT over WebSocket with mTLS certificate verification methods  | ocsp                         |
| MGATE_MQTT_WS_WITH_MTLS_OCSP_RESPONDER_URL        | MQTT over WebSocket with mTLS OCSP responder URL                | <http://localhost:8080/ocsp> |
| MGATE_HTTP_WITHOUT_TLS_HOST                       | HTTP without TLS inbound bind host                              | (empty)                      |
| MGATE_HTTP_WITHOUT_TLS_PORT                       | HTTP without TLS inbound port                                   | 8086                         |
| MGATE_HTTP_WITHOUT_TLS_PATH_PREFIX                | HTTP without TLS inbound path prefix                            | /mgate-http                  |
| MGATE_HTTP_WITHOUT_TLS_TARGET_PROTOCOL            | HTTP without TLS outbound protocol                              | http                         |
| MGATE_HTTP_WITHOUT_TLS_TARGET_HOST                | HTTP without TLS outbound host                                  | localhost                    |
| MGATE_HTTP_WITHOUT_TLS_TARGET_PORT                | HTTP without TLS outbound port                                  | 8888                         |
| MGATE_HTTP_WITHOUT_TLS_TARGET_PATH                | HTTP without TLS outbound path                                  | /messages                    |
| MGATE_HTTP_WITH_TLS_HOST                          | HTTP with TLS inbound bind host                                 | (empty)                      |
| MGATE_HTTP_WITH_TLS_PORT                          | HTTP with TLS inbound port                                      | 8087                         |
| MGATE_HTTP_WITH_TLS_PATH_PREFIX                   | HTTP with TLS inbound path prefix                               | /mgate-http                  |
| MGATE_HTTP_WITH_TLS_TARGET_PROTOCOL               | HTTP with TLS outbound protocol                                 | http                         |
| MGATE_HTTP_WITH_TLS_TARGET_HOST                   | HTTP with TLS outbound host                                     | localhost                    |
| MGATE_HTTP_WITH_TLS_TARGET_PORT                   | HTTP with TLS outbound port                                     | 8888                         |
| MGATE_HTTP_WITH_TLS_TARGET_PATH                   | HTTP with TLS outbound path                                     | /messages                    |
| MGATE_HTTP_WITH_TLS_CERT_FILE                     | HTTP with TLS certificate file path                             | ssl/certs/server.crt         |
| MGATE_HTTP_WITH_TLS_KEY_FILE                      | HTTP with TLS key file path                                     | ssl/certs/server.key         |
| MGATE_HTTP_WITH_TLS_SERVER_CA_FILE                | HTTP with TLS server CA file path                               | ssl/certs/ca.crt             |
| MGATE_HTTP_WITH_MTLS_HOST                         | HTTP with mTLS inbound bind host                                | (empty)                      |
| MGATE_HTTP_WITH_MTLS_PORT                         | HTTP with mTLS inbound port                                     | 8088                         |
| MGATE_HTTP_WITH_MTLS_PATH_PREFIX                  | HTTP with mTLS inbound path prefix                              | /mgate-http                  |
| MGATE_HTTP_WITH_MTLS_TARGET_PROTOCOL              | HTTP with mTLS outbound protocol                                | http                         |
| MGATE_HTTP_WITH_MTLS_TARGET_HOST                  | HTTP with mTLS outbound host                                    | localhost                    |
| MGATE_HTTP_WITH_MTLS_TARGET_PORT                  | HTTP with mTLS outbound port                                    | 8888                         |
| MGATE_HTTP_WITH_MTLS_TARGET_PATH                  | HTTP with mTLS outbound path                                    | /messages                    |
| MGATE_HTTP_WITH_MTLS_CERT_FILE                    | HTTP with mTLS certificate file path                            | ssl/certs/server.crt         |
| MGATE_HTTP_WITH_MTLS_KEY_FILE                     | HTTP with mTLS key file path                                    | ssl/certs/server.key         |
| MGATE_HTTP_WITH_MTLS_SERVER_CA_FILE               | HTTP with mTLS server CA file path                              | ssl/certs/ca.crt             |
| MGATE_HTTP_WITH_MTLS_CLIENT_CA_FILE               | HTTP with mTLS client CA file path                              | ssl/certs/ca.crt             |
| MGATE_HTTP_WITH_MTLS_CERT_VERIFICATION_METHODS    | HTTP with mTLS certificate verification methods                 | ocsp                         |
| MGATE_HTTP_WITH_MTLS_OCSP_RESPONDER_URL           | HTTP with mTLS OCSP responder URL                               | <http://localhost:8080/ocsp> |
| MGATE_COAP_WITHOUT_DTLS_HOST                      | CoAP without DTLS inbound bind host                             | (empty)                      |
| MGATE_COAP_WITHOUT_DTLS_PORT                      | CoAP without DTLS inbound port                                  | 5682                         |
| MGATE_COAP_WITHOUT_DTLS_TARGET_PROTOCOL           | CoAP without DTLS outbound protocol                             | (missing)                    |
| MGATE_COAP_WITHOUT_DTLS_TARGET_HOST               | CoAP without DTLS outbound host                                 | localhost                    |
| MGATE_COAP_WITHOUT_DTLS_TARGET_PORT               | CoAP without DTLS outbound port                                 | 5683                         |
| MGATE_COAP_WITH_DTLS_HOST                         | CoAP with DTLS inbound bind host                                | localhost                    |
| MGATE_COAP_WITH_DTLS_PORT                         | CoAP with DTLS inbound port                                     | 5684                         |
| MGATE_COAP_WITH_DTLS_TARGET_PROTOCOL              | CoAP with DTLS outbound protocol                                | (missing)                    |
| MGATE_COAP_WITH_DTLS_TARGET_HOST                  | CoAP with DTLS outbound host                                    | localhost                    |
| MGATE_COAP_WITH_DTLS_TARGET_PORT                  | CoAP with DTLS outbound port                                    | 5683                         |
| MGATE_COAP_WITH_DTLS_CERT_FILE                    | CoAP with DTLS certificate file path                            | ssl/certs/server.crt         |
| MGATE_COAP_WITH_DTLS_KEY_FILE                     | CoAP with DTLS key file path                                    | ssl/certs/server.key         |
| MGATE_COAP_WITH_DTLS_SERVER_CA_FILE               | CoAP with DTLS server CA file path                              | ssl/certs/ca.crt             |
| MGATE_COAP_WITH_DTLS_CLIENT_CA_FILE               | CoAP with DTLS client CA file path                              | ssl/certs/ca.crt             |

## mGate Configuration Environment Variables

### Server Configuration Keys (used under a prefix)

- `HOST`: Inbound bind host (empty binds all interfaces).
- `PORT`: Inbound port (required; example values listed above).
- `TARGET_HOST`: Backend host (required; example `localhost`).
- `TARGET_PORT`: Backend port (required; example values listed above).
- `TARGET_PROTOCOL`: Backend scheme (`mqtt`, `ws`, `http`, `coap`); required even if unused by a protocol.
- `TARGET_PATH`: Backend path suffix (used for WebSocket/HTTP).
- `PATH_PREFIX`: Optional inbound path prefix (used for HTTP and MQTT over WebSocket).

### TLS Configuration Keys

- `CERT_FILE`: TLS certificate file.
- `KEY_FILE`: TLS private key file.
- `SERVER_CA_FILE`: Server CA bundle.
- `CLIENT_CA_FILE`: Client CA bundle (for mTLS only).
- `CERT_VERIFICATION_METHODS`: Comma-separated `ocsp` and/or `crl`.
  - OCSP: If AIA lacks a responder URL or you prefer a custom endpoint, set `OCSP_RESPONDER_URL`.
  - CRL: If distribution points are missing or unavailable, use overrides and/or offline files.

#### OCSP Keys

- `OCSP_DEPTH`: Verification depth (0 = no limit; verify all certs).
- `OCSP_RESPONDER_URL`: Override OCSP responder when AIA is missing or overridden.

#### CRL Keys

- `CRL_DEPTH`: Verification depth (default `1`, leaf only).
- `CRL_DISTRIBUTION_POINTS`: CRL URL override.
- `CRL_DISTRIBUTION_POINTS_ISSUER_CERT_FILE`: Issuer cert to verify CRL signature.
- `OFFLINE_CRL_FILE`: Offline CRL file.
- `OFFLINE_CRL_ISSUER_CERT_FILE`: Issuer cert to verify offline CRL.

## Adding Prefix to Environmental Variables

mGate relies on the [caarlos0/env](https://github.com/caarlos0/env) package to load environmental variables into its [configuration](config.go#L15).
You can control how these variables are loaded by passing `env.Options` to `mgate.NewConfig`.

To add a prefix to environmental variables, use `env.Options{Prefix: "MGATE_"}` from the [caarlos0/env](https://github.com/caarlos0/env) package. For example:

```go
package main
import (
  "github.com/caarlos0/env/v11"
  "github.com/absmach/mgate"
)

cfg, err := mgate.NewConfig(env.Options{Prefix: "MGATE_"})
if err != nil {
    panic(err)
}
fmt.Printf("%+v\n", cfg)
```

In the above snippet, `mgate.NewConfig` expects all environmental variables with the prefix `MGATE_`.
For instance:

- MGATE_HOST
- MGATE_PORT
- MGATE_PATH_PREFIX
- MGATE_TARGET_PROTOCOL
- MGATE_TARGET_HOST
- MGATE_TARGET_PORT
- MGATE_TARGET_PATH
- MGATE_CERT_FILE
- MGATE_KEY_FILE
- MGATE_SERVER_CA_FILE
- MGATE_CLIENT_CA_FILE
- MGATE_CERT_VERIFICATION_METHODS
- MGATE_OCSP_DEPTH
- MGATE_OCSP_RESPONDER_URL
- MGATE_CRL_DEPTH
- MGATE_CRL_DISTRIBUTION_POINTS
- MGATE_CRL_DISTRIBUTION_POINTS_ISSUER_CERT_FILE
- MGATE_OFFLINE_CRL_FILE
- MGATE_OFFLINE_CRL_ISSUER_CERT_FILE

## Troubleshooting & FAQ

- Cert chain errors: Ensure `SERVER_CA_FILE` (and `CLIENT_CA_FILE` for mTLS) include required intermediates.
- OCSP responder missing/unreachable: Set `OCSP_RESPONDER_URL` to a reachable endpoint; verify firewall rules.
- CRL retrieval failures: Use `CRL_DISTRIBUTION_POINTS` overrides or provide `OFFLINE_CRL_FILE` and `OFFLINE_CRL_ISSUER_CERT_FILE`.
- WebSocket path mismatches: Confirm client path matches server prefix (default `/mgate-ws`) and backend target path.
- HTTP prefix issues: Verify inbound path `/mgate-http/messages` (PATH_PREFIX + TARGET_PATH) and backend routing.
- Target connectivity: Confirm the backend (MQTT/HTTP/WS/CoAP) is listening and ports are open.

## License

[Apache-2.0](LICENSE)
