package common

import (
	"fmt"
	"log/slog"
	"net/http"
)

type OriginChecker struct {
	enabled        bool
	logger         *slog.Logger
	allowedOrigins map[string]struct{}
}

func NewOriginChecker(logger *slog.Logger, allowedOrigins []string) OriginChecker {
	o := OriginChecker{}
	if len(allowedOrigins) > 0 {
		o.enabled = true
	}
	for _, allowedOrigin := range allowedOrigins {
		o.allowedOrigins[allowedOrigin] = struct{}{}
	}

	return o
}
func (o *OriginChecker) CheckOrigin(r *http.Request) bool {
	if o.enabled {
		origin := r.Header.Get("Origin")
		_, allowed := o.allowedOrigins[origin]
		if !allowed {
			o.logger.Debug(fmt.Sprintf("Blocked connection from origin: %s", origin))
		}
		return allowed
	}

	return true
}
