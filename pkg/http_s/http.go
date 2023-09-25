package http

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/mainflux/mproxy/pkg/logger"
	"github.com/mainflux/mproxy/pkg/session"
)

func (p Proxy) handleTunneling(w http.ResponseWriter, r *http.Request) {
	dest_conn, err := net.DialTimeout("tcp", p.target, 10*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	client_conn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
	go transfer(dest_conn, client_conn)
	go transfer(client_conn, dest_conn)
}
func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
}
func handleRequest(clientConn net.Conn) {
	defer clientConn.Close()

	clientReader := bufio.NewReader(clientConn)
	request, err := http.ReadRequest(clientReader)
	if err != nil {
		fmt.Println("Failed to read request:", err)
		return
	}
	if err != nil {
		fmt.Println("Failed to read request:", err)
		return
	}

	fmt.Printf("Proxying request to: %s %s\n", request.Method, request.URL)

	serverConn, err := net.Dial("tcp", "localhost:8081")
	if err != nil {
		fmt.Println("Failed to connect to the server:", err)
		return
	}
	defer serverConn.Close()

	err = request.Write(serverConn)
	if err != nil {
		fmt.Println("Failed to write request to the server:", err)
		return
	}

	go func() {
		_, err := io.Copy(clientConn, serverConn)
		if err != nil {
			fmt.Println("Failed to copy from server to client:", err)
		}
	}()

	_, err = io.Copy(serverConn, clientConn)
	if err != nil {
		fmt.Println("Failed to copy from client to server:", err)
	}
}

func (p Proxy) defaultHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p.handleTunneling(w, r)
	}
}

// Proxy represents HTTP Proxy.
type Proxy struct {
	address string
	target  string
	event   session.Handler
	logger  logger.Logger
}

func (p Proxy) Listen() error {
	server := &http.Server{
		Addr:    p.address,
		Handler: p.defaultHandler(),
		// Disable HTTP/2.
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}
	return server.ListenAndServe()
}

func (p Proxy) ListenTLS(cert, key string) error {
	server := &http.Server{
		Addr:    p.address,
		Handler: p.defaultHandler(),
		// Disable HTTP/2.
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}
	return server.ListenAndServeTLS(cert, key)
}
