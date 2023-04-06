package middleware

import (
	"net"
	"net/http"

	"github.com/nickzhog/devops-tool/pkg/logging"
)

func CheckIP(trustedSubnet *net.IPNet, logger *logging.Logger) func(next http.Handler) http.Handler {
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

func isIPInSubnet(ip string, subnet *net.IPNet) bool {
	checkIP := net.ParseIP(ip)
	return subnet.Contains(checkIP)
}
