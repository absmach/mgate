package tls

import (
	"github.com/absmach/mproxy/pkg/tls/verifier"
	"github.com/caarlos0/env/v10"
)

type Config struct {
	CertFile     string `env:"CERT_FILE"                                  envDefault:""`
	KeyFile      string `env:"KEY_FILE"                                   envDefault:""`
	ServerCAFile string `env:"SERVER_CA_FILE"                             envDefault:""`
	ClientCAFile string `env:"CLIENT_CA_FILE"                             envDefault:""`
	Validator    verifier.Validator
}

func NewConfig(opts env.Options) (Config, error) {
	c := Config{}
	var err error
	if err = env.ParseWithOptions(&c, opts); err != nil {
		return Config{}, err
	}
	c.Validator, err = verifier.New(opts)
	if err != nil {
		return Config{}, err
	}
	return c, nil
}
