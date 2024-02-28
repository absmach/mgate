package mproxy

type Config struct {
	Address      string
	Target       string
	CertFile     string
	KeyFile      string
	ServerCAFile string
	ClientCAFile string
}
