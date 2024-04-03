// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package mproxy

import (
	"crypto/tls"

	mptls "github.com/absmach/mproxy/pkg/tls"
	"github.com/caarlos0/env/v10"
)

type Config struct {
	Address    string `env:"ADDRESS"                        envDefault:""`
	PrefixPath string `env:"PREFIX_PATH"                    envDefault:""`
	Target     string `env:"TARGET"                         envDefault:""`
	TLSConfig  *tls.Config
}

func NewConfig(opts env.Options) (Config, error) {
	c := Config{}
	if err := env.ParseWithOptions(&c, opts); err != nil {
		return Config{}, err
	}
	vfs, err := mptls.NewVerification(opts)
	if err != nil {
		return Config{}, err
	}
	mptlsConfig, err := mptls.NewConfig(opts, vfs)
	if err != nil {
		return Config{}, err
	}

	c.TLSConfig, err = mptls.Load(&mptlsConfig)
	if err != nil {
		return Config{}, err
	}
	return c, nil
}
