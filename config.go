// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package mgate

import (
	"crypto/tls"

	mptls "github.com/absmach/mgate/pkg/tls"
	"github.com/caarlos0/env/v11"
)

type Config struct {
	Host           string `env:"HOST"            envDefault:""`
	Port           string `env:"PORT,required"            envDefault:""`
	PathPrefix     string `env:"PATH_PREFIX"              envDefault:""`
	TargetHost     string `env:"TARGET_HOST,required"     envDefault:""`
	TargetPort     string `env:"TARGET_PORT,required"     envDefault:""`
	TargetProtocol string `env:"TARGET_PROTOCOL,required" envDefault:""`
	TargetPath     string `env:"TARGET_PATH"              envDefault:""`
	TLSConfig      *tls.Config
}

func NewConfig(opts env.Options) (Config, error) {
	c := Config{}
	if err := env.ParseWithOptions(&c, opts); err != nil {
		return Config{}, err
	}

	cfg, err := mptls.NewConfig(opts)
	if err != nil {
		return Config{}, err
	}

	c.TLSConfig, err = mptls.Load(&cfg)
	if err != nil {
		return Config{}, err
	}
	return c, nil
}
