package server

import (
	"fmt"
	"net"
	"net/http"
	"regexp"
)

const xRealIP = "X-Real-IP"

var subnet *net.IPNet

func setCIDR(cidr string) error {
	if cidr == "" {
		return nil
	}
	_, s, err := net.ParseCIDR(cidr)
	if err != nil {
		return err
	}
	subnet = s

	return nil
}

func trustedSubnet(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		match, err := regexp.MatchString("^/api/internal", r.URL.Path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if !match {
			next.ServeHTTP(w, r)
			return
		}

		if subnet == nil {
			http.Error(w, "trsusted subnet not set", http.StatusForbidden)
			return
		}

		ipStr := r.Header.Get(xRealIP)
		ip := net.ParseIP(ipStr)
		if ip == nil {
			http.Error(w, "header X-Real-IP is empty", http.StatusForbidden)
			return
		}

		if subnet.Contains(ip) {
			next.ServeHTTP(w, r)
			return
		}

		http.Error(w, fmt.Sprintf("trusted subnet does not contain IP %s", ip), http.StatusForbidden)
	})
}
