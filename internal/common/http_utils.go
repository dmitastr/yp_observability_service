package common

import (
	"cmp"
	"errors"
	"net"
	"net/http"
	"net/netip"
)

type SenderInfo struct {
}

// ExtractIP check header and remote address for IP value of a request
func ExtractIP(r *http.Request) (ip string) {
	forwarded := r.Header.Get("X-Forwarded-For")
	realIP := r.Header.Get("X-Real-IP")
	return cmp.Or(forwarded, realIP, r.RemoteAddr)
}

// ExtractIPFromAddress extracts IP address from
func ExtractIPFromAddress(r *http.Request) (net.IP, error) {
	ip := ExtractIP(r)
	if ip == "" {
		return nil, errors.New("no IP address found")
	}

	addr, err := netip.ParseAddrPort(ip)
	if err != nil {
		return nil, err
	}

	return addr.Addr().AsSlice(), nil

}
