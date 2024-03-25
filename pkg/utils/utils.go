package utils

import (
	"fmt"
	"strings"

	"github.com/absmach/mproxy/pkg/tls"
)

func SecurityStatus(c tls.Config) string {
	if c.CertFile == "" && c.KeyFile == "" {
		return "TLS"
	}
	if c.ClientCAFile != "" {
		if len(c.Verifier.ValidationMethods) > 0 {
			validations := []string{}
			for _, v := range c.Verifier.ValidationMethods {
				validations = append(validations, v.String())
			}
			return fmt.Sprintf("mTLS with client verification %s", strings.Join(validations, ", "))
		}
		return "mTLS"
	}
	return "no TLS"
}
