// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package mproxy

import (
	"reflect"
	"strings"

	mptls "github.com/absmach/mproxy/pkg/tls"
	"github.com/absmach/mproxy/pkg/tls/verifier"
	"github.com/caarlos0/env/v10"
)

type Config struct {
	Address    string `env:"ADDRESS"                        envDefault:""`
	Target     string `env:"TARGET"                         envDefault:""`
	PrefixPath string `env:"PREFIX_PATH"                    envDefault:""`
	TLSConfig  mptls.Config
}

func (c *Config) EnvParse(opts env.Options) error {
	switch {
	case len(opts.FuncMap) == 0:
		opts.FuncMap = map[reflect.Type]env.ParserFunc{
			reflect.TypeOf(c.TLSConfig.Verifier.ValidationMethods): parseSliceValidateMethod,
			reflect.TypeOf(verifier.OCSP):                          parseValidateMethod,
		}
	default:
		opts.FuncMap[reflect.TypeOf(c.TLSConfig.Verifier.ValidationMethods)] = parseSliceValidateMethod
		opts.FuncMap[reflect.TypeOf(verifier.OCSP)] = parseValidateMethod
	}
	return env.ParseWithOptions(c, opts)
}

func parseSliceValidateMethod(v string) (interface{}, error) {
	var vms []verifier.ValidateMethod
	v = strings.TrimSpace(v)
	vmss := strings.Split(v, ",")
	for _, vm := range vmss {
		v, err := verifier.ParseValidateMethod(vm)
		if err != nil {
			return nil, err
		}
		vms = append(vms, v)
	}
	return vms, nil
}

func parseValidateMethod(v string) (interface{}, error) {
	return verifier.ParseValidateMethod(v)
}
