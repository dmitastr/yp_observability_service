package ipchecker

import (
	"fmt"
	"net"
	"net/http"

	"github.com/dmitastr/yp_observability_service/internal/common"
)

type IPValidator struct {
	trusted *net.IPNet
}

func New(trustedAddr string) (*IPValidator, error) {
	validator := &IPValidator{}
	if trustedAddr == "" {
		return validator, nil
	}

	_, trusted, err := net.ParseCIDR(trustedAddr)
	if err != nil {
		return nil, fmt.Errorf("error parsing trusted IP address: %w", err)
	}
	validator.trusted = trusted
	return validator, nil
}

func (i *IPValidator) CheckIP(ip net.IP) bool {
	return i.trusted.Contains(ip)
}

func (i *IPValidator) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if i.trusted != nil {
			ip, err := common.ExtractIPFromAddress(r)
			if err != nil || !i.CheckIP(ip) {
				http.Error(w, "IP address is not valid", http.StatusForbidden)
			}
		}
		next.ServeHTTP(w, r)
	})
}
