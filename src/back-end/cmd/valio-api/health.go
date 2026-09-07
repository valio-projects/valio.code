package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"
)

// apiHealth checks the running API rather than opening a database connection or
// initializing Fx. The configured listener contributes only its port/IP family;
// the destination is always loopback and redirects are rejected.
func apiHealth() error {
	host, port, err := net.SplitHostPort(apiConfig().Address)
	if err != nil || port == "" {
		return errors.New("invalid API listening address")
	}
	destination := "127.0.0.1"
	if ip := net.ParseIP(host); ip != nil && ip.To4() == nil {
		destination = "::1"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+net.JoinHostPort(destination, port)+"/readyz", nil)
	if err != nil {
		return errors.New("cannot create readiness request")
	}
	transport := &http.Transport{Proxy: nil}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: 4 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	response, err := client.Do(request)
	if err != nil {
		return errors.New("API is unreachable")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return errors.New("API is not ready")
	}
	return nil
}
