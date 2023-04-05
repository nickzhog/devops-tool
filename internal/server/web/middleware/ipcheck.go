package middleware

import (
	"net"
	"net/http"

	"github.com/nickzhog/devops-tool/pkg/logging"
)

func CheckIP(trustedSubnet string, logger *logging.Logger) func(next http.Handler) http.Handler {
	fn := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			ip := r.Header.Get("X-Real-IP")
			if ip != "" && isIPInSubnet(ip, trustedSubnet) {
				next.ServeHTTP(w, r)
			} else {
				logger.Tracef("wrong ip: %s", ip)
				http.Error(w, "Forbidden", http.StatusForbidden)
			}
		})
	}
	return fn
}

func isIPInSubnet(ip, subnet string) bool {
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return false
	}
	checkIP := net.ParseIP(ip)
	return ipNet.Contains(checkIP)
}
