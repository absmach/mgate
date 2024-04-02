package validation

import (
	"errors"
	"reflect"
	"strings"

	"github.com/absmach/mproxy/pkg/tls/verifier"
	"github.com/absmach/mproxy/pkg/tls/verifier/crl"
	"github.com/absmach/mproxy/pkg/tls/verifier/ocsp"
	"github.com/caarlos0/env/v10"
)

var ErrInvalidClientValidation = errors.New("invalid client validation method")

type validation int

const (
	OCSP validation = iota + 1
	CRL
)

func parseValidation(v string) (validation, error) {
	v = strings.ToUpper(strings.TrimSpace(v))
	switch v {
	case "OCSP":
		return OCSP, nil
	case "CRL":
		return CRL, nil
	default:
		return 0, ErrInvalidClientValidation
	}
}

type config struct {
	Validations []validation `env:"CLIENT_CERT_VALIDATION_METHODS"             envDefault:""`
}

func NewVerifiers(opts env.Options) ([]verifier.Verifier, error) {
	var c config
	if opts.FuncMap == nil {
		opts.FuncMap = make(map[reflect.Type]env.ParserFunc)
	}
	opts.FuncMap[reflect.TypeOf(make([]validation, 0))] = envParseSliceValidate
	opts.FuncMap[reflect.TypeOf(new(validation))] = envParseValidation
	if err := env.ParseWithOptions(&c, opts); err != nil {
		return nil, err
	}
	if len(c.Validations) == 0 {
		return nil, nil
	}
	return c.newValidationMethods(opts)
}

func (c *config) newValidationMethods(opts env.Options) ([]verifier.Verifier, error) {
	var vms []verifier.Verifier
	for _, v := range c.Validations {
		switch v {
		case OCSP:
			vm, err := ocsp.New(opts)
			if err != nil {
				return nil, err
			}
			vms = append(vms, vm)
		case CRL:
			vm, err := crl.New(opts)
			if err != nil {
				return nil, err
			}
			vms = append(vms, vm)
		default:
			return nil, ErrInvalidClientValidation
		}
	}
	return vms, nil
}

func envParseSliceValidate(v string) (interface{}, error) {
	var vms []validation
	v = strings.TrimSpace(v)
	vmss := strings.Split(v, ",")
	for _, vm := range vmss {
		v, err := parseValidation(vm)
		if err != nil {
			return nil, err
		}
		vms = append(vms, v)
	}
	return vms, nil
}

func envParseValidation(v string) (interface{}, error) {
	return parseValidation(v)
}
