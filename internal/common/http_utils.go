package common

import "net/http"

func ExtractIP(r *http.Request) (ip string) {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		ip = forwarded
	} else {
		ip = r.RemoteAddr
	}
	return
}
