package net_scan

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"
	"unicode"

	"github.com/cockroachdb/errors"
	"github.com/teadove/teasutils/utils/redact_utils"
)

func (r *Service) protoDetection(ctx context.Context, host string, port uint16) string {
	server, err := r.tryHttp(ctx, "https", host, port)
	if err == nil {
		return "https/" + stripServer(server)
	}

	server, err = r.tryHttp(ctx, "http", host, port)
	if err == nil {
		return "http/" + stripServer(server)
	}

	server, err = r.tryTcp(ctx, host, port)
	if err == nil {
		return "tcp/" + stripServer(server)
	}

	return ""
}

func stripServer(server string) string {
	fields := slices.Collect(strings.Lines(strings.TrimSpace(server)))
	if len(fields) == 0 {
		return ""
	}

	server = fields[0]

	server = strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) {
			return r
		}

		return -1
	}, server)

	return redact_utils.TrimSized(server, 70)
}

func (r *Service) tryHttp(ctx context.Context, proto string, host string, port uint16) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s://%s:%d/", proto, host, port), nil)
	if err != nil {
		return "", errors.WithStack(err)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return "", errors.WithStack(err)
	}
	defer resp.Body.Close()

	return serverInHeaders(resp.Header), nil
}

var headersToTry = [3]string{"Server", "Content-Type", "X-Server-Hostname"}

func serverInHeaders(headers http.Header) string {
	var server string
	for _, h := range headersToTry {
		server = headers.Get(h)
		if server != "" {
			return server
		}
	}

	return ""
}

func (r *Service) tryTcp(ctx context.Context, host string, port uint16) (string, error) {
	conn, err := r.dialer.DialContext(ctx, "tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return "", errors.WithStack(err)
	}

	defer conn.Close()

	err = conn.SetDeadline(time.Now().Add(500 * time.Millisecond))
	if err != nil {
		return "", errors.WithStack(err)
	}

	resp := make([]byte, 256)

	n, err := conn.Read(resp)
	if err != nil {
		return "", errors.WithStack(err)
	}

	if n != 0 {
		return string(resp[:n]), nil
	}

	return "", errors.New("empty response")
}
