package ipchecker

import (
	"fmt"
	"net"
	"net/http"
	"net/netip"

	"github.com/dmitastr/yp_observability_service/internal/common"
	"github.com/dmitastr/yp_observability_service/internal/logger"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
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

// IPValidatorInterceptor logs the incoming request and outgoing response.
func (i *IPValidator) IPValidatorInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if i.trusted != nil {
		logger.Info("Checking IP address")

		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			ipAddr := md.Get("x-real-ip")
			if len(ipAddr) > 0 {
				addr, err := netip.ParseAddrPort(ipAddr[0])
				if err != nil || !i.CheckIP(addr.Addr().AsSlice()) {
					return nil, status.Errorf(codes.PermissionDenied, `IP is not in trusted subnet`)
				}
				logger.Infof("Processing request from IP: %s", addr.String())

			}
		}
	}

	// Log the incoming request
	logger.Infof("Incoming request: Method: %s, Request: %+v", info.FullMethod, req)

	// Call the next handler in the chain
	resp, err := handler(ctx, req)

	return resp, err
}
