package common

import "net/http"

type SenderInfo struct {
	ip string
}

func ExtractIP(r *http.Request) (ip string) {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		ip = forwarded
	} else {
		ip = r.RemoteAddr
	}
	return ip
}
