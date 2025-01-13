package middlewares

import (
	"net/http"

	"github.com/evildead81/metrics-and-alerts/internal/server/helpers"
	"github.com/evildead81/metrics-and-alerts/internal/server/logger"
)

func CheckClientIpMiddleware(subnet string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.Header.Get("X-Real-IP")
			if ok, err := helpers.CheckIPInSubnet(subnet, ip); !ok {
				if err != nil {
					logger.Logger.Error(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				logger.Logger.Error("Client IP Address is not in subnet")
				w.WriteHeader(http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
