// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package mproxy

import (
	"reflect"
	"strings"

	mptls "github.com/absmach/mproxy/pkg/tls"
	"github.com/caarlos0/env/v10"
)

type Config struct {
	Address    string `env:"ADDRESS"                        envDefault:""`
	Target     string `env:"TARGET"                         envDefault:""`
	PrefixPath string `env:"PREFIX_PATH"                    envDefault:""`
	TLSConfig  mptls.Config
}

func (c *Config) EnvParse(opts env.Options) error {
	if len(opts.FuncMap) == 0 {
		opts.FuncMap = map[reflect.Type]env.ParserFunc{
			reflect.TypeOf(c.TLSConfig.ClientCertValidationMethods): parseSliceValidateMethod,
			reflect.TypeOf(mptls.OCSP):                              parseValidateMethod,
		}
	} else {
		opts.FuncMap[reflect.TypeOf(c.TLSConfig.ClientCertValidationMethods)] = parseSliceValidateMethod
		opts.FuncMap[reflect.TypeOf(mptls.OCSP)] = parseValidateMethod
	}
	return env.ParseWithOptions(c, opts)
}

func parseSliceValidateMethod(v string) (interface{}, error) {
	var vms []mptls.ValidateMethod
	v = strings.TrimSpace(v)
	vmss := strings.Split(v, ",")
	for _, vm := range vmss {
		v, err := mptls.ParseValidateMethod(vm)
		if err != nil {
			return nil, err
		}
		vms = append(vms, v)
	}
	return vms, nil
}

func parseValidateMethod(v string) (interface{}, error) {
	return mptls.ParseValidateMethod(v)
}
