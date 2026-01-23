package http_supplier

import (
	"context"
	"io"
	"net"
	"net/http"

	"github.com/cockroachdb/errors"
)

func (r *Supplier) MyIP(ctx context.Context) (net.IP, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.ipify.org/", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build request to get ip")
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch ip")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read body")
	}

	return net.ParseIP(string(body)), nil
}
