package ipchecker

import (
	"fmt"
	"net"
	"net/http"
)

type IPValidator struct {
	trusted *net.IPNet
}

func New(trustedAddr string) (*IPValidator, error) {
	_, trusted, err := net.ParseCIDR(trustedAddr)
	if err != nil {
		return nil, fmt.Errorf("error parsing trusted IP address: %w", err)
	}
	return &IPValidator{trusted: trusted}, nil
}

func (i *IPValidator) CheckIP(ip net.IP) bool {
	return i.trusted.Contains(ip)
}

func (i *IPValidator) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := net.ParseIP(r.Header.Get("X-Real-Ip"))
		if ip == nil || !i.CheckIP(ip) {
			http.Error(w, "IP address not found in X-Real-Ip header", http.StatusForbidden)
		}
		next.ServeHTTP(w, r)
	})
}
