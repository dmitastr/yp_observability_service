package common

import "net/http"

type SenderInfo struct {
}

// ExtractIP check header and remote address for IP value of a request
func ExtractIP(r *http.Request) (ip string) {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		ip = forwarded
	} else {
		ip = r.RemoteAddr
	}
	return ip
}
