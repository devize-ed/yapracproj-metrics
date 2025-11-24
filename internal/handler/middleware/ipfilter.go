package handler

import (
	"fmt"
	"net"
	"net/http"

	"go.uber.org/zap"
)

// IPFilterMiddleware is a middleware that filters requests by IP address.
func IPFilterMiddleware(trustedSubnet string, logger *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Debugf("IP filter middleware with trusted subnet %s", trustedSubnet)
			// If the trusted subnet is not empty, check if the IP address is in the trusted subnet.
			if trustedSubnet != "" {
				// Get the IP address from the header.
				ipStr := r.Header.Get("X-Real-IP")
				if ipStr == "" {
					logger.Debugf("IP address is empty")
					http.Error(w, "IP address empty", http.StatusForbidden)
					return
				}
				// Parse the IP address.
				ip := net.ParseIP(ipStr)
				if ip == nil {
					logger.Debugf("IP address %s is not valid", ipStr)
					http.Error(w, "IP address empty", http.StatusForbidden)
					return
				}
				// Check if the IP address is in the trusted subnet.
				ok, err := ipInSubnet(ip, trustedSubnet)
				if err != nil {
					logger.Debugf("error checking if IP address is in trusted subnet: %w", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				if !ok {
					logger.Debugf("IP address %s is not in trusted subnet", ipStr)
					http.Error(w, "IP address not in trusted subnet", http.StatusForbidden)
					return
				}
			}
			logger.Info("IP address is in trusted subnet")
			next.ServeHTTP(w, r)
		})
	}
}

// ipInSubnet checks if the given IP address is in the trusted subnet.
func ipInSubnet(ip net.IP, trustedSubnet string) (bool, error) {
	// Parse the trusted subnet.
	_, subnet, err := net.ParseCIDR(trustedSubnet)
	if err != nil {
		return false, fmt.Errorf("failed to parse trusted subnet: %w", err)
	}
	// Check if the IP address is in the trusted subnet.
	return subnet.Contains(ip), nil
}
