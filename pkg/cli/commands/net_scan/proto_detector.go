package net_scan

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode"

	"github.com/pkg/errors"
	"github.com/teadove/teasutils/utils/redact_utils"
)

func (r *NetSystem) protoDetection(ctx context.Context, host string, port uint16) string {
	server, err := r.tryHttp(ctx, "https", host, port)
	if err == nil {
		return server
	}

	server, err = r.tryHttp(ctx, "http", host, port)
	if err == nil {
		return server
	}

	server, err = r.tryTcp(ctx, host, port)
	if err == nil {
		return server
	}

	return ""
}

func (r *NetSystem) tryHttp(ctx context.Context, proto string, host string, port uint16) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s://%s:%d/", proto, host, port), nil)
	if err != nil {
		return "", errors.WithStack(err)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return "", errors.WithStack(err)
	}
	defer resp.Body.Close()

	server := resp.Header.Get("Server")
	if server == "" {
		server = resp.Header.Get("Content-Type")
	}

	return fmt.Sprintf("%s/%s", proto, server), nil
}

func (r *NetSystem) tryTcp(ctx context.Context, host string, port uint16) (string, error) {
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
		clean := strings.Map(func(r rune) rune {
			if unicode.IsPrint(r) {
				return r
			}

			return -1
		}, strings.TrimSpace(string(resp[:n])))

		return "tcp/" + redact_utils.TrimSized(clean, 70), nil
	}

	return "", errors.New("empty response")
}
