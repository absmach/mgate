package main

import (
	"log"

	"github.com/inetaf/tcpproxy"
)

func main() {
	var p tcpproxy.Proxy
	p.AddRoute(":1884", tcpproxy.To("localhost:1883")) // fallback
	p.AddRoute(":8083", tcpproxy.To("localhost:8000")) // fallback
	// p.AddSNIRoute(":443", "foo.com", tcpproxy.To("10.0.0.1:4431"))
	// p.AddSNIRoute(":443", "bar.com", tcpproxy.To("10.0.0.2:4432"))
	// p.AddRoute(":443", tcpproxy.To("10.0.0.1:4431")) // fallback
	log.Fatal(p.Run())
}
