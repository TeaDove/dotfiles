package http_supplier

import (
	"context"
	"github.com/pkg/errors"
	"io"
	"net/http"
)

func (r *Supplier) MyIP(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.ipify.org/", nil)
	if err != nil {
		return "", errors.Wrap(err, "failed to build request to get ip")
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "failed to fetch ip")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed to read body")
	}

	return string(body), nil
}
